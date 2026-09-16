package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// The review under a definition.
//
// The browser says what a table is. What it does not say is whether the
// definition is any good - whether the partition grows without bound, whether
// the clustering serves the query the table is obviously for. Cassandra
// punishes those months later, by which time the table has data in it.

// reviewModel is the SCHEMA view on a table, with a provider configured and
// room to draw a review under the definition.
func reviewModel(t *testing.T) *MainModel {
	t.Helper()

	m := schemaModel(t)
	m.windowWidth, m.windowHeight = 120, 30
	m.historyViewport = viewport.New(viewport.WithWidth(120), viewport.WithHeight(20))
	m.aiConfig = &config.AIConfig{Provider: "anthropic", APIKey: "sk-ant-test"}

	// On the table rather than the keyspace: a definition is what is reviewed.
	m.schema.expanded["my_keyspace"] = true
	m.schema.selected = 2 // my_keyspace, events, users
	m.showSchemaDetail()
	return m
}

// TestTheDefinitionPaneHasAButtonToReviewIt.
func TestTheDefinitionPaneHasAButtonToReviewIt(t *testing.T) {
	m := reviewModel(t)

	drawn := stripAnsi(m.viewSchema(m.windowWidth, m.schemaHeight()))
	assert.Contains(t, drawn, reviewButton)

	// And it is where it is drawn.
	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	assert.True(t, m.schemaReviewButtonAt(g.buttonFrom, tabBarHeight))
	assert.True(t, m.schemaReviewButtonAt(g.buttonTo, tabBarHeight))
	assert.False(t, m.schemaReviewButtonAt(g.buttonFrom-1, tabBarHeight))
	assert.False(t, m.schemaReviewButtonAt(g.buttonFrom, tabBarHeight+1), "the row under it is the definition")
}

// TestAltAReviewsTheDefinitionShowing, since a button is no use to the keyboard.
func TestAltAReviewsTheDefinitionShowing(t *testing.T) {
	m := reviewModel(t)

	m, cmd := m.handleKeyboardInput(tea.KeyPressMsg{Code: 'a', Mod: tea.ModAlt})
	require.NotNil(t, cmd, "the definition should have been sent")

	assert.True(t, m.schema.review.open)
	assert.True(t, m.schema.review.running)
	assert.Equal(t, "my_keyspace.users", m.schema.review.of)

	drawn := stripAnsi(m.viewSchema(m.windowWidth, m.schemaHeight()))
	assert.Contains(t, drawn, reviewingNow)
	assert.Contains(t, drawn, "Reading the definition")
}

// TestTheAnswerAppearsUnderTheDefinition.
func TestTheAnswerAppearsUnderTheDefinition(t *testing.T) {
	m := reviewModel(t)
	m, _ = m.startSchemaReview()

	m, _ = m.schemaReviewed(schemaReviewedMsg{
		of:   "my_keyspace.users",
		text: "WHAT IT IS\nOne row per user.\n\nRISKS\n- Nothing that will bite.",
	})

	assert.False(t, m.schema.review.running)

	lines := strings.Split(stripAnsi(m.viewSchema(m.windowWidth, m.schemaHeight())), "\n")
	assert.GreaterOrEqual(t, lineHolding(lines, "CREATE TABLE"), 0, "the definition is still there")
	assert.Greater(t, lineHolding(lines, "One row per user"), lineHolding(lines, "CREATE TABLE"),
		"and the answer is under it")
	assert.GreaterOrEqual(t, lineHolding(lines, reviewHeading), 0, "with the line between them")
}

// TestAnAnswerAboutAnotherTableIsNotShownUnderThisOne.
//
// The reviews take a while and the tree does not: by the time one comes back
// the reader may be somewhere else entirely.
func TestAnAnswerAboutAnotherTableIsNotShownUnderThisOne(t *testing.T) {
	m := reviewModel(t)
	m, _ = m.startSchemaReview()

	m, _ = m.schemaReviewed(schemaReviewedMsg{of: "my_keyspace.events", text: "About another table."})

	assert.Empty(t, m.schema.review.text)
	assert.NotContains(t, stripAnsi(m.viewSchema(m.windowWidth, m.schemaHeight())), "About another table")
}

