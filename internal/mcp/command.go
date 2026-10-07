package mcp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/policy"
	"github.com/axonops/cqlai/internal/session"
)

// RunHeadless is `cqlai mcp --headless`: an MCP server on stdin and stdout,
// for an AI client that starts it as a child process. It returns the exit
// code.
//
// Nothing may be written to stdout but protocol messages, and nothing read
// from stdin but them: there is no password prompt here, and everything said
// to a person goes to stderr.
func RunHeadless(o Options, version string) int {
	if o.Debug || os.Getenv("CQLAI_DEBUG") == "true" || os.Getenv("CQLAI_DEBUG") == "1" {
		logger.SetDebugEnabled(true)
	}
	fail := func(format string, a ...any) int {
		fmt.Fprintf(os.Stderr, "cqlai mcp: "+format+"\n", a...)
		return 1
	}

	cfg, err := config.LoadConfig(o.ConfigFile)
	if err != nil {
		cfg = &config.Config{Host: "127.0.0.1", Port: 9042}
	}

	// Which connection: the one named, or the first saved one, or the
	// settings at the top of the file when there are none saved. A name that
	// is not there is a mistake in the settings, which stops the server: it
	// cannot know what it was meant to serve.
	conn, name := *cfg, ""
	switch {
	case o.Connection != "":
		saved, ok := cfg.Connection(o.Connection)
		if !ok {
			return fail("there is no saved connection called %q in %s", o.Connection, cfg.SavePath())
		}
		conn, name = saved, config.ConnectionName(saved)
	default:
		if saved, ok := cfg.DefaultConnection(); ok {
			conn, name = saved, config.ConnectionName(saved)
		}
	}
	if o.Host != "" {
		conn.Host = o.Host
	}
	if o.PortOverride != 0 {
		conn.Port = o.PortOverride
	}
	if o.Username != "" {
		conn.Username = o.Username
	}
	if env := os.Getenv("CQLAI_PASSWORD"); env != "" {
		conn.Password = env
	}
	if conn.Username != "" && conn.Password == "" {
		return fail("connection %q has a username and no password: save one with the connection, or set CQLAI_PASSWORD", name)
	}
	if o.SSL {
		if conn.SSL == nil {
			conn.SSL = &config.SSLConfig{}
		}
		conn.SSL.Enabled = true
	}
	if conn.Host == "" {
		conn.Host = "127.0.0.1"
	}

	timeout := time.Duration(o.RequestTimeout) * time.Second
	pol, err := policy.Load(cfg, name, conn.Password, timeout, o.Flags)
	if err != nil {
		return fail("%v", err)
	}

	where := describeConnection(conn, name, cfg.SavePath())
	mgr := session.NewManager(cfg)
	link := newCluster(conn, cfg.Consistency, o, where, pol.Scrub, mgr)
	defer link.close()

	audit, err := policy.OpenAudit(pol.AuditLog())
	if err != nil {
		return fail("%v", err)
	}
	defer audit.Close()
	if pol.AuditLog() == policy.AuditOff {
		fmt.Fprintln(os.Stderr, "cqlai mcp: warning: the audit log is off")
	}
	fmt.Fprintf(os.Stderr, "cqlai mcp: serving %s; permitted: %s; audit log: %s\n",
		where, strings.Join(pol.Permitted(), ", "), pol.AuditLog())

	link.start()
	go func() {
		// Say on stderr how the first attempt went, for whoever reads the
		// client's log. The server is serving either way.
		link.waitForFirstAttempt(context.Background())
		if _, _, reason := link.connected(); reason != "" {
			fmt.Fprintf(os.Stderr, "cqlai mcp: %s; trying again when a tool is called\n", reason)
		} else {
			fmt.Fprintf(os.Stderr, "cqlai mcp: connected to %s\n", where)
		}
	}()

	server := NewServer(fixed{pol: pol, link: link}, policy.NewLimiter(pol.CallsPerMinute()), audit, version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go server.WatchSchema(ctx, 5*time.Second)
	if err := server.Serve(ctx, os.Stdin, os.Stdout); err != nil && ctx.Err() == nil {
		return fail("%v", err)
	}
	return 0
}

// describeConnection says which connection, and where it came from: with
// several configuration files, the answer is not obvious.
func describeConnection(conn config.Config, name, file string) string {
	if name != "" {
		return fmt.Sprintf("%s:%d (saved connection %q in %s)", conn.Host, portOr9042(conn.Port), name, file)
	}
	return fmt.Sprintf("%s:%d (the settings at the top of %s)", conn.Host, portOr9042(conn.Port), file)
}

// newCluster is a connection the server makes when it can, to conn.
func newCluster(conn config.Config, consistency string, o Options, where string, scrub func(string) string, mgr *session.Manager) *cluster {
	return &cluster{
		opts: db.SessionOptions{
			Host:           conn.Host,
			Port:           conn.Port,
			Keyspace:       conn.Keyspace,
			Username:       conn.Username,
			Password:       conn.Password,
			Consistency:    consistency,
			SSL:            conn.SSL,
			BatchMode:      true,
			ConnectTimeout: o.ConnectTimeout,
			RequestTimeout: o.RequestTimeout,
			ConfigFile:     o.ConfigFile,
		},
		where: where,
		scrub: scrub,
		mgr:   mgr,
		retry: 5 * time.Second,
	}
}

// portOr9042 is the port a connection uses, which is 9042 when it does not
// say.
func portOr9042(port int) int {
	if port == 0 {
		return 9042
	}
	return port
}
