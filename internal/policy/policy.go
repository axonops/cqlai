// Package policy decides what the MCP server lets a model do.
//
// Every security decision is here, so it can be read and tested in one place:
// which commands may run, which keyspaces and tables can be seen, which
// values are hidden, and how much one model may ask of a cluster.
//
// The model is not trusted. It makes mistakes, and it can be steered by text
// it reads from the cluster. So nothing here depends on what the model says a
// statement is or means; it is decided from the statement and from settings
// only the user can write.
package policy

import (
	"fmt"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/validation"
)

// Policy is what one server, on one connection, may do.
//
// The zero value allows nothing: no command is permitted, no keyspace is
// visible, and the limits are zero. A code path that forgets to pass a policy
// fails closed. Every field is named so that its zero value is the strict one.
type Policy struct {
	permit map[string]bool // the commands that may run

	allKeyspaces bool // every keyspace is visible, but for deny
	// systemVisible lets every keyspace include the system ones. Without it,
	// a policy that allows every keyspace does not allow these. They describe
	// the other keyspaces: their names, their tables and columns, their sizes,
	// and the statements traced against them. Reading them would show what
	// deny hides. A keyspace list that names one still shows it.
	systemVisible bool
	keyspaces     map[string]bool // the visible ones, when not all of them

	deny   []string // keyspace, or keyspace.table, never visible
	redact []string // keyspace.table.column, with * for any part

	allowScans bool
	autoFetch  bool // a query returns every row, not one page

	maxRows        int
	maxValueBytes  int
	callsPerMinute int
	timeout        time.Duration

	connection string // the connection this policy is for
	auditLog   string

	// secrets are strings that must never be returned: the connection's
	// password. Driver errors pass through Scrub before they go out.
	secrets []string

	// partitionKey says which columns make up a table's partition key, for
	// telling a scan from a read of one partition. Nil, or an empty answer,
	// means it cannot be told, and the statement is treated as a scan.
	partitionKey func(keyspace, table string) []string

	// columns says which columns a table has, for telling whether a redaction
	// pattern applies to it. Nil, or an empty answer, means it cannot be told,
	// and any pattern that could apply is taken to.
	columns func(keyspace, table string) []string
}

// alwaysDenied is never visible, whatever the settings say. system_auth holds
// the password hashes.
var alwaysDenied = []string{"system_auth"}

// virtualKeyspaces are computed by the node and are small. Reading one is not
// a scan.
var virtualKeyspaces = map[string]bool{"system_views": true, "system_virtual_schema": true}

// Shell is the policy for cqlai's own CHAT view, which shows what the shell
// shows. It is asked for by name: nothing gets it by default.
func Shell() Policy {
	permit := map[string]bool{}
	for _, c := range validation.Commands {
		permit[c.Name] = true
	}
	return Policy{
		permit:         permit,
		allKeyspaces:   true,
		systemVisible:  true,
		allowScans:     true,
		maxRows:        1 << 30,
		maxValueBytes:  1 << 30,
		callsPerMinute: 1 << 30,
	}
}

// Permits reports whether a command may run.
func (p Policy) Permits(command string) bool {
	return p.permit[strings.ToUpper(command)]
}

// Permitted is the commands that may run, in the order of the table.
func (p Policy) Permitted() []string {
	var names []string
	for _, c := range validation.Commands {
		if p.permit[c.Name] {
			names = append(names, c.Name)
		}
	}
	return names
}

// PermitsAnyOf reports whether any command of a kind may run.
func (p Policy) PermitsAnyOf(kinds ...validation.Kind) bool {
	for _, c := range validation.Commands {
		for _, k := range kinds {
			if c.Kind == k && p.permit[c.Name] {
				return true
			}
		}
	}
	return false
}

// Visible reports whether a keyspace can be seen.
//
// A keyspace in the visible list has to match it exactly, as Cassandra
// stores the name; a denied one matches ignoring case. Both lean the same
// way: a near miss hides a keyspace rather than shows one.
func (p Policy) Visible(keyspace string) bool {
	if keyspace == "" || p.denied(keyspace, "") {
		return false
	}
	if p.allKeyspaces {
		return p.systemVisible || !IsSystemKeyspace(keyspace)
	}
	return p.keyspaces[keyspace]
}

// IsSystemKeyspace reports whether a keyspace is one of Cassandra's own:
// system, or any whose name starts system_. Taking every system_ name, not
// a list, hides one a newer Cassandra adds as well.
func IsSystemKeyspace(keyspace string) bool {
	k := strings.ToLower(keyspace)
	return k == "system" || strings.HasPrefix(k, "system_")
}

