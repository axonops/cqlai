package mcp

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/pflag"

	"github.com/axonops/cqlai/internal/policy"
)

// Options is what `cqlai mcp` was started with.
type Options struct {
	// Headless is a server on stdin and stdout and nothing else. Without it,
	// cqlai mcp is the terminal app, serving MCP over HTTP on this machine
	// for whatever connection is picked in it.
	Headless bool

	Flags      policy.Flags
	Port       int // HTTP, for the terminal app; 0 is the settings', or 7845
	ConfigFile string

	// The terminal app's HTTP, over the settings: where it listens, whether
	// the token is required (nil is the settings'), and TLS.
	Listen      string
	Token       *bool
	TLSCert     string
	TLSKey      string
	TLSClientCA string

	// Headless only: which connection, and overrides for it.
	Connection     string
	Host           string
	PortOverride   int
	Username       string
	SSL            bool
	ConnectTimeout int
	RequestTimeout int
	Debug          bool
}

// DefaultPort is where the terminal app serves MCP unless told otherwise.
const DefaultPort = 7845

// ParseOptions reads the flags after `cqlai mcp`. It reports whether to carry
// on, and the exit code when not: 0 after --help, 2 for a flag it cannot read.
func ParseOptions(args []string, stderr io.Writer) (Options, bool, int) {
	flags := pflag.NewFlagSet("cqlai mcp", pflag.ContinueOnError)
	flags.SetOutput(stderr)

	var o Options
	var permit, keyspaces []string
	flags.BoolVar(&o.Headless, "headless", false, "No terminal app: an MCP server on stdin and stdout, for a client that starts it")
	flags.IntVar(&o.Port, "port", 0, "Terminal app: the port MCP is served on (default: the settings', or 7845)")
	flags.StringVar(&o.Listen, "listen", "", "Terminal app: the address MCP is served on (default: the settings', or 127.0.0.1)")
	token := flags.Bool("token", false, "Terminal app: every request has to carry the token in ~/.cassandra/cqlai_mcp_token (default: the settings', or off)")
	flags.StringVar(&o.TLSCert, "tls-cert", "", "Terminal app: serve HTTPS with this certificate (PEM)")
	flags.StringVar(&o.TLSKey, "tls-key", "", "Terminal app: the private key for --tls-cert (PEM)")
	flags.StringVar(&o.TLSClientCA, "tls-client-ca", "", "Terminal app: ask each client for a certificate this CA signed (PEM)")
	flags.StringSliceVar(&permit, "permit", nil, "Permit only these CQL commands, of those the config file permits")
	flags.StringSliceVar(&keyspaces, "keyspaces", nil, "Only these keyspaces are visible, of those the config file allows")
	flags.IntVar(&o.Flags.MaxRows, "max-rows", 0, "The page size, if lower than the config file's")
	flags.StringVar(&o.Flags.AuditLog, "audit-log", "", "Where the audit log goes (\"-\" turns it off)")
	flags.StringVar(&o.ConfigFile, "config-file", "", "Path to config file (overrides default locations)")
	flags.StringVar(&o.Connection, "connection", "", "Headless: the saved connection to use (default: the first saved one)")
	flags.StringVar(&o.Host, "host", "", "Headless: Cassandra host (overrides the connection)")
	flags.IntVar(&o.PortOverride, "cassandra-port", 0, "Headless: Cassandra port (overrides the connection)")
	flags.StringVarP(&o.Username, "username", "u", "", "Headless: username (overrides the connection)")
	flags.BoolVar(&o.SSL, "ssl", false, "Headless: enable SSL/TLS")
	flags.IntVar(&o.ConnectTimeout, "connect-timeout", 10, "Connection timeout in seconds")
	flags.IntVar(&o.RequestTimeout, "request-timeout", 10, "Request timeout in seconds; also the longest one tool call may run")
	flags.BoolVar(&o.Debug, "debug", false, "Write the debug log")
	help := flags.BoolP("help", "h", false, "Show this help")

	if err := flags.Parse(args); err != nil {
		return o, false, 2
	}
	if *help {
		fmt.Fprintln(stderr, "cqlai mcp - Cassandra for an AI client, over the Model Context Protocol")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Usage: cqlai mcp [options]              the terminal app, serving MCP at http://127.0.0.1:7845/mcp unless set")
		fmt.Fprintln(stderr, "       cqlai mcp --headless [options]   an MCP server on stdin and stdout, for a client that starts it")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "What the server may do is set in PREFERENCES and CONNECT, under MCP SERVER.")
		fmt.Fprintln(stderr, "These options can only narrow it.")
		fmt.Fprintln(stderr)
		flags.PrintDefaults()
		return o, false, 0
	}

	if flags.Changed("permit") {
		o.Flags.Permit = permit
		if o.Flags.Permit == nil {
			o.Flags.Permit = []string{} // --permit= permits nothing
		}
	}
	o.Flags.Keyspaces = keyspaces
	if flags.Changed("token") {
		o.Token = token
	}
	if o.ConfigFile == "" {
		o.ConfigFile = os.Getenv("CQLAI_CONFIG_FILE")
	}
	return o, true, 0
}
