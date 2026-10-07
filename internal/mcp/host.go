package mcp

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/policy"
	"github.com/axonops/cqlai/internal/session"
)

// Host is the MCP server inside the terminal app.
//
// The app owns the terminal, so the protocol cannot run over stdin and
// stdout. It runs over HTTP instead, on 127.0.0.1 only, and every request has
// to carry the token: something else on the machine - a web page in a
// browser, another user's process - must not be able to use the cluster
// through it.
//
// It serves whichever connection the app is using, under that connection's
// policy, and follows the app when another one is picked. It keeps a session
// of its own to that connection: a model's queries must not change the
// shell's session, nor the shell's change theirs.
type Host struct {
	opts    Options
	version string
	port    int
	token   string

	listener net.Listener
	http     *http.Server
	server   *Server
	audit    *policy.Auditor
	limiter  *policy.Limiter

	mu         sync.Mutex
	pol        policy.Policy
	link       *cluster
	connection string // what the app is connected to, in words
	failed     string // why the server is not serving, when it is not
}

// TokenFile is where the token is kept, readable only by its owner. It is
// kept rather than made afresh each time so the client's configuration does
// not have to change every time the app starts.
func TokenFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cqlai_mcp_token"
	}
	return filepath.Join(home, ".cqlai_mcp_token")
}

// StartHost starts serving MCP for the terminal app. It never fails the app:
// what went wrong - a port in use, a policy that cannot be read - is kept, and
// shown by Status.
func StartHost(o Options, version string) *Host {
	h := &Host{opts: o, version: version}

	file, err := config.LoadConfig(o.ConfigFile)
	if err != nil {
		file = &config.Config{}
	}
	h.port = o.Port
	if h.port == 0 && file.MCP != nil {
		h.port = file.MCP.Port
	}
	if h.port == 0 {
		h.port = DefaultPort
	}

	// Until the app is connected, the policy is the top of the file's.
	pol, err := policy.Load(file, "", "", time.Duration(o.RequestTimeout)*time.Second, o.Flags)
	if err != nil {
		h.failed = err.Error()
		return h
	}
	h.pol = pol

	if h.token, err = loadToken(TokenFile()); err != nil {
		h.failed = err.Error()
		return h
	}
	if h.audit, err = policy.OpenAudit(pol.AuditLog()); err != nil {
		h.failed = err.Error()
		return h
	}
	h.limiter = policy.NewLimiter(pol.CallsPerMinute())

	h.server = NewServer(h, h.limiter, h.audit, version)
	h.server.requireInit = false // each HTTP request stands alone

	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(h.port)))
	if err != nil {
		h.failed = fmt.Sprintf("cannot serve MCP on 127.0.0.1:%d: %v", h.port, err)
		return h
	}
	h.listener = listener
	h.http = &http.Server{
		Handler:           h.handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() { _ = h.http.Serve(listener) }()
	return h
}

// Use serves conn, which the app has just connected to. name is the saved
// connection's name, or "" for the settings at the top of the file.
func (h *Host) Use(conn config.Config, name string) {
	file, err := config.LoadConfig(h.opts.ConfigFile)
	if err != nil {
		file = &config.Config{}
	}
	if _, saved := file.Connection(name); !saved {
		name = ""
	}

	pol, err := policy.Load(file, name, conn.Password, time.Duration(h.opts.RequestTimeout)*time.Second, h.opts.Flags)

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.link != nil {
		go h.link.close()
		h.link = nil
	}
	h.connection = describeConnection(conn, name, file.SavePath())
	if err != nil {
		// Settings that cannot be read are not served at all.
		h.pol = policy.Policy{}
		h.failed = err.Error()
		return
	}
	h.pol = pol
	if h.failed != "" && h.listener != nil {
		h.failed = ""
	}
	h.link = newCluster(conn, file.Consistency, h.opts, h.connection, pol.Scrub, session.NewManager(file))
	h.link.start()
}

// Drop stops serving the connection the app has left.
func (h *Host) Drop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.link != nil {
		go h.link.close()
		h.link = nil
	}
	h.connection = ""
}

