package ui

import (
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Working the PREFERENCES window: moving between settings, filling them in, and
// pressing the buttons.

// movePrefFocus moves the cursor through the settings and on to the buttons.
//
// It stops at the ends rather than wrapping, so holding a key down does not
// take you from the last chat setting back to Host.
func (m *MainModel) movePrefFocus(delta int) (*MainModel, tea.Cmd) {
	p := &m.preferences
	if len(p.fields) == 0 {
		return m, nil
	}

	switch p.onButton {
	case onRun:
		if delta > 0 {
			p.onButton = onCancel
			return m, nil
		}
		p.focusField(len(p.fields) - 1)
		return m, nil
	case onCancel:
		if delta < 0 {
			p.onButton = onRun
		}
		return m, nil
	}

	next := p.focus + delta
	switch {
	case next < 0:
		return m, nil
	case next >= len(p.fields):
		p.blurAll()
		p.onButton = onRun
		return m, nil
	}
	p.focusField(next)
	return m, nil
}

func (p *preferences) blurAll() {
	for i := range p.fields {
		p.fields[i].input.Blur()
	}
}

// togglePrefField changes a yes/no setting.
func (m *MainModel) togglePrefField() (*MainModel, tea.Cmd) {
	if field := m.preferences.current(); field != nil && field.spec.kind == prefYesNo {
		m.preferences.fields[m.preferences.focus].yes = !field.yes
	}
	return m, nil
}

// completePrefField fills in as much as the candidates agree on, and lists them
// when that was not enough to pick one.
func (m *MainModel) completePrefField() (*MainModel, tea.Cmd) {
	field := m.preferences.current()
	if field == nil {
		return m, nil
	}

	switch field.spec.kind {
	case prefPath:
		got := completePath(field.input.Value())
		m.setPrefField(m.preferences.focus, got.Completed)
		m.preferences.clearMatches()
		if got.worthListing() {
			m.preferences.matches = got.Matches
		}

	case prefChoice:
		m.completePrefFrom(field.spec.choices())
	}
	return m, nil
}

// completePrefFrom is completion against a fixed set of values.
func (m *MainModel) completePrefFrom(names []string) {
	typed := strings.ToLower(strings.TrimSpace(m.preferences.fields[m.preferences.focus].input.Value()))

	var found []string
	for _, name := range names {
		if strings.HasPrefix(strings.ToLower(name), typed) {
			found = append(found, name)
		}
	}

	m.preferences.clearMatches()
	switch len(found) {
	case 0:
		// Nothing matches what is there. Tab is not where you find out a value
		// is wrong - the hint line says that - so what was typed is left alone
		// and the whole set is offered instead.
		m.preferences.matches = names
	case 1:
		m.setPrefField(m.preferences.focus, found[0])
	default:
		m.setPrefField(m.preferences.focus, commonPrefix(found))
		m.preferences.matches = found
	}
}

// setPrefField puts a value in a setting, with the view of it moved to the end
// so a path longer than the box shows its far end rather than its start.
func (m *MainModel) setPrefField(i int, value string) {
	if i < 0 || i >= len(m.preferences.fields) {
		return
	}
	m.preferences.fields[i].input.SetValue(value)
	m.preferences.fields[i].input.CursorEnd()
}

// movePrefMatch moves the highlight in the candidate list and keeps it showing.
func (m *MainModel) movePrefMatch(delta int) (*MainModel, tea.Cmd) {
	p := &m.preferences
	if len(p.matches) == 0 {
		return m, nil
	}

	p.match = min(max(p.match+delta, 0), len(p.matches)-1)
	p.scrollTop = min(p.scrollTop, p.match)
	p.scrollTop = max(p.scrollTop, p.match-p.matchRows+1)
	p.scrollTop = max(p.scrollTop, 0)
	return m, nil
}

// usePrefMatch puts the highlighted candidate in the setting.
func (m *MainModel) usePrefMatch() (*MainModel, tea.Cmd) {
	p := &m.preferences
	if p.match < 0 || p.match >= len(p.matches) {
		return m, nil
	}

	field := p.current()
	if field == nil {
		return m, nil
	}

	value := p.matches[p.match]
	if field.spec.kind == prefPath {
		// The candidates are names inside a directory, so the directory typed
		// so far stays in front of the one picked - except the way up, which
		// replaces it.
		dir, _ := splitPath(field.input.Value())
		if value == parentEntry {
			value = parentDir(dir)
		} else {
			value = dir + value
		}
	}

	m.setPrefField(p.focus, value)
	p.clearMatches()

	// Stepping into a directory, or back out of one, shows what is in it
	// rather than leaving another Tab to find out.
	if field.spec.kind == prefPath && strings.HasSuffix(value, string(filepath.Separator)) {
		return m.completePrefField()
	}
	return m, nil
}

// scrollPrefMatches moves the candidate list under the wheel, bringing the
// highlight with it rather than leaving it off-screen.
func (m *MainModel) scrollPrefMatches(delta int) (*MainModel, tea.Cmd) {
	p := &m.preferences
	if len(p.matches) <= p.matchRows {
		return m, nil
	}

	p.scrollTop = min(max(p.scrollTop+delta, 0), len(p.matches)-p.matchRows)
	p.match = min(max(p.match, p.scrollTop), p.scrollTop+p.matchRows-1)
	return m, nil
}

// scrollPreferences moves the settings list under the wheel.
//
// The cursor stays where it was: this is looking around rather than moving,
// and typing carries on going where it was going.
func (m *MainModel) scrollPreferences(delta int) (*MainModel, tea.Cmd) {
	p := &m.preferences
	lines := len(p.lines())
	if lines <= p.rows {
		return m, nil
	}

	p.scroll = min(max(p.scroll+delta, 0), lines-p.rows)
	return m, nil
}

// handlePreferencesKey is every key press while the window is open.
func (m *MainModel) handlePreferencesKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	// The candidate list takes the keys while it is showing: it is what was
	// just asked for, and Esc goes back to the setting rather than closing the
	// window out from under you.
	if len(m.preferences.matches) > 0 {
		switch msg.String() {
		case "esc":
			m.preferences.clearMatches()
			return m, nil
		case "up":
			return m.movePrefMatch(-1)
		case "down":
			return m.movePrefMatch(1)
		case "enter":
			return m.usePrefMatch()
		}
	}

	switch msg.String() {
	case "esc":
		m.closePreferences()
		return m, nil
	case "up":
		return m.movePrefFocus(-1)
	case "down":
		return m.movePrefFocus(1)
	case "pgup":
		return m.movePrefFocus(-m.preferences.rows)
	case "pgdown":
		return m.movePrefFocus(m.preferences.rows)
	case "left":
		if m.preferences.onButton != onNoButton {
			return m.movePrefFocus(-1)
		}
	case "right":
		if m.preferences.onButton != onNoButton {
			return m.movePrefFocus(1)
		}
	case "tab":
		if field := m.preferences.current(); field != nil && field.spec.kind == prefYesNo {
			return m.togglePrefField()
		}
		return m.completePrefField()
	case "shift+tab":
		return m.movePrefFocus(-1)
	case " ", "space":
		if field := m.preferences.current(); field != nil && field.spec.kind == prefYesNo {
			return m.togglePrefField()
		}
	case "enter":
		// Only from a button. Enter is what you press to finish typing, and
		// writing the file by reflex on the way past a setting is not what it
		// should mean.
		switch m.preferences.onButton {
		case onRun:
			return m.savePreferences()
		case onCancel:
			m.closePreferences()
			return m, nil
		}
		return m.movePrefFocus(1)
	}

	// Anything else is typing.
	if field := m.preferences.current(); field != nil && field.spec.kind != prefYesNo {
		var cmd tea.Cmd
		m.preferences.fields[m.preferences.focus].input, cmd = field.input.Update(msg)
		m.preferences.clearMatches()
		return m, cmd
	}
	return m, nil
}

// handlePreferencesClick is a press inside the window.
func (m *MainModel) handlePreferencesClick(col, row int) (*MainModel, tea.Cmd) {
	if button, ok := m.prefButtonAt(m.windowWidth, m.windowHeight, col, row); ok {
		m.preferences.blurAll()
		switch button {
		case "save":
			m.preferences.onButton = onRun
			return m.savePreferences()
		case "cancel":
			m.closePreferences()
			return m, nil
		}
	}

	if i, ok := m.prefMatchAt(m.windowWidth, m.windowHeight, col, row); ok {
		m.preferences.match = i
		return m.usePrefMatch()
	}

	if i, ok := m.prefFieldAt(m.windowWidth, m.windowHeight, col, row); ok {
		m.preferences.focusField(i)
		if m.preferences.fields[i].spec.kind == prefYesNo {
			return m.togglePrefField()
		}
		return m, nil
	}
	return m, nil
}
