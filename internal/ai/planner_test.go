package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The warning above a plan the modal is about to run.
//
// It was set by a PlanValidator that only the unreachable half of this package
// called, so it never appeared: the model could put a warning on a plan
// through its tool call, and for DROP and DELETE nothing else did. It is set
// on the way to the CQL now, which is the path that runs.

// TestADestructiveOperationIsWarnedAbout.
func TestADestructiveOperationIsWarnedAbout(t *testing.T) {
	for _, operation := range []string{"DROP", "DELETE", "TRUNCATE"} {
		plan := &AIResult{Operation: operation, Table: "users", Keyspace: "app"}

		warnAboutPlan(plan)

		assert.Contains(t, plan.Warning, "permanently delete data", "%s", operation)
	}
}

// TestTheWarningIsSetOnTheWayToTheCQL, which is the path the modal takes: it
// renders the plan and then draws the warning above it.
func TestTheWarningIsSetOnTheWayToTheCQL(t *testing.T) {
	plan := &AIResult{
		Operation: "DELETE",
		Keyspace:  "app",
		Table:     "users",
		Where:     []WhereClause{{Column: "id", Operator: "=", Value: 1}},
	}

	cql, err := RenderCQL(plan)

	require.NoError(t, err)
	assert.Contains(t, cql, "DELETE FROM app.users")
	assert.Contains(t, plan.Warning, "permanently delete data")
}

// TestAlterSaysItChangesTheSchema.
func TestAlterSaysItChangesTheSchema(t *testing.T) {
	plan := &AIResult{Operation: "ALTER", Table: "users"}

	warnAboutPlan(plan)

	assert.Contains(t, plan.Warning, "modify the schema")
}

// TestReadingIsNotWarnedAbout.
func TestReadingIsNotWarnedAbout(t *testing.T) {
	plan := &AIResult{Operation: "SELECT", Table: "users", ReadOnly: true}

	_, err := RenderCQL(plan)

	require.NoError(t, err)
	assert.Empty(t, plan.Warning)
}

// TestAPlanThatSaysItIsSafeIsLeftAlone: ReadOnly is the model's word for a
// statement that changes nothing, and it is taken.
func TestAPlanThatSaysItIsSafeIsLeftAlone(t *testing.T) {
	plan := &AIResult{Operation: "DELETE", Table: "users", ReadOnly: true}

	warnAboutPlan(plan)

	assert.Empty(t, plan.Warning)
}

// TestTheModelsOwnWarningIsNotOverwritten: it knows what the statement does to
// this schema, which a list of operation names does not.
func TestTheModelsOwnWarningIsNotOverwritten(t *testing.T) {
	plan := &AIResult{
		Operation: "DROP",
		Table:     "users",
		Warning:   "users is the only copy of the customer records",
	}

	warnAboutPlan(plan)

	assert.Equal(t, "users is the only copy of the customer records", plan.Warning)
}