// WithSystemKeyspaces is the policy for a tool that writes its own statement
// against a system keyspace and filters what comes back, such as
// node_status. With every keyspace allowed, the system ones are allowed too.
// A deny entry, or a keyspace list that leaves one out, still holds.
func (p Policy) WithSystemKeyspaces() Policy {
	p.systemVisible = true
	return p
}

// VisibleTable reports whether a table can be seen.
func (p Policy) VisibleTable(keyspace, table string) bool {
	return p.Visible(keyspace) && !p.denied(keyspace, table)
}

func (p Policy) denied(keyspace, table string) bool {
	return Policy{deny: append(append([]string{}, alwaysDenied...), p.deny...)}.deniedBySettings(keyspace, table)
}

// deniedBySettings is denied by p.deny alone.
func (p Policy) deniedBySettings(keyspace, table string) bool {
	for _, d := range p.deny {
		ks, tbl, hasTable := strings.Cut(d, ".")
		if !strings.EqualFold(ks, keyspace) {
			continue
		}
		if !hasTable || (table != "" && strings.EqualFold(tbl, table)) {
			return true
		}
	}
	return false
}

// Redacted reports whether a column's values are hidden.
func (p Policy) Redacted(keyspace, table, column string) bool {
	for _, pattern := range p.redact {
		parts := strings.Split(pattern, ".")
		if len(parts) != 3 {
			continue
		}
		if matchPart(parts[0], keyspace) && matchPart(parts[1], table) && matchPart(parts[2], column) {
			return true
		}
	}
	return false
}

func matchPart(pattern, name string) bool {
	return pattern == "*" || strings.EqualFold(pattern, name)
}

// RedactedValue is what a hidden value is replaced with.
const RedactedValue = "[redacted]"

