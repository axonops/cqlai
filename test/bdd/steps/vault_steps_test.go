package steps_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/vault"
)

var vaultEnvNames = []string{
	"VAULT_ADDR", "VAULT_TOKEN", "VAULT_CACERT", "VAULT_NAMESPACE",
	"BAO_ADDR", "BAO_TOKEN", "BAO_CACERT", "BAO_NAMESPACE",
}

// fakeVault answers one database credentials path and lease revocation.
type fakeVault struct {
	mu          sync.Mutex
	server      *httptest.Server
	path        string
	username    string
	password    string
	leaseID     string
	denyRead    bool
	denyRevoke  bool
	reads       int
	tokens      []string
	revocations map[string]int
}

func (f *fakeVault) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokens = append(f.tokens, r.Header.Get("X-Vault-Token"))

	switch {
	case r.Method == http.MethodPut && r.URL.Path == "/v1/sys/leases/revoke":
		var body struct {
			LeaseID string `json:"lease_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.revocations[body.LeaseID]++
		if f.denyRevoke {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"errors":["1 error occurred:\n\t* permission denied\n\n"]}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/"):
		f.reads++
		if f.denyRead {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"errors":["permission denied"]}`))
			return
		}
		if r.URL.Path != "/v1/"+f.path {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":[]}`))
			return
		}
		data := map[string]string{"username": f.username}
		if f.password != "" {
			data["password"] = f.password
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"lease_id":       f.leaseID,
			"lease_duration": 3600,
			"renewable":      true,
			"data":           data,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type vaultWorld struct {
	fake       *fakeVault
	home       string
	client     *vault.Client
	creds      *vault.Credentials
	err        error
	configFile string
	sessionErr error
}

func (w *vaultWorld) configWithVaultPath(path string) error {
	w.configFile = filepath.Join(w.home, "cqlai.json")
	data, err := json.Marshal(map[string]string{"host": "127.0.0.1", "vaultPath": path})
	if err != nil {
		return err
	}
	return os.WriteFile(w.configFile, data, 0o600)
}

// closedPort is a local port nothing listens on.
func closedPort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	return port, l.Close()
}

func (w *vaultWorld) openSession(path string) error {
	port, err := closedPort()
	if err != nil {
		return err
	}
	sess, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:           "127.0.0.1",
		Port:           port,
		BatchMode:      true,
		ConnectTimeout: 1,
		RequestTimeout: 1,
		ConfigFile:     w.configFile,
		VaultPath:      path,
	})
	if sess != nil {
		sess.Close()
	}
	w.sessionErr = err
	return nil
}

func (w *vaultWorld) openSessionNoPath() error { return w.openSession("") }

func (w *vaultWorld) sessionFails() error {
	if w.sessionErr == nil {
		return fmt.Errorf("expected the session to fail")
	}
	return nil
}

func (w *vaultWorld) readCount(n int) error {
	w.fake.mu.Lock()
	defer w.fake.mu.Unlock()
	if w.fake.reads != n {
		return fmt.Errorf("server received %d read requests, want %d", w.fake.reads, n)
	}
	return nil
}

func (w *vaultWorld) fakeServer(username, password, leaseID, path string) error {
	w.fake = &fakeVault{
		path:        path,
		username:    username,
		password:    password,
		leaseID:     leaseID,
		revocations: map[string]int{},
	}
	w.fake.server = httptest.NewServer(w.fake)
	return nil
}

func (w *vaultWorld) envPointsAt(prefix, token string) error {
	if err := os.Setenv(prefix+"_ADDR", w.fake.server.URL); err != nil {
		return err
	}
	return os.Setenv(prefix+"_TOKEN", token)
}

func (w *vaultWorld) envPointsAtNoToken() error {
	return os.Setenv("VAULT_ADDR", w.fake.server.URL)
}

func (w *vaultWorld) noAddress() error { return nil }

func (w *vaultWorld) tokenFile(token string) error {
	return os.WriteFile(filepath.Join(w.home, ".vault-token"), []byte(token+"\n"), 0o600)
}

func (w *vaultWorld) denyReading() error {
	w.fake.denyRead = true
	return nil
}

func (w *vaultWorld) denyRevocation() error {
	w.fake.denyRevoke = true
	return nil
}

func (w *vaultWorld) withoutPassword(string) error {
	w.fake.password = ""
	return nil
}

func (w *vaultWorld) read(path string) error {
	w.client, w.err = vault.FromEnv()
	if w.err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.creds, w.err = w.client.ReadCredentials(ctx, path)
	return nil
}

