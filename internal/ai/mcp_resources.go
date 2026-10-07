package ai

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axonops/cqlai/internal/policy"
)

// MCP resources and prompts.
//
// A resource is something a client can read and attach without calling a
// tool: here, a keyspace's or a table's definition. A prompt is a request the
// user can pick in the client. Both go through the same policy as the tools.

// SchemaResourceURI is a keyspace's or a table's resource.
func SchemaResourceURI(keyspace, table string) string {
	if table == "" {
		return "cql://schema/" + keyspace
	}
	return "cql://schema/" + keyspace + "/" + table
}

// SchemaResourceTemplate is the resource of any table.
const SchemaResourceTemplate = "cql://schema/{keyspace}/{table}"

// SchemaResources is the keyspaces a client is offered as resources: the
// visible ones, when DESCRIBE is permitted and there is a cluster.
func SchemaResources(env ToolEnv) []string {
	if env.Session == nil || env.Schema == nil || !env.Policy.Permits("DESCRIBE") {
		return nil
	}
	env.Schema.refreshIfChanged()
	keyspaces := visibleKeyspaces(env.Policy, env.Schema)
	sort.Strings(keyspaces)
	return keyspaces
}

// ReadSchemaResource is a resource's text: what describe gives for it.
func ReadSchemaResource(env ToolEnv, uri string) ToolOutcome {
	rest, ok := strings.CutPrefix(uri, "cql://schema/")
	if !ok {
		return errorOutcome("there is no resource " + uri)
	}
	keyspace, table, _ := strings.Cut(rest, "/")
	if !env.Policy.Permits("DESCRIBE") {
		return refusedOutcome("DESCRIBE is not permitted on this server")
	}
	if o, ok := needsCluster(env); !ok {
		return o
	}

	params := DescribeParams{Kind: "keyspace", Keyspace: keyspace}
	if table != "" {
		params = DescribeParams{Kind: "table", Keyspace: keyspace, Name: table}
	}
	if err := params.Validate(); err != nil {
		return errorOutcome("there is no resource " + uri)
	}
	return describe(env, params)
}

// PromptArgument is one argument of a prompt.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// PromptDefinition is a prompt a client can offer the user.
type PromptDefinition struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments"`
	needs       ToolNeed
}

func promptDefinitions() []PromptDefinition {
	return []PromptDefinition{
		{
			Name:        "review_table",
			Description: "Review a table's definition: what it is for, what will go wrong with it, and what to change",
			Arguments:   []PromptArgument{{Name: "table", Description: "keyspace.table", Required: true}},
			needs:       NeedDescribe,
		},
		{
			Name:        "diagnose_query",
			Description: "Trace a query and explain where its time goes, and what to change",
			Arguments:   []PromptArgument{{Name: "cql", Description: "One CQL SELECT, with tables named as keyspace.table", Required: true}},
			needs:       NeedSelect,
		},
	}
}

// MCPPrompts is the prompts a client is offered under a policy.
func MCPPrompts(p policy.Policy) []PromptDefinition {
	var prompts []PromptDefinition
	for _, d := range promptDefinitions() {
		if offered(ToolDefinition{Needs: d.needs}, p) {
			prompts = append(prompts, d)
		}
	}
	return prompts
}

// GetMCPPrompt is a prompt's text, with its arguments filled in. The
// instructions are the ones the shell's own Alt+A uses, so the two read a
// definition or a trace the same way.
func GetMCPPrompt(env ToolEnv, name string, args map[string]string) (string, error) {
	var def *PromptDefinition
	for _, d := range MCPPrompts(env.Policy) {
		if d.Name == name {
			def = &d
		}
	}
	if def == nil {
		return "", fmt.Errorf("there is no prompt called %s", name)
	}
	for _, a := range def.Arguments {
		if a.Required && strings.TrimSpace(args[a.Name]) == "" {
			return "", fmt.Errorf("%s needs %s", name, a.Name)
		}
	}

	switch name {
	case "review_table":
		keyspace, table, ok := strings.Cut(strings.TrimSpace(args["table"]), ".")
		params := DescribeParams{Kind: "table", Keyspace: keyspace, Name: table}
		if !ok || params.Validate() != nil {
			return "", fmt.Errorf("table has to be keyspace.table")
		}
		if !env.Policy.VisibleTable(keyspace, table) {
			return "", fmt.Errorf("%s.%s is not visible to this server", keyspace, table)
		}
		// The definition, when it can be had now. Without a cluster the
		// model is asked to fetch it with describe itself.
		definition := describe(env, params)
		if env.Session == nil || definition.IsError {
			return SchemaInstructions + "\n\nUse the describe tool to read the definition of " +
				keyspace + "." + table + ", then review it.", nil
		}
		return SchemaInstructions + "\n\n" + definition.Text, nil

	case "diagnose_query":
		return TraceInstructions + "\n\nRun this query with the trace_query tool, then read its trace:\n\n" +
			strings.TrimSpace(args["cql"]), nil
	}
	return "", fmt.Errorf("there is no prompt called %s", name)
}
