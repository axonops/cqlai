package policy

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func permitList(names ...string) *[]string { return &names }

func load(t *testing.T, file *config.Config, connection string, flags Flags) Policy {
	t.Helper()
	p, err := Load(file, connection, "", 10*time.Second, flags)
	require.NoError(t, err)
	return p
}

// TestTheZeroValueAllowsNothing: a code path that forgets the policy fails
// closed.
func TestTheZeroValueAllowsNothing(t *testing.T) {
	var p Policy
	assert.Empty(t, p.Permitted())
	assert.False(t, p.Visible("shop"))
	_, err := p.Check("SELECT * FROM shop.orders")
	assert.Error(t, err)
}

// TestNothingConfiguredIsTheReadCommands, with system_auth hidden.
func TestNothingConfiguredIsTheReadCommands(t *testing.T) {
	p := load(t, &config.Config{}, "", Flags{})

	assert.Equal(t, []string{"SELECT", "DESCRIBE", "LIST"}, p.Permitted())
	assert.True(t, p.Visible("shop"))
	assert.False(t, p.Visible("system_auth"))
	assert.False(t, p.AllowScans())
	assert.Equal(t, DefaultMaxRows, p.MaxRows())
}

// TestAnEmptyPermitIsNothing, which is not the same as leaving it out.
func TestAnEmptyPermitIsNothing(t *testing.T) {
	p := load(t, &config.Config{MCP: &config.MCPConfig{Permit: permitList()}}, "", Flags{})
	assert.Empty(t, p.Permitted())
}

// TestATypoStopsTheServer rather than permitting something else.
func TestATypoStopsTheServer(t *testing.T) {
	for _, permit := range []*[]string{
		permitList("SELECT", "INSRET"),
		permitList("GRANT"),
		permitList("USE"),
	} {
		_, err := Load(&config.Config{MCP: &config.MCPConfig{Permit: permit}}, "", "", 0, Flags{})
		assert.Error(t, err, "%v", *permit)
	}
	_, err := Load(&config.Config{}, "", "", 0, Flags{Permit: []string{"DROPP"}})
	assert.Error(t, err)

	_, err = Load(&config.Config{MCP: &config.MCPConfig{Redact: []string{"shop.email"}}}, "", "", 0, Flags{})
	assert.Error(t, err, "a redaction has three parts")
}

// TestTheStrictestOfThreeWins, part by part, each place narrowing the others.
func TestTheStrictestOfThreeWins(t *testing.T) {
	file := &config.Config{
		MCP: &config.MCPConfig{
			Permit:     permitList("SELECT", "DESCRIBE", "LIST"),
			Keyspaces:  []string{"shop", "catalog", "audit"},
			Deny:       []string{"shop.secrets"},
			Redact:     []string{"*.*.card_number"},
			AllowScans: true,
			MaxRows:    500,
		},
		Connections: []config.Config{
			{Name: "prod", MCP: &config.MCPConfig{
				Permit:    permitList("SELECT", "DESCRIBE"),
				Keyspaces: []string{"shop", "catalog"},
				Redact:    []string{"shop.customers.email"},
				MaxRows:   50,
			}},
			{Name: "local"},
		},
	}

	prod := load(t, file, "prod", Flags{})
	assert.Equal(t, []string{"SELECT", "DESCRIBE"}, prod.Permitted(), "LIST is not in prod")
	assert.True(t, prod.Visible("shop"))
	assert.False(t, prod.Visible("audit"), "prod narrowed the keyspaces")
	assert.False(t, prod.VisibleTable("shop", "secrets"), "the top's deny still holds")
	assert.True(t, prod.Redacted("shop", "customers", "email"))
	assert.True(t, prod.Redacted("shop", "payments", "card_number"), "the top's redaction still holds")
	assert.False(t, prod.AllowScans(), "prod's block did not turn scans on")
	assert.Equal(t, 50, prod.MaxRows())

	// A connection with no block is what the top says.
	local := load(t, file, "local", Flags{})
	assert.Equal(t, []string{"SELECT", "DESCRIBE", "LIST"}, local.Permitted())
	assert.True(t, local.AllowScans())
	assert.Equal(t, 500, local.MaxRows())

	// The flags narrow both, and never widen.
	flagged := load(t, file, "local", Flags{Permit: []string{"SELECT", "DESCRIBE"}, Keyspaces: []string{"shop", "elsewhere"}, MaxRows: 1000})
	assert.Equal(t, []string{"SELECT", "DESCRIBE"}, flagged.Permitted())
	assert.True(t, flagged.Visible("shop"))
	assert.False(t, flagged.Visible("elsewhere"), "a keyspace the file does not allow stays hidden")
	assert.Equal(t, 500, flagged.MaxRows(), "a higher limit on the command line does not raise it")

	file.Connections[1].MCP = &config.MCPConfig{Permit: permitList("SELECT")}
	flagged = load(t, file, "local", Flags{Permit: []string{"SELECT", "DESCRIBE"}})
	assert.Equal(t, []string{"SELECT"}, flagged.Permitted(), "--permit cannot add what the file does not permit")
}

