package ai

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/policy"
	"github.com/axonops/cqlai/internal/validation"
)

// The tools only MCP clients are offered, and the one way every MCP call runs.
//
// They are in this registry, beside the CHAT view's, so a tool is described
// and run one way whoever asks for it. The MCP server in internal/mcp speaks
// the protocol and nothing else; what a call may do is decided here and in
// internal/policy.

// mcpToolDefinitions is the tools only MCP clients are offered.
func mcpToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name: ToolConnectionInfo.String(),
			Description: "What this server is connected to - the cluster, its version and data centres - and what it may do there: " +
				"the CQL commands permitted, the keyspaces visible, the columns hidden and the limits. Call it first.",
			Parameters: map[string]any{},
			MCP:        true, Needs: NeedNothing, ReadOnly: true,
		},
		{
			Name: ToolDescribe.String(),
			Description: "The CQL definition of a keyspace, table, type, index, materialized view, function or aggregate, " +
				"or of the cluster or the whole schema, exactly as DESCRIBE prints it in cqlai.",
			Parameters: map[string]any{
				"kind": map[string]any{
					"type": "string",
					"enum": []string{"keyspace", "table", "type", "index", "materialized_view", "function", "aggregate", "cluster", "schema"},
				},
				"keyspace": map[string]any{"type": "string", "description": "The keyspace; needed for every kind but cluster and schema"},
				"name":     map[string]any{"type": "string", "description": "The object's name within the keyspace; needed for every kind but keyspace, cluster and schema"},
			},
			Required: []string{"kind"},
			MCP:      true, Needs: NeedDescribe, ReadOnly: true,
		},
		{
			Name: ToolQuery.String(),
			Description: "Run one CQL SELECT and return a page of its rows as JSON, with each column's type. " +
				"Name tables as keyspace.table. Use next_page_token for the page after. " +
				"Other statements are refused; ALLOW FILTERING and aggregates across partitions may be refused too.",
			Parameters: queryParameters(),
			Required:   []string{"cql"},
			MCP:        true, Needs: NeedSelect, ReadOnly: true,
		},
		{
			Name: ToolTraceQuery.String(),
			Description: "Run one CQL SELECT with tracing on, and return its first page of rows and the trace: " +
				"the coordinator, how long it took, and each step on each node. For finding out why a query is slow.",
			Parameters: queryParameters(),
			Required:   []string{"cql"},
			MCP:        true, Needs: NeedSelect, ReadOnly: true,
		},
	}
}

func queryParameters() map[string]any {
	return map[string]any{
		"cql":         map[string]any{"type": "string", "description": "One CQL SELECT, with tables named as keyspace.table"},
		"consistency": map[string]any{"type": "string", "enum": db.ConsistencyLevels(), "description": "For this query only"},
		"page_size":   map[string]any{"type": "integer", "minimum": 1, "description": "Rows per page, up to the server's limit"},
		"page_token":  map[string]any{"type": "string", "description": "next_page_token from the page before, with the same cql"},
	}
}

// MCPToolDefinitions is the tools an MCP client is offered under a policy. A
// tool the policy does not allow is left out, not just refused when called: a
// model cannot be talked into a tool it cannot see.
func MCPToolDefinitions(p policy.Policy) []ToolDefinition {
	var defs []ToolDefinition
	for _, def := range allToolDefinitions() {
		if def.MCP && offered(def, p) {
			defs = append(defs, def)
		}
	}
	return defs
}

func offered(def ToolDefinition, p policy.Policy) bool {
	switch def.Needs {
	case NeedNothing:
		return true
	case NeedDescribe:
		return p.Permits("DESCRIBE")
	case NeedSelect:
		return p.Permits("SELECT")
	case NeedList:
		return p.Permits("LIST")
	case NeedChange:
		return p.PermitsAnyOf(validation.KindWrite, validation.KindSchema)
	}
	return false
}

