package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The buttons that read what is on screen with the AI.
//
// Alt+A was bound and nowhere written down in the schema browser: the trace
// view had a line of text beside its button saying so, and the definition pane
// had no room for one, so nothing on screen said the button had a key at all.
// The key is on both buttons now.
//
// And being told why the button did nothing used to cost you the view you were
// in: m.report writes into the console and switches to it. It is said over the
// view now.

// TestBothButtonsNameTheKeyThatWorksThem.
func TestBothButtonsNameTheKeyThatWorksThem(t *testing.T) {
	for _, button := range []string{reviewButton, analyseButton} {
		assert.Contains(t, button, aiKeyLabel, "%q should name its key", button)
	}
}

// TestTheKeyOnTheButtonIsTheKeyThatIsBound.
//
// A button naming a key nothing answers, and a key no button mentions, are the
// same defect from either end. One pair of constants, and the binding reads the
// same one.
func TestTheKeyOnTheButtonIsTheKeyThatIsBound(t *testing.T) {
	assert.Equal(t, aiKey, strings.ToLower(aiKeyLabel))
	assert.True(t, isShellKey(aiKey), "it has to reach the handler from every view")
}

// TestTheKeyReviewsTheDefinition, which is what the button does.
func TestTheKeyReviewsTheDefinition(t *testing.T) {
	m := selectableSchema(t)
	m.aiConfig = configuredAI()

	m, _ = m.handleKeyboardInput(keyPress(aiKey))

	assert.True(t, m.schema.review.running, "%s should start the review", aiKeyLabel)
	assert.Equal(t, ModalNone, m.modal.Type, "with nothing in the way of it")
}

// TestTheButtonIsDrawnWithItsKey, at a width that has room for both.
func TestTheButtonIsDrawnWithItsKey(t *testing.T) {
	m := selectableSchema(t)

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	drawn := stripAnsi(m.schemaDetailHeadingRow(g, m.styles.AccentText))

	assert.Contains(t, drawn, aiKeyLabel)
	assert.Contains(t, drawn, "Review Schema")
}

// TestNoProviderIsSaidOverTheSchemaViewRatherThanInTheConsole.
func TestNoProviderIsSaidOverTheSchemaViewRatherThanInTheConsole(t *testing.T) {
	m := selectableSchema(t)
	m.aiConfig = nil
	before := m.fullHistoryContent

	m, _ = m.startSchemaReview()

	assert.Equal(t, ModalMessage, m.modal.Type)
	assert.Contains(t, m.modal.Title, "No AI provider is configured")
	assert.Contains(t, m.modal.Message, "PREFERENCES", "and says where the setting is")

	assert.Equal(t, "schema", m.viewMode, "the view you were in is the view you are left in")
	assert.Equal(t, before, m.fullHistoryContent, "and nothing went into the console")
	assert.False(t, m.schema.review.open, "no empty review pane either")
}

// TestNoProviderIsSaidOverTheTraceViewToo: the same button, the same answer.
func TestNoProviderIsSaidOverTheTraceViewToo(t *testing.T) {
	m := traceModel(t)
	m.aiConfig = nil

	m, _ = m.startTraceAnalysis()

	assert.Equal(t, ModalMessage, m.modal.Type)
	assert.Equal(t, "trace", m.viewMode)
	assert.False(t, m.trace.open)
}

// TestNothingToReviewIsSaidTheSameWay.
func TestNothingToReviewIsSaidTheSameWay(t *testing.T) {
	m := schemaModel(t)
	m.aiConfig = configuredAI()
	m.schema.detail = nil

	m, _ = m.startSchemaReview()

	assert.Equal(t, ModalMessage, m.modal.Type)
	assert.Contains(t, m.modal.Title, "Nothing to review")
	assert.Equal(t, "schema", m.viewMode)
}

// TestAMessageModalIsDismissedByEitherKey. It asks nothing, so there is nothing
// to answer.
func TestAMessageModalIsDismissedByEitherKey(t *testing.T) {
	for _, key := range []string{"enter", "esc"} {
		m := selectableSchema(t)
		m.aiConfig = nil
		m, _ = m.startSchemaReview()
		require.Equal(t, ModalMessage, m.modal.Type)

		switch key {
		case "enter":
			m, _ = m.handleModalConfirmation("")
		case "esc":
			m, _ = m.handleEscapeKey()
		}

		assert.Equal(t, ModalNone, m.modal.Type, "%s should put it away", key)
		assert.Equal(t, "schema", m.viewMode, "%s should leave the view alone", key)
	}
}

// TestAMessageModalHasOneButton: it is not a question.
func TestAMessageModalHasOneButton(t *testing.T) {
	modal := NewMessageModal("A title", "Something worth saying.")

	assert.Equal(t, ModalMessage, modal.Type)
	require.Len(t, modal.Choices, 1)
	assert.Equal(t, "OK", modal.Choices[0])
}