// TestAConnectionCannotWidenTheTop.
func TestAConnectionCannotWidenTheTop(t *testing.T) {
	file := &config.Config{
		Connections: []config.Config{{Name: "wide", MCP: &config.MCPConfig{
			Permit:     permitList("SELECT", "DESCRIBE"),
			AllowScans: true,
		}}},
	}
	p := load(t, file, "wide", Flags{})
	assert.Equal(t, []string{"SELECT", "DESCRIBE"}, p.Permitted())
	assert.False(t, p.AllowScans(), "the top did not turn scans on")
}

func gate(t *testing.T, permit ...string) Policy {
	t.Helper()
	return load(t, &config.Config{MCP: &config.MCPConfig{
		Permit:    &permit,
		Keyspaces: []string{"shop", "Shop2", "system_auth", "system_views"},
		Deny:      []string{"shop.secrets"},
	}}, "", Flags{}).WithPartitionKey(func(keyspace, table string) []string {
		if keyspace == "shop" && table == "orders" {
			return []string{"customer", "day"}
		}
		return nil
	})
}

// TestTheGateRefusesWhatIsNotPermitted.
func TestTheGateRefusesWhatIsNotPermitted(t *testing.T) {
	p := gate(t, "SELECT")

	for _, cql := range []string{
		"SELECT * FROM shop.orders WHERE customer = 'a'",
	} {
		_, err := p.Check(cql)
		assert.NoError(t, err, cql)
	}

	for cql, why := range map[string]string{
		"UPDATE shop.orders SET x = 1 WHERE customer = 'a'":                        "UPDATE is not permitted",
		"DROP TABLE shop.orders":                                                   "DROP is not permitted",
		"BEGIN BATCH INSERT INTO shop.orders (customer) VALUES ('a'); APPLY BATCH": "BATCH is not permitted",
		"SELECT * FROM orders":                                                     "name the keyspace",
		"SELECT * FROM billing.invoices":                                           "keyspace billing is not visible",
		"SELECT * FROM shop.secrets":                                               "shop.secrets is not visible",
		"SELECT * FROM SHOP.Secrets":                                               "shop.secrets is not visible",
		"SELECT * FROM system_auth.roles":                                          "system_auth is not visible",
		`SELECT * FROM "system_auth".roles`:                                        "system_auth is not visible",
		`SELECT * FROM "SYSTEM_AUTH".roles`:                                        "not visible",
		"SELECT * FROM shop2.orders":                                               "keyspace shop2 is not visible",
		"GRANT ALL ON ALL KEYSPACES TO bob":                                        "GRANT is never permitted",
		"SELECT * FROM shop.orders; DROP TABLE shop.orders":                        "one statement at a time",
	} {
		_, err := p.Check(cql)
		require.Error(t, err, cql)
		assert.Contains(t, err.Error(), why, cql)
		var refusal Refusal
		assert.True(t, errors.As(err, &refusal), "a refusal is a Refusal: %s", cql)
	}

	// A quoted keyspace with capitals is a different keyspace from the
	// unquoted one, and is visible only as it is listed.
	_, err := p.Check(`SELECT * FROM "Shop2".orders`)
	assert.NoError(t, err)
}

