package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// menuModel is the FILE menu open, with no session: PREFERENCES and QUIT have
// nothing to do with one, so both have to be reachable without it.
func menuModel(t *testing.T) *MainModel {
	t.Helper()

	m := helpModel()
	m.windowWidth = 120
	m.windowHeight = 40
	m.openFileMenu(0)
	require.True(t, m.fileMenu.active)
	return m
}

// menuEntry is where an entry sits in the menu.
func menuEntry(t *testing.T, label string) int {
	t.Helper()

	for i, item := range fileMenuItems() {
		if item.label == label {
			return i
		}
	}
	t.Fatalf("no entry called %q", label)
	return -1
}

// TestTheMenuIsThreeGroups: what moves data in and out, what changes how cqlai
// starts, and leaving. Each line marks a change of subject.
func TestTheMenuIsThreeGroups(t *testing.T) {
	items := fileMenuItems()

	var groups [][]string
	group := []string{}
	for _, item := range items {
		if item.rule {
			groups = append(groups, group)
			group = []string{}
			continue
		}
		group = append(group, item.label)
	}
	groups = append(groups, group)

	assert.Equal(t, [][]string{
		{"SAVE RESULTS", "AUTOSAVE", "SOURCE", "COPY TO", "COPY FROM"},
		{"PREFERENCES"},
		{"QUIT"},
	}, groups)
}

// TestASeparatorCannotBeChosen: by arrow key, by Enter, or by clicking it.
func TestASeparatorCannotBeChosen(t *testing.T) {
	items := fileMenuItems()

	for rule, item := range items {
		if !item.rule {
			continue
		}

		m := menuModel(t)

		// Down from the entry above the line lands past it.
		m.fileMenu.selected = rule - 1
		m, _ = m.moveFileMenu(1)
		assert.Greater(t, m.fileMenu.selected, rule)

		// Up from the entry below it lands above it, on whichever of those can
		// be picked - the two COPY entries need a session and there is none.
		m.fileMenu.selected = rule + 1
		m, _ = m.moveFileMenu(-1)
		assert.Less(t, m.fileMenu.selected, rule)
		assert.True(t, m.fileItemAvailable(items[m.fileMenu.selected]))

		// And landing on it another way still does nothing.
		m.fileMenu.selected = rule
		m, _ = m.chooseFileMenuItem()
		assert.True(t, m.fileMenu.active, "the separator closed the menu")
		assert.False(t, m.preferences.active)
	}
}

// TestChoosingPreferencesOpensTheWindow.
func TestChoosingPreferencesOpensTheWindow(t *testing.T) {
	m := menuModel(t)
	m.fileMenu.selected = menuEntry(t, "PREFERENCES")

	m, _ = m.chooseFileMenuItem()
	assert.False(t, m.fileMenu.active, "the menu stayed open behind the window")
	assert.True(t, m.preferences.active)
	assert.NotEmpty(t, m.preferences.fields)
}

// TestChoosingQuitAsksFirst: the entry is one press from the one above it and
// sits under a button on the tab line, so it raises the question rather than
// answering it.
func TestChoosingQuitAsksFirst(t *testing.T) {
	m := menuModel(t)
	m.fileMenu.selected = menuEntry(t, "QUIT")

	m, cmd := m.chooseFileMenuItem()
	assert.False(t, m.fileMenu.active)
	assert.Nil(t, cmd, "QUIT left without asking")
	require.Equal(t, ModalConfirmQuit, m.modal.Type)

	// On Cancel to begin with, and cancelling stays.
	require.Equal(t, 0, m.modal.Selected)
	cancelled, cmd := m.answerModal(0)
	assert.Nil(t, cmd)
	assert.Equal(t, ModalNone, cancelled.modal.Type)

	// Answering the other way leaves.
	m.modal = NewQuitModal()
	m.modal.NextChoice()
	left, cmd := m.handleModalConfirmation("")
	require.NotNil(t, cmd, "the dialog did not quit")
	assert.IsType(t, tea.QuitMsg{}, cmd())
	assert.Equal(t, ModalNone, left.modal.Type)
}

