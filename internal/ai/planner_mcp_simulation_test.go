package ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test what happens when we simulate the exact MCP flow
func TestMCP_SimulateJSONFlow_NumericList(t *testing.T) {
	// Simulate JSON from MCP client
	jsonArgs := `{
		"operation": "INSERT",
		"keyspace": "cqlai_test",
		"table": "users",
		"values": {
			"id": 2000,
			"name": "Test",
			"scores": [95, 87, 92]
		},
		"value_types": {
			"scores": "list<int>"
		}
	}`

	// Unmarshal like MCP does
	var args map[string]interface{}
	err := json.Unmarshal([]byte(jsonArgs), &args)
	assert.NoError(t, err)

	// Check what type scores becomes
	values := args["values"].(map[string]interface{})
	scores := values["scores"]
	t.Logf("Type of scores after JSON unmarshal: %T", scores)
	t.Logf("Value of scores: %v", scores)

	// Parse into SubmitQueryPlanParams (simulate parseSubmitQueryPlanParams)
	params := SubmitQueryPlanParams{
		Operation: args["operation"].(string),
		Keyspace:  args["keyspace"].(string),
		Table:     args["table"].(string),
		Values:    values,
	}

	if valueTypes, ok := args["value_types"].(map[string]interface{}); ok {
		params.ValueTypes = make(map[string]string)
		for k, v := range valueTypes {
			params.ValueTypes[k] = v.(string)
		}
	}

	// Convert to AIResult
	aiResult := params.ToQueryPlan()

	// Generate CQL
	cql, err := RenderCQL(aiResult)
	assert.NoError(t, err)

	t.Logf("Generated CQL: %s", cql)

	// Check if it has commas
	assert.Contains(t, cql, "[95, 87, 92]", "List should have commas")
}
