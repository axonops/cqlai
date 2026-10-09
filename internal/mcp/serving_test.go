package mcp

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// How the terminal app serves MCP: the token, TLS and the address.

const ping = `{"jsonrpc":"2.0","id":1,"method":"ping"}`

// TestTheTokenIsOffUnlessAskedFor: a request without one gets in, the
// client's configuration has no header, and no token file is made.
func TestTheTokenIsOffUnlessAskedFor(t *testing.T) {
	h := startHostWith(t, "")
	require.False(t, h.TokenRequired())

	resp, err := http.Post(h.URL(), "application/json", strings.NewReader(ping))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.NotContains(t, h.ClientConfig(), "Authorization")
	assert.Equal(t, "http://127.0.0.1:"+itoa(h.Port())+"/mcp", h.URL())
	assert.NoFileExists(t, TokenFile())
}

// TestAWebPageCannotUseItWithoutTheTokenEither: the token is off by default,
// so the origin and host checks are what keep a browser out.
func TestAWebPageCannotUseItWithoutTheTokenEither(t *testing.T) {
	h := startHostWith(t, "")
	port := itoa(h.Port())

	for name, mutate := range map[string]func(*http.Request){
		"another origin": func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") },
		"another name":   func(r *http.Request) { r.Host = "evil.example:" + port },
	} {
		r, err := http.NewRequest(http.MethodPost, h.URL(), strings.NewReader(ping))
		require.NoError(t, err)
		mutate(r)
		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode, name)
	}
}

// TestTheTokenIsRequiredWhenAskedFor, and the client configuration carries it.
func TestTheTokenIsRequiredWhenAskedFor(t *testing.T) {
	h := startHostWith(t, `"token": true`)
	require.True(t, h.TokenRequired())

	resp, err := http.Post(h.URL(), "application/json", strings.NewReader(ping))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, h.ClientConfig(), "Bearer "+h.token)
	assert.FileExists(t, TokenFile())
}

// TestTheSettingsThatCannotWork are refused before anything is served.
func TestTheSettingsThatCannotWork(t *testing.T) {
	for name, s := range map[string]serving{
		"a certificate without its key":   {listen: "127.0.0.1", cert: "c.pem"},
		"a key without its certificate":   {listen: "127.0.0.1", key: "k.pem"},
		"a client CA without TLS":         {listen: "127.0.0.1", clientCA: "ca.pem"},
		"no address":                      {listen: " "},
		"the network, without TLS":        {listen: "0.0.0.0", token: true},
		"the network, with nobody proven": {listen: "10.0.0.5", cert: "c.pem", key: "k.pem"},
	} {
		assert.Error(t, s.check(), name)
	}

	for name, s := range map[string]serving{
		"this machine, plain":              {listen: "127.0.0.1"},
		"this machine by name":             {listen: "localhost"},
		"this machine over IPv6":           {listen: "::1"},
		"the network, TLS and the token":   {listen: "0.0.0.0", cert: "c.pem", key: "k.pem", token: true},
		"the network, TLS and a client CA": {listen: "cqlai.example", cert: "c.pem", key: "k.pem", clientCA: "ca.pem"},
	} {
		assert.NoError(t, s.check(), name)
	}
}

// TestASettingThatCannotWorkSaysSoInTheStatus, and nothing is served.
func TestASettingThatCannotWorkSaysSoInTheStatus(t *testing.T) {
	h := startHostFailing(t, `"listen": "0.0.0.0"`)
	assert.False(t, h.Serving())
	assert.Contains(t, h.Status(), "only over TLS")
}

// TestTheURLIsWhereAClientConnects: HTTPS with TLS, and 127.0.0.1 for an
// address that is every address.
func TestTheURLIsWhereAClientConnects(t *testing.T) {
	assert.Equal(t, "http://127.0.0.1:7845/mcp", serving{listen: "127.0.0.1", port: 7845}.url())
	assert.Equal(t, "https://127.0.0.1:7845/mcp", serving{listen: "0.0.0.0", port: 7845, cert: "c"}.url())
	assert.Equal(t, "https://[::1]:9000/mcp", serving{listen: "::1", port: 9000, cert: "c"}.url())
	assert.Equal(t, "https://cqlai.example:7845/mcp", serving{listen: "cqlai.example", port: 7845, cert: "c"}.url())
}

// TestTheFlagsGoOverTheSettings.
func TestTheFlagsGoOverTheSettings(t *testing.T) {
	file := &config.Config{MCP: &config.MCPConfig{
		Listen: "127.0.0.1", Port: 9000, Token: true, TLSCert: "file.pem", TLSKey: "file.key",
	}}
	off := false
	s := servingFrom(file, Options{Listen: "::1", Port: 9001, Token: &off, TLSCert: "flag.pem"})
	assert.Equal(t, serving{listen: "::1", port: 9001, token: false, cert: "flag.pem", key: "file.key"}, s)

	s = servingFrom(&config.Config{}, Options{})
	assert.Equal(t, serving{listen: DefaultListen, port: DefaultPort}, s, "the defaults: this machine, no token, no TLS")
}