// TestMovingToAnotherDefinitionPutsTheReviewAway.
func TestMovingToAnotherDefinitionPutsTheReviewAway(t *testing.T) {
	m := reviewModel(t)
	m, _ = m.startSchemaReview()
	m, _ = m.schemaReviewed(schemaReviewedMsg{of: "my_keyspace.users", text: "WHAT IT IS\nOne row per user."})
	require.True(t, m.schema.review.open)

	m, _ = m.moveSchemaSelection(-1)

	assert.False(t, m.schema.review.open, "the answer was about the table before this one")
	assert.NotContains(t, stripAnsi(m.viewSchema(m.windowWidth, m.schemaHeight())), "One row per user")
}

// TestTheLineUnderTheDefinitionIsDragged.
func TestTheLineUnderTheDefinitionIsDragged(t *testing.T) {
	m := reviewModel(t)
	m, _ = m.startSchemaReview()

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	rule := g.ruleRow + tabBarHeight + schemaHeaderRows

	require.True(t, m.schemaReviewRuleAt(g.treeWidth+2, rule))
	assert.False(t, m.schemaReviewRuleAt(1, rule), "the tree side of it is the tree")

	m, _ = m.handleMousePress(tea.Mouse{X: g.treeWidth + 2, Y: rule, Button: tea.MouseLeft})
	require.True(t, m.schema.review.dragging)

	m.dragSchemaRule(rule - 3)
	assert.Equal(t, g.ruleRow-3, m.schemaGeometry(m.windowWidth, m.schemaHeight()).ruleRow)

	// Neither half is dragged away.
	m.dragSchemaRule(-50)
	after := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	assert.GreaterOrEqual(t, after.detailRows, definitionMin)
	assert.GreaterOrEqual(t, after.reviewRows, reviewMin)

	m.dragSchemaRule(500)
	after = m.schemaGeometry(m.windowWidth, m.schemaHeight())
	assert.GreaterOrEqual(t, after.detailRows, definitionMin)
	assert.GreaterOrEqual(t, after.reviewRows, reviewMin)
}

// TestThePaneCanBePutAwayWithEscape.
func TestThePaneCanBePutAwayWithEscape(t *testing.T) {
	m := reviewModel(t)
	m, _ = m.startSchemaReview()
	m, _ = m.schemaReviewed(schemaReviewedMsg{of: "my_keyspace.users", text: "One row per user."})

	m, _ = m.handleEscapeKey()

	assert.False(t, m.schema.review.open)
	assert.Equal(t, -1, m.schemaGeometry(m.windowWidth, m.schemaHeight()).ruleRow)
}

// TestTheWheelOverTheReviewMovesTheReview, and over the definition moves that.
func TestTheWheelOverTheReviewMovesTheReview(t *testing.T) {
	m := reviewModel(t)
	m, _ = m.startSchemaReview()
	m, _ = m.schemaReviewed(schemaReviewedMsg{
		of:   "my_keyspace.users",
		text: strings.Repeat("a line of the review\n", 30),
	})

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	over := g.treeWidth + 2

	assert.True(t, m.inSchemaReview(over, g.reviewTop+tabBarHeight+schemaHeaderRows))
	assert.False(t, m.inSchemaReview(over, g.ruleRow+tabBarHeight+schemaHeaderRows), "the rule is not the review")

	m.scrollSchemaReview(4)
	assert.Equal(t, 4, m.schema.review.scroll)

	// And it stops at the end of what there is.
	m.scrollSchemaReview(500)
	assert.Equal(t, len(m.schema.review.lines)-g.reviewRows, m.schema.review.scroll)
}

// TestThereIsNothingToReviewWithoutADefinitionOrAProvider.
func TestThereIsNothingToReviewWithoutADefinitionOrAProvider(t *testing.T) {
	m := reviewModel(t)
	m.schema.detail = nil

	m, cmd := m.startSchemaReview()
	assert.Nil(t, cmd)
	assert.Contains(t, m.fullHistoryContent, "no definition to review")

	m = reviewModel(t)
	m.aiConfig = nil
	m, cmd = m.startSchemaReview()
	assert.Nil(t, cmd)
	assert.Contains(t, m.fullHistoryContent, "No AI provider is configured")
}
