package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/validation"
)

// What the MCP server may do, in PREFERENCES and CONNECT.
//
// The same settings in both windows: in PREFERENCES they are the top of the
// file, the most any connection may do; in CONNECT they are the connection's
// own, which can only narrow that. The server reads them when it starts, so a
// change applies the next time an MCP client starts it.
//
// The window is reached only from the shell, by the person at the keyboard.
// The MCP server has no tool that writes the file, so nothing a model is told
// can change what it is allowed.

const (
	mcpSection         = "MCP SERVER"
	mcpCommandsSection = "MCP SERVER - PERMITTED COMMANDS"
)

// mcpRestartNote goes after "saved" when the window holds the MCP settings.
const mcpRestartNote = " The MCP settings apply when an MCP client next starts cqlai mcp."

// mcpSpecs is the two MCP sections. The command rows come from the table the
// statement gate classifies by, so the window cannot offer a command the gate
// does not know.
func mcpSpecs() []prefSpec {
	specs := []prefSpec{
		{section: mcpSection, path: "MCP.Keyspaces", label: "Keyspaces", kind: prefList, hint: "the keyspaces the model can see, separated by commas; empty is all but the system ones"},
		{path: "MCP.Deny", label: "Hidden", kind: prefList, hint: "keyspace or keyspace.table never visible; system_auth always is"},
		{path: "MCP.Redact", label: "Redacted columns", kind: prefList, hint: "keyspace.table.column, with * for any part: the values are replaced"},
		{path: "MCP.AllowScans", label: "Allow scans", kind: prefYesNo, hint: "ALLOW FILTERING, and aggregates across partitions"},
		{path: "MCP.MaxRows", label: "Max rows", kind: prefNumber, hint: "the most rows one call returns; 100 unless set"},
		{path: "MCP.MaxValueBytes", label: "Max value bytes", kind: prefNumber, hint: "longer values are cut; 4096 unless set"},
		{path: "MCP.MaxCallsPerMinute", label: "Calls per minute", kind: prefNumber, hint: "60 unless set"},
		{path: "MCP.AuditLog", label: "Audit log", kind: prefPath, topOnly: true, hint: "~/.cassandra/cqlai_mcp_audit.log unless set; - turns it off"},
		{path: "MCP.Port", label: "Port", kind: prefNumber, topOnly: true, hint: "where `cqlai mcp` serves MCP, on 127.0.0.1 only; 7845 unless set"},
	}
	// The read commands only. A change is never run by the MCP server - a
	// model proposes it, and the user runs it - so there is nothing to permit.
	first := true
	for _, c := range validation.Commands {
		if c.Kind != validation.KindRead {
			continue
		}
		spec := prefSpec{
			path: "MCP.Permit", label: c.Name, kind: prefMember, member: c.Name,
			hint: c.Covers + " - changes are only ever proposed, for you to run",
		}
		if first {
			first = false
			spec.section = mcpCommandsSection
		}
		specs = append(specs, spec)
	}
	return specs
}

// permitOf is a configuration's permitted commands, or nil when it does not
// say.
func permitOf(cfg *config.Config) *[]string {
	if cfg == nil || cfg.MCP == nil {
		return nil
	}
	return cfg.MCP.Permit
}

// inheritedCommands is what a configuration that does not say permits: the
// defaults at the top of the file, and for a connection whatever the top of
// the file permits.
func (p preferences) inheritedCommands() []string {
	if p.purpose == connecting {
		if top := permitOf(p.cfg); top != nil {
			return *top
		}
	}
	return validation.DefaultCommands()
}

// loadMembers ticks the command rows for a configuration, and in CONNECT
// disables the ones the top of the file does not permit.
func (p *preferences) loadMembers(cfg *config.Config) {
	inherited := p.inheritedCommands()
	permitted := inherited
	if list := permitOf(cfg); list != nil {
		permitted = *list
	}
	for i := range p.fields {
		f := &p.fields[i]
		if f.spec.kind != prefMember {
			continue
		}
		f.yes = containsFold(permitted, f.spec.member)
		f.kept = f.yes && permitOf(cfg) != nil
		f.disabled = ""
		if p.purpose == connecting && !containsFold(inherited, f.spec.member) {
			f.yes = false
			f.disabled = "not permitted in PREFERENCES"
		}
	}
}