// TestAChangeCanOnlyBeProposed: no setting permits one to run, and a
// proposal still cannot reach what the policy hides.
func TestAChangeCanOnlyBeProposed(t *testing.T) {
	for _, command := range []string{"INSERT", "UPDATE", "DELETE", "BATCH", "CREATE", "ALTER", "DROP", "TRUNCATE"} {
		_, err := Load(&config.Config{MCP: &config.MCPConfig{Permit: permitList("SELECT", command)}}, "", "", 0, Flags{})
		require.Error(t, err, command)
		assert.Contains(t, err.Error(), "never runs", command)
	}

	p := gate(t, "SELECT")
	_, err := p.Check("INSERT INTO shop.orders (customer) VALUES ('a')")
	assert.Error(t, err, "and the gate refuses to run one")

	for _, cql := range []string{
		"INSERT INTO shop.orders (customer) VALUES ('a')",
		"DROP TABLE shop.orders",
		"ALTER TABLE shop.orders ADD note text",
		"BEGIN BATCH INSERT INTO shop.orders (customer) VALUES ('a'); DELETE FROM shop.orders WHERE customer = 'b'; APPLY BATCH",
	} {
		_, err := p.CheckProposal(cql)
		assert.NoError(t, err, cql)
	}
	for cql, why := range map[string]string{
		"SELECT * FROM shop.orders":         "does not change anything",
		"DROP TABLE shop.secrets":           "not visible",
		"DROP TABLE orders":                 "name the keyspace",
		"DROP KEYSPACE billing":             "not visible",
		"GRANT ALL ON ALL KEYSPACES TO bob": "never permitted",
		"BEGIN BATCH INSERT INTO shop.secrets (id) VALUES (1); APPLY BATCH": "not visible",
		"DROP TABLE shop.orders; DROP TABLE shop.customers":                 "one statement",
	} {
		_, err := p.CheckProposal(cql)
		require.Error(t, err, cql)
		assert.Contains(t, err.Error(), why, cql)
	}
}

// TestScansAreRefusedUnlessAllowed.
func TestScansAreRefusedUnlessAllowed(t *testing.T) {
	p := gate(t, "SELECT")

	for cql, why := range map[string]string{
		"SELECT * FROM shop.orders WHERE total > 1 ALLOW FILTERING": "ALLOW FILTERING",
		"SELECT COUNT(*) FROM shop.orders":                          "whole partition key (customer, day)",
		"SELECT COUNT(*) FROM shop.orders WHERE customer = 'a'":     "whole partition key",
		"SELECT COUNT(*) FROM shop.items WHERE id = 1":              "could not be found",
	} {
		_, err := p.Check(cql)
		require.Error(t, err, cql)
		assert.Contains(t, err.Error(), why, cql)
	}

	for _, cql := range []string{
		"SELECT * FROM shop.orders", // paged, and stops at max_rows
		"SELECT * FROM shop.orders LIMIT 10",
		"SELECT COUNT(*) FROM shop.orders WHERE customer = 'a' AND day IN ('x', 'y')",
		"SELECT * FROM system_views.clients WHERE address = '1' ALLOW FILTERING", // virtual tables are small
	} {
		_, err := p.Check(cql)
		assert.NoError(t, err, cql)
	}

	allowed := load(t, &config.Config{MCP: &config.MCPConfig{AllowScans: true}}, "", Flags{})
	_, err := allowed.Check("SELECT COUNT(*) FROM shop.orders")
	assert.NoError(t, err)
}

// TestSecretsDoNotGoBack.
func TestSecretsDoNotGoBack(t *testing.T) {
	p, err := Load(&config.Config{}, "", "hunter2", 0, Flags{})
	require.NoError(t, err)
	assert.Equal(t, "auth failed for password "+RedactedValue, p.Scrub("auth failed for password hunter2"))

	assert.True(t, MaskedSetting("server_encryption_options_keystore_password"))
	assert.True(t, MaskedSetting("SomeSecretThing"))
	assert.False(t, MaskedSetting("concurrent_reads"))
}

