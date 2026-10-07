package mcp

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// The terminal app's server, over HTTP on this machine.

// freePort is a port nothing is listening on.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return port
}

// startHost starts a host with a configuration file of its own and a token
// in a home directory of its own.
func startHost(t *testing.T) *Host {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	file := filepath.Join(home, "cqlai.json")
	require.NoError(t, os.WriteFile(file, []byte(`{"mcp": {"auditLog": "-"}}`), 0o600))

	h := StartHost(Options{ConfigFile: file, Port: freePort(t), RequestTimeout: 5}, "test")
	t.Cleanup(h.Close)
	require.True(t, h.Serving(), h.Status())
	return h
}

// bearer adds the token to every request.
type bearer struct {
	token string
	next  http.RoundTripper
}

func (b bearer) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return b.next.RoundTrip(r)
}

// TestTheSDKClientConnectsOverHTTP, with the token, and is told there is no
// cluster until the app has one.
func TestTheSDKClientConnectsOverHTTP(t *testing.T) {
	h := startHost(t)

	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "0"}, nil)
	session, err := client.Connect(context.Background(), &sdk.StreamableClientTransport{
		Endpoint:   h.URL(),
		HTTPClient: &http.Client{Transport: bearer{token: h.token, next: http.DefaultTransport}},
	}, nil)
	require.NoError(t, err)
	defer session.Close()

	tools, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, tools.Tools)

	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{
		Name: "query", Arguments: map[string]any{"cql": "SELECT * FROM shop.orders"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*sdk.TextContent).Text, "pick one in FILE > CONNECT")
}

func post(t *testing.T, h *Host, mutate func(*http.Request)) int {
	t.Helper()
	body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	r, err := http.NewRequest(http.MethodPost, h.URL(), strings.NewReader(body))
	require.NoError(t, err)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+h.token)
	mutate(r)
	resp, err := http.DefaultClient.Do(r)
	require.NoError(t, err)
	_ = resp.Body.Close()
	return resp.StatusCode
}

// TestOnlyTheClientWithTheTokenGetsIn.
func TestOnlyTheClientWithTheTokenGetsIn(t *testing.T) {
	h := startHost(t)

	assert.Equal(t, http.StatusOK, post(t, h, func(*http.Request) {}))
	assert.Equal(t, http.StatusUnauthorized, post(t, h, func(r *http.Request) { r.Header.Del("Authorization") }))
	assert.Equal(t, http.StatusUnauthorized, post(t, h, func(r *http.Request) { r.Header.Set("Authorization", "Bearer wrong") }))
	assert.Equal(t, http.StatusUnauthorized, post(t, h, func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+h.token[:len(h.token)-1])
	}))
}

// TestAWebPageCannotUseIt: not from another origin, and not under another
// name for this machine.
func TestAWebPageCannotUseIt(t *testing.T) {
	h := startHost(t)
	port := strconv.Itoa(h.Port())

	assert.Equal(t, http.StatusForbidden, post(t, h, func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") }))
	assert.Equal(t, http.StatusForbidden, post(t, h, func(r *http.Request) { r.Host = "evil.example:" + port }))
	assert.Equal(t, http.StatusOK, post(t, h, func(r *http.Request) { r.Header.Set("Origin", "http://localhost:"+port) }))
}

// TestTheTokenIsKeptAndPrivate: the same token next time, in a file only its
// owner can read, and one anyone can read is refused.
func TestTheTokenIsKeptAndPrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	first, err := loadToken(path)
	require.NoError(t, err)
	again, err := loadToken(path)
	require.NoError(t, err)
	assert.Equal(t, first, again)
	assert.Len(t, first, 64)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	require.NoError(t, os.Chmod(path, 0o644))
	_, err = loadToken(path)
	assert.Error(t, err)
}

// TestAPortInUseDoesNotStopTheApp: the host says why it is not serving.
func TestAPortInUseDoesNotStopTheApp(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	h := StartHost(Options{Port: l.Addr().(*net.TCPAddr).Port, ConfigFile: filepath.Join(t.TempDir(), "none.json")}, "test")
	defer h.Close()
	assert.False(t, h.Serving())
	assert.Contains(t, h.Status(), "cannot serve MCP")
}

// TestTheHostFollowsTheConnection: its policy is the connection's.
func TestTheHostFollowsTheConnection(t *testing.T) {
	h := startHost(t)
	readOnly := []string{"SELECT"}
	file := h.opts.ConfigFile
	require.NoError(t, os.WriteFile(file, []byte(`{"mcp": {"auditLog": "-"}, "connections": [
		{"name": "prod", "host": "10.0.0.5", "mcp": {"permit": ["SELECT"]}}]}`), 0o600))

	h.Use(config.Config{Name: "prod", Host: "10.0.0.5"}, "prod")
	assert.Equal(t, readOnly, h.Policy().Permitted())
	assert.Contains(t, h.Status(), `saved connection "prod"`)

	h.Drop()
	assert.Contains(t, h.Status(), "pick one in FILE > CONNECT")
}