// ToolEnv is what an MCP call runs against.
type ToolEnv struct {
	Policy  policy.Policy
	Session *db.Session
	// Schema is the schema cache the schema tools answer from, on Session.
	Schema *AI

	// Describe runs a DESCRIBE through cqlai's own, so describe prints what
	// the shell prints. It is passed in because the router imports this
	// package.
	Describe func(statement string) interface{}

	// NotConnected says why there is no Session, when there is none. The
	// server starts without a cluster and connects when it can.
	NotConnected string
}

// ToolOutcome is what a call came to: what goes back to the client, and what
// goes in the audit log.
type ToolOutcome struct {
	Text       string // what the client reads
	Structured any    // the same, as JSON, when it has a shape
	IsError    bool

	Statement string // the CQL run, for the audit log
	Refused   bool
	Reason    string
	Rows      int
}

// CallMCPTool runs one call from an MCP client. Everything a client can
// reach goes through here: the tool has to be one the policy offers, its
// arguments have to parse and validate, and what it runs passes the statement
// gate.
func CallMCPTool(ctx context.Context, env ToolEnv, name string, args map[string]any) ToolOutcome {
	tool := ParseToolName(name)
	def, ok := toolDefinition(tool)
	if !ok || !def.MCP || !offered(def, env.Policy) {
		// The same answer for a tool that does not exist and one the policy
		// does not offer.
		return refusedOutcome(fmt.Sprintf("there is no tool called %s", name))
	}

	params, err := ParseToolParamsFromMap(tool, args)
	if err != nil {
		return errorOutcome(err.Error())
	}
	if err := params.Validate(); err != nil {
		return errorOutcome(fmt.Sprintf("invalid arguments for %s: %v", name, err))
	}

	switch tool {
	case ToolFuzzySearch, ToolGetSchema, ToolListKeyspaces, ToolListTables:
		if o, ok := needsCluster(env); !ok {
			return o
		}
		env.Schema.refreshIfChanged()
		result := executeCommandFor(env.Policy, env.Schema, tool, schemaToolArg(tool, params))
		if result.Error != nil {
			return errorOutcome(env.Policy.Scrub(result.Error.Error()))
		}
		return ToolOutcome{Text: result.Data}
	case ToolConnectionInfo:
		return connectionInfo(env)
	case ToolDescribe:
		if o, ok := needsCluster(env); !ok {
			return o
		}
		return describe(env, params.(DescribeParams))
	case ToolQuery:
		return runQuery(ctx, env, params.(QueryParams), false)
	case ToolTraceQuery:
		return runQuery(ctx, env, params.(QueryParams), true)
	}
	return refusedOutcome(fmt.Sprintf("there is no tool called %s", name))
}

// schemaToolArg is the argument the CHAT view's command processor takes for
// one of the schema tools.
func schemaToolArg(tool ToolName, params ToolParams) string {
	switch p := params.(type) {
	case FuzzySearchParams:
		return p.Query
	case GetSchemaParams:
		return p.Keyspace + "." + p.Table
	case ListTablesParams:
		return p.Keyspace
	}
	return ""
}

// needsCluster is the outcome of a call made with no cluster to run it on,
// and false; or true when there is one.
func needsCluster(env ToolEnv) (ToolOutcome, bool) {
	if env.Session != nil {
		return ToolOutcome{}, true
	}
	reason := env.NotConnected
	if reason == "" {
		reason = "not connected"
	}
	return errorOutcome(reason + "; the server tries again on the next call"), false
}

func refusedOutcome(reason string) ToolOutcome {
	return ToolOutcome{Text: "refused: " + reason, IsError: true, Refused: true, Reason: reason}
}

func errorOutcome(message string) ToolOutcome {
	return ToolOutcome{Text: message, IsError: true, Reason: message}
}

// gateOutcome is the outcome of a statement the gate refused, or that failed.
func gateOutcome(env ToolEnv, statement string, err error) ToolOutcome {
	var refusal policy.Refusal
	if errors.As(err, &refusal) {
		o := refusedOutcome(refusal.Reason)
		o.Statement = statement
		return o
	}
	o := errorOutcome(env.Policy.Scrub(err.Error()))
	o.Statement = statement
	return o
}

