// Package vault reads dynamic database credentials from Vault or OpenBao.
//
// The address and token come from the environment, as the vault and bao
// command line tools read them, so a user who is logged in with either has
// nothing more to set up.
package vault

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/logger"
)

// Credentials are what a database secrets engine hands out for one role.
type Credentials struct {
	Username string
	Password string
	LeaseID  string
}

// Client talks to one Vault or OpenBao server with one token.
type Client struct {
	addr      string
	token     string
	namespace string
	http      *http.Client
}

// getenv is the first of names that is set. OpenBao's names go first: whoever
// set them meant OpenBao.
func getenv(names ...string) string {
	for _, name := range names {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	return ""
}

// FromEnv is a client for the server and token in the environment:
// BAO_ADDR or VAULT_ADDR, and BAO_TOKEN, VAULT_TOKEN or ~/.vault-token.
// BAO_CACERT or VAULT_CACERT, and BAO_NAMESPACE or VAULT_NAMESPACE, are used
// when set.
func FromEnv() (*Client, error) {
	addr := getenv("BAO_ADDR", "VAULT_ADDR")
	if addr == "" {
		return nil, errors.New("set VAULT_ADDR or BAO_ADDR to read credentials from Vault")
	}

	token := getenv("BAO_TOKEN", "VAULT_TOKEN")
	if token == "" {
		if home, err := os.UserHomeDir(); err == nil {
			if data, err := os.ReadFile(filepath.Join(home, ".vault-token")); err == nil { //nolint:gosec // G304: the CLI's own token file
				token = strings.TrimSpace(string(data))
			}
		}
	}
	if token == "" {
		return nil, errors.New("set VAULT_TOKEN or BAO_TOKEN, or log in, to read credentials from Vault")
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if caFile := getenv("BAO_CACERT", "VAULT_CACERT"); caFile != "" {
		pem, err := os.ReadFile(caFile) //nolint:gosec // G304: the path is the user's own setting
		if err != nil {
			return nil, fmt.Errorf("reading the Vault CA certificate: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("no certificates in %s", caFile)
		}
		tlsConfig.RootCAs = pool
	}

	return &Client{
		addr:      strings.TrimRight(addr, "/"),
		token:     token,
		namespace: getenv("BAO_NAMESPACE", "VAULT_NAMESPACE"),
		http: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: tlsConfig},
			// A redirect would carry X-Vault-Token to wherever it points.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

// cleanPath is path as the API wants it, without the /v1/ the CLI leaves off.
func cleanPath(path string) (string, error) {
	path = strings.Trim(strings.TrimSpace(path), "/")
	path = strings.TrimPrefix(path, "v1/")
	if path == "" {
		return "", errors.New("the Vault path is empty")
	}
	if strings.ContainsAny(path, "?#") {
		return "", fmt.Errorf("the Vault path %q cannot have a query or fragment", path)
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("the Vault path %q is not a valid path", path)
		}
	}
	return path, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.addr+"/v1/"+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("X-Vault-Request", "true")
	if c.namespace != "" {
		req.Header.Set("X-Vault-Namespace", c.namespace)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.http.Do(req)
}

// apiError is the first message of a Vault error response, or the status.
func apiError(res *http.Response, data []byte) string {
	var parsed struct {
		Errors []string `json:"errors"`
	}
	if json.Unmarshal(data, &parsed) == nil && len(parsed.Errors) > 0 && parsed.Errors[0] != "" {
		return parsed.Errors[0]
	}
	return res.Status
}

// ReadCredentials reads a username and password from path, such as
// database/creds/myrole.
func (c *Client) ReadCredentials(ctx context.Context, path string) (*Credentials, error) {
	path, err := cleanPath(path)
	if err != nil {
		return nil, err
	}

	res, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("reading %s from Vault: %w", path, err)
	}
	defer func() { _ = res.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading %s from Vault: %w", path, err)
	}

	switch {
	case res.StatusCode == http.StatusForbidden:
		return nil, fmt.Errorf("reading %s from Vault: permission denied", path)
	case res.StatusCode == http.StatusNotFound:
		return nil, fmt.Errorf("reading %s from Vault: no such path", path)
	case res.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("reading %s from Vault: %s", path, apiError(res, data))
	}

	var parsed struct {
		LeaseID string `json:"lease_id"`
		Data    struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("reading %s from Vault: the response is not JSON", path)
	}
	if parsed.Data.Username == "" || parsed.Data.Password == "" {
		return nil, fmt.Errorf("reading %s from Vault: the response has no username and password", path)
	}

	creds := &Credentials{
		Username: parsed.Data.Username,
		Password: parsed.Data.Password,
		LeaseID:  parsed.LeaseID,
	}
	track(c, creds.LeaseID)
	return creds, nil
}

// Revoke ends a lease. A token that may read credentials is often not allowed
// to revoke them, and the lease then runs out on its own; so a failure is
// written to the debug log and otherwise ignored.
func (c *Client) Revoke(ctx context.Context, leaseID string) {
	if leaseID == "" || !untrack(leaseID) {
		return
	}
	res, err := c.do(ctx, http.MethodPut, "sys/leases/revoke", map[string]string{"lease_id": leaseID})
	if err != nil {
		logger.DebugfToFile("Vault", "Revoking lease %s: %v", leaseID, err)
		return
	}
	defer func() { _ = res.Body.Close() }()
	data, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode/100 != 2 {
		logger.DebugfToFile("Vault", "Revoking lease %s: %s", leaseID, apiError(res, data))
		return
	}
	logger.DebugfToFile("Vault", "Revoked lease %s", leaseID)
}

// The leases handed out and not yet revoked, so that whatever way cqlai
// leaves, RevokeAll can end the ones still open.
var (
	leasesMu sync.Mutex
	leases   = map[string]*Client{}
)

func track(c *Client, leaseID string) {
	if leaseID == "" {
		return
	}
	leasesMu.Lock()
	defer leasesMu.Unlock()
	leases[leaseID] = c
}

// untrack reports whether leaseID was still open, so it is revoked once.
func untrack(leaseID string) bool {
	leasesMu.Lock()
	defer leasesMu.Unlock()
	if _, ok := leases[leaseID]; !ok {
		return false
	}
	delete(leases, leaseID)
	return true
}

// RevokeAll revokes every lease still open.
func RevokeAll() {
	leasesMu.Lock()
	open := make(map[string]*Client, len(leases))
	for id, c := range leases {
		open[id] = c
	}
	leasesMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for id, c := range open {
		c.Revoke(ctx, id)
	}
}
