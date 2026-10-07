package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/policy"
)

// The MCP tools about the cluster rather than its data: how the nodes are
// doing, how big a table is, and who may do what.

// Tool names.
const (
	ToolNodeStatus ToolName = "node_status"
	ToolTableSize  ToolName = "table_size"
	ToolListRoles  ToolName = "list_roles"
)

// nodeSections are what node_status can show, and the virtual table each one
// reads. The tool description lists them, so a model knows what each is for.
var nodeSections = []struct {
	name, table, about string
	since              int // the Cassandra major version that has it
}{
	{"settings", "settings", "the node's configuration, as cassandra.yaml and the startup flags set it", 4},
	{"thread_pools", "thread_pools", "each thread pool: active, pending, blocked and completed tasks", 4},
	{"clients", "clients", "the connected clients: address, user, driver, protocol, requests", 4},
	{"caches", "caches", "the key, row, counter and chunk caches: size, entries, hit ratio", 4},
	{"compaction", "sstable_tasks", "the compactions and other SSTable tasks running now", 4},
	{"disk_usage", "disk_usage", "the space each table takes on disk", 5},
}

func clusterToolDefinitions() []ToolDefinition {
	var sections, about []string
	for _, s := range nodeSections {
		sections = append(sections, s.name)
		about = append(about, s.name+": "+s.about)
	}
	return []ToolDefinition{
		{
			Name: ToolNodeStatus.String(),
			Description: "How the node cqlai is connected to is doing, from its virtual tables (Cassandra 4.0 and later). Sections: " +
				strings.Join(about, "; ") + ". filter keeps the rows whose name contains it.",
			Parameters: map[string]any{
				"section": map[string]any{"type": "string", "enum": sections},
				"filter":  map[string]any{"type": "string", "description": "Only rows whose name (the first column) contains this, ignoring case"},
			},
			Required: []string{"section"},
			MCP:      true, Needs: NeedSelect, ReadOnly: true,
		},
		{
			Name: ToolTableSize.String(),
			Description: "Cassandra's estimate of how many partitions a table has and how big they are on average, " +
				"from system.size_estimates. For judging whether a partition key makes partitions too wide.",
			Parameters: map[string]any{
				"keyspace": map[string]any{"type": "string"},
				"table":    map[string]any{"type": "string"},
			},
			Required: []string{"keyspace", "table"},
			MCP:      true, Needs: NeedSelect, ReadOnly: true,
		},
		{
			Name:        ToolListRoles.String(),
			Description: "The roles, and with role given, every permission that role has. Never their passwords.",
			Parameters: map[string]any{
				"role": map[string]any{"type": "string", "description": "A role to list the permissions of"},
			},
			MCP: true, Needs: NeedList, ReadOnly: true,
		},
	}
}

// NodeStatusParams is which section of node_status, and what to keep.
type NodeStatusParams struct {
	Section string `json:"section"`
	Filter  string `json:"filter,omitempty"`
}

func (p NodeStatusParams) Validate() error {
	for _, s := range nodeSections {
		if s.name == p.Section {
			return nil
		}
	}
	var names []string
	for _, s := range nodeSections {
		names = append(names, s.name)
	}
	return fmt.Errorf("section has to be one of %s", strings.Join(names, ", "))
}

// TableSizeParams is the table to estimate.
type TableSizeParams struct {
	Keyspace string `json:"keyspace"`
	Table    string `json:"table"`
}

func (p TableSizeParams) Validate() error {
	if !cqlName.MatchString(p.Keyspace) || !cqlName.MatchString(p.Table) {
		return fmt.Errorf("keyspace and table are names: letters, digits and underscores")
	}
	return nil
}

// ListRolesParams is a role, or none.
type ListRolesParams struct {
	Role string `json:"role,omitempty"`
}

func (p ListRolesParams) Validate() error {
	if len(p.Role) > 256 {
		return fmt.Errorf("role is too long")
	}
	return nil
}