// connection_info.

func connectionInfo(env ToolEnv) ToolOutcome {
	p := env.Policy
	info := map[string]any{
		"connection": p.Connection(),
		"connected":  env.Session != nil,
		"policy": map[string]any{
			"permitted_commands":    p.Permitted(),
			"never_permitted":       validation.NeverPermitted,
			"all_keyspaces_visible": p.AllKeyspacesVisible(),
			"visible_keyspaces":     p.VisibleKeyspaces(),
			// Whether anything is hidden, and not what: naming a hidden table
			// would tell the model it exists.
			"some_tables_hidden":      len(p.Deny()) > 1,
			"redacted_columns":        p.RedactPatterns(),
			"scans_allowed":           p.AllowScans(),
			"changes_need_the_user":   p.ConfirmChanges(),
			"max_rows":                p.MaxRows(),
			"max_value_bytes":         p.MaxValueBytes(),
			"max_calls_per_minute":    p.CallsPerMinute(),
			"timeout_seconds":         int(p.Timeout().Seconds()),
			"names_must_be_qualified": true,
		},
	}

	if env.Session == nil {
		info["not_connected"] = env.NotConnected
	}
	if s := env.Session; s != nil && s.Session != nil {
		if cluster, err := s.DescribeClusterQuery(); err == nil {
			info["cluster_name"] = cluster.ClusterName
			info["cassandra_version"] = cluster.Version
		}
		dcs := map[string]int{}
		var dc string
		iter := s.Query("SELECT data_center FROM system.local").Iter()
		for iter.Scan(&dc) {
			dcs[dc]++
		}
		_ = iter.Close()
		iter = s.Query("SELECT data_center FROM system.peers").Iter()
		for iter.Scan(&dc) {
			dcs[dc]++
		}
		_ = iter.Close()
		info["data_centres"] = dcs
	}
	return jsonOutcome(info)
}

// describe.

// describeStatements is the DESCRIBE each kind is, as cqlai's own DESCRIBE
// takes it.
var describeStatements = map[string]string{
	"keyspace":          "DESCRIBE KEYSPACE %s",
	"table":             "DESCRIBE TABLE %s.%s",
	"type":              "DESCRIBE TYPE %s.%s",
	"index":             "DESCRIBE INDEX %s.%s",
	"materialized_view": "DESCRIBE MATERIALIZED VIEW %s.%s",
	"function":          "DESCRIBE FUNCTION %s.%s",
	"aggregate":         "DESCRIBE AGGREGATE %s.%s",
}

func describe(env ToolEnv, p DescribeParams) ToolOutcome {
	pol := env.Policy

	switch p.Kind {
	case "cluster":
		return describeText(env, "DESCRIBE CLUSTER")
	case "schema":
		// The visible keyspaces, each described: DESCRIBE SCHEMA would
		// describe them all.
		keyspaces, err := env.Session.DescribeKeyspacesQuery()
		if err != nil {
			return errorOutcome(pol.Scrub(err.Error()))
		}
		var parts []string
		for _, ks := range keyspaces {
			if ks.Virtual || strings.HasPrefix(ks.Name, "system") || !pol.Visible(ks.Name) {
				continue
			}
			o := describeText(env, fmt.Sprintf(describeStatements["keyspace"], ks.Name))
			if !o.IsError {
				parts = append(parts, o.Text)
			}
		}
		return ToolOutcome{Text: strings.Join(parts, "\n\n")}
	}

	// The same answer for a keyspace that does not exist and one that is
	// hidden, so a refusal says nothing a statement would not.
	if !pol.Visible(p.Keyspace) {
		return refusedOutcome(fmt.Sprintf("keyspace %s is not visible to this server", p.Keyspace))
	}
	if p.Kind == "table" || p.Kind == "materialized_view" {
		if !pol.VisibleTable(p.Keyspace, p.Name) {
			return refusedOutcome(fmt.Sprintf("%s.%s is not visible to this server", p.Keyspace, p.Name))
		}
	}

	statement := fmt.Sprintf(describeStatements[p.Kind], p.Keyspace, p.Name)
	if p.Kind == "keyspace" {
		statement = fmt.Sprintf(describeStatements[p.Kind], p.Keyspace)
	}
	return describeText(env, statement)
}

