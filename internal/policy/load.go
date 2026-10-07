package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/validation"
)

// The defaults for the limits, when nothing sets them.
const (
	DefaultMaxRows        = 100
	DefaultMaxValueBytes  = 4096
	DefaultCallsPerMinute = 60
)

// Flags is what the command line says. It can only narrow what the file
// allows.
type Flags struct {
	Permit    []string // --permit: only these, of what the file permits; nil narrows nothing
	Keyspaces []string // --keyspaces
	MaxRows   int      // --max-rows; 0 narrows nothing
	AuditLog  string   // --audit-log; "-" is off
}

// Load builds the policy for one connection from three places: the top of the
// file, the connection's own block, and the flags. Each part is the strictest
// of the three. A connection can narrow the top of the file and the flags can
// narrow both; nothing can widen them.
//
// connection is the saved connection's name, or "" when the file has none and
// its top level is the connection. password is that connection's, so it can
// be kept out of anything returned.
func Load(file *config.Config, connection string, password string, timeout time.Duration, flags Flags) (Policy, error) {
	var top, own *config.MCPConfig
	if file != nil {
		top = file.MCP
		if connection != "" {
			if conn, ok := file.Connection(connection); ok {
				own = conn.MCP
			}
		}
	}

	p := Policy{connection: connection, timeout: timeout}
	if password != "" {
		p.secrets = []string{password}
	}

	// Commands: permitted only where every place permits them.
	permit, err := commandSet(top)
	if err != nil {
		return Policy{}, fmt.Errorf("the mcp block in the file: %w", err)
	}
	// A connection that does not say which commands leaves them as the top of
	// the file has them; one that does can only narrow them.
	if own != nil && own.Permit != nil {
		ownSet, err := namedSet(*own.Permit)
		if err != nil {
			return Policy{}, fmt.Errorf("connection %s: %w", connection, err)
		}
		permit = intersect(permit, ownSet)
	}
	if flags.Permit != nil {
		flagSet, err := namedSet(flags.Permit)
		if err != nil {
			return Policy{}, fmt.Errorf("--permit: %w", err)
		}
		permit = intersect(permit, flagSet)
	}
	p.permit = permit

	// Keyspaces: the overlap of every list given. No list anywhere is all.
	lists := [][]string{}
	for _, m := range []*config.MCPConfig{top, own} {
		if m != nil && len(m.Keyspaces) > 0 {
			lists = append(lists, m.Keyspaces)
		}
	}
	if len(flags.Keyspaces) > 0 {
		lists = append(lists, flags.Keyspaces)
	}
	if len(lists) == 0 {
		p.allKeyspaces = true
	} else {
		visible := toSet(lists[0])
		for _, list := range lists[1:] {
			visible = intersect(visible, toSet(list))
		}
		p.keyspaces = visible
	}

	// Deny and redact: everything any place lists.
	for _, m := range []*config.MCPConfig{top, own} {
		if m == nil {
			continue
		}
		p.deny = append(p.deny, trimmed(m.Deny)...)
		for _, pattern := range trimmed(m.Redact) {
			if parts := strings.Split(pattern, "."); len(parts) != 3 || slicesHasEmpty(parts) {
				return Policy{}, fmt.Errorf("redact %q has to be keyspace.table.column, with * for any part", pattern)
			}
			p.redact = append(p.redact, pattern)
		}
	}

	// A yes/no that loosens is on only if the top turns it on, and the
	// connection's block, when it has one, does too.
	p.allowScans = top != nil && top.AllowScans && (own == nil || own.AllowScans)

	// Limits: the lowest that is set, or the default.
	p.maxRows = lowest(DefaultMaxRows, field(top, own, func(m *config.MCPConfig) int { return m.MaxRows }), flags.MaxRows)
	p.maxValueBytes = lowest(DefaultMaxValueBytes, field(top, own, func(m *config.MCPConfig) int { return m.MaxValueBytes }))
	p.callsPerMinute = lowest(DefaultCallsPerMinute, field(top, own, func(m *config.MCPConfig) int { return m.MaxCallsPerMinute }))

	// Server-wide: the audit log.
	p.auditLog = DefaultAuditLog()
	if top != nil && top.AuditLog != "" {
		p.auditLog = top.AuditLog
	}
	if flags.AuditLog != "" {
		p.auditLog = flags.AuditLog
	}
	return p, nil
}

// DefaultAuditLog is where the audit log goes unless the file or the flags say
// otherwise: beside cqlai's other files in the home directory.
func DefaultAuditLog() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cqlai_mcp_audit.log"
	}
	return filepath.Join(home, ".cqlai_mcp_audit.log")
}

// commandSet is the commands a block permits: the defaults when it does not
// say, and nothing when it says an empty list.
func commandSet(m *config.MCPConfig) (map[string]bool, error) {
	if m == nil || m.Permit == nil {
		return toSet(validation.DefaultCommands()), nil
	}
	return namedSet(*m.Permit)
}

// namedSet is a set of commands by name. A name that is not a command, or is
// one that is never permitted, is an error: a typo must not quietly permit or
// refuse something other than what was meant.
func namedSet(names []string) (map[string]bool, error) {
	set := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		command, ok := validation.CommandNamed(name)
		if !ok {
			return nil, fmt.Errorf("%q is not a command that can be permitted (%s)", name, validation.NeverPermitted)
		}
		if command.Kind != validation.KindRead {
			return nil, fmt.Errorf("%s changes data or schema, which the MCP server never runs: it proposes the statement for you to run. "+
				"Take %s out of the permitted commands", command.Name, command.Name)
		}
		set[command.Name] = true
	}
	return set, nil
}

func field(top, own *config.MCPConfig, get func(*config.MCPConfig) int) int {
	n := 0
	for _, m := range []*config.MCPConfig{top, own} {
		if m != nil && get(m) > 0 && (n == 0 || get(m) < n) {
			n = get(m)
		}
	}
	return n
}

// lowest is the smallest of the values that are set, or the default.
func lowest(def int, values ...int) int {
	n := 0
	for _, v := range values {
		if v > 0 && (n == 0 || v < n) {
			n = v
		}
	}
	if n == 0 {
		return def
	}
	return n
}

func toSet(names []string) map[string]bool {
	set := map[string]bool{}
	for _, n := range trimmed(names) {
		set[n] = true
	}
	return set
}

func intersect(a, b map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k := range a {
		if b[k] {
			out[k] = true
		}
	}
	return out
}

func trimmed(list []string) []string {
	var out []string
	for _, s := range list {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func slicesHasEmpty(parts []string) bool {
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			return true
		}
	}
	return false
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