// nodeStatus reads one of the node's virtual tables, through the same gate
// as any SELECT a model writes, so system_views has to be visible to it.
func nodeStatus(ctx context.Context, env ToolEnv, p NodeStatusParams) ToolOutcome {
	var table string
	var since int
	for _, s := range nodeSections {
		if s.name == p.Section {
			table, since = s.table, s.since
		}
	}
	if !env.Session.IsVersion4OrHigher() {
		return errorOutcome("node_status reads virtual tables, which Cassandra has from 4.0; this cluster is " + env.Session.CassandraVersion())
	}

	statement := "SELECT * FROM system_views." + table
	if _, err := env.Policy.Check(statement); err != nil {
		return gateOutcome(env, statement, err)
	}

	page, err := env.Session.QueryPage(ctx, db.PageRequest{Statement: statement, PageSize: 5000})
	if err != nil {
		if since > 4 {
			return gateOutcome(env, statement, fmt.Errorf("%s needs Cassandra %d.0 or later: %w", p.Section, since, err))
		}
		return gateOutcome(env, statement, err)
	}

	// Kept by name, then shaped as any result is: redacted, cut, and limited
	// to the server's rows.
	if p.Filter != "" && len(page.Columns) > 0 {
		first := page.Columns[0].Name
		want := strings.ToLower(p.Filter)
		kept := page.Rows[:0]
		for _, row := range page.Rows {
			if strings.Contains(strings.ToLower(fmt.Sprint(row[first])), want) {
				kept = append(kept, row)
			}
		}
		page.Rows = kept
	}
	more := len(page.Rows) > env.Policy.MaxRows()
	if more {
		page.Rows = page.Rows[:env.Policy.MaxRows()]
	}
	rows, cut := shapeRows(env.Policy, "system_views", table, page)

	// A setting that holds a secret is hidden whether or not Cassandra hides
	// it.
	if table == "settings" {
		for _, row := range rows {
			if name, ok := row["name"].(string); ok && policy.MaskedSetting(name) {
				row["value"] = policy.RedactedValue
			}
		}
	}

	out := map[string]any{"section": p.Section, "columns": page.Columns, "rows": rows}
	if more {
		out["note"] = fmt.Sprintf("only the first %d rows: use filter to narrow it", env.Policy.MaxRows())
	}
	if len(cut) > 0 {
		out["truncated"] = cut
	}
	o := jsonOutcome(out)
	o.Statement, o.Rows = statement, len(rows)
	return o
}

// tableSize sums system.size_estimates over the token ranges of one table.
func tableSize(ctx context.Context, env ToolEnv, p TableSizeParams) ToolOutcome {
	if !env.Policy.VisibleTable(p.Keyspace, p.Table) {
		return refusedOutcome(fmt.Sprintf("%s.%s is not visible to this server", p.Keyspace, p.Table))
	}

	// A fixed statement with the names bound, not written in: it reads one
	// table's estimates and nothing else, so it does not need system to be
	// visible.
	iter := env.Session.Query(`SELECT partitions_count, mean_partition_size
	                           FROM system.size_estimates WHERE keyspace_name = ? AND table_name = ?`,
		p.Keyspace, p.Table).IterContext(ctx)
	var partitions, total, ranges int64
	var count, mean int64
	for iter.Scan(&count, &mean) {
		partitions += count
		total += count * mean
		ranges++
	}
	if err := iter.Close(); err != nil {
		return errorOutcome(env.Policy.Scrub(err.Error()))
	}

	out := map[string]any{"keyspace": p.Keyspace, "table": p.Table, "token_ranges": ranges}
	if ranges == 0 || partitions == 0 {
		out["note"] = "no estimate yet: Cassandra writes them every few minutes, and for tables with data on disk"
		return jsonOutcome(out)
	}
	out["estimated_partitions"] = partitions
	out["mean_partition_bytes"] = total / partitions
	out["note"] = "estimates for the token ranges this node holds as primary, refreshed every few minutes"
	return jsonOutcome(out)
}

// listRoles runs LIST ROLES, or LIST ALL PERMISSIONS OF a role, through the
// gate. Neither returns a password hash; reading system_auth would, and it
// is never visible.
func listRoles(ctx context.Context, env ToolEnv, p ListRolesParams) ToolOutcome {
	statement := "LIST ROLES"
	if p.Role != "" {
		statement = "LIST ALL PERMISSIONS OF '" + strings.ReplaceAll(p.Role, "'", "''") + "'"
	}
	if _, err := env.Policy.Check(statement); err != nil {
		return gateOutcome(env, statement, err)
	}
	page, err := env.Session.QueryPage(ctx, db.PageRequest{Statement: statement, PageSize: env.Policy.MaxRows()})
	if err != nil {
		return gateOutcome(env, statement, err)
	}
	rows, _ := shapeRows(env.Policy, "system_auth", "roles", page)
	o := jsonOutcome(map[string]any{"columns": page.Columns, "rows": rows})
	o.Statement, o.Rows = statement, len(rows)
	return o
}