// TestTheLimiterBoundsTheLoad: one call at a time, and so many a minute.
func TestTheLimiterBoundsTheLoad(t *testing.T) {
	l := NewLimiter(2)
	now := time.Unix(1000, 0)
	l.now = func() time.Time { return now }

	release, err := l.Acquire(context.Background())
	require.NoError(t, err)

	// The second waits for the first.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = l.Acquire(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded, "one call at a time")
	release()

	// That wait counted as a call, so the minute's two are used.
	_, err = l.Acquire(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "more than 2 calls a minute")

	now = now.Add(61 * time.Second)
	release, err = l.Acquire(context.Background())
	require.NoError(t, err, "a minute later there is room again")
	release()
}

// TestTheAuditLogHoldsNoValues, and is readable only by its owner.
func TestTheAuditLogHoldsNoValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	a, err := OpenAudit(path)
	require.NoError(t, err)

	a.Log(Entry{Tool: "query", Statement: "SELECT * FROM shop.customers WHERE email = 'ann@example.com' AND age = 42", Decision: "allowed", Rows: 1})
	a.Log(Entry{Tool: "query", Statement: "DROP TABLE shop.customers", Decision: "refused", Reason: "DROP is not permitted"})
	// A uuid, a boolean, and a driver error that repeats the value.
	a.Log(Entry{Tool: "query", Statement: "SELECT * FROM shop.customers WHERE id = 123e4567-e89b-12d3-a456-426614174000 AND vip = true",
		Decision: "failed", Reason: "Invalid UUID constant (123e4567-e89b-12d3-a456-426614174000)"})
	a.Log(Entry{Tool: "query", Statement: "SELECT * FROM shop.customers WHERE age = 'hunter2'",
		Decision: "failed", Reason: "Invalid STRING constant (hunter2) for \"age\" of type int"})
	require.NoError(t, a.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	for _, value := range []string{"ann@example.com", "42", "e89b", "a456", "true", "hunter2"} {
		assert.NotContains(t, string(data), value)
	}
	assert.Contains(t, string(data), `for \"age\" of type int`, "the rest of the error stays")

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	require.Len(t, lines, 4)
	var first Entry
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &first))
	assert.Equal(t, "select * from shop . customers where email = ? and age = ?", first.Statement)
	assert.Contains(t, lines[1], `"decision":"refused"`)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	off, err := OpenAudit(AuditOff)
	require.NoError(t, err)
	off.Log(Entry{Tool: "query"}) // writes nothing, and does not fail
}

// TestRedactedValuesCannotComeBackAnotherWay: under another name, as JSON,
// through a function, or by testing guesses in WHERE.
func TestRedactedValuesCannotComeBackAnotherWay(t *testing.T) {
	p := load(t, &config.Config{MCP: &config.MCPConfig{
		Redact: []string{"shop.customers.email"},
	}}, "", Flags{}).WithColumns(func(keyspace, table string) []string {
		if table == "customers" {
			return []string{"id", "name", "email"}
		}
		return []string{"id", "total"}
	})

	for _, cql := range []string{
		"SELECT * FROM shop.customers WHERE id = 1",
		"SELECT id, name, email FROM shop.customers",
		"SELECT DISTINCT id FROM shop.customers",
		"SELECT JSON * FROM shop.orders", // no hidden columns there
		"SELECT total AS t FROM shop.orders",
	} {
		_, err := p.Check(cql)
		assert.NoError(t, err, cql)
	}

	for cql, why := range map[string]string{
		"SELECT JSON * FROM shop.customers":                             "hidden columns",
		"SELECT email AS e FROM shop.customers":                         "hidden columns",
		"SELECT toJson(email) FROM shop.customers":                      "hidden columns",
		"SELECT writetime(email) FROM shop.customers":                   "hidden columns",
		"SELECT id FROM shop.customers WHERE email = 'a@b.c'":           "email is hidden",
		`SELECT id FROM shop.customers WHERE "email" > 'a'`:             "email is hidden",
		"SELECT id FROM shop.customers WHERE id = 1 AND EMAIL IN ('x')": "is hidden",
	} {
		_, err := p.Check(cql)
		require.Error(t, err, cql)
		assert.Contains(t, err.Error(), why, cql)
	}

	// When the table's columns cannot be found, a pattern that could apply
	// is taken to.
	unknown := load(t, &config.Config{MCP: &config.MCPConfig{Redact: []string{"*.*.card_number"}}}, "", Flags{})
	_, err := unknown.Check("SELECT JSON * FROM shop.anything")
	assert.Error(t, err)
}

