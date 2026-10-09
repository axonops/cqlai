package ui

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/policy"
	"github.com/axonops/cqlai/internal/validation"
)

// commandRow is the position of a command's row.
func commandRow(t *testing.T, m *MainModel, command string) int {
	t.Helper()
	for i, f := range m.preferences.fields {
		if f.spec.kind == prefMember && f.spec.member == command {
			return i
		}
	}
	t.Fatalf("no row for %s", command)
	return -1
}

func toggleCommand(t *testing.T, m *MainModel, command string) *MainModel {
	t.Helper()
	m.preferences.focusField(commandRow(t, m, command))
	return press(m, " ")
}

func ticked(m *MainModel) []string {
	var names []string
	for _, f := range m.preferences.fields {
		if f.spec.kind == prefMember && f.yes {
			names = append(names, f.spec.member)
		}
	}
	return names
}

// savedMCP saves the window and reads back the file's mcp block as written.
func savedMCP(t *testing.T, m *MainModel) map[string]any {
	t.Helper()
	path := m.preferences.path
	m.savePreferences()
	require.False(t, m.preferences.active, "the window should have saved and closed")

	data, err := os.ReadFile(path) //nolint:gosec // the test wrote it
	require.NoError(t, err)
	var file map[string]any
	require.NoError(t, json.Unmarshal(data, &file))
	block, _ := file["mcp"].(map[string]any)
	return block
}

// TestTheCommandRowsAreTheGatesTable: the window offers exactly the read
// commands the statement gate knows, in its order. A change is never run by
// the MCP server, so there is no row to permit one.
func TestTheCommandRowsAreTheGatesTable(t *testing.T) {
	m := prefModel(t, &config.Config{})

	var rows []string
	for _, f := range m.preferences.fields {
		if f.spec.kind == prefMember {
			rows = append(rows, f.spec.member)
		}
	}
	var table []string
	for _, c := range validation.Commands {
		if c.Kind == validation.KindRead {
			table = append(table, c.Name)
		}
	}
	assert.Equal(t, table, rows)
	assert.Equal(t, []string{"SELECT", "DESCRIBE", "LIST"}, rows)

	for _, path := range []string{"MCP.Keyspaces", "MCP.Deny", "MCP.Redact", "MCP.AllowScans",
		"MCP.MaxRows", "MCP.AuditLog", "MCP.Listen", "MCP.Port", "MCP.Token",
		"MCP.TLSCert", "MCP.TLSKey", "MCP.TLSClientCA"} {
		prefIndex(t, m, path)
	}
}

// TestUnsetCommandsShowTheDefaultsAndStayUnset: opening and saving the window
// must not write the defaults into the file, or they would stop being
// defaults.
func TestUnsetCommandsShowTheDefaultsAndStayUnset(t *testing.T) {
	m := prefModel(t, &config.Config{})
	assert.Equal(t, validation.DefaultCommands(), ticked(m))

	assert.Nil(t, savedMCP(t, m), "nothing was changed, so nothing is written")
}

// TestUntickingEverythingPermitsNothing, which the file has to keep apart from
// leaving the list out.
func TestUntickingEverythingPermitsNothing(t *testing.T) {
	m := prefModel(t, &config.Config{})
	for _, c := range validation.DefaultCommands() {
		m = toggleCommand(t, m, c)
	}
	assert.Empty(t, ticked(m))

	block := savedMCP(t, m)
	require.NotNil(t, block)
	assert.Equal(t, []any{}, block["permit"], "an empty list, not a missing one")

	// And the server reads it as nothing.
	file, err := config.LoadConfig(m.config.SourcePath)
	require.NoError(t, err)
	p, err := policy.Load(file, "", "", time.Second, policy.Flags{})
	require.NoError(t, err)
	assert.Empty(t, p.Permitted())
}

// TestUntickingACommandIsSaved.
func TestUntickingACommandIsSaved(t *testing.T) {
	m := prefModel(t, &config.Config{})
	m = toggleCommand(t, m, "LIST")

	block := savedMCP(t, m)
	assert.Equal(t, []any{"SELECT", "DESCRIBE"}, block["permit"])
}

// TestNothingAboutChangesIsOffered: no row for a write or schema command,
// nothing to confirm, and no connection switching. A command of that kind
// written into the file by hand is dropped on saving: the server refuses to
// start with it there.
func TestNothingAboutChangesIsOffered(t *testing.T) {
	permit := []string{"SELECT", "INSERT"}
	m := prefModel(t, &config.Config{MCP: &config.MCPConfig{Permit: &permit}})
	for _, f := range m.preferences.fields {
		assert.NotContains(t, []string{"INSERT", "DROP", "TRUNCATE"}, f.spec.member)
		assert.NotEqual(t, "MCP.SkipConfirm", f.spec.path)
		assert.NotEqual(t, "MCP.Connections", f.spec.path)
	}
	m = toggleCommand(t, m, "LIST")

	block := savedMCP(t, m)
	assert.Equal(t, []any{"SELECT", "LIST"}, block["permit"])
}