// TestTheTokenFlagIsOnlyAnOverrideWhenGiven: without --token the settings
// decide.
func TestTheTokenFlagIsOnlyAnOverrideWhenGiven(t *testing.T) {
	o, ok, _ := ParseOptions(nil, io.Discard)
	require.True(t, ok)
	assert.Nil(t, o.Token)

	o, ok, _ = ParseOptions([]string{"--token", "--listen", "0.0.0.0", "--tls-cert", "c.pem", "--tls-key", "k.pem", "--tls-client-ca", "ca.pem"}, io.Discard)
	require.True(t, ok)
	require.NotNil(t, o.Token)
	assert.True(t, *o.Token)
	assert.Equal(t, []string{"0.0.0.0", "c.pem", "k.pem", "ca.pem"}, []string{o.Listen, o.TLSCert, o.TLSKey, o.TLSClientCA})

	o, ok, _ = ParseOptions([]string{"--token=false"}, io.Discard)
	require.True(t, ok)
	require.NotNil(t, o.Token)
	assert.False(t, *o.Token)
}

// TestMCPOverTLS: a client that trusts the CA gets in over HTTPS, and plain
// HTTP does not.
func TestMCPOverTLS(t *testing.T) {
	pki := newPKI(t)
	h := startHostWith(t, `"tlsCert": "`+pki.serverCert+`", "tlsKey": "`+pki.serverKey+`"`)
	require.True(t, strings.HasPrefix(h.URL(), "https://"), h.URL())

	assert.Equal(t, http.StatusOK, pki.post(t, h.URL(), nil))

	plain := strings.Replace(h.URL(), "https://", "http://", 1)
	resp, err := http.Post(plain, "application/json", strings.NewReader(ping))
	if err == nil {
		_ = resp.Body.Close()
		assert.NotEqual(t, http.StatusOK, resp.StatusCode, "plain HTTP is not MCP")
	}
}

// TestMCPOverMutualTLS: without a certificate the CA signed, the handshake
// fails; with one, the client gets in.
func TestMCPOverMutualTLS(t *testing.T) {
	pki := newPKI(t)
	h := startHostWith(t, `"tlsCert": "`+pki.serverCert+`", "tlsKey": "`+pki.serverKey+`", "tlsClientCA": "`+pki.caFile+`"`)
	require.True(t, h.ClientCertRequired())

	_, err := pki.client(nil).Post(h.URL(), "application/json", strings.NewReader(ping))
	assert.Error(t, err, "no client certificate")

	stranger := newPKI(t)
	_, err = pki.client(&stranger.clientPair).Post(h.URL(), "application/json", strings.NewReader(ping))
	assert.Error(t, err, "a certificate another CA signed")

	assert.Equal(t, http.StatusOK, pki.post(t, h.URL(), &pki.clientPair))
}

// startHostFailing starts a host whose settings stop it serving.
func startHostFailing(t *testing.T, settings string) *Host {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	file := filepath.Join(home, "cqlai.json")
	require.NoError(t, os.WriteFile(file, []byte(`{"mcp": {"auditLog": "-", `+settings+`}}`), 0o600))
	h := StartHost(Options{ConfigFile: file, Port: freePort(t), RequestTimeout: 5}, "test")
	t.Cleanup(h.Close)
	return h
}

func itoa(n int) string { return big.NewInt(int64(n)).String() }

// pki is a CA, a server certificate for this machine, and a client
// certificate, all made for the test.
type pki struct {
	caFile, serverCert, serverKey string
	pool                          *x509.CertPool
	clientPair                    tls.Certificate
}

func newPKI(t *testing.T) pki {
	t.Helper()
	dir := t.TempDir()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test CA"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)
	ca, err := x509.ParseCertificate(caDER)
	require.NoError(t, err)

	issue := func(serial int64, usage x509.ExtKeyUsage, ips []net.IP) ([]byte, []byte) {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		template := &x509.Certificate{
			SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "test"},
			NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
			KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{usage},
			IPAddresses: ips, DNSNames: []string{"localhost"},
		}
		der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
		require.NoError(t, err)
		keyDER, err := x509.MarshalECPrivateKey(key)
		require.NoError(t, err)
		return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
			pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	}

	write := func(name string, data []byte) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, data, 0o600))
		return path
	}

	p := pki{caFile: write("ca.pem", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}))}
	p.pool = x509.NewCertPool()
	p.pool.AddCert(ca)

	certPEM, keyPEM := issue(2, x509.ExtKeyUsageServerAuth, []net.IP{net.ParseIP("127.0.0.1"), net.IPv6loopback})
	p.serverCert, p.serverKey = write("server.pem", certPEM), write("server.key", keyPEM)

	certPEM, keyPEM = issue(3, x509.ExtKeyUsageClientAuth, nil)
	p.clientPair, err = tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)
	return p
}

// client is an HTTPS client that trusts the CA, with a certificate of its own
// when one is given.
func (p pki) client(cert *tls.Certificate) *http.Client {
	cfg := &tls.Config{RootCAs: p.pool, MinVersion: tls.VersionTLS12}
	if cert != nil {
		cfg.Certificates = []tls.Certificate{*cert}
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}, Timeout: 5 * time.Second}
}

func (p pki) post(t *testing.T, url string, cert *tls.Certificate) int {
	t.Helper()
	resp, err := p.client(cert).Post(url, "application/json", bytes.NewBufferString(ping))
	require.NoError(t, err)
	_ = resp.Body.Close()
	return resp.StatusCode
}