// describeText runs a DESCRIBE through cqlai's own, and leaves out every
// statement in what it prints that is about a hidden table.
func describeText(env ToolEnv, statement string) ToolOutcome {
	if env.Describe == nil {
		return errorOutcome("describe is not available")
	}
	result := env.Describe(statement)
	var text string
	switch r := result.(type) {
	case string:
		text = r
	case error:
		return errorOutcome(env.Policy.Scrub(r.Error()))
	case nil:
		text = ""
	default:
		text = fmt.Sprint(r)
	}
	return ToolOutcome{Text: withoutHidden(env.Policy, text)}
}

// withoutHidden drops each statement of a description that names a table the
// policy hides: a keyspace's description holds every table in it, and the
// indexes and views on them.
//
// A statement this cannot read is dropped too if its text mentions a hidden
// table, so nothing is kept for want of being understood.
func withoutHidden(p policy.Policy, text string) string {
	chunks := strings.Split(text, ";\n")
	kept := chunks[:0]
	for _, chunk := range chunks {
		if !mentionsHidden(p, chunk) {
			kept = append(kept, chunk)
		}
	}
	return strings.Join(kept, ";\n")
}

func mentionsHidden(p policy.Policy, chunk string) bool {
	if s, err := validation.Classify(chunk); err == nil {
		for _, name := range s.Names {
			if name.Table != "" && !p.VisibleTable(name.Keyspace, name.Table) {
				return true
			}
		}
		// An index statement names its table after ON, which Classify read.
		return false
	}
	lower := strings.ToLower(chunk)
	for _, d := range p.Deny() {
		if strings.Contains(d, ".") && strings.Contains(lower, strings.ToLower(d)) {
			return true
		}
	}
	return false
}

// query and trace_query.

// pageToken is where a query's next page starts, tied to the statement so it
// cannot be used with another.
type pageToken struct {
	State []byte `json:"s"`
	Hash  string `json:"h"`
}

func statementHash(cql string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(cql)))
	return base64.RawURLEncoding.EncodeToString(sum[:12])
}