// TestTheListsRoundTrip through the file.
func TestTheListsRoundTrip(t *testing.T) {
	m := prefModel(t, &config.Config{})
	m.setPrefField(prefIndex(t, m, "MCP.Keyspaces"), "shop,  catalog ,")
	m.setPrefField(prefIndex(t, m, "MCP.Redact"), "shop.customers.email, *.*.card_number")
	m.preferences.focusField(prefIndex(t, m, "MCP.AllowScans"))
	m = press(m, " ")

	block := savedMCP(t, m)
	assert.Equal(t, []any{"shop", "catalog"}, block["keyspaces"])
	assert.Equal(t, []any{"shop.customers.email", "*.*.card_number"}, block["redact"])
	assert.Equal(t, true, block["allowScans"])
	assert.Nil(t, block["permit"], "the commands were not touched")

	reopened := prefModel(t, &config.Config{SourcePath: m.config.SourcePath, MCP: &config.MCPConfig{
		Keyspaces: []string{"shop", "catalog"},
	}})
	assert.Equal(t, "shop, catalog", reopened.preferences.fields[prefIndex(t, reopened, "MCP.Keyspaces")].value())
}

// TestAWrongListSaysWhy and the window will not save it.
func TestAWrongListSaysWhy(t *testing.T) {
	for path, value := range map[string]string{
		"MCP.Keyspaces": "shop.orders",
		"MCP.Deny":      "a.b.c",
		"MCP.Redact":    "shop.email",
	} {
		m := prefModel(t, &config.Config{})
		i := prefIndex(t, m, path)
		m.setPrefField(i, value)
		assert.Contains(t, m.preferenceErrors(), i, "%s = %q", path, value)
		assert.False(t, m.preferencesReady())
	}
}

// TestAConnectionCanOnlyNarrow: in CONNECT, a command the top of the file does
// not permit cannot be ticked, and a connection that is not changed keeps
// following the top of the file.
func TestAConnectionCanOnlyNarrow(t *testing.T) {
	top := []string{"SELECT", "LIST"}
	m := connectModel(t, &config.Config{
		MCP:         &config.MCPConfig{Permit: &top},
		Connections: []config.Config{{Name: "prod", Host: "10.0.0.5"}},
	})

	assert.Equal(t, []string{"SELECT", "LIST"}, ticked(m), "a connection that does not say follows the top")
	describe := m.preferences.fields[commandRow(t, m, "DESCRIBE")]
	assert.Equal(t, "not permitted in PREFERENCES", describe.disabled)

	m = toggleCommand(t, m, "DESCRIBE")
	assert.NotContains(t, ticked(m), "DESCRIBE", "a disabled row does not tick")

	// Unchanged, the connection is stored without a list of its own.
	m.storeConnection()
	assert.Nil(t, permitOf(&m.preferences.connections[m.preferences.chosen]))

	// Narrowed, it is stored with one.
	m = toggleCommand(t, m, "LIST")
	m.storeConnection()
	stored := permitOf(&m.preferences.connections[m.preferences.chosen])
	require.NotNil(t, stored)
	assert.Equal(t, []string{"SELECT"}, *stored)

	// What is the whole server's is not offered per connection.
	for _, f := range m.preferences.fields {
		assert.NotEqual(t, "MCP.Connections", f.spec.path)
		for _, serverWide := range []string{"MCP.AuditLog", "MCP.Listen", "MCP.Port", "MCP.Token",
			"MCP.TLSCert", "MCP.TLSKey", "MCP.TLSClientCA"} {
			assert.NotEqual(t, serverWide, f.spec.path)
		}
	}
}

// TestAConnectionsBlockIsKeptWithIt when it is saved, and not copied to the top
// when it is made the default.
func TestAConnectionsBlockIsKeptWithIt(t *testing.T) {
	permit := []string{"SELECT"}
	conn := config.Config{Name: "prod", Host: "10.0.0.5", MCP: &config.MCPConfig{Permit: &permit}}

	kept := config.ConnectionSettings(conn)
	require.NotNil(t, kept.MCP)
	assert.Equal(t, []string{"SELECT"}, *kept.MCP.Permit)
	(*kept.MCP.Permit)[0] = "DROP"
	assert.Equal(t, "SELECT", permit[0], "a copy, sharing nothing")

	file := &config.Config{}
	file.MakeDefault(conn)
	assert.Nil(t, file.MCP, "the top of the file is every connection's, not the default one's")
}

