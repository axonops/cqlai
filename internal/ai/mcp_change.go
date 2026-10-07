package ai

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/validation"
)

// Proposing a change.
//
// The MCP server never runs a statement that changes data or schema. A model
// that wants one proposes it here, and gets back the statement, where it would
// run, and what it would do - from fixed sentences in cqlai, not the model's
// own account of it. The user reads that, and runs it, or not.

// ToolProposeChange is the name of the tool.
const ToolProposeChange ToolName = "propose_change"

func changeToolDefinitions() []ToolDefinition {
	return []ToolDefinition{{
		Name: ToolProposeChange.String(),
		Description: "Propose a change - INSERT, UPDATE, DELETE, BATCH, CREATE, ALTER, DROP or TRUNCATE - for the user to " +
			"review and run themselves. This server never runs it. Returns the statement, where it would run, and what it will do, " +
			"which you must show the user in full.",
		Parameters: map[string]any{
			"cql":    map[string]any{"type": "string", "description": "One CQL statement, with tables named as keyspace.table"},
			"reason": map[string]any{"type": "string", "description": "Why, in a sentence, for the user"},
		},
		Required: []string{"cql"},
		MCP:      true, Needs: NeedNothing, ReadOnly: true,
	}}
}

// ProposeChangeParams is the statement proposed, and why.
type ProposeChangeParams struct {
	CQL    string `json:"cql"`
	Reason string `json:"reason,omitempty"`
}

func (p ProposeChangeParams) Validate() error {
	if strings.TrimSpace(p.CQL) == "" {
		return fmt.Errorf("cql is required")
	}
	if len(p.Reason) > 1000 {
		return fmt.Errorf("reason is too long: a sentence")
	}
	return nil
}

// proposeChange checks a proposed change and says what it will do. Nothing
// is run.
func proposeChange(env ToolEnv, p ProposeChangeParams) ToolOutcome {
	statement := strings.TrimRight(strings.TrimSpace(p.CQL), "; \n\t") + ";"
	s, err := env.Policy.CheckProposal(statement)
	if err != nil {
		return gateOutcome(env, statement, err)
	}

	// What is known about the table, for saying how much a DELETE removes.
	var facts validation.Facts
	if env.Session != nil && len(s.Names) > 0 && s.Names[0].Table != "" {
		facts.PartitionKey, facts.ClusteringKey = env.Session.PrimaryKey(s.Names[0].Keyspace, s.Names[0].Table)
	}

	changes := "data"
	if s.Kind == validation.KindSchema {
		changes = "schema"
	}
	where := env.Policy.Connection()
	if where == "" {
		where = "the connection this server is serving"
	}

	out := map[string]any{
		"statement":    statement,
		"connection":   where,
		"changes":      changes,
		"implications": validation.Implications(s, facts),
		"not_run":      "cqlai has not run this, and will not. Show the user the statement and every implication, then let them decide whether to run it themselves.",
	}
	if strings.TrimSpace(p.Reason) != "" {
		out["reason"] = strings.TrimSpace(p.Reason)
	}

	// In the shell, the statement is put in front of the user too.
	if env.Propose != nil {
		env.Propose(Handover{Statement: statement, Notes: validation.Implications(s, facts)})
		out["in_cqlai"] = "It has also been put in cqlai's prompt, unrun, or in its Console if something was being typed there."
	}

	o := jsonOutcome(out)
	o.Statement = statement
	o.Proposed = true
	return o
}