// MaskedSetting reports whether a node setting's value is hidden, by its
// name: one that holds a password, a secret or a credential. Cassandra masks
// some of these itself; this does not rely on that.
func MaskedSetting(name string) bool {
	lower := strings.ToLower(name)
	for _, word := range []string{"password", "secret", "credential"} {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

// Scrub removes anything secret from text going back to the model: a driver
// error can repeat what it was given.
func (p Policy) Scrub(text string) string {
	for _, secret := range p.secrets {
		if secret != "" {
			text = strings.ReplaceAll(text, secret, RedactedValue)
		}
	}
	return text
}

// The limits.

func (p Policy) MaxRows() int           { return p.maxRows }
func (p Policy) MaxValueBytes() int     { return p.maxValueBytes }
func (p Policy) CallsPerMinute() int    { return p.callsPerMinute }
func (p Policy) Timeout() time.Duration { return p.timeout }
func (p Policy) AllowScans() bool       { return p.allowScans }

// AutoFetch reports whether a query returns every row, paging through them
// itself, rather than one page and a token for the next.
func (p Policy) AutoFetch() bool            { return p.autoFetch }
func (p Policy) Connection() string         { return p.connection }
func (p Policy) AuditLog() string           { return p.auditLog }
func (p Policy) Deny() []string             { return append(append([]string{}, alwaysDenied...), p.deny...) }
func (p Policy) RedactPatterns() []string   { return append([]string(nil), p.redact...) }
func (p Policy) AllKeyspacesVisible() bool  { return p.allKeyspaces }
func (p Policy) VisibleKeyspaces() []string { return sortedKeys(p.keyspaces) }

// WithColumns is the policy, told how to find a table's columns.
func (p Policy) WithColumns(lookup func(keyspace, table string) []string) Policy {
	p.columns = lookup
	return p
}

// HasRedactedColumns reports whether any column of a table is redacted. When
// the table's columns cannot be found, a pattern that names the table, or
// matches any table, is taken to apply.
func (p Policy) HasRedactedColumns(keyspace, table string) bool {
	var columns []string
	if p.columns != nil {
		columns = p.columns(keyspace, table)
	}
	for _, pattern := range p.redact {
		parts := strings.Split(pattern, ".")
		if len(parts) != 3 || !matchPart(parts[0], keyspace) || !matchPart(parts[1], table) {
			continue
		}
		if parts[2] == "*" || len(columns) == 0 {
			return true
		}
		for _, c := range columns {
			if matchPart(parts[2], c) {
				return true
			}
		}
	}
	return false
}

// WithPartitionKey is the policy, told how to find a table's partition key.
func (p Policy) WithPartitionKey(lookup func(keyspace, table string) []string) Policy {
	p.partitionKey = lookup
	return p
}

// Check is the statement gate. Every statement a model sends passes through
// it before it reaches Cassandra, and it stops at the first rule refused.
//
// The error is written for the model to act on, and names nothing the policy
// hides beyond what the statement itself named.
func (p Policy) Check(text string) (validation.Statement, error) {
	s, err := validation.Classify(text)
	if err != nil {
		return s, refused(err.Error())
	}

	if !p.Permits(s.Command) {
		return s, refused(fmt.Sprintf("%s is not permitted on this server", s.Command))
	}
	for _, inner := range s.Inner {
		if !p.Permits(inner.Command) {
			return s, refused(fmt.Sprintf("the BATCH holds %s, which is not permitted on this server", inner.Command))
		}
	}

	for _, name := range s.Names {
		if !name.Qualified {
			return s, refused(fmt.Sprintf("name the keyspace: keyspace.%s, not %s", name.Table, name.Table))
		}
		if !p.Visible(name.Keyspace) {
			return s, refused(fmt.Sprintf("keyspace %s is not visible to this server", name.Keyspace))
		}
		if name.Table != "" && !p.VisibleTable(name.Keyspace, name.Table) {
			return s, refused(fmt.Sprintf("%s.%s is not visible to this server", name.Keyspace, name.Table))
		}
	}

	if err := p.checkRedaction(s); err != nil {
		return s, err
	}
	if s.Command == "SELECT" && len(s.Names) > 0 {
		n := s.Names[0]
		if err := p.checkNamingColumns(n.Keyspace+"."+n.Table, n.Keyspace, n.Table, s.PlainSelection); err != nil {
			return s, err
		}
	}
	if err := p.checkScan(s); err != nil {
		return s, err
	}
	return s, nil
}

// checkRedaction keeps a redacted column's values from coming back by another
// route. A value is hidden by its column's name, so a SELECT on a table with
// redacted columns has to return columns under their own names: no JSON, no
// alias, no function. And its WHERE clause cannot mention a redacted column,
// which would let the model test guesses at the values.
func (p Policy) checkRedaction(s validation.Statement) error {
	if s.Command != "SELECT" || len(s.Names) == 0 {
		return nil
	}
	name := s.Names[0]
	if !p.HasRedactedColumns(name.Keyspace, name.Table) {
		return nil
	}
	if !s.PlainSelection {
		return refused(fmt.Sprintf("%s.%s has hidden columns: select with * or by column name, with no JSON, alias or function",
			name.Keyspace, name.Table))
	}
	for _, column := range s.WhereNames {
		if p.Redacted(name.Keyspace, name.Table, column) {
			return refused(fmt.Sprintf("%s is hidden on this server and cannot be used in WHERE", column))
		}
	}
	return nil
}

// checkScan refuses a statement that reads far more than it returns, unless
// scans are allowed.
func (p Policy) checkScan(s validation.Statement) error {
	if p.allowScans || s.Command != "SELECT" || len(s.Names) == 0 {
		return nil
	}
	name := s.Names[0]
	if virtualKeyspaces[name.Keyspace] {
		return nil
	}

	if s.AllowFiltering {
		return refused("ALLOW FILTERING is not permitted on this server: it can read a whole table to return a few rows")
	}
	if !s.Aggregate {
		return nil
	}

	var key []string
	if p.partitionKey != nil {
		key = p.partitionKey(name.Keyspace, name.Table)
	}
	if len(key) == 0 {
		return refused("an aggregate needs the whole partition key in its WHERE clause, and this table's could not be found")
	}
	restricted := map[string]bool{}
	for _, c := range s.Restricted {
		restricted[c] = true
	}
	for _, column := range key {
		if !restricted[column] {
			return refused(fmt.Sprintf("an aggregate has to fix the whole partition key (%s) with = or IN: across partitions it reads every row first",
				strings.Join(key, ", ")))
		}
	}
	return nil
}

// CheckProposal checks a change a model proposes for the user to run. It is
// never run here, so the permitted commands do not apply; but it has to be one
// statement that changes data or schema, with every table named in full and
// visible, so a proposal cannot be the way to reach what the policy hides.
func (p Policy) CheckProposal(text string) (validation.Statement, error) {
	s, err := validation.Classify(text)
	if err != nil {
		return s, refused(err.Error())
	}
	if s.Kind == validation.KindRead {
		return s, refused(fmt.Sprintf("%s does not change anything: run it with query or describe", s.Command))
	}
	for _, name := range s.Names {
		if !name.Qualified {
			return s, refused(fmt.Sprintf("name the keyspace: keyspace.%s, not %s", name.Table, name.Table))
		}
		if !p.Visible(name.Keyspace) {
			return s, refused(fmt.Sprintf("keyspace %s is not visible to this server", name.Keyspace))
		}
		if name.Table != "" && !p.VisibleTable(name.Keyspace, name.Table) {
			return s, refused(fmt.Sprintf("%s.%s is not visible to this server", name.Keyspace, name.Table))
		}
	}
	return s, nil
}

// Refusal is the gate saying no. It is an error so callers cannot miss it.
type Refusal struct{ Reason string }

func (r Refusal) Error() string { return "refused: " + r.Reason }

func refused(reason string) error { return Refusal{Reason: reason} }