// TestADenyEntryMatchesWhateverItsCase: hiding leans towards hiding, so a
// denied name written in other capitals still hides the table.
func TestADenyEntryMatchesWhateverItsCase(t *testing.T) {
	p := load(t, &config.Config{MCP: &config.MCPConfig{Deny: []string{"SHOP.Secrets", "Billing"}}}, "", Flags{})
	assert.False(t, p.VisibleTable("shop", "secrets"))
	assert.False(t, p.Visible("billing"))
	_, err := p.Check("SELECT * FROM shop.secrets")
	assert.Error(t, err)
}

// denyOnly allows every keyspace but billing and shop.secrets: the settings
// with a deny list and no keyspace list.
func denyOnly(t *testing.T) Policy {
	t.Helper()
	permit := []string{"SELECT", "DESCRIBE"}
	return load(t, &config.Config{MCP: &config.MCPConfig{
		Permit: &permit,
		Deny:   []string{"billing", "shop.secrets"},
	}}, "", Flags{})
}

// TestTheSystemKeyspacesDoNotShowWhatDenyHides: with every keyspace allowed,
// the system ones are not. They list the hidden keyspaces, tables and
// columns, size them, and hold the statements traced against them.
func TestTheSystemKeyspacesDoNotShowWhatDenyHides(t *testing.T) {
	p := denyOnly(t)
	require.True(t, p.AllKeyspacesVisible())

	for _, q := range []string{
		"SELECT keyspace_name, table_name FROM system_schema.tables",
		"SELECT * FROM system_schema.columns WHERE keyspace_name = 'billing'",
		"SELECT * FROM system_schema.keyspaces",
		"SELECT * FROM system_traces.sessions",
		"SELECT * FROM system_traces.events",
		"SELECT * FROM system.size_estimates WHERE keyspace_name = 'billing' AND table_name = 'invoices'",
		"SELECT * FROM system_views.settings",
		"SELECT * FROM system_distributed.repair_history",
		"SELECT * FROM System_Schema.tables",
	} {
		_, err := p.Check(q)
		assert.Error(t, err, q)
	}

	// What deny does not hide is still there.
	_, err := p.Check("SELECT * FROM shop.orders WHERE customer = 1")
	assert.NoError(t, err)
	assert.True(t, p.Visible("shop"))
}

// TestAKeyspaceListCanStillNameASystemKeyspace: listing one is asking for it.
func TestAKeyspaceListCanStillNameASystemKeyspace(t *testing.T) {
	permit := []string{"SELECT"}
	p := load(t, &config.Config{MCP: &config.MCPConfig{
		Permit:    &permit,
		Keyspaces: []string{"shop", "system_schema"},
	}}, "", Flags{})

	_, err := p.Check("SELECT keyspace_name FROM system_schema.keyspaces")
	assert.NoError(t, err)
	assert.False(t, p.Visible("system_traces"), "only the one named")
}

// TestTheChatViewStillSeesTheSystemKeyspaces: it shows what the shell shows.
func TestTheChatViewStillSeesTheSystemKeyspaces(t *testing.T) {
	assert.True(t, Shell().Visible("system_schema"))
	assert.True(t, Shell().Visible("system_traces"))
	assert.False(t, Shell().Visible("system_auth"), "the password hashes, never")
}

// TestNodeStatusReadsTheVirtualTablesUnlessDenied: its statement is its own
// and it hides the secrets, so it reads system_views when every keyspace is
// allowed. A deny entry or a keyspace list without it still refuses it.
func TestNodeStatusReadsTheVirtualTablesUnlessDenied(t *testing.T) {
	_, err := denyOnly(t).WithSystemKeyspaces().Check("SELECT * FROM system_views.thread_pools")
	assert.NoError(t, err)

	permit := []string{"SELECT"}
	denied := load(t, &config.Config{MCP: &config.MCPConfig{Permit: &permit, Deny: []string{"system_views"}}}, "", Flags{})
	_, err = denied.WithSystemKeyspaces().Check("SELECT * FROM system_views.thread_pools")
	assert.Error(t, err)

	listed := load(t, &config.Config{MCP: &config.MCPConfig{Permit: &permit, Keyspaces: []string{"shop"}}}, "", Flags{})
	_, err = listed.WithSystemKeyspaces().Check("SELECT * FROM system_views.thread_pools")
	assert.Error(t, err)

	assert.False(t, IsSystemKeyspace("systems"), "a keyspace merely starting with the word is not one")
	assert.True(t, IsSystemKeyspace("system"))
}

