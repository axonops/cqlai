package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// systemTables is enough of the system tables' keys and columns to tell them
// apart: rows about keyspaces, rows that only mention one, and neither.
func systemTables(p Policy) Policy {
	keys := map[string][]string{
		"system_schema.tables":     {"keyspace_name"},
		"system_schema.columns":    {"keyspace_name"},
		"system.size_estimates":    {"keyspace_name"},
		"system_views.clients":     {"address"},
		"system.local":             {"key"},
		"system_views.thread_pool": {"name"},
	}
	columns := map[string][]string{
		"system_schema.tables":     {"keyspace_name", "table_name", "comment"},
		"system_schema.columns":    {"keyspace_name", "table_name", "column_name", "type"},
		"system.size_estimates":    {"keyspace_name", "table_name", "range_start", "partitions_count"},
		"system_views.clients":     {"address", "port", "keyspace_name", "username"},
		"system.local":             {"key", "cluster_name"},
		"system_views.thread_pool": {"name", "active_tasks"},
	}
	return p.
		WithPartitionKey(func(ks, t string) []string { return keys[ks+"."+t] }).
		WithColumns(func(ks, t string) []string { return columns[ks+"."+t] })
}

func listingSystemSchema(t *testing.T) Policy {
	t.Helper()
	permit := []string{"SELECT"}
	return systemTables(load(t, &config.Config{MCP: &config.MCPConfig{
		Permit:    &permit,
		Keyspaces: []string{"shop", "system_schema", "system", "system_views"},
		Deny:      []string{"shop.secrets"},
	}}, "", Flags{}))
}

// TestRowsAboutHiddenKeyspacesAreLeftOut, and the rest are kept.
func TestRowsAboutHiddenKeyspacesAreLeftOut(t *testing.T) {
	p := listingSystemSchema(t)

	for _, row := range []map[string]any{
		{"keyspace_name": "billing", "table_name": "invoices"},
		{"keyspace_name": "shop", "table_name": "secrets"},
		{"keyspace_name": "", "table_name": "x"},
	} {
		assert.False(t, p.RowVisible("system_schema", "tables", row), "%v", row)
	}
	assert.True(t, p.RowVisible("system_schema", "tables", map[string]any{"keyspace_name": "shop", "table_name": "orders"}))
	assert.True(t, p.RowVisible("system_schema", "keyspaces", map[string]any{"keyspace_name": "system_schema"}))
	assert.True(t, p.RowVisible("shop", "orders", map[string]any{"keyspace_name": "billing"}), "a user table's rows are its own")

	// system_auth's schema is the same in every cluster: describing it hides
	// nothing. Its data is another matter, and stays out of reach.
	assert.True(t, p.RowVisible("system_schema", "tables", map[string]any{"keyspace_name": "system_auth", "table_name": "roles"}))
	_, err := p.Check("SELECT * FROM system_auth.roles")
	assert.Error(t, err)
}

// TestNothingHiddenNothingFiltered: with every keyspace and the system ones
// visible, and nothing denied, a system table is read as it is, with any
// selection.
func TestNothingHiddenNothingFiltered(t *testing.T) {
	permit := []string{"SELECT"}
	p := systemTables(load(t, &config.Config{MCP: &config.MCPConfig{Permit: &permit, SystemKeyspaces: true, AllowScans: true}}, "", Flags{}))

	assert.True(t, p.Visible("system_schema"))
	assert.False(t, p.Visible("system_auth"), "its data, never")
	_, err := p.Check("SELECT table_name, column_name FROM system_schema.columns")
	assert.NoError(t, err)
	_, err = p.Check("SELECT COUNT(*) FROM system_schema.columns")
	assert.NoError(t, err)
	assert.Empty(t, p.MissingNamingColumns("system_schema", "columns", []string{"column_name"}))
	assert.True(t, p.RowVisible("system_schema", "columns", map[string]any{"table_name": "x"}))

	// A deny list brings the filter back.
	denying := systemTables(load(t, &config.Config{MCP: &config.MCPConfig{Permit: &permit, SystemKeyspaces: true, Deny: []string{"billing"}}}, "", Flags{}))
	_, err = denying.Check("SELECT table_name AS keyspace_name FROM system_schema.columns")
	assert.Error(t, err)
	assert.False(t, denying.RowVisible("system_schema", "tables", map[string]any{"keyspace_name": "billing", "table_name": "invoices"}))
}

// TestSystemKeyspacesAddsThemToEveryKeyspace, and only when the top of the
// file turns it on.
func TestSystemKeyspacesAddsThemToEveryKeyspace(t *testing.T) {
	assert.False(t, load(t, &config.Config{}, "", Flags{}).Visible("system_schema"))
	on := load(t, &config.Config{MCP: &config.MCPConfig{SystemKeyspaces: true}}, "", Flags{})
	assert.True(t, on.Visible("system_schema"))
	assert.True(t, on.Visible("system_traces"))
	assert.True(t, on.Visible("shop"))
	assert.False(t, on.Visible("system_auth"))

	listed := load(t, &config.Config{MCP: &config.MCPConfig{SystemKeyspaces: true, Keyspaces: []string{"shop"}}}, "", Flags{})
	assert.False(t, listed.Visible("system_schema"), "a keyspace list is the list")
}