// Policy is the policy of the connection being served. Source.
func (h *Host) Policy() policy.Policy {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.pol
}

// Env is what a call runs against now. Source.
func (h *Host) Env() ai.ToolEnv {
	h.mu.Lock()
	pol, link := h.pol, h.link
	h.mu.Unlock()
	if link == nil {
		return ai.ToolEnv{Policy: pol, NotConnected: "cqlai is not connected to a cluster: pick one in FILE > CONNECT"}
	}
	return link.env(pol)
}

// Serving reports whether MCP is being served.
func (h *Host) Serving() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.listener != nil && h.failed == ""
}

// Port is where MCP is served.
func (h *Host) Port() int { return h.port }

// URL is the address a client connects to.
func (h *Host) URL() string { return fmt.Sprintf("http://127.0.0.1:%d/mcp", h.port) }

// Status says what the server is doing, in words, for the app to show.
func (h *Host) Status() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	switch {
	case h.failed != "":
		return "MCP is not being served: " + h.failed
	case h.connection == "":
		return "MCP is served at " + h.URL() + ", with no cluster: pick one in FILE > CONNECT"
	}
	return "MCP is served at " + h.URL() + " for " + h.connection +
		"; permitted: " + strings.Join(h.pol.Permitted(), ", ")
}

// ClientConfig is what to put in a client's MCP configuration.
func (h *Host) ClientConfig() string {
	cfg := map[string]any{
		"mcpServers": map[string]any{
			"cqlai": map[string]any{
				"type":    "http",
				"url":     h.URL(),
				"headers": map[string]string{"Authorization": "Bearer " + h.token},
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

// Close stops serving.
func (h *Host) Close() {
	if h.http != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = h.http.Shutdown(ctx)
	}
	h.Drop()
	_ = h.audit.Close()
}

// handler is the one endpoint, /mcp, and the checks every request passes
// before it is read.
func (h *Host) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if !h.sameMachine(r) {
			// A page in a browser can reach 127.0.0.1 under another name
			// (DNS rebinding), or post to it from its own origin. Neither
			// is a client this server is for.
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !h.authorized(r) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="cqlai"`)
			http.Error(w, "the Authorization header has to carry the token cqlai shows", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			// No stream from the server: it sends nothing it was not asked
			// for.
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxMessageBytes+1))
		if err != nil || len(body) > maxMessageBytes {
			http.Error(w, "the message is too large", http.StatusRequestEntityTooLarge)
			return
		}
		answer := h.server.Process(r.Context(), body)
		if answer == nil {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(answer)
	})
	return mux
}

// sameMachine checks that a request was made for this server by name, and not
// from a web page of another origin.
func (h *Host) sameMachine(r *http.Request) bool {
	port := strconv.Itoa(h.port)
	allowedHost := map[string]bool{"127.0.0.1:" + port: true, "localhost:" + port: true, "[::1]:" + port: true}
	if !allowedHost[r.Host] {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // not a browser
	}
	return origin == "http://127.0.0.1:"+port || origin == "http://localhost:"+port
}

// authorized checks the token, in time that does not depend on how much of
// it was right.
func (h *Host) authorized(r *http.Request) bool {
	got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || h.token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(got)), []byte(h.token)) == 1
}

// loadToken reads the token, making one the first time. A file anyone else
// can read is not used: the token would not be a secret.
func loadToken(path string) (string, error) {
	data, err := os.ReadFile(path) // #nosec G304 - the token file in the user's home directory
	if err == nil {
		info, statErr := os.Stat(path)
		if statErr == nil && info.Mode().Perm()&0o077 != 0 {
			return "", fmt.Errorf("%s can be read by others: make it readable only by you (chmod 600)", path)
		}
		token := strings.TrimSpace(string(data))
		if len(token) < 32 {
			return "", fmt.Errorf("%s does not hold a token: delete it, and cqlai makes a new one", path)
		}
		return token, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("cannot keep the MCP token in %s: %w", path, err)
	}
	return token, nil
}