func encodePageToken(state []byte, cql string) string {
	if len(state) == 0 {
		return ""
	}
	data, _ := json.Marshal(pageToken{State: state, Hash: statementHash(cql)})
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodePageToken(token, cql string) ([]byte, error) {
	if token == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(token)
	var t pageToken
	if err != nil || json.Unmarshal(data, &t) != nil {
		return nil, fmt.Errorf("page_token is not one this server gave")
	}
	if t.Hash != statementHash(cql) {
		return nil, fmt.Errorf("page_token belongs to a different statement")
	}
	return t.State, nil
}

func runQuery(ctx context.Context, env ToolEnv, p QueryParams, trace bool) ToolOutcome {
	pol := env.Policy
	statement, err := pol.Check(p.CQL)
	if err != nil {
		return gateOutcome(env, p.CQL, err)
	}
	if statement.Command != "SELECT" {
		return gateOutcome(env, p.CQL, policy.Refusal{Reason: fmt.Sprintf(
			"query runs SELECT only; %s is not one (use describe for the schema)", statement.Command)})
	}

	state, err := decodePageToken(p.PageToken, p.CQL)
	if err != nil {
		return gateOutcome(env, p.CQL, err)
	}
	// After the gate: a statement that would be refused says so whether or
	// not there is a cluster to run it on.
	if o, ok := needsCluster(env); !ok {
		o.Statement = p.CQL
		return o
	}

	size := pol.MaxRows()
	if p.PageSize > 0 && p.PageSize < size {
		size = p.PageSize
	}

	if timeout := pol.Timeout(); timeout > 0 {
		if trace {
			timeout *= 2
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	page, err := env.Session.QueryPage(ctx, db.PageRequest{
		Statement:   p.CQL,
		Consistency: p.Consistency,
		PageSize:    size,
		PageState:   state,
		Trace:       trace,
	})
	if err != nil {
		return gateOutcome(env, p.CQL, err)
	}

	table := statement.Names[0]
	rows, truncated := shapeRows(pol, table.Keyspace, table.Table, page)

	out := map[string]any{
		"columns": page.Columns,
		"rows":    rows,
	}
	if token := encodePageToken(page.PageState, p.CQL); token != "" {
		out["next_page_token"] = token
	}
	if len(page.Warnings) > 0 {
		out["warnings"] = page.Warnings
	}
	if len(truncated) > 0 {
		out["truncated"] = truncated
		out["truncated_note"] = fmt.Sprintf("values longer than %d bytes were cut", pol.MaxValueBytes())
	}
	if trace {
		out["trace"] = traceOf(env.Session)
	}

	o := jsonOutcome(out)
	o.Statement = p.CQL
	o.Rows = len(rows)
	return o
}

// shapeRows turns a page's rows into what goes back: JSON values, the hidden
// columns replaced, and long values cut. It reports the columns that were cut.
func shapeRows(p policy.Policy, keyspace, table string, page db.QueryPageResult) ([]map[string]any, []string) {
	limit := p.MaxValueBytes()
	cut := map[string]bool{}
	rows := make([]map[string]any, 0, len(page.Rows))
	for _, raw := range page.Rows {
		row := make(map[string]any, len(raw))
		for column, value := range raw {
			if p.Redacted(keyspace, table, column) {
				row[column] = policy.RedactedValue
				continue
			}
			v := db.JSONValue(value)
			if limited, wasCut := limitValue(v, limit); wasCut {
				v = limited
				cut[column] = true
			}
			row[column] = v
		}
		rows = append(rows, row)
	}
	var columns []string
	for c := range cut {
		columns = append(columns, c)
	}
	sort.Strings(columns)
	return rows, columns
}

// limitValue cuts a value whose JSON is longer than limit bytes. A string is
// cut where it is; anything else is replaced by the start of its JSON.
func limitValue(v any, limit int) (any, bool) {
	if s, ok := v.(string); ok {
		if len(s) <= limit {
			return v, false
		}
		return cutUTF8(s, limit) + "…", true
	}
	data, err := json.Marshal(v)
	if err != nil || len(data) <= limit {
		return v, false
	}
	return cutUTF8(string(data), limit) + "…", true
}

func cutUTF8(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && n < len(s) && (s[n]&0xC0) == 0x80 {
		n--
	}
	return s[:n]
}

// traceOf is the trace of the query just run. Cassandra writes a trace after
// it answers, so it is asked for a few times before giving up.
func traceOf(s *db.Session) map[string]any {
	for attempt := 0; attempt < 10; attempt++ {
		rows, headers, info, err := s.GetTraceData()
		if err == nil && info != nil {
			events := make([]map[string]string, 0, len(rows))
			for _, row := range rows {
				event := map[string]string{}
				for i, h := range headers {
					if i < len(row) {
						event[h] = row[i]
					}
				}
				events = append(events, event)
			}
			return map[string]any{
				"coordinator": info.Coordinator,
				"duration_us": info.Duration,
				"requests":    info.Pages,
				"events":      events,
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return map[string]any{"note": "Cassandra had not written the trace yet; run trace_query again"}
}

func jsonOutcome(v any) ToolOutcome {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errorOutcome(err.Error())
	}
	return ToolOutcome{Text: string(data), Structured: v}
}

// MCPCanChange reports whether any MCP tool can change data or schema. Until
// one can, permitting a write or schema command would do nothing, and the
// preferences window says so rather than offer it.
func MCPCanChange() bool {
	for _, def := range allToolDefinitions() {
		if def.MCP && def.Needs == NeedChange {
			return true
		}
	}
	return false
}
