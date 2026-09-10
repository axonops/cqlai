package ui

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// formModel is an open form on a terminal with room for it, with the schema
// already fetched so the fields have something to check against.
func formModel(t *testing.T, action formAction) *MainModel {
	t.Helper()

	m := helpModel()
	m.windowWidth = 120
	m.windowHeight = 40
	m.openFileForm(action)
	m.form.keyspaces = []string{"my_keyspace", "other"}
	m.form.tables["my_keyspace"] = []string{"users", "events"}
	m.form.columns = []string{"id", "name", "created_at"}
	return m
}

// set fills a field in by label.
func set(t *testing.T, m *MainModel, label, value string) {
	t.Helper()

	for i := range m.form.fields {
		if m.form.fields[i].label == label {
			m.form.fields[i].set(value)
			return
		}
	}
	t.Fatalf("no field called %q", label)
}

// focus puts the cursor on a field by label.
func focus(t *testing.T, m *MainModel, label string) {
	t.Helper()

	for i := range m.form.fields {
		if m.form.fields[i].label == label {
			m.form.focusField(i)
			return
		}
	}
	t.Fatalf("no field called %q", label)
}

// TestEachFormBuildsItsCommand, which is the command typing it would run.
func TestEachFormBuildsItsCommand(t *testing.T) {
	source := formModel(t, sourcing)
	set(t, source, "File", "/tmp/schema.cql")
	assert.Equal(t, "SOURCE '/tmp/schema.cql'", source.form.command())

	to := formModel(t, copyingTo)
	set(t, to, "Keyspace", "my_keyspace")
	set(t, to, "Table", "users")
	set(t, to, "File", "/tmp/users.parquet")
	assert.Equal(t, "COPY my_keyspace.users TO '/tmp/users.parquet'", to.form.command())

	set(t, to, "Columns", "id, name")
	set(t, to, "COMPRESSION", "gzip")
	assert.Equal(t,
		"COPY my_keyspace.users (id, name) TO '/tmp/users.parquet' WITH COMPRESSION='gzip'",
		to.form.command())

	from := formModel(t, copyingFrom)
	set(t, from, "Keyspace", "my_keyspace")
	set(t, from, "Table", "users")
	set(t, from, "File", "/tmp/users.csv")
	set(t, from, "SKIPROWS", "2")
	assert.Equal(t,
		"COPY my_keyspace.users FROM '/tmp/users.csv' WITH HEADER=true AND SKIPROWS=2",
		from.form.command())
}

// TestEnterDoesNotRunFromAField.
//
// COPY FROM writes to a table and COPY TO writes over a file, and both take a
// while to find out they were pointed somewhere wrong. Enter is what you press
// to finish typing; it should not also be what does that.
func TestEnterDoesNotRunFromAField(t *testing.T) {
	m := formModel(t, copyingTo)
	set(t, m, "Keyspace", "my_keyspace")
	set(t, m, "Table", "users")
	set(t, m, "File", "/tmp/users.csv")
	focus(t, m, "File")
	require.True(t, m.formReady(), "the form has everything it needs")

	m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.True(t, m.form.active, "the form should still be open")
	assert.Equal(t, onNoButton, m.form.onButton)
}

// TestTheButtonsAreReachedByWalkingPastTheFields, and Enter presses the one you
// are on.
func TestTheButtonsAreReachedByWalkingPastTheFields(t *testing.T) {
	m := formModel(t, copyingTo)
	set(t, m, "Keyspace", "my_keyspace")
	set(t, m, "Table", "users")
	set(t, m, "File", "/tmp/users.csv")

	// From the first field, one step per field lands on the button after the
	// last of them.
	m.form.focusField(0)
	for range len(m.form.fields) {
		m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	require.Equal(t, onRun, m.form.onButton, "past the last field is the button")
	assert.Nil(t, m.form.current(), "and no field has the cursor")

	m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, onCancel, m.form.onButton)

	m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.False(t, m.form.active, "Cancel closes it")
}

// TestAWrongFieldIsMarkedAndExplained.
func TestAWrongFieldIsMarkedAndExplained(t *testing.T) {
	m := formModel(t, copyingTo)
	set(t, m, "Keyspace", "my_keyspace")
	set(t, m, "Table", "nosuch")
	set(t, m, "File", "/tmp/users.csv")
	set(t, m, "PAGESIZE", "lots")

	assert.False(t, m.formReady())

	drawn := stripAnsiForTest(mustRender(t, m))
	assert.Contains(t, drawn, "! Table", "marked even when you are not on it")
	assert.Contains(t, drawn, "! PAGESIZE")

	focus(t, m, "PAGESIZE")
	assert.Contains(t, stripAnsiForTest(mustRender(t, m)), "PAGESIZE: a whole number")
}