// TestAReportOfSeveralLinesIsNotPadded: each line as it is, so a long one does
// not make the others wrap onto blank lines.
func TestAReportOfSeveralLinesIsNotPadded(t *testing.T) {
	m := prefModel(t, &config.Config{})
	m.closePreferences()
	m.fullHistoryContent = ""
	m, _ = m.report("a long first line, much longer than the rest of them\n{\n}")
	for _, line := range strings.Split(ansi.Strip(m.fullHistoryContent), "\n")[1:] {
		assert.Equal(t, strings.TrimRight(line, " "), line, "no padding after %q", line)
	}
}

// TestCompletingAListKeepsWhatIsInIt: Tab completes the item after the last
// comma, and leaves the ones before it - it used to replace the whole value.
func TestCompletingAListKeepsWhatIsInIt(t *testing.T) {
	names := []string{"catalog", "shop", "shop_archive", "system_views"}
	for _, c := range []struct {
		value, want string
		matches     []string
	}{
		{"ca", "catalog", nil},
		{"catalog, sh", "catalog, shop", []string{"shop", "shop_archive"}},
		{"catalog,sy", "catalog, system_views", nil},
		{"catalog, shop, sh", "catalog, shop, shop_archive", nil},                     // shop is in the list already
		{"catalog, ", "catalog, s", []string{"shop", "shop_archive", "system_views"}}, // as far as they agree
		{"catalog, zz", "catalog, zz", nil},
	} {
		value, matches := completeListValue(c.value, names)
		assert.Equal(t, c.want, value, c.value)
		assert.Equal(t, c.matches, matches, c.value)
	}
}

// TestPickingFromTheListKeepsWhatIsInIt: choosing one of the names offered
// replaces the item being typed, not the list.
func TestPickingFromTheListKeepsWhatIsInIt(t *testing.T) {
	m := prefModel(t, &config.Config{})
	i := prefIndex(t, m, "MCP.Keyspaces")
	m.preferences.focusField(i)
	m.setPrefField(i, "catalog, sh")
	m.preferences.matches = []string{"shop", "shop_archive"}
	m.preferences.match = 1

	m.usePrefMatch()

	assert.Equal(t, "catalog, shop_archive", m.preferences.fields[i].input.Value())
}

// TestTheListenAddressIsAnAddressOrAHostName, and the window will not save
// anything else.
func TestTheListenAddressIsAnAddressOrAHostName(t *testing.T) {
	for value, wrong := range map[string]bool{
		"127.0.0.1": false, "0.0.0.0": false, "::1": false, "[::1]": false,
		"localhost": false, "cqlai.example.com": false,
		"http://127.0.0.1": true, "a b": true, "-bad": true, "127.0.0.1:7845": true,
	} {
		m := prefModel(t, &config.Config{})
		i := prefIndex(t, m, "MCP.Listen")
		m.setPrefField(i, value)
		_, got := m.preferenceErrors()[i]
		assert.Equal(t, wrong, got, "%q", value)
	}
}

// TestTheServingSettingsAreSaved under the keys cqlai mcp reads.
func TestTheServingSettingsAreSaved(t *testing.T) {
	m := prefModel(t, &config.Config{})
	m.setPrefField(prefIndex(t, m, "MCP.Listen"), "0.0.0.0")
	m.setPrefField(prefIndex(t, m, "MCP.TLSCert"), "/etc/cqlai/server.pem")
	m.setPrefField(prefIndex(t, m, "MCP.TLSKey"), "/etc/cqlai/server.key")
	m.setPrefField(prefIndex(t, m, "MCP.TLSClientCA"), "/etc/cqlai/ca.pem")
	m.preferences.focusField(prefIndex(t, m, "MCP.Token"))
	m.togglePrefField()

	block := savedMCP(t, m)
	require.NotNil(t, block)
	assert.Equal(t, "0.0.0.0", block["listen"])
	assert.Equal(t, true, block["token"])
	assert.Equal(t, "/etc/cqlai/server.pem", block["tlsCert"])
	assert.Equal(t, "/etc/cqlai/server.key", block["tlsKey"])
	assert.Equal(t, "/etc/cqlai/ca.pem", block["tlsClientCA"])
}

// TestTheTokenIsNotWrittenWhenLeftOff: off is the default, so an untouched
// window does not write it.
func TestTheTokenIsNotWrittenWhenLeftOff(t *testing.T) {
	m := prefModel(t, &config.Config{})
	assert.False(t, m.preferences.fields[prefIndex(t, m, "MCP.Token")].yes)
	assert.Nil(t, savedMCP(t, m))
}
