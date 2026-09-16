package ui

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
)

// Reading a definition with help.
//
// The browser shows what a table is. What it does not say is whether the
// definition is any good - whether the partition grows without bound, whether
// the clustering serves the query the table is obviously for, whether the
// compaction strategy matches how the data is written. Cassandra punishes those
// months later, by which time the table has data in it and the answer is to
// write it all again somewhere else.
//
// The answer goes under the definition it is about, in the same pane, the way
// the trace analysis goes under the trace.

// The rows of the definition pane that are neither definition nor review.
const (
	reviewRuleRows = 1

	// What each half is never squeezed below, so the line between them cannot
	// be dragged until one of them is a single row.
	definitionMin = 3
	reviewMin     = 3
)

// reviewButton is what the button says, and what it says while it waits.
const (
	reviewButton  = "[ Review Schema ]"
	reviewingNow  = "[ Reviewing... ]"
	reviewHeading = " Review "
)

// schemaReview is the pane under the definition, and what is in it.
type schemaReview struct {
	open    bool
	running bool
	failed  bool
	text    string

	// of is the row the review is about. Moving to another table puts the pane
	// away rather than leaving an answer about one table under another's
	// definition.
	of string

	lines  []string // as drawn, wrapped to the pane
	scroll int
	rows   int // how many rows the pane has, the rule aside

	dragging bool
}

// schemaReviewedMsg is the answer coming back.
type schemaReviewedMsg struct {
	of     string
	text   string
	failed bool
}

// reviewSchema asks the configured provider what it makes of a definition.
func reviewSchema(providerConfig *config.AIConfig, of, definition string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		said, err := ai.ReviewSchema(ctx, providerConfig, definition)
		if err != nil {
			logger.DebugfToFile("Schema", "Review of %s failed: %v", of, err)
			return schemaReviewedMsg{of: of, text: err.Error(), failed: true}
		}
		return schemaReviewedMsg{of: of, text: said}
	}
}

// startSchemaReview sends the definition showing off to be read.
func (m *MainModel) startSchemaReview() (*MainModel, tea.Cmd) {
	row, ok := m.schema.current()

	switch {
	case !ok || len(m.schema.detail) == 0:
		return m.report("There is no definition to review. Pick a keyspace or a table on the left.")
	case m.aiConfig == nil || m.aiConfig.Provider == "":
		return m.report("No AI provider is configured. FILE > PREFERENCES has the settings, under CHAT.")
	case m.schema.review.running:
		return m, nil
	}

	// Half the pane the first time; after that, wherever the line was left.
	if m.schema.review.rows == 0 {
		rows := m.schemaGeometry(m.windowWidth, m.schemaHeight()).height
		m.schema.review.rows = max((rows-reviewRuleRows)/2, reviewMin)
	}

	m.schema.review.open = true
	m.schema.review.running = true
	m.schema.review.failed = false
	m.schema.review.text = ""
	m.schema.review.of = row.key()
	m.schema.review.scroll = 0
	m.wrapSchemaReview()

	return m, reviewSchema(m.aiConfig, row.key(), strings.Join(m.schema.detail, "\n"))
}

// schemaReviewed puts the answer in the pane.
func (m *MainModel) schemaReviewed(msg schemaReviewedMsg) (*MainModel, tea.Cmd) {
	// The answer to a question about a table nobody is looking at any more is
	// not shown under the one they are.
	if msg.of != m.schema.review.of {
		m.schema.review.running = false
		return m, nil
	}

	m.schema.review.running = false
	m.schema.review.failed = msg.failed
	m.schema.review.text = msg.text
	m.schema.review.open = true
	m.schema.review.scroll = 0
	m.wrapSchemaReview()
	return m, nil
}

// closeSchemaReview puts the pane away, giving the definition the room back.
func (m *MainModel) closeSchemaReview() {
	m.schema.review = schemaReview{rows: m.schema.review.rows}
}

// wrapSchemaReview lays the answer out for the pane it is drawn in.
func (m *MainModel) wrapSchemaReview() {
	width := max(m.schemaGeometry(m.windowWidth, m.schemaHeight()).detailWidth-1, 20)

	if m.schema.review.running {
		m.schema.review.lines = []string{"Reading the definition..."}
		return
	}

	lines := []string{}
	for _, line := range strings.Split(m.schema.review.text, "\n") {
		lines = append(lines, wrapLine(line, width)...)
	}
	m.schema.review.lines = lines
}

// reviewIsAbout reports whether the review showing belongs to the row the tree
// is on: moving to another table puts it away.
func (m *MainModel) reviewIsAbout(key string) bool {
	return m.schema.review.of == key
}

// Where a press lands in the definition pane.
//
// One description of the rows, in schemaGeometry, used to draw them and to
// work out what was pressed: two copies of that is how a drag grabs the row
// above the line being pointed at.

// schemaReviewButtonAt reports whether a press landed on the button.
func (m *MainModel) schemaReviewButtonAt(col, row int) bool {
	if m.viewMode != "schema" || row != tabBarHeight {
		return false
	}

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	return col >= g.buttonFrom && col <= g.buttonTo
}

// schemaReviewRuleAt reports whether a press landed on the line between the
// definition and the review, which is in the right-hand pane only.
func (m *MainModel) schemaReviewRuleAt(col, row int) bool {
	if m.viewMode != "schema" || !m.schema.review.open {
		return false
	}

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	return col > g.treeWidth && row-tabBarHeight-schemaHeaderRows == g.ruleRow
}

// inSchemaReview reports whether a screen row is in the review pane.
func (m *MainModel) inSchemaReview(col, row int) bool {
	if m.viewMode != "schema" || !m.schema.review.open {
		return false
	}

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	return col > g.treeWidth && row-tabBarHeight-schemaHeaderRows > g.ruleRow
}

// dragSchemaRule moves the line between the definition and the review to where
// the pointer is.
//
// Each half keeps whatever it is given between its own smallest size and the
// one that leaves the other its smallest, so the line cannot be dragged off
// the top or the bottom of the pane.
func (m *MainModel) dragSchemaRule(row int) {
	if !m.schema.review.dragging {
		return
	}

	// The rows inside the panes, which is the screen less the tab line and the
	// two heading rows - the same height the geometry lays out.
	height := m.schemaGeometry(m.windowWidth, m.schemaHeight()).height

	rule := min(max(row-tabBarHeight-schemaHeaderRows, definitionMin), height-reviewRuleRows-reviewMin)
	m.schema.review.rows = height - rule - reviewRuleRows
	m.wrapSchemaReview()
}

// scrollSchemaReview moves the review under the definition.
func (m *MainModel) scrollSchemaReview(by int) {
	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())

	furthest := max(len(m.schema.review.lines)-g.reviewRows, 0)
	m.schema.review.scroll = min(max(m.schema.review.scroll+by, 0), furthest)
}