// TestWhatIsChecked, field by field.
func TestWhatIsChecked(t *testing.T) {
	m := formModel(t, copyingFrom)
	set(t, m, "Keyspace", "my_keyspace")
	set(t, m, "Table", "users")
	set(t, m, "File", "/tmp/users.csv")
	require.True(t, m.formReady())

	for _, tt := range []struct{ field, value, wrong string }{
		{"Keyspace", "", "a keyspace is needed"},
		{"Keyspace", "nosuch", "no keyspace of that name"},
		{"Table", "nosuch", "no table of that name in this keyspace"},
		{"File", pathRoot, "a file is needed"},
		{"File", "/tmp/", "that is a directory"},
		{"SKIPROWS", "many", "a whole number"},
		{"DELIMITER", ";;", "one character"},
		{"ENCODING", "klingon", "one of utf8"},
	} {
		m := formModel(t, copyingFrom)
		set(t, m, "Keyspace", "my_keyspace")
		set(t, m, "Table", "users")
		set(t, m, "File", "/tmp/users.csv")
		set(t, m, tt.field, tt.value)

		assert.False(t, m.formReady(), "%s=%q should stop it", tt.field, tt.value)
		assert.Contains(t, collect(m.formErrors()), tt.wrong)
	}
}

// TestAnOptionLeftAloneIsNotWrong: empty means unanswered, not invalid.
func TestAnOptionLeftAloneIsNotWrong(t *testing.T) {
	m := formModel(t, copyingFrom)
	set(t, m, "Keyspace", "my_keyspace")
	set(t, m, "Table", "users")
	set(t, m, "File", "/tmp/users.csv")

	assert.True(t, m.formReady())
	assert.Empty(t, m.formErrors())
	assert.NotContains(t, m.form.command(), "SKIPROWS", "an empty option is not in the command")
}

// TestTheOptionsAreTheOnesTheDirectionReads.
//
// CHUNKSIZE means something to an import and nothing to the CSV export beside
// it. An option you can set and then watch do nothing is worse than one that is
// not offered.
func TestTheOptionsAreTheOnesTheDirectionReads(t *testing.T) {
	to := labelsOf(formModel(t, copyingTo))
	from := labelsOf(formModel(t, copyingFrom))

	assert.NotContains(t, to, "CHUNKSIZE", "an export does not read it")
	assert.Contains(t, from, "CHUNKSIZE")

	assert.Contains(t, to, "COMPRESSION", "which only a Parquet write uses")
	assert.NotContains(t, from, "COMPRESSION")

	assert.NotContains(t, from, "HEADER", "the import form has a field of its own for it")
	assert.Contains(t, from, "Header row")
	assert.Contains(t, to, "HEADER")

	assert.Empty(t, optionsOf(formModel(t, sourcing)), "SOURCE takes none")
}

// TestTheDefaultsComeFromTheCommand rather than from a second list of numbers
// that has to be kept in step with it.
func TestTheDefaultsComeFromTheCommand(t *testing.T) {
	m := formModel(t, copyingFrom)
	defaults := router.CopyOptionDefaults()

	for _, i := range optionsOf(m) {
		field := m.form.fields[i]
		if want := defaults[field.label]; want != "" {
			assert.Equal(t, want, field.input.Placeholder,
				"%s shows what leaving it alone does", field.label)
		}
	}

	assert.NotEmpty(t, defaults["CHUNKSIZE"], "the command has a default for it")
}

// TestChangingTheKeyspaceClearsWhatIsUnderIt.
func TestChangingTheKeyspaceClearsWhatIsUnderIt(t *testing.T) {
	m := formModel(t, copyingTo)
	set(t, m, "Keyspace", "my_keyspace")
	set(t, m, "Table", "users")
	set(t, m, "Columns", "id, name")
	m.form.tableKeyspace, m.form.columnsTable = "my_keyspace", "users"

	focus(t, m, "Keyspace")
	m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyBackspace})

	assert.Empty(t, m.form.field("Table"), "a table belongs to a keyspace")
	assert.Empty(t, m.form.field("Columns"), "and columns to a table")
}