// TestTheAuditLogIsKeptInCassandra, and one an older cqlai wrote in the home
// directory is moved there with what it holds.
func TestTheAuditLogIsKeptInCassandra(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	old := filepath.Join(home, ".cqlai_mcp_audit.log")
	require.NoError(t, os.WriteFile(old, []byte("{\"tool\":\"query\"}\n"), 0o600))

	path := filepath.Join(home, ".cassandra", "cqlai_mcp_audit.log")
	require.Equal(t, path, DefaultAuditLog())

	a, err := OpenAudit(DefaultAuditLog())
	require.NoError(t, err)
	a.Log(Entry{Tool: "describe", Decision: "allowed"})
	require.NoError(t, a.Close())

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), `"tool":"query"`, "the old entries came with it")
	assert.Contains(t, string(got), `"tool":"describe"`)
	assert.NoFileExists(t, old)
}

// TestAViewIsHiddenAsItsBaseTableIs: a materialized view holds its base
// table's rows, so a deny entry or a redaction for the table holds for every
// view of it.
func TestAViewIsHiddenAsItsBaseTableIs(t *testing.T) {
	views := func(keyspace, name string) string {
		switch {
		case keyspace == "shop" && name == "secrets_by_owner":
			return "secrets"
		case keyspace == "shop" && name == "customers_by_email":
			return "customers"
		}
		return ""
	}
	permit := []string{"SELECT"}
	p := load(t, &config.Config{MCP: &config.MCPConfig{
		Permit: &permit,
		Deny:   []string{"shop.secrets"},
		Redact: []string{"shop.customers.email"},
	}}, "", Flags{}).WithViewBase(views)

	assert.False(t, p.VisibleTable("shop", "secrets_by_owner"))
	_, err := p.Check("SELECT * FROM shop.secrets_by_owner WHERE owner = 'a'")
	assert.Error(t, err)

	assert.True(t, p.Redacted("shop", "customers_by_email", "email"))
	assert.False(t, p.Redacted("shop", "customers_by_email", "name"))
	_, err = p.Check("SELECT JSON * FROM shop.customers_by_email WHERE email = 'a@b.c'")
	assert.Error(t, err)

	// A table that is not a view is as it was.
	assert.True(t, p.VisibleTable("shop", "orders"))
	assert.False(t, p.Redacted("shop", "orders", "email"))
}

// TestASecretSettingDoesNotComeBack: system_views.settings holds the node's
// configuration, passwords among it. The value is hidden in a row for one,
// and cannot come back as JSON, under an alias, or as the answer to a guess.
func TestASecretSettingDoesNotComeBack(t *testing.T) {
	secret := map[string]any{"name": "client_encryption_options_keystore_password", "value": "hunter2"}
	plain := map[string]any{"name": "cluster_name", "value": "Test Cluster"}
	assert.True(t, SettingHidden("system_views", "settings", "value", secret))
	assert.False(t, SettingHidden("system_views", "settings", "name", secret))
	assert.False(t, SettingHidden("system_views", "settings", "value", plain))
	assert.True(t, SettingHidden("system_views", "settings", "value", map[string]any{"value": "x"}),
		"without the name it cannot be told")
	assert.False(t, SettingHidden("shop", "settings", "value", secret), "only the node's settings")

	p := load(t, &config.Config{}, "", Flags{}).WithSystemKeyspaces()
	for _, cql := range []string{
		"SELECT * FROM system_views.settings",
		"SELECT name, value FROM system_views.settings WHERE name = 'cluster_name'",
	} {
		_, err := p.Check(cql)
		assert.NoError(t, err, cql)
	}
	for _, cql := range []string{
		"SELECT JSON * FROM system_views.settings",
		"SELECT value AS v FROM system_views.settings",
		"SELECT name FROM system_views.settings WHERE value = 'hunter2' ALLOW FILTERING",
	} {
		_, err := p.Check(cql)
		assert.Error(t, err, cql)
	}
}