// TestTheNamingColumnsHaveToComeBack, under their own names.
func TestTheNamingColumnsHaveToComeBack(t *testing.T) {
	p := listingSystemSchema(t)

	assert.Equal(t, []string{"keyspace_name", "table_name"},
		p.MissingNamingColumns("system_schema", "columns", []string{"column_name", "type"}))
	assert.Empty(t, p.MissingNamingColumns("system_schema", "columns", []string{"keyspace_name", "table_name", "column_name"}))
	assert.Empty(t, p.MissingNamingColumns("system_views", "clients", []string{"address"}), "clients' rows are not about keyspaces")
	assert.Empty(t, p.MissingNamingColumns("system", "local", []string{"cluster_name"}))

	_, err := p.Check("SELECT table_name AS keyspace_name FROM system_schema.columns")
	assert.Error(t, err, "an alias could give another column the name")
	_, err = p.Check("SELECT keyspace_name, table_name, column_name FROM system_schema.columns")
	assert.NoError(t, err)
}

// TestASystemTableThatCannotBeLookedUpIsTakenToBeAboutKeyspaces.
func TestASystemTableThatCannotBeLookedUpIsTakenToBeAboutKeyspaces(t *testing.T) {
	p := listingSystemSchema(t)
	assert.Equal(t, []string{"keyspace_name"}, p.MissingNamingColumns("system", "unknown_table", []string{"x"}))
}

// TestAHiddenKeyspaceNameIsBlankedWhereTheRowStays: clients says which
// keyspace each client is using.
func TestAHiddenKeyspaceNameIsBlankedWhereTheRowStays(t *testing.T) {
	p := listingSystemSchema(t)

	hidden := map[string]any{"address": "10.0.0.1", "keyspace_name": "billing"}
	assert.True(t, p.RowVisible("system_views", "clients", hidden))
	assert.True(t, p.NameHidden("system_views", "clients", "keyspace_name", hidden))
	assert.False(t, p.NameHidden("system_views", "clients", "address", hidden))

	shown := map[string]any{"address": "10.0.0.1", "keyspace_name": "shop"}
	assert.False(t, p.NameHidden("system_views", "clients", "keyspace_name", shown))
	none := map[string]any{"address": "10.0.0.1", "keyspace_name": nil}
	assert.False(t, p.NameHidden("system_views", "clients", "keyspace_name", none))
	assert.False(t, p.NameHidden("system_schema", "tables", "keyspace_name", hidden), "there the row is dropped instead")
}

// TestAutoFetchIsOnlyOnWhenTheTopTurnsItOn, like the other settings that
// loosen.
func TestAutoFetchIsOnlyOnWhenTheTopTurnsItOn(t *testing.T) {
	assert.False(t, load(t, &config.Config{}, "", Flags{}).AutoFetch())
	assert.True(t, load(t, &config.Config{MCP: &config.MCPConfig{AutoFetch: true}}, "", Flags{}).AutoFetch())

	own := load(t, &config.Config{
		MCP:         &config.MCPConfig{AutoFetch: true},
		Connections: []config.Config{{Name: "prod", MCP: &config.MCPConfig{}}},
	}, "prod", Flags{})
	assert.False(t, own.AutoFetch(), "a connection's block that leaves it off")

	require.False(t, load(t, &config.Config{Connections: []config.Config{{Name: "prod", MCP: &config.MCPConfig{AutoFetch: true}}}}, "prod", Flags{}).AutoFetch(),
		"a connection cannot turn on what the top does not")
}

// TestAPermissionOnAHiddenThingIsNotShown: LIST PERMISSIONS names the
// resource each permission is on, so a row about a hidden keyspace or table
// would name it.
func TestAPermissionOnAHiddenThingIsNotShown(t *testing.T) {
	p := denyOnly(t)
	for resource, visible := range map[string]bool{
		"<all keyspaces>":            true,
		"<keyspace shop>":            true,
		"<table shop.orders>":        true,
		"<all tables in shop>":       true,
		"<function shop.total(int)>": true,
		"<role analyst>":             true,
		"<all roles>":                true,
		"<keyspace billing>":         false,
		"<table shop.secrets>":       false,
		"<table billing.invoices>":   false,
		"<all tables in billing>":    false,
		"<all functions in billing>": false,
		"<function billing.f(int)>":  false,
		"<keyspace system_auth>":     false,
		`<table "shop"."secrets">`:   false,
		"<table shopnodot>":          false,
	} {
		assert.Equal(t, visible, p.ResourceVisible(resource), resource)
	}
}
