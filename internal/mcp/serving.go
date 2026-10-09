package mcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/axonops/cqlai/internal/config"
)

// DefaultListen is where the terminal app serves MCP unless told otherwise:
// this machine only.
const DefaultListen = "127.0.0.1"

// serving is how the terminal app serves MCP over HTTP: where, whether every
// request has to carry the token, and with what TLS.
type serving struct {
	listen   string // an address or a host name, without the port
	port     int
	token    bool
	cert     string // TLS: the server's certificate and key, PEM
	key      string
	clientCA string // TLS: the CA a client's certificate has to be signed by
}

// servingFrom is the settings at the top of the file, with the flags over
// them.
func servingFrom(file *config.Config, o Options) serving {
	s := serving{listen: DefaultListen, port: DefaultPort}
	if m := file.MCP; m != nil {
		if m.Listen != "" {
			s.listen = m.Listen
		}
		if m.Port != 0 {
			s.port = m.Port
		}
		s.token = m.Token
		s.cert, s.key, s.clientCA = m.TLSCert, m.TLSKey, m.TLSClientCA
	}
	if o.Listen != "" {
		s.listen = o.Listen
	}
	if o.Port != 0 {
		s.port = o.Port
	}
	if o.Token != nil {
		s.token = *o.Token
	}
	if o.TLSCert != "" {
		s.cert = o.TLSCert
	}
	if o.TLSKey != "" {
		s.key = o.TLSKey
	}
	if o.TLSClientCA != "" {
		s.clientCA = o.TLSClientCA
	}
	return s
}

// tls reports whether MCP is served over HTTPS.
func (s serving) tls() bool { return s.cert != "" }

// local reports whether the address is this machine's own, which nothing
// else on the network can reach.
func (s serving) local() bool {
	return isLoopback(s.listen)
}

// isLoopback reports whether an address or host name is this machine's own.
// A wildcard - 0.0.0.0 or :: - is every address, and is not.
func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

// check says what is wrong with the settings, or nil.
//
// Off this machine, the network can reach the server. Then it has to be
// HTTPS, so the token and the results are not sent in the clear, and a
// client has to prove who it is, with the token or a client certificate.
func (s serving) check() error {
	switch {
	case (s.cert == "") != (s.key == ""):
		return errors.New("MCP over TLS needs both tlsCert and tlsKey")
	case s.clientCA != "" && !s.tls():
		return errors.New("tlsClientCA needs tlsCert and tlsKey: client certificates are part of TLS")
	case strings.TrimSpace(s.listen) == "":
		return errors.New("listen has to be an address or a host name")
	case !s.local() && !s.tls():
		return fmt.Errorf("MCP is served on %s, which the network can reach, only over TLS: set tlsCert and tlsKey", s.listen)
	case !s.local() && !s.token && s.clientCA == "":
		return fmt.Errorf("MCP is served on %s, which the network can reach, only to a client that proves who it is: turn on the token, or set tlsClientCA", s.listen)
	}
	return nil
}

// address is where to listen.
func (s serving) address() string {
	return net.JoinHostPort(strings.Trim(s.listen, "[]"), strconv.Itoa(s.port))
}

// url is the address a client connects to. A wildcard listen is every
// address this machine has, of which 127.0.0.1 is the one sure to be there.
func (s serving) url() string {
	host := strings.Trim(s.listen, "[]")
	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		host = DefaultListen
	}
	scheme := "http"
	if s.tls() {
		scheme = "https"
	}
	return scheme + "://" + net.JoinHostPort(host, strconv.Itoa(s.port)) + "/mcp"
}

// tlsConfig is the server's TLS, from the files the settings name.
func (s serving) tlsConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(s.cert, s.key)
	if err != nil {
		return nil, fmt.Errorf("cannot read the MCP TLS certificate and key: %w", err)
	}
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{cert}}
	if s.clientCA != "" {
		pem, err := os.ReadFile(s.clientCA) // #nosec G304 - the CA file the user named in the settings
		if err != nil {
			return nil, fmt.Errorf("cannot read the MCP client CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("%s holds no PEM certificate", s.clientCA)
		}
		cfg.ClientCAs = pool
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return cfg, nil
}