// applyPrefFields writes the window's settings into a configuration.
//
// The command rows are written together, as one list. A list the window
// started without, and was not changed from what that meant, is left out of
// the file: an unset list follows the defaults, or for a connection the top of
// the file, and writing it out would stop it following them.
func applyPrefFields(cfg *config.Config, fields []prefField, inherited []string) error {
	var ticked []string
	members := false
	for _, f := range fields {
		if f.spec.kind == prefMember {
			members = true
			if (f.yes && f.disabled == "") || (f.disabled != "" && f.kept) {
				ticked = append(ticked, f.spec.member)
			}
			continue
		}
		if err := setPrefValue(cfg, f.spec.path, f.value()); err != nil {
			return err
		}
	}
	if !members {
		return nil
	}
	if permitOf(cfg) == nil && sameNames(ticked, inherited) {
		return nil
	}
	if cfg.MCP == nil {
		cfg.MCP = &config.MCPConfig{}
	}
	list := append([]string{}, ticked...) // an empty list, not nil: nothing permitted
	cfg.MCP.Permit = &list
	return nil
}

// splitPrefList reads a comma-separated setting.
func splitPrefList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func containsFold(list []string, name string) bool {
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), name) {
			return true
		}
	}
	return false
}

func sameNames(a, b []string) bool {
	for _, x := range a {
		if !containsFold(b, x) {
			return false
		}
	}
	for _, x := range b {
		if !containsFold(a, x) {
			return false
		}
	}
	return true
}

// cqlNamePattern is what a keyspace or table name can be.
var cqlNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// prefListError is what is wrong with a list setting, or "".
func (p preferences) prefListError(spec prefSpec, value string) string {
	for _, item := range splitPrefList(value) {
		parts := strings.Split(item, ".")
		switch spec.path {
		case "MCP.Keyspaces":
			if len(parts) != 1 || !cqlNamePattern.MatchString(item) {
				return fmt.Sprintf("%s is not a keyspace name", item)
			}
		case "MCP.Deny":
			if len(parts) > 2 || !allMatch(parts, cqlNamePattern) {
				return fmt.Sprintf("%s has to be keyspace or keyspace.table", item)
			}
		case "MCP.Redact":
			if len(parts) != 3 {
				return fmt.Sprintf("%s has to be keyspace.table.column", item)
			}
			for _, part := range parts {
				if part != "*" && !cqlNamePattern.MatchString(part) {
					return fmt.Sprintf("%s has to be keyspace.table.column, with * for any part", item)
				}
			}
		}
	}
	return ""
}

func allMatch(parts []string, pattern *regexp.Regexp) bool {
	for _, part := range parts {
		if !pattern.MatchString(part) {
			return false
		}
	}
	return true
}

// listChoices is what Tab completes the last item of a list setting from.
func (m *MainModel) listChoices(spec prefSpec) []string {
	switch spec.path {
	case "MCP.Keyspaces", "MCP.Deny":
		return m.keyspaceChoices()
	}
	return nil
}

// splitListTail splits a list setting into what is already in it - up to and
// including its last comma - and the item being typed after that. Completing
// the item, by Tab or by picking from the list, keeps the head: replacing the
// whole value lost every item before the one being typed.
func splitListTail(value string) (head, last string) {
	i := strings.LastIndex(value, ",")
	if i < 0 {
		return "", strings.TrimSpace(value)
	}
	return strings.TrimRight(value[:i+1], " ") + " ", strings.TrimSpace(value[i+1:])
}

// completePrefList completes the item being typed at the end of a list.
func (m *MainModel) completePrefList(field *prefField) {
	value, matches := completeListValue(field.input.Value(), m.listChoices(field.spec))
	m.preferences.clearMatches()
	if value != field.input.Value() {
		m.setPrefField(m.preferences.focus, value)
	}
	m.preferences.matches = matches
}

// completeListValue completes the item being typed at the end of a list from
// the names given, keeping every item before it. It reports the new value,
// and the names to choose from when more than one would do. A name already in
// the list is not offered again.
func completeListValue(value string, names []string) (string, []string) {
	head, last := splitListTail(value)
	already := splitPrefList(head)

	var found []string
	for _, name := range names {
		if containsFold(already, name) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(last)) {
			found = append(found, name)
		}
	}
	switch len(found) {
	case 0:
		return value, nil
	case 1:
		return head + found[0], nil
	}
	return head + commonPrefix(found), found
}