// TestTheQuitDialogIsAnsweredWhereItIsDrawn: the entry was picked with the
// mouse, so the question has to be answerable with it.
func TestTheQuitDialogIsAnsweredWhereItIsDrawn(t *testing.T) {
	m := menuModel(t)
	m.fileMenu.selected = menuEntry(t, "QUIT")
	m, _ = m.chooseFileMenuItem()
	require.Equal(t, ModalConfirmQuit, m.modal.Type)

	g, ok := m.modal.geometry(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	row := g.y + 1 + g.buttonRow
	starts := m.modal.buttonStarts(g.inner)

	// Every answer is where the dialog draws it.
	for i := range m.modal.Choices {
		got, hit := m.modal.buttonAt(m.windowWidth, m.windowHeight, g.x+3+starts[i], row)
		require.True(t, hit, "nothing on the %s button", m.modal.Choices[i])
		assert.Equal(t, i, got)
	}

	// A press between them is not an answer, and neither is one outside the
	// dialog: a question about quitting is not waved away by a stray click.
	_, hit := m.modal.buttonAt(m.windowWidth, m.windowHeight, g.x+3+starts[1]-2, row)
	assert.False(t, hit)

	stray, cmd := m.handleMousePress(tea.Mouse{X: 0, Y: 0, Button: tea.MouseLeft})
	assert.Nil(t, cmd)
	assert.Equal(t, ModalConfirmQuit, stray.modal.Type, "a click outside answered it")

	// Clicking Quit leaves.
	clicked, cmd := m.handleMousePress(tea.Mouse{
		X: g.x + 3 + starts[1], Y: row, Button: tea.MouseLeft,
	})
	require.NotNil(t, cmd, "clicking Quit did nothing")
	assert.IsType(t, tea.QuitMsg{}, cmd())
	assert.Equal(t, ModalNone, clicked.modal.Type)
}

// TestEveryWayOutGivesTheTerminalBack: the wheel and the buttons are borrowed
// from the terminal, and a shell left with mouse reporting on is not something
// the person who typed EXIT can undo.
func TestEveryWayOutGivesTheTerminalBack(t *testing.T) {
	menu := menuModel(t)
	menu.fileMenu.selected = menuEntry(t, "QUIT")
	menu, _ = menu.chooseFileMenuItem()
	menu.modal.NextChoice() // on to Quit
	_, fromMenu := menu.handleModalConfirmation("")

	typed := helpModel()
	_, fromCommand, handled := typed.handleSpecialCommands("QUIT")
	require.True(t, handled)

	ctrlD := helpModel()
	ctrlD.confirmExit = true
	_, fromCtrlD := ctrlD.handleCtrlD()

	for name, cmd := range map[string]tea.Cmd{
		"the menu": fromMenu, "the command": fromCommand, "Ctrl+D": fromCtrlD,
	} {
		require.NotNil(t, cmd, "%s did not quit", name)
		assert.IsType(t, tea.QuitMsg{}, cmd(), "%s did not quit", name)
	}
}

// TestTheQuitDialogTakesTheKeyboard: the arrows move between the answers and
// Enter gives the one showing, through the ordinary key route rather than by
// calling the dialog's own handler.
func TestTheQuitDialogTakesTheKeyboard(t *testing.T) {
	m := menuModel(t)
	m.fileMenu.selected = menuEntry(t, "QUIT")
	m, _ = m.chooseFileMenuItem()
	require.Equal(t, ModalConfirmQuit, m.modal.Type)

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.Equal(t, 1, m.modal.Selected, "the arrow keys do not reach the dialog")

	_, cmd := m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd, "Enter did not answer the dialog")
	assert.IsType(t, tea.QuitMsg{}, cmd())

	// And Escape is a way out of the question itself.
	m = menuModel(t)
	m.fileMenu.selected = menuEntry(t, "QUIT")
	m, _ = m.chooseFileMenuItem()
	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Equal(t, ModalNone, m.modal.Type)
}

// TestTheMenuIsDrawnWhereItIsClicked, the separators included: a menu that
// counts its rows twice puts a click on the entry above the one pointed at.
func TestTheMenuIsDrawnWhereItIsClicked(t *testing.T) {
	m := menuModel(t)

	layer, ok := m.viewFileMenu(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	lines := strings.Split(layer.Content, "\n")
	require.Equal(t, len(fileMenuItems())+2, len(lines), "the menu is not as tall as its entries")

	for i, item := range fileMenuItems() {
		row := layer.Y + 1 + i
		got, hit := m.fileMenuItemAt(m.windowWidth, m.windowHeight, layer.X+1, row)
		require.True(t, hit, "nothing at row %d", row)
		assert.Equal(t, i, got)

		drawn := stripAnsi(lines[i+1])
		if item.rule {
			assert.Contains(t, drawn, "───", "the separator is not a line")
			continue
		}
		assert.Contains(t, drawn, item.label)
	}
}

// TestSaveResultsIsThereBeforeThereAreResults.
//
// It was dimmed until a query had run - a fair description of the state and a
// poor thing to look at, since a greyed-out entry on a freshly started shell
// reads as one that is not there.
func TestSaveResultsIsThereBeforeThereAreResults(t *testing.T) {
	m := menuModel(t)
	save := fileMenuItems()[menuEntry(t, "SAVE RESULTS")]
	require.Empty(t, m.lastTableData)

	assert.True(t, m.fileItemAvailable(save), "it should be pickable from the start")

	// Picking it says why rather than opening a window that can only be
	// cancelled.
	m.fileMenu.selected = menuEntry(t, "SAVE RESULTS")
	m, _ = m.chooseFileMenuItem()
	assert.False(t, m.capture.active)
	assert.Contains(t, stripAnsiForTest(m.fullHistoryContent), "Execute a query first")

	// And with something on screen it opens the window.
	m = menuModel(t)
	m.lastTableData = [][]string{{"id"}, {"1"}}
	m.fileMenu.selected = menuEntry(t, "SAVE RESULTS")
	m, _ = m.chooseFileMenuItem()
	assert.True(t, m.capture.active)
	assert.Equal(t, saving, m.capture.kind)
}
