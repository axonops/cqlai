package ui

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
)

// Reading a trace with help.
//
// A trace is forty rows of microsecond timings and node names, and what anyone
// wants from it is which step was slow and what to do about it. That is work to
// read out of the table, and it is work a model is good at.
//
// The answer belongs beside the trace it is about rather than in the CHAT tab:
// the two are read together, and the line between them moves so that whichever
// one is being read can have the room.

// The rows of the view that are not the trace or the analysis.
const (
	traceHeaderRows = 1 // the button
	traceRuleRows   = 1 // the line between the panes

	// What each pane is never squeezed below, so that dragging the line cannot
	// leave either of them a single row.
	traceMin    = 3
	analysisMin = 3
)

// traceAnalysis is the pane under the trace, and what is in it.
type traceAnalysis struct {
	open     bool // the pane is showing
	running  bool // the answer has been asked for and has not arrived
	failed   bool // what is in it is what went wrong
	text     string
	rows     int // how many rows the pane has, the rule aside
	viewport viewport.Model

	// dragging is set while the line between the panes is being moved, which
	// is what makes the motion a drag rather than a pointer going past.
	dragging bool

	// drawn is the view as it was last drawn, which is what a selection is
	// taken from.
	drawn []string
}

// traceAnalysedMsg is the answer coming back.
type traceAnalysedMsg struct {
	text   string
	failed bool
}

// analyseTrace asks the configured provider what the trace shows.
func analyseTrace(providerConfig *config.AIConfig, trace string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		said, err := ai.AnalyseTrace(ctx, providerConfig, trace)
		if err != nil {
			logger.DebugfToFile("Trace", "Analysis failed: %v", err)
			return traceAnalysedMsg{text: err.Error(), failed: true}
		}
		return traceAnalysedMsg{text: said}
	}
}

// startTraceAnalysis sends the trace off to be read, and opens the pane it
// will come back into.
func (m *MainModel) startTraceAnalysis() (*MainModel, tea.Cmd) {
	switch {
	case !m.hasTrace:
		return m.report("There is no trace to analyse. Turn tracing on with TRACING ON and run a query.")
	case m.aiConfig == nil || m.aiConfig.Provider == "":
		return m.report("No AI provider is configured. FILE > PREFERENCES has the settings, under CHAT.")
	case m.trace.running:
		return m, nil
	}

	// Half the view the first time, which is what there is room for while the
	// answer is being read against the trace it is about. After that it is
	// wherever the line between them was left.
	if m.trace.rows == 0 {
		m.trace.rows = max((m.traceHeight()-traceHeaderRows-traceRuleRows)/2, analysisMin)
	}

	m.trace.open = true
	m.trace.running = true
	m.trace.failed = false
	m.trace.text = ""
	m.fitTracePanes()

	return m, analyseTrace(m.aiConfig, m.traceAsText())
}

// traceAnalysed puts the answer in the pane.
func (m *MainModel) traceAnalysed(msg traceAnalysedMsg) (*MainModel, tea.Cmd) {
	m.trace.running = false
	m.trace.failed = msg.failed
	m.trace.text = msg.text
	m.trace.open = true
	m.fitTracePanes()
	return m, nil
}

// traceAsText is the trace as the model is given it.
//
// The rows as they were read from system_traces rather than the drawn table:
// the columns of the drawn one are cut to the width of the screen, and what is
// cut off is the end of the activity - which is the part that says what
// happened.
func (m *MainModel) traceAsText() string {
	var said strings.Builder

	if m.traceInfo != nil {
		fmt.Fprintf(&said, "Coordinator: %s\nTotal duration: %d microseconds\n\n",
			m.traceInfo.Coordinator, m.traceInfo.Duration)
	}

	for _, row := range m.traceData {
		said.WriteString(strings.Join(row, " | "))
		said.WriteString("\n")
	}
	return said.String()
}

// fitTracePanes gives each pane its rows.
//
// One description of the split, used to draw the view, to place the line
// between the panes and to work out what a press landed on.
func (m *MainModel) fitTracePanes() {
	height := m.traceHeight()

	// A column narrower than the window: the scrollbar lives there, the way it
	// does in the console.
	paneWidth := max(m.windowWidth-scrollbarWidth, 1)

	if !m.trace.open {
		m.traceViewport.SetHeight(max(height-traceHeaderRows, 1))
		m.traceViewport.SetWidth(paneWidth)
		return
	}

	// What the pane was asked for is kept as it was asked for, and squeezed to
	// fit here. A terminal made small and then large again gives the pane back
	// the size it had: writing the squeezed size back would lose it, and there
	// is no way to drag a line that has been pushed off the bottom.
	room := max(height-traceHeaderRows-traceRuleRows, traceMin+analysisMin)
	rows := min(max(m.trace.rows, analysisMin), room-traceMin)

	m.traceViewport.SetWidth(paneWidth)
	m.traceViewport.SetHeight(room - rows)

	m.trace.viewport.SetWidth(paneWidth)
	m.trace.viewport.SetHeight(rows)
	m.trace.viewport.SetContent(m.wrapTraceAnalysis())
}

// wrapTraceAnalysis is what the pane holds: the answer, what went wrong, or
// the line that says it has been asked for.
//
// The answer comes back in sections - TIME, FINDINGS, WHAT TO DO - and the
// headings are drawn as headings, so that the shape of it can be seen without
// reading it.
func (m *MainModel) wrapTraceAnalysis() string {
	switch {
	case m.trace.running:
		return m.styles.MutedText.Render("  Reading the trace...")
	case m.trace.failed:
		return m.styles.ErrorText.Render(m.wrapAIText(m.trace.text))
	}

	heading := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)

	lines := strings.Split(m.trace.text, "\n")
	drawn := make([]string, 0, len(lines))
	for _, line := range lines {
		if isHeadingLine(line) {
			drawn = append(drawn, heading.Render(line))
			continue
		}
		drawn = append(drawn, m.wrapAIText(line))
	}
	return strings.Join(drawn, "\n")
}

// isHeadingLine reports whether a line is one of the answer's headings: a line
// of capitals, and nothing else on it.
func isHeadingLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	for _, letter := range trimmed {
		if !unicode.IsUpper(letter) && letter != ' ' {
			return false
		}
	}
	return true
}

// traceHeight is how many rows the view has under its tabs.
func (m *MainModel) traceHeight() int {
	return m.viewHeight()
}

// closeTraceAnalysis puts the pane away, giving the trace the whole view back.
func (m *MainModel) closeTraceAnalysis() {
	m.trace.open = false
	m.trace.dragging = false
	m.fitTracePanes()
}