func (w *vaultWorld) revokeLease() error {
	if w.creds == nil {
		return fmt.Errorf("no credentials were read: %v", w.err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.client.Revoke(ctx, w.creds.LeaseID)
	return nil
}

func (w *vaultWorld) exits() error {
	vault.RevokeAll()
	return nil
}

func (w *vaultWorld) credsAre(username, password string) error {
	if w.err != nil {
		return fmt.Errorf("expected credentials, got error: %v", w.err)
	}
	if w.creds.Username != username || w.creds.Password != password {
		return fmt.Errorf("got %q/%q, want %q/%q", w.creds.Username, w.creds.Password, username, password)
	}
	return nil
}

func (w *vaultWorld) sawToken(token string) error {
	w.fake.mu.Lock()
	defer w.fake.mu.Unlock()
	for _, t := range w.fake.tokens {
		if t != token {
			return fmt.Errorf("server saw token %q, want %q", t, token)
		}
	}
	if len(w.fake.tokens) == 0 {
		return fmt.Errorf("server saw no request")
	}
	return nil
}

func (w *vaultWorld) revokedOnce(leaseID string) error {
	w.fake.mu.Lock()
	defer w.fake.mu.Unlock()
	if n := w.fake.revocations[leaseID]; n != 1 {
		return fmt.Errorf("lease %q revoked %d times, want 1", leaseID, n)
	}
	return nil
}

func (w *vaultWorld) noError() error {
	if w.err != nil {
		return fmt.Errorf("unexpected error: %v", w.err)
	}
	return nil
}

func (w *vaultWorld) failsWith(needle string) error {
	if w.err == nil {
		return fmt.Errorf("expected an error containing %q, got none", needle)
	}
	if !strings.Contains(w.err.Error(), needle) {
		return fmt.Errorf("error %q does not contain %q", w.err.Error(), needle)
	}
	return nil
}

func (w *vaultWorld) errorLacks(secret string) error {
	if w.err != nil && strings.Contains(w.err.Error(), secret) {
		return fmt.Errorf("error %q contains %q", w.err.Error(), secret)
	}
	return nil
}

func (w *vaultWorld) noReads() error {
	w.fake.mu.Lock()
	defer w.fake.mu.Unlock()
	if w.fake.reads != 0 {
		return fmt.Errorf("server received %d read requests, want 0", w.fake.reads)
	}
	return nil
}

func initVaultScenario(t *testing.T) func(*godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		w := &vaultWorld{}
		saved := map[string]*string{}

		ctx.Before(func(c context.Context, _ *godog.Scenario) (context.Context, error) {
			*w = vaultWorld{home: t.TempDir()}
			for _, name := range append(vaultEnvNames, "HOME") {
				if v, ok := os.LookupEnv(name); ok {
					saved[name] = &v
				} else {
					saved[name] = nil
				}
				_ = os.Unsetenv(name)
			}
			return c, os.Setenv("HOME", w.home)
		})
		ctx.After(func(c context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			vault.RevokeAll()
			if w.fake != nil {
				w.fake.server.Close()
			}
			for name, v := range saved {
				if v == nil {
					_ = os.Unsetenv(name)
				} else {
					_ = os.Setenv(name, *v)
				}
			}
			return c, nil
		})

		ctx.Step(`^a fake Vault server that issues user "([^"]*)" with password "([^"]*)" and lease "([^"]*)" at "([^"]*)"$`, w.fakeServer)
		ctx.Step(`^the environment points (VAULT|BAO)_ADDR at the fake server with (?:VAULT|BAO)_TOKEN "([^"]*)"$`, w.envPointsAt)
		ctx.Step(`^the environment points VAULT_ADDR at the fake server with no token variable$`, w.envPointsAtNoToken)
		ctx.Step(`^the environment has no Vault or OpenBao address$`, w.noAddress)
		ctx.Step(`^the file ~/\.vault-token contains "([^"]*)"$`, w.tokenFile)
		ctx.Step(`^the fake server denies reading$`, w.denyReading)
		ctx.Step(`^the fake server denies revocation$`, w.denyRevocation)
		ctx.Step(`^the fake server answers "([^"]*)" without a password$`, w.withoutPassword)

		ctx.Step(`^I read credentials from "([^"]*)"$`, w.read)
		ctx.Step(`^the lease is revoked$`, w.revokeLease)
		ctx.Step(`^cqlai exits$`, w.exits)

		ctx.Step(`^the credentials have username "([^"]*)" and password "([^"]*)"$`, w.credsAre)
		ctx.Step(`^the fake server saw the token "([^"]*)"$`, w.sawToken)
		ctx.Step(`^the fake server was asked to revoke lease "([^"]*)" once$`, w.revokedOnce)
		ctx.Step(`^no error is reported$`, w.noError)
		ctx.Step(`^reading fails with an error containing "([^"]*)"$`, w.failsWith)
		ctx.Step(`^the error does not contain "([^"]*)"$`, w.errorLacks)
		ctx.Step(`^the fake server received no read request$`, w.noReads)
		ctx.Step(`^a config file with vaultPath "([^"]*)"$`, w.configWithVaultPath)
		ctx.Step(`^I open a session to an unreachable cluster with Vault path "([^"]*)"$`, w.openSession)
		ctx.Step(`^I open a session to an unreachable cluster with no Vault path$`, w.openSessionNoPath)
		ctx.Step(`^opening the session fails$`, w.sessionFails)
		ctx.Step(`^the fake server received (\d+) read requests?$`, w.readCount)
	}
}

func TestVaultCredentials_BDD(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "vault-credentials",
		ScenarioInitializer: initVaultScenario(t),
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features/vault-credentials.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("BDD scenarios failed — see output above for details")
	}
}