// TestTabListsAnOptionsValues, for the options that have a set of them.
func TestTabListsAnOptionsValues(t *testing.T) {
	m := formModel(t, copyingTo)

	focus(t, m, "COMPRESSION")
	m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, "COMPRESSION", m.form.awaiting)
	assert.Equal(t, []string{"snappy", "gzip", "lz4", "zstd", "none"}, m.form.matches)

	m.form.match = 1
	m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, "gzip", m.form.field("COMPRESSION"))
	assert.Empty(t, m.form.awaiting, "and the list is done with")

	// One that takes anything shows nothing rather than an empty list.
	focus(t, m, "PAGESIZE")
	m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Empty(t, m.form.matches)
}

// TestTheFormFitsEveryTerminalWorthDrawingOn.
//
// COPY FROM has thirteen options. In one column with no scrolling the window
// wanted twenty-nine rows and would not open on anything smaller.
func TestTheFormFitsEveryTerminalWorthDrawingOn(t *testing.T) {
	for _, action := range []formAction{sourcing, copyingTo, copyingFrom} {
		for _, height := range []int{20, 24, 30, 40} {
			m := helpModel()
			m.windowWidth = 120
			m.windowHeight = height
			m.openFileForm(action)

			layer, ok := m.viewFileForm(m.windowWidth, m.windowHeight)
			require.True(t, ok, "action %d at %d rows", action, height)
			assert.LessOrEqual(t, strings.Count(layer.Content, "\n")+1, height,
				"action %d at %d rows", action, height)
		}
	}
}

// TestALongValueIsNotDrawnBlank.
//
// SetValue leaves the textinput's view window past the end of the text, so a
// path longer than the box renders as nothing at all. Every write goes through
// set, which puts the cursor at the end.
func TestALongValueIsNotDrawnBlank(t *testing.T) {
	m := formModel(t, copyingTo)
	set(t, m, "File", "/var/lib/cassandra/exports/users.parquet")

	drawn := stripAnsiForTest(mustRender(t, m))
	assert.Contains(t, drawn, "users.parquet")
}

// mustRender draws the form.
func mustRender(t *testing.T, m *MainModel) string {
	t.Helper()

	layer, ok := m.viewFileForm(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	return layer.Content
}

// labelsOf is every field label on a form.
func labelsOf(m *MainModel) []string {
	labels := make([]string, 0, len(m.form.fields))
	for _, field := range m.form.fields {
		labels = append(labels, field.label)
	}
	return labels
}

// optionsOf is the indices of the option fields.
func optionsOf(m *MainModel) []int {
	var options []int
	for i, field := range m.form.fields {
		if field.kind == fieldOption {
			options = append(options, i)
		}
	}
	return options
}

// collect is the error messages, for asserting one is among them.
func collect(errors map[int]string) []string {
	var all []string
	for _, err := range errors {
		all = append(all, err)
	}
	slices.Sort(all)
	return all
}

// TestEveryFieldCanBeFocused.
//
// The Header row field was declared without a textinput, and a zero textinput
// has no cursor: focusing one dereferences it. Clicking that field crashed
// cqlai, and the panic came from three layers down in bubbles, which is not
// where anyone would look.
//
// Every field, on every form, by click and by key - because the one that was
// wrong was the one field that never draws its input, and nothing else touched
// it.
func TestEveryFieldCanBeFocused(t *testing.T) {
	for _, action := range []formAction{sourcing, copyingTo, copyingFrom} {
		m := formModel(t, action)

		for i := range m.form.fields {
			label := m.form.fields[i].label
			require.NotPanics(t, func() { m.form.focusField(i) },
				"focusing %s should not panic", label)
			require.NotPanics(t, func() { m.form.blurAll() },
				"blurring past %s should not panic", label)
		}

		// And by walking with the arrows, which is how you reach it without a
		// mouse.
		require.NotPanics(t, func() {
			m.form.focusField(0)
			for range len(m.form.fields) + 2 {
				m.handleFormKey(tea.KeyPressMsg{Code: tea.KeyDown})
			}
		}, "walking the whole form should not panic")
	}
}

// TestEveryFieldHasAnInput, which is what makes the above true rather than
// happening to be true.
func TestEveryFieldHasAnInput(t *testing.T) {
	for _, action := range []formAction{sourcing, copyingTo, copyingFrom} {
		m := formModel(t, action)
		for i, field := range m.form.fields {
			assert.NotPanics(t, func() { m.form.fields[i].input.Focus() },
				"%s has no usable input", field.label)
		}
	}
}
