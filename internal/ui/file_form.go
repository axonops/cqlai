package ui

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/ui/completion"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// The forms behind SOURCE, COPY TO and COPY FROM.
//
// The window Save and Capture share asks one question at a time - pick a
// format, then say where - which works because those two commands are a format
// and a path and nothing else.
//
// COPY is not. It needs a table as well as a path, and whether a CSV has a
// header row is the difference between importing your data and importing your
// column names as a row. Asked one at a time that is a wizard, and you cannot
// see what you already answered. So these show their fields together.

// buttonFocus is which button the cursor is on, if any.
type buttonFocus int

const (
	onNoButton buttonFocus = iota
	onRun
	onCancel
)

// formAction is which command a form builds.
type formAction int

const (
	sourcing formAction = iota
	copyingTo
	copyingFrom
)

// formFieldKind decides what Tab does in a field.
type formFieldKind int

const (
	fieldText     formFieldKind = iota
	fieldPath                   // completes against the filesystem
	fieldKeyspace               // completes against the cluster
	fieldTable                  // completes against the keyspace named above
	fieldColumns                // ticked off a list, or typed comma separated
	fieldOption                 // one of the WITH options, with its value
	fieldYesNo                  // Tab and Space turn it over
)

// formField is one labelled input.
type formField struct {
	label    string
	kind     formFieldKind
	input    textinput.Model
	yes      bool // for fieldYesNo
	hint     string
	optional bool // the form runs without it
}

// fileForm is the open form, if any.
type fileForm struct {
	active bool
	action formAction
	fields []formField
	focus  int

	// The schema, read from the cluster rather than typed. Keyspaces are
	// fetched when the form opens; a keyspace's tables the first time you
	// complete them, because which keyspace that is depends on the field above
	// and can change while the form is open. Both are queries, and the answers
	// do not change while you type.
	keyspaces []string
	tables    map[string][]string

	// What the fields below were chosen under. Each field is only meaningful
	// for the one above it: a table belongs to a keyspace and columns belong to
	// a table, so when one changes the ones under it are no longer answers to
	// anything. Left behind they build a command naming a table that is not
	// there - or worse, one that is.
	tableKeyspace string
	columnsTable  string

	// columns are the named table's columns, fetched the first time the list is
	// opened. ticked is what the list is showing as chosen, so a click knows
	// what it is turning off.
	columns []string
	ticked  map[string]bool

	// awaiting is the option whose value is being chosen, so the list knows
	// whether it is showing names or the values of one of them. An option
	// picked from a list should not then leave you guessing what goes after the
	// "=".
	awaiting string

	// matches are the candidates from the last Tab, shown as a list to pick
	// from. Joined onto one line they run off the edge with no way to reach the
	// rest, which is the mistake the capture window already made once.
	matches     []string
	match       int
	matchScroll int

	// onButton is which button the cursor is on once it has walked past the last
	// field: none, the one that acts, or Cancel.
	//
	// The buttons are part of the walk rather than Enter running the command
	// from wherever you happen to be. COPY FROM writes to a table and COPY TO
	// writes over a file, and a key you press to finish typing should not also
	// be the one that does that.
	onButton buttonFocus

	// rows is how many candidates the list shows, worked out from the terminal
	// when the form opens. See matchRows.
	rows int

	// How many options are on screen at once, and which is at the top. A short
	// terminal shows a window onto the list rather than refusing to draw the
	// form at all.
	optionRows   int
	optionScroll int
}

// paneDivider separates the columns. The form's own fields are on the left and
// the options beside them rather than under them: the options are a set of
// adjustments to what the left-hand side says, and running them together made a
// list you had to scroll past to reach the button.
const paneDivider = " │ "

// panes splits the fields into columns, by index: the form's own fields first,
// then the options in as many columns as they need.
func (f fileForm) panes() [][]int {
	var own, options []int
	for i, field := range f.fields {
		if field.kind == fieldOption {
			options = append(options, i)
			continue
		}
		own = append(own, i)
	}

	panes := [][]int{own}
	if len(options) == 0 {
		return panes
	}

	// One column, scrolled when the terminal is too short for all of it.
	// COPY FROM has thirteen options, and a window tall enough for every one of
	// them would not open on a twenty-four row terminal at all.
	if f.optionRows > 0 && len(options) > f.optionRows {
		first := min(f.optionScroll, len(options)-f.optionRows)
		options = options[first : first+f.optionRows]
	}
	return append(panes, options)
}

// optionWindow is the range of options showing, for the scrollbar beside them.
func (f fileForm) optionWindow() (first, total int) {
	total = 0
	for _, field := range f.fields {
		if field.kind == fieldOption {
			total++
		}
	}
	if f.optionRows <= 0 || total <= f.optionRows {
		return 0, total
	}
	return min(f.optionScroll, total-f.optionRows), total
}

// paneStarts is the column each pane begins at, inside the window's padding.
func (f fileForm) paneStarts(panes [][]int) []int {
	starts := make([]int, len(panes))
	at := 0
	for i, pane := range panes {
		starts[i] = at
		at += f.paneWidth(pane) + lipgloss.Width(paneDivider)
	}
	return starts
}

// paneLabelWidth is the label column of one pane, so its values line up.
func (f fileForm) paneLabelWidth(pane []int) int {
	labels := 0
	for _, i := range pane {
		labels = max(labels, lipgloss.Width(f.fields[i].label))
	}
	return labels
}

// paneWidth is how wide a column has to be for its fields.
func (f fileForm) paneWidth(pane []int) int {
	labels := f.paneLabelWidth(pane)

	width := 0
	for _, i := range pane {
		value := max(f.fields[i].input.Width(), lipgloss.Width(f.fields[i].display()))
		width = max(width, 2+labels+2+value)
	}
	return width
}

// fieldRow draws one field of a column, or nothing where that column has run
// out of fields.
func (m *MainModel) fieldRow(pane []int, n int) string {
	if n >= len(pane) {
		return ""
	}

	i := pane[n]
	field := m.form.fields[i]
	labels := m.form.paneLabelWidth(pane)

	marker := "  "
	switch {
	case i == m.form.focus && m.form.onButton == onNoButton:
		marker = "> "
	case m.fieldWrong(i):
		// Marked even when you are not on it, so a form that will not run says
		// where rather than only that.
		marker = "! "
	}

	return marker + pad(field.label, labels) + "  " + field.display()
}

// focusedHint is the line under the fields saying what the one you are on is
// for.
//
// On its own line rather than beside the field. Beside it, the longest of them
// - "csv, json or parquet, rather than from the extension" - set the width of
// the whole window, which then stayed that wide for every other field.
func (m *MainModel) focusedHint() string {
	field := m.form.current()
	if field == nil {
		return ""
	}

	// What is wrong with it comes before what it is for: you already know what
	// you were trying to put there.
	if m.fieldWrong(m.form.focus) {
		return field.label + ": " + m.fieldError(m.form.focus)
	}
	if field.hint != "" {
		return field.label + ": " + field.hint
	}
	return ""
}

// The rows the candidate list gets. They are reserved whether or not there is a
// list: a window that changes size every time you press Tab is one you have to
// find again afterwards, and the rows are cheaper than that.
//
// How many depends on the terminal, because the options are a single column and
// COPY FROM has thirteen of them. Worked out once for a given screen, so the
// window is still one size for as long as it is open.
const (
	formMatchRowsMax = 6
	formMatchRowsMin = 3
)

// formChrome is every row of the window that is not a field or a candidate: the
// title, what it does and a blank; a blank above the candidates and one below;
// the buttons; the keys; and the border top and bottom.
const formChrome = 3 + 1 + 1 + 1 + 1 + 1 + 1 + 2

// fitHeights works out how many option rows and how many candidate rows this
// screen has room for.
//
// Both are settled when the form opens, so the window is one size for as long
// as it is open. The candidates get what they can up to six; the options take
// the rest, and scroll when there are more of them than that.
func (m *MainModel) fitHeights(screenHeight int) (optionRows, matchRows int) {
	own, options := 0, 0
	for _, field := range m.form.fields {
		if field.kind == fieldOption {
			options++
			continue
		}
		own++
	}

	matchRows = formMatchRowsMax
	room := screenHeight - 1 - formChrome - matchRows

	// The form's own fields are beside the options, so the taller column
	// decides. Below the minimum the candidates give up their rows first: an
	// option you cannot see is worse than a shorter list of candidates.
	for matchRows > formMatchRowsMin && room < max(own, min(options, own)) {
		matchRows--
		room++
	}

	optionRows = max(room, 1)
	if options <= optionRows {
		optionRows = options
	}
	return optionRows, matchRows
}

// formKeyLines is every line of keys the window can show. The widest decides
// the window's width, so pressing Tab does not widen it.
var formKeyLines = []string{
	"Tab: Complete   ↑↓: Field   Esc: Close",
	"Space: Change   ↑↓: Field   Esc: Close",
	"↑↓: Field   Esc: Close",
	"↑↓: Move   Enter: Press   Esc: Close",
	"↑↓/wheel: Move   Enter: Use   Esc: Back",
	"↑↓/wheel: Move   Space: Tick   Esc: Done",
	"Tab: List   ↑↓: Field   Esc: Close",
	"Tab: Values   ↑↓: Field   Esc: Close",
}

// title is the line at the top of the window.
func (f fileForm) title() string {
	switch f.action {
	case copyingTo:
		return "Copy a table to a file"
	case copyingFrom:
		return "Copy a file into a table"
	}
	return "Run CQL from a file"
}

// runLabel is the word on the button that does it.
func (f fileForm) runLabel() string {
	if f.action == sourcing {
		return "Run"
	}
	return "Copy"
}

// The buttons at the foot of the window. Every other way to work cqlai can be
// clicked, and a form you can only finish from the keyboard is the odd one out.
const (
	formButtonGap = "   "
	formCancel    = "Cancel"
)

// describe says what the action does, since the title alone does not.
func (f fileForm) describe() string {
	switch f.action {
	case copyingTo:
		return "CSV or Parquet, from the file extension."
	case copyingFrom:
		return "CSV or Parquet, from the file extension."
	}
	return "The statements run as if you had typed them."
}

// pathRoot is where the File field starts.
//
// Completion used to begin wherever cqlai was started from, which is rarely
// where the file is and is not somewhere you can see: the field looked empty
// and Tab produced a directory listing you had no reason to expect. From the
// root the field says where it is looking, and every path from there is one you
// could have typed.
const pathRoot = "/"

// formPlaceholderColour is the grey an empty field's instruction is drawn in:
// dim enough to read as a prompt rather than as something you typed, and the
// same grey the dimmed tabs and menu entries use.
const formPlaceholderColour = "#585858"

// newOptionInput is an option's value box, narrower than the form's own fields:
// what goes in one is a word, a number or a single character, and the pane is
// beside the form rather than under it.
func newOptionInput(placeholder string) textinput.Model {
	in := newFormInput("", placeholder)
	in.SetWidth(12)
	return in
}

// newFormInput is an input sized for the form.
//
// The placeholder says what the field takes and, where there is completion,
// that Tab does it. An empty box with a label beside it does not say whether it
// wants a name, a path, or something you have to know already.
func newFormInput(value, placeholder string) textinput.Model {
	in := textinput.New()
	in.Prompt = ""
	in.CharLimit = 512
	// Wide enough for a path worth reading. The options beside it need far
	// less, so they have their own narrower box rather than both being sized
	// for whichever needs more.
	in.SetWidth(34)
	in.Placeholder = placeholder

	grey := lipgloss.NewStyle().Foreground(lipgloss.Color(formPlaceholderColour))
	styles := in.Styles()
	styles.Focused.Placeholder = grey
	styles.Blurred.Placeholder = grey
	in.SetStyles(styles)

	in.SetValue(value)
	in.CursorEnd()
	return in
}

// formFields is what an action asks for.
//
// COPY's other options are not here. They are things you set once and rarely -
// MAXBATCHSIZE, MAXPARSEERRORS, CHUNKSIZE - and a form with eighteen fields is
// a worse way to reach them than typing the command, which still works.
func formFields(action formAction, keyspace string) []formField {
	return append(actionFields(action, keyspace), optionFields(action)...)
}

// optionFields is a row for every option the direction reads.
//
// All of them, rather than a row that adds one at a time: there is room for
// them beside the form's own fields, and a list you can see is a list you can
// choose from. An option left empty is one that was not set.
func optionFields(action formAction) []formField {
	if action == sourcing {
		return nil
	}

	names := fileForm{action: action}.copyOptionsFor()
	defaults := router.CopyOptionDefaults()
	fields := make([]formField, 0, len(names))
	for _, name := range names {
		// An option with a list says how to reach it; one that takes anything
		// says so instead. Telling someone to press Tab for a value that is a
		// number they have to know is an instruction that does nothing.
		// What it is when you leave it alone, from the command that reads it.
		// An empty box says nothing about whether the option does anything by
		// default, and for most of these it does: CHUNKSIZE is five thousand
		// rows whether or not you ever type a number.
		// What it is when you leave it alone, from the command that reads it. An
		// empty box says nothing about whether the option does anything by
		// default, and for most of these it does: CHUNKSIZE is five thousand
		// rows whether or not you ever type a number.
		//
		// Where there is no default, an option with a set of values says how to
		// reach them instead. FORMAT has no default worth printing - it is read
		// from the file extension - and knowing that Tab lists csv, json and
		// parquet is the more useful thing to be told.
		placeholder := defaults[name]
		switch {
		case placeholder != "":
		case len(copyOptionValues[name]) > 0:
			placeholder = "Tab for list"
		default:
			placeholder = "not set"
		}

		fields = append(fields, formField{
			label:    name,
			kind:     fieldOption,
			input:    newOptionInput(placeholder),
			hint:     copyOptionHelp[name],
			optional: true,
		})
	}
	return fields
}

func actionFields(action formAction, keyspace string) []formField {
	switch action {
	case copyingTo:
		return []formField{
			{label: "Keyspace", kind: fieldKeyspace, input: newFormInput(keyspace, "Tab to list keyspaces")},
			{label: "Table", kind: fieldTable, input: newFormInput("", "Tab to list tables")},
			{label: "File", kind: fieldPath, input: newFormInput(pathRoot, ""), hint: "Tab completes"},
			{label: "Columns", kind: fieldColumns, input: newFormInput("", "all of them"), hint: "or tick them off", optional: true},
		}
	case copyingFrom:
		return []formField{
			{label: "Keyspace", kind: fieldKeyspace, input: newFormInput(keyspace, "Tab to list keyspaces")},
			{label: "Table", kind: fieldTable, input: newFormInput("", "Tab to list tables")},
			{label: "File", kind: fieldPath, input: newFormInput(pathRoot, ""), hint: "Tab completes"},
			// On the FROM form and not the TO form because this is where
			// getting it wrong costs you: a CSV whose first row is column
			// names, read with HEADER=false, inserts that row as data.
			{label: "Header row", kind: fieldYesNo, yes: true, hint: "the file's first row is column names"},
		}
	}
	return []formField{
		{label: "File", kind: fieldPath, input: newFormInput(pathRoot, ""), hint: "Tab completes"},
	}
}

// openFileForm opens the form for an action.
func (m *MainModel) openFileForm(action formAction) (*MainModel, tea.Cmd) {
	m.form = fileForm{
		active:    true,
		action:    action,
		fields:    formFields(action, m.currentKeyspace()),
		keyspaces: m.keyspaceChoices(),
		tables:    map[string][]string{},
	}
	m.form.tableKeyspace = m.form.field("Keyspace")
	m.form.columnsTable = m.form.field("Table")
	m.form.optionRows, m.form.rows = m.fitHeights(m.windowHeight)

	// Open on the first field with nothing in it. With a keyspace already set
	// that is the table, which is the first thing you have to decide.
	m.form.focusField(0)
	for i, field := range m.form.fields {
		if field.value() == "" {
			m.form.focusField(i)
			break
		}
	}
	return m, nil
}

// closeFileForm dismisses it without running anything.
func (m *MainModel) closeFileForm() {
	m.form = fileForm{}
}

// focusField moves the cursor to a field, and only that one has it.
func (f *fileForm) focusField(i int) {
	if i < 0 || i >= len(f.fields) {
		return
	}
	f.focus = i
	f.onButton = onNoButton
	f.clearMatches()
	f.showFocusedOption()
	for j := range f.fields {
		if j == i {
			f.fields[j].input.Focus()
		} else {
			f.fields[j].input.Blur()
		}
	}
}

// showFocusedOption scrolls the option column so the focused option is on it.
//
// The column shows a window onto the options when the terminal is short, and
// moving to one outside that window would otherwise take the cursor somewhere
// there is nothing drawn.
func (f *fileForm) showFocusedOption() {
	if f.optionRows <= 0 || f.focus >= len(f.fields) || f.fields[f.focus].kind != fieldOption {
		return
	}

	at, total := 0, 0
	for i, field := range f.fields {
		if field.kind != fieldOption {
			continue
		}
		if i == f.focus {
			at = total
		}
		total++
	}
	if total <= f.optionRows {
		return
	}

	switch {
	case at < f.optionScroll:
		f.optionScroll = at
	case at >= f.optionScroll+f.optionRows:
		f.optionScroll = at - f.optionRows + 1
	}
	f.optionScroll = min(max(f.optionScroll, 0), total-f.optionRows)
}

// clearMatches takes the candidate list down.
func (f *fileForm) clearMatches() {
	f.matches = nil
	f.match = 0
	f.matchScroll = 0
}

// abandonMatches takes the list down and forgets what it was for, which
// clearMatches deliberately does not: picking an option keeps the list up to
// ask for its value.
func (f *fileForm) abandonMatches() {
	f.clearMatches()
	f.awaiting = ""
}

// moveFocus steps between fields, stopping at the ends rather than wrapping:
// a form is a thing you fill in from the top, not a carousel.
func (m *MainModel) moveFocus(delta int) (*MainModel, tea.Cmd) {
	// The fields, then the two buttons: one walk through everything on the
	// window, so the way to reach the button is the way you reach anything.
	at := m.form.focus + int(m.form.onButton)*0
	switch m.form.onButton {
	case onRun:
		at = len(m.form.fields)
	case onCancel:
		at = len(m.form.fields) + 1
	}

	at = min(max(at+delta, 0), len(m.form.fields)+1)
	switch {
	case at < len(m.form.fields):
		m.form.onButton = onNoButton
		m.form.focusField(at)
	case at == len(m.form.fields):
		m.form.onButton = onRun
		m.form.blurAll()
	default:
		m.form.onButton = onCancel
		m.form.blurAll()
	}
	return m, nil
}

// blurAll takes the cursor out of every field, for when it is on a button.
func (f *fileForm) blurAll() {
	f.clearMatches()
	for i := range f.fields {
		f.fields[i].input.Blur()
	}
}

// current is the field with the cursor.
func (f *fileForm) current() *formField {
	if f.onButton != onNoButton || f.focus < 0 || f.focus >= len(f.fields) {
		return nil
	}
	return &f.fields[f.focus]
}

// field finds a field by its label, so the command does not depend on the order
// they happen to be drawn in.
func (f fileForm) field(label string) string {
	for _, field := range f.fields {
		if field.label == label {
			return field.value()
		}
	}
	return ""
}

// qualifiedTable is the table as the command names it, with its keyspace where
// one is given.
func (f fileForm) qualifiedTable() string {
	table := f.field("Table")
	if keyspace := f.field("Keyspace"); keyspace != "" {
		return keyspace + "." + table
	}
	return table
}

// set puts a value in a field.
//
// Always with the cursor at the end, because the textinput works out which part
// of a long value it is showing from where the cursor is - and SetValue on its
// own leaves that window past the end of the text, so a path longer than the
// box renders as nothing at all.
func (f *formField) set(value string) {
	f.input.SetValue(value)
	f.input.CursorEnd()
}

// value is what a field holds, as the command wants it.
func (f formField) value() string {
	if f.kind == fieldYesNo {
		if f.yes {
			return "true"
		}
		return "false"
	}
	return strings.TrimSpace(f.input.Value())
}

// display is what a field shows.
func (f formField) display() string {
	if f.kind == fieldYesNo {
		if f.yes {
			return "yes"
		}
		return "no"
	}
	return f.input.View()
}

// command is what the form runs: the command you could have typed.
//
// Both routes go through it, so neither can drift from the other - the rule the
// capture window already follows.
func (f fileForm) command() string {
	file := quotePath(f.field("File"))

	switch f.action {
	case copyingTo:
		copy := fmt.Sprintf("COPY %s TO %s", f.qualifiedTable(), file)
		if columns := f.field("Columns"); columns != "" {
			copy = fmt.Sprintf("COPY %s (%s) TO %s", f.qualifiedTable(), columns, file)
		}
		return copy + f.withClause()

	case copyingFrom:
		// HEADER has a field of its own, so it goes in ahead of whatever else
		// was set - and it is left out of the option list, or there would be
		// two places saying what it is.
		copy := fmt.Sprintf("COPY %s FROM %s WITH HEADER=%s",
			f.qualifiedTable(), file, f.field("Header row"))
		if with := f.withClause(); with != "" {
			copy += " AND " + strings.TrimPrefix(with, " WITH ")
		}
		return copy
	}
	return "SOURCE " + file
}

// withClause is the WITH the option rows describe, or nothing.
//
// Quoted where the parser wants it quoted: a format and a compression are
// strings, a count and a boolean are not.
func (f fileForm) withClause() string {
	var set []string
	for _, field := range f.fields {
		if field.kind != fieldOption || field.value() == "" {
			continue
		}
		set = append(set, field.label+"="+quoteOptionValue(field.label, field.value()))
	}
	if len(set) == 0 {
		return ""
	}
	return " WITH " + strings.Join(set, " AND ")
}

// quotedOptions are the options whose values are strings.
var quotedOptions = map[string]bool{
	"FORMAT": true, "COMPRESSION": true, "ENCODING": true,
	"DELIMITER": true, "QUOTE": true, "ESCAPE": true, "NULLVAL": true,
	"PARTITION": true, "PARTITION_FILTER": true,
}

func quoteOptionValue(name, value string) string {
	// A value with a quote anywhere in it is left alone. PARTITION_FILTER takes
	// a whole expression - day='2026-09-10' - and wrapping that in quotes again
	// makes a string of the first three characters and nonsense of the rest.
	if strings.Contains(value, "'") || !quotedOptions[strings.ToUpper(name)] {
		return value
	}
	return "'" + value + "'"
}

// quotePath wraps a path in the quotes these commands expect, unless it is
// already quoted.
func quotePath(path string) string {
	if path == "" {
		return "''"
	}
	if strings.HasPrefix(path, "'") || strings.HasPrefix(path, `"`) {
		return path
	}
	return "'" + path + "'"
}

// ready reports whether the form has enough to run.
func (m *MainModel) formReady() bool {
	return len(m.formErrors()) == 0
}

// submitFileForm runs the command the form describes.
//
// It goes to the Console afterwards. What COPY and SOURCE have to say - how
// many rows moved, or which line of the file would not parse - is written
// there, and running one of these from the Results view left the answer on a
// page you were not looking at. The command was echoed and answered into a view
// you had to know to go and find.
func (m *MainModel) submitFileForm() (*MainModel, tea.Cmd) {
	if !m.formReady() {
		return m, nil
	}

	command := m.form.command()
	m.closeFileForm()

	m.viewMode = "history"
	updated, cmd := m.runCommand(command)
	updated.historyViewport.GotoBottom()
	return updated, cmd
}

// currentKeyspace is the keyspace queries go to, or "" if none is set.
func (m *MainModel) currentKeyspace() string {
	if m.sessionManager == nil {
		return ""
	}
	return m.sessionManager.CurrentKeyspace()
}

// tableChoices is the tables in a keyspace, for the table field.
//
// An empty list means the field simply does not complete, which is better than
// a form that will not open.
func (m *MainModel) tableChoices(keyspace string) []string {
	if m.session == nil || keyspace == "" {
		return nil
	}

	tables, err := m.session.DescribeTablesQuery(keyspace)
	if err != nil {
		logger.DebugfToFile("FileForm", "Listing tables for the form: %v", err)
		return nil
	}

	names := make([]string, 0, len(tables))
	for _, t := range tables {
		names = append(names, t.Name)
	}
	return names
}

// formGeometry is where the window sits. Drawing and clicking both come through
// here, so a click cannot land on a different row from the one drawn there.
type formGeometry struct {
	x, y          int
	width, height int
	rows          []string
	fieldRow      int // the row the first field is on, inside the border
	matchRow      int // the row the first candidate is on, or -1 for none

	// bars is the scrollbar character for the rows that have one, drawn against
	// the right-hand edge of the window rather than beside the list.
	bars map[int]rune

	buttonRow int // the row the buttons are on, inside the border
	headerRow int // the COMMAND/OPTIONS heading
	paneRows  int // how many rows of fields follow it
}

// fixedWidth is the window's content width, from the parts of it that do not
// change while it is open.
func (f fileForm) fixedWidth() int {
	width := max(lipgloss.Width(f.title()), lipgloss.Width(f.describe()))
	for _, keys := range formKeyLines {
		width = max(width, lipgloss.Width(keys))
	}

	panes := f.panes()
	starts := f.paneStarts(panes)

	for i, pane := range panes {
		width = max(width, starts[i]+f.paneWidth(pane))
	}
	return width
}

// formRows is the window's content, one line each.
func (m *MainModel) formRows() (rows []string, fieldRow, matchRow int, bars map[int]rune) {
	f := m.form

	rows = []string{f.title(), f.describe(), ""}
	bars = map[int]rune{}

	// A heading over each column: what the command needs on the left, what
	// adjusts it on the right. Without saying so the options read as more
	// fields of the form.
	//
	// Columns and Header row sit under REQUIRED without being required. They
	// are part of what the command says rather than settings for how it runs,
	// which is the line the two columns are drawn along; both show what leaving
	// them alone does - "all of them", "yes" - so neither reads as unanswered.
	if panes := f.panes(); len(panes) > 1 {
		rows = append(rows, pad("  REQUIRED", f.paneWidth(panes[0]))+paneDivider+"  OPTIONS")
	}
	fieldRow = len(rows)

	panes := f.panes()

	tallest := 0
	for _, pane := range panes {
		tallest = max(tallest, len(pane))
	}

	first, total := f.optionWindow()
	thumb := scrollbarColumn(min(f.optionRows, total), first, total)
	scrolls := f.optionRows > 0 && total > f.optionRows

	for n := range tallest {
		row := ""
		for i, pane := range panes {
			if i > 0 {
				row += paneDivider
			}
			row += pad(m.fieldRow(pane, n), f.paneWidth(pane))
		}
		rows = append(rows, strings.TrimRight(row, " "))

		// A bar down the edge while the column is a window onto a longer list.
		if scrolls && n < len(thumb) {
			bar := '░'
			if thumb[n] {
				bar = '█'
			}
			bars[len(rows)-1] = bar
		}
	}

	rows = append(rows, "", m.focusedHint())
	if f.awaiting != "" {
		rows[len(rows)-1] = f.awaiting + " ="
	}

	matchRow = len(rows)

	if len(f.matches) == 0 {
		// The rows are there either way, so the window is one size.
		for range f.rows {
			rows = append(rows, "")
		}
		return append(rows, "", m.formButtons(), m.formKeys()), fieldRow, -1, bars
	}

	{
		first, last := f.matchWindow()
		width := 0
		for _, name := range f.matches {
			width = max(width, lipgloss.Width(name))
		}

		// A bar down the right-hand edge of the window - not tucked in beside
		// the longest name, where it reads as part of the list. A keyspace can
		// hold hundreds of tables, and without it there is no sign of how much
		// of the list you are looking at or where in it you are.
		thumb := scrollbarColumn(last-first, first, len(f.matches))
		scrolls := len(f.matches) > f.rows

		for i := first; i < last; i++ {
			marker := "  "
			if i == f.match {
				marker = "> "
			}
			rows = append(rows, marker+pad(f.matches[i], width))

			bar := ' '
			if scrolls {
				bar = '░'
				if thumb[i-first] {
					bar = '█'
				}
			}
			bars[len(rows)-1] = bar
		}
		// Short lists still take the whole area, for the same reason.
		for range f.rows - (last - first) {
			rows = append(rows, "")
		}
		keys := "↑↓/wheel: Move   Enter: Use   Esc: Back"
		if field := f.current(); field != nil && field.kind == fieldColumns {
			keys = "↑↓/wheel: Move   Space: Tick   Esc: Done"
		}
		if f.awaiting != "" {
			keys = "↑↓/wheel: Move   Enter: Use   Esc: Type it"
		}
		return append(rows, "", m.formButtons(), keys), fieldRow, matchRow, bars
	}
}

// formButtons is the button row.
func (m *MainModel) formButtons() string {
	return "  [ " + m.form.runLabel() + " ]" + formButtonGap + "  [ " + formCancel + " ]"
}

// formButtonSpans is where the two buttons sit inside the row, so drawing and
// clicking cannot disagree about which one was pressed.
func (f fileForm) formButtonSpans() (runEnd, cancelStart, cancelEnd int) {
	// Two columns in front of each, for the marker that shows which one the
	// cursor is on.
	runEnd = 2 + lipgloss.Width("[ "+f.runLabel()+" ]")
	cancelStart = runEnd + len(formButtonGap)
	return runEnd, cancelStart, cancelStart + 2 + lipgloss.Width("[ "+formCancel+" ]")
}

// formKeys is the line of keys, which changes with the field you are on.
func (m *MainModel) formKeys() string {
	if m.form.onButton != onNoButton {
		return "↑↓: Move   Enter: Press   Esc: Close"
	}

	keys := "↑↓: Field   Esc: Close"
	if field := m.form.current(); field != nil {
		switch field.kind {
		case fieldPath, fieldKeyspace, fieldTable:
			keys = "Tab: Complete   " + keys
		case fieldColumns:
			keys = "Tab: List   " + keys
		case fieldOption:
			if len(copyOptionValues[field.label]) > 0 {
				keys = "Tab: Values   " + keys
			}
		case fieldYesNo:
			keys = "Space: Change   " + keys
		}
	}
	return keys
}

// matchWindow is the slice of candidates showing, so a long list scrolls rather
// than filling the screen.
func (f fileForm) matchWindow() (first, last int) {
	if len(f.matches) <= f.rows {
		return 0, len(f.matches)
	}
	first = min(f.matchScroll, len(f.matches)-f.rows)
	return first, first + f.rows
}

// formGeometry places the window in the middle of the screen.
func (m *MainModel) formGeometry(screenWidth, screenHeight int) (formGeometry, bool) {
	if !m.form.active {
		return formGeometry{}, false
	}

	rows, fieldRow, matchRow, bars := m.formRows()

	paneRows := 0
	for _, pane := range m.form.panes() {
		paneRows = max(paneRows, len(pane))
	}

	// Worked out from what the form always has - its title, its fields at their
	// full width, and the longest line of keys - rather than from what happens
	// to be on screen. A window that grows when a long table name turns up in
	// the list, and shrinks again when you pick one, moves under the pointer.
	inner := min(m.form.fixedWidth()+2, max(screenWidth-2, 0))

	width := inner + 2
	height := len(rows) + 2
	if width < 20 || height > screenHeight-1 {
		return formGeometry{}, false
	}

	return formGeometry{
		x:         max((screenWidth-width)/2, 0),
		y:         max((screenHeight-height)/2, 0),
		width:     width,
		height:    height,
		rows:      rows,
		fieldRow:  fieldRow,
		matchRow:  matchRow,
		bars:      bars,
		buttonRow: len(rows) - 2, // the keys line is last
		headerRow: fieldRow - 1,
		paneRows:  paneRows,
	}, true
}

// viewFileForm draws it.
func (m *MainModel) viewFileForm(screenWidth, screenHeight int) (Layer, bool) {
	g, ok := m.formGeometry(screenWidth, screenHeight)
	if !ok {
		return Layer{}, false
	}

	titleStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())

	thumbStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	inner := g.width - 2
	lines := make([]string, 0, len(g.rows))
	for i, row := range g.rows {
		if bar, ok := g.bars[i]; ok {
			// One column short, so the bar has the edge to itself.
			text := textStyle.Render(pad(" "+row, inner-1))
			switch bar {
			case '█':
				text += thumbStyle.Render("█")
			case '░':
				text += trackStyle.Render("░")
			default:
				text += " "
			}
			lines = append(lines, text)
			continue
		}

		if i == g.buttonRow {
			lines = append(lines, m.renderFormButtons(inner))
			continue
		}
		// The columns and the rule between them are drawn separately, so the
		// rule is the same colour down its whole length. Rendered with the row,
		// the heading's accent coloured its piece of the rule too and it
		// changed colour a row down.
		if i == g.headerRow {
			lines = append(lines, m.renderPaneRow(row, titleStyle, inner))
			continue
		}
		if i > g.headerRow && i <= g.headerRow+g.paneRows {
			lines = append(lines, m.renderPaneRow(row, textStyle, inner))
			continue
		}

		text := pad(" "+row, inner)
		if i == 0 {
			lines = append(lines, titleStyle.Render(text))
			continue
		}
		lines = append(lines, textStyle.Render(text))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Accent).
		Render(strings.Join(lines, "\n"))

	return Layer{
		Content: box,
		X:       g.x,
		Y:       g.y,
		Width:   g.width,
		Height:  g.height,
		ZIndex:  260,
	}, true
}

// formFieldAt returns the field a press covers.
//
// Which column the press is in decides which list of fields it indexes, so a
// click on an option lands on that option rather than on the form field drawn
// beside it.
func (m *MainModel) formFieldAt(screenWidth, screenHeight, col, row int) (int, bool) {
	g, ok := m.formGeometry(screenWidth, screenHeight)
	if !ok {
		return 0, false
	}
	if col < g.x || col >= g.x+g.width {
		return 0, false
	}

	n := row - g.y - 1 - g.fieldRow
	if n < 0 {
		return 0, false
	}

	// One for the border, one for the row's own padding. The last pane whose
	// start is at or before the press is the column it was drawn in.
	at := col - g.x - 2
	panes := m.form.panes()
	pane := panes[0]
	for i, start := range m.form.paneStarts(panes) {
		if at >= start {
			pane = panes[i]
		}
	}

	if n >= len(pane) {
		return 0, false
	}
	return pane[n], true
}

// inFileForm reports whether a press landed inside the window.
func (m *MainModel) inFileForm(screenWidth, screenHeight, col, row int) bool {
	g, ok := m.formGeometry(screenWidth, screenHeight)
	if !ok {
		return false
	}
	return col >= g.x && col < g.x+g.width && row >= g.y && row < g.y+g.height
}

// completeField fills in as much as the candidates agree on, and lists them
// when that was not enough to pick one.
//
// Paths go through the same completion as everywhere else (#130); a table
// completes against the schema read when the form opened.
func (m *MainModel) completeField() (*MainModel, tea.Cmd) {
	field := m.form.current()
	if field == nil {
		return m, nil
	}

	switch field.kind {
	case fieldPath:
		got := completePath(field.input.Value())
		field.set(got.Completed)
		m.form.clearMatches()
		if len(got.Matches) > 1 {
			m.form.matches = got.Matches
		}

	case fieldKeyspace:
		m.completeFrom(field, m.form.keyspaces)

	case fieldTable:
		m.completeFrom(field, m.tablesIn(m.form.field("Keyspace")))

	case fieldColumns:
		// There is nothing to complete from a half-typed column list, so Tab
		// shows the list to tick instead.
		return m.listField()

	case fieldOption:
		m.showOptionValues()
	}
	return m, nil
}

// completeFrom fills a field in from a list of names, the way the path
// completion does: as far as the candidates agree, and then a list to pick from
// when that was not enough to choose one.
func (m *MainModel) completeFrom(field *formField, names []string) {
	typed := strings.ToLower(strings.TrimSpace(field.input.Value()))

	var found []string
	for _, name := range names {
		if strings.HasPrefix(strings.ToLower(name), typed) {
			found = append(found, name)
		}
	}

	m.form.clearMatches()
	switch len(found) {
	case 0:
		// Nothing matches. Tab is not where you learn a name is wrong, so this
		// leaves what was typed alone rather than emptying the field.
	case 1:
		field.set(found[0])
	default:
		field.set(commonPrefix(found))
		m.form.matches = found
	}
}

// tablesIn is the tables of a keyspace, fetched the first time they are asked
// for and kept.
//
// Not fetched when the form opens, because which keyspace this is depends on
// the field above and can change while the form is open.
func (m *MainModel) tablesIn(keyspace string) []string {
	if keyspace == "" {
		keyspace = m.currentKeyspace()
	}
	if keyspace == "" {
		return nil
	}
	if names, ok := m.form.tables[keyspace]; ok {
		return names
	}

	names := m.tableChoices(keyspace)
	m.form.tables[keyspace] = names
	return names
}

// listField shows everything a field can hold, whatever is already in it.
//
// This is what a click does, and it is not what Tab does. Tab is shell-like: it
// fills in from what you have typed, and with one match it just completes.
// Clicking a field that already says "my_keyspace" is not asking to have
// "my_keyspace" completed - it is asking what the other keyspaces are, and
// filtering by what is there answers with the one thing you were trying to
// change.
func (m *MainModel) listField() (*MainModel, tea.Cmd) {
	field := m.form.current()
	if field == nil {
		return m, nil
	}

	var names []string
	var current string

	switch field.kind {
	case fieldKeyspace:
		names, current = m.form.keyspaces, field.value()
	case fieldTable:
		names, current = m.tablesIn(m.form.field("Keyspace")), field.value()
	case fieldPath:
		// The directory the path is in, listed whole: the file already named
		// is the one you are replacing.
		dir, name := splitPath(field.input.Value())
		names, current = completePath(dir).Matches, name

	case fieldColumns:
		// Not a list you pick one from: a list you tick several off. Everything
		// already in the field arrives ticked, so typing them by hand and
		// clicking them are the same thing seen two ways.
		if m.columnsIn(m.form.field("Keyspace"), m.form.field("Table")) == nil {
			return m, nil
		}
		m.form.ticked = m.form.tickedColumns()
		names = m.columnRows()

	default:
		return m, nil
	}

	m.form.abandonMatches()
	if len(names) == 0 {
		return m, nil
	}
	m.form.matches = names

	// Open on what the field already says, so the list arrives showing where
	// you are rather than at its alphabetical start.
	for i, name := range names {
		if name == current || strings.TrimSuffix(name, "/") == current {
			m.form.match = i
			m.form.matchScroll = max(min(i-m.form.rows/2, len(names)-m.form.rows), 0)
			break
		}
	}
	return m, nil
}

// useMatchInForm puts the highlighted candidate into the field.
func (m *MainModel) useMatchInForm() (*MainModel, tea.Cmd) {
	if m.form.match >= len(m.form.matches) {
		return m, nil
	}

	field := m.form.current()
	if field == nil {
		return m, nil
	}

	chosen := m.form.matches[m.form.match]
	if field.kind == fieldPath {
		// The candidates are names inside a directory, so the directory the
		// user typed has to stay in front of the one they picked.
		dir, _ := splitPath(field.input.Value())
		chosen = dir + chosen
	}

	field.set(chosen)
	m.form.clearMatches()
	return m, nil
}

// moveMatchInForm moves the highlight through the candidate list.
func (m *MainModel) moveMatchInForm(delta int) (*MainModel, tea.Cmd) {
	if len(m.form.matches) == 0 {
		return m, nil
	}

	m.form.match = min(max(m.form.match+delta, 0), len(m.form.matches)-1)
	switch {
	case m.form.match < m.form.matchScroll:
		m.form.matchScroll = m.form.match
	case m.form.match >= m.form.matchScroll+m.form.rows:
		m.form.matchScroll = m.form.match - m.form.rows + 1
	}
	return m, nil
}

// toggleField turns a yes/no field over.
func (m *MainModel) toggleField() (*MainModel, tea.Cmd) {
	if field := m.form.current(); field != nil && field.kind == fieldYesNo {
		field.yes = !field.yes
	}
	return m, nil
}

// clearDependentFields empties the fields under one that has changed.
//
// Called after anything that can edit a field rather than at each of the places
// a value can change - picking from the list, completing with Tab, typing -
// because that is three places to remember and one to check.
//
// The keyspace is checked before the table, so changing a keyspace clears the
// table and the cleared table then clears the columns, in one pass.
func (m *MainModel) clearDependentFields() {
	if !m.form.active {
		return
	}

	if keyspace := m.form.field("Keyspace"); keyspace != m.form.tableKeyspace {
		m.form.tableKeyspace = keyspace
		m.form.clearField("Table")
	}
	if table := m.form.field("Table"); table != m.form.columnsTable {
		m.form.columnsTable = table
		m.form.clearField("Columns")

		// The columns themselves go with the table, not just the choice of
		// them: kept, they are the old table's columns offered for the new one.
		m.form.columns = nil
		m.form.ticked = nil
	}
}

// clearField empties a field by label, if the form has one.
func (f *fileForm) clearField(label string) {
	for i := range f.fields {
		if f.fields[i].label == label {
			f.fields[i].set("")
		}
	}
}

// handleFormKey is every key while a form is open. It is over the view, so it
// is what a keypress is aimed at.
func (m *MainModel) handleFormKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	updated, cmd := m.formKey(msg)
	updated.clearDependentFields()
	return updated, cmd
}

// formKey is the key handling itself.
func (m *MainModel) formKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	// The candidate list takes the keys while it is showing: it is the thing
	// just asked for, and Esc goes back to the field rather than closing the
	// window out from under you.
	if len(m.form.matches) > 0 {
		ticking := false
		if field := m.form.current(); field != nil {
			ticking = field.kind == fieldColumns
		}

		switch msg.String() {
		case "esc":
			m.form.abandonMatches()
			return m, nil
		case "up":
			return m.moveMatchInForm(-1)
		case "down":
			return m.moveMatchInForm(1)
		case " ", "space":
			if ticking {
				return m.toggleColumn()
			}
		case "enter":
			switch {
			case ticking:
				return m.toggleColumn()
			case m.form.awaiting != "":
				return m.useOptionValue()
			}
			return m.useMatchInForm()
		}
	}

	switch msg.String() {
	case "esc":
		m.closeFileForm()
		return m, nil
	case "up":
		return m.moveFocus(-1)
	case "down":
		return m.moveFocus(1)
	case "tab":
		if field := m.form.current(); field != nil && field.kind == fieldYesNo {
			return m.toggleField()
		}
		return m.completeField()
	case "shift+tab":
		return m.moveFocus(-1)
	case " ", "space":
		if field := m.form.current(); field != nil && field.kind == fieldYesNo {
			return m.toggleField()
		}
	case "enter":
		// Only from the button. Enter is what you press to finish typing, and
		// on this form that would also be what overwrites a file or writes to a
		// table - too easy to do by reflex.
		switch m.form.onButton {
		case onRun:
			return m.submitFileForm()
		case onCancel:
			m.closeFileForm()
			return m, nil
		}
		return m.moveFocus(1)
	}

	// Anything else is typing.
	if field := m.form.current(); field != nil && field.kind != fieldYesNo {
		var cmd tea.Cmd
		field.input, cmd = field.input.Update(msg)
		m.form.clearMatches()
		return m, cmd
	}
	return m, nil
}

// formMatchAt returns the candidate a screen row covers, so a click picks the
// one drawn there rather than one worked out a second time.
func (m *MainModel) formMatchAt(screenWidth, screenHeight, col, row int) (int, bool) {
	g, ok := m.formGeometry(screenWidth, screenHeight)
	if !ok || g.matchRow < 0 {
		return 0, false
	}
	if col < g.x || col >= g.x+g.width {
		return 0, false
	}

	first, last := m.form.matchWindow()
	i := first + (row - g.y - 1 - g.matchRow)
	if i < first || i >= last {
		return 0, false
	}
	return i, true
}

// scrollMatches moves the candidate list under the wheel.
//
// The highlight comes with it rather than being left behind off-screen: the
// list is a thing you pick from, and a selection you cannot see is worse than
// none.
func (m *MainModel) scrollMatches(delta int) (*MainModel, tea.Cmd) {
	if len(m.form.matches) <= m.form.rows {
		return m, nil
	}

	m.form.matchScroll = min(max(m.form.matchScroll+delta, 0), len(m.form.matches)-m.form.rows)
	m.form.match = min(max(m.form.match, m.form.matchScroll), m.form.matchScroll+m.form.rows-1)
	return m, nil
}

// columnsIn is the columns of a table, fetched the first time they are wanted.
func (m *MainModel) columnsIn(keyspace, table string) []string {
	if keyspace == "" {
		keyspace = m.currentKeyspace()
	}
	if m.session == nil || keyspace == "" || table == "" {
		return nil
	}
	if m.form.columns != nil {
		return m.form.columns
	}

	info, err := m.session.DescribeTableQuery(keyspace, table)
	if err != nil || info == nil {
		logger.DebugfToFile("FileForm", "Listing columns for the form: %v", err)
		return nil
	}

	names := make([]string, 0, len(info.Columns))
	for _, col := range info.Columns {
		names = append(names, col.Name)
	}
	m.form.columns = names
	return names
}

// tickedColumns reads the field back into the set the list shows, so typing a
// list by hand and ticking one off are the same thing seen two ways.
func (f fileForm) tickedColumns() map[string]bool {
	ticked := map[string]bool{}
	for _, name := range strings.Split(f.field("Columns"), ",") {
		if name = strings.TrimSpace(name); name != "" {
			ticked[name] = true
		}
	}
	return ticked
}

// writeTickedColumns puts the ticked columns back into the field, in the order
// the table declares them rather than the order they were clicked.
func (m *MainModel) writeTickedColumns() {
	var chosen []string
	for _, name := range m.form.columns {
		if m.form.ticked[name] {
			chosen = append(chosen, name)
		}
	}

	for i := range m.form.fields {
		if m.form.fields[i].kind == fieldColumns {
			m.form.fields[i].set(strings.Join(chosen, ", "))
		}
	}
}

// toggleColumn turns the highlighted column on or off.
func (m *MainModel) toggleColumn() (*MainModel, tea.Cmd) {
	if m.form.match >= len(m.form.matches) {
		return m, nil
	}

	name := columnName(m.form.matches[m.form.match])
	m.form.ticked[name] = !m.form.ticked[name]
	m.writeTickedColumns()

	// The list stays open. Ticking columns is a several-at-a-time job, and a
	// list that closed on each one would have to be reopened for the next.
	m.form.matches = m.columnRows()
	return m, nil
}

// columnRows is the list as it is drawn: a box and a name.
func (m *MainModel) columnRows() []string {
	rows := make([]string, 0, len(m.form.columns))
	for _, name := range m.form.columns {
		box := "[ ] "
		if m.form.ticked[name] {
			box = "[x] "
		}
		rows = append(rows, box+name)
	}
	return rows
}

// columnName takes the name back out of a drawn row.
func columnName(row string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(row, "[x] "), "[ ] "))
}

// renderFormButtons draws the button row. The one that acts is dimmed until the
// form has enough to act on, the same as SAVE RESULTS with nothing to save:
// offering to do something and then reporting that it cannot is worse than
// saying so on the button.
func (m *MainModel) renderFormButtons(inner int) string {
	ready := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(formPlaceholderColour))
	plain := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())

	mark := func(label string, on bool) string {
		if on {
			return "> " + label
		}
		return "  " + label
	}

	run := mark("[ "+m.form.runLabel()+" ]", m.form.onButton == onRun)
	if m.formReady() {
		run = ready.Render(run)
	} else {
		run = dim.Render(run)
	}

	cancel := plain.Render(mark("[ "+formCancel+" ]", m.form.onButton == onCancel))
	row := " " + run + formButtonGap + cancel
	if pad := inner - lipgloss.Width(stripAnsi(row)); pad > 0 {
		row += strings.Repeat(" ", pad)
	}
	return row
}

// formButtonAt returns which button a press landed on.
func (m *MainModel) formButtonAt(screenWidth, screenHeight, col, row int) (string, bool) {
	g, ok := m.formGeometry(screenWidth, screenHeight)
	if !ok || row != g.y+1+g.buttonRow {
		return "", false
	}

	// One for the border, one for the row's own padding.
	at := col - g.x - 2
	runEnd, cancelStart, cancelEnd := m.form.formButtonSpans()

	switch {
	case at >= 0 && at < runEnd:
		return "run", true
	case at >= cancelStart && at < cancelEnd:
		return "cancel", true
	}
	return "", false
}

// copyOptionHelp is a line for the options that have one.
//
// The names themselves come from completion.CopyOptions, so an option added
// there appears here whether or not anyone writes it a line. Keeping the names
// in one place and the descriptions in another means a new option turns up
// undescribed rather than missing, which is the right way round.
var copyOptionHelp = map[string]string{
	"HEADER":           "write or read a row of column names",
	"DELIMITER":        "what separates the fields",
	"NULLVAL":          "what an empty value looks like",
	"QUOTE":            "what wraps a field containing the delimiter",
	"ESCAPE":           "what escapes a quote inside a field",
	"ENCODING":         "the file's character encoding",
	"PAGESIZE":         "rows fetched from Cassandra at a time",
	"MAXREQUESTS":      "queries in flight at once",
	"MAXROWS":          "stop after this many",
	"SKIPROWS":         "ignore this many at the start",
	"MAXPARSEERRORS":   "give up after this many unreadable rows",
	"MAXINSERTERRORS":  "give up after this many failed writes",
	"MAXBATCHSIZE":     "rows per write, at most",
	"MINBATCHSIZE":     "rows per write, at least",
	"CHUNKSIZE":        "rows read from the file at a time",
	"FORMAT":           "csv, json or parquet, rather than from the extension",
	"COMPRESSION":      "how the Parquet file is compressed",
	"PARTITION":        "columns to split the output into directories by",
	"PARTITION_FILTER": "which partitions to read",
	"MAX_FILE_SIZE":    "split the output at this size",
}

// copyOptionValues is the values an option accepts, where it accepts a known
// set. The rest - a delimiter, a row count, a filename - are anything.
//
// Taken from what reads them: the formats from the same list the prompt
// completes against, the compressions from ParseCompression. An option whose
// values are written out here and changed there is the drift this file keeps
// avoiding elsewhere.
var copyOptionValues = map[string][]string{
	"HEADER":      {"true", "false"},
	"FORMAT":      lowered(completion.CopyFormats),
	"COMPRESSION": {"snappy", "gzip", "lz4", "zstd", "none"},
	"ENCODING":    {"utf8"},
}

// lowered is a copy of a list in lower case, which is how these options are
// written in a command.
func lowered(names []string) []string {
	out := make([]string, len(names))
	for i, name := range names {
		out[i] = strings.ToLower(name)
	}
	return out
}

// copyOptionsFor is the options a direction reads.
//
// From the router, which is what reads them, and per direction rather than the
// union the prompt completes with: CHUNKSIZE means something to an import and
// nothing to the CSV export beside it, and an option you can set and then watch
// do nothing is worse than one that is not offered.
func (f fileForm) copyOptionsFor() []string {
	var names []string
	if f.action == copyingFrom {
		names = router.CopyFromOptions()
	} else {
		names = router.CopyToOptions()
	}

	kept := names[:0]
	for _, name := range names {
		// HEADER has a field of its own on the import form. Two places setting
		// one thing is how they end up disagreeing.
		if f.action == copyingFrom && name == "HEADER" {
			continue
		}
		kept = append(kept, name)
	}
	sort.Strings(kept)
	return kept
}

// showOptionValues offers the values an option accepts, where it accepts a
// known set. One that takes anything shows nothing: an empty list in front of a
// field you have to type into is worse than no list.
func (m *MainModel) showOptionValues() {
	field := m.form.current()
	if field == nil || field.kind != fieldOption {
		return
	}

	m.form.abandonMatches()
	if values := copyOptionValues[field.label]; len(values) > 0 {
		m.form.awaiting = field.label
		m.form.matches = values
	}
}

// useOptionValue puts the highlighted value after the "=".
func (m *MainModel) useOptionValue() (*MainModel, tea.Cmd) {
	if m.form.match >= len(m.form.matches) {
		return m, nil
	}

	field := m.form.current()
	if field == nil {
		return m, nil
	}

	field.set(m.form.matches[m.form.match])

	m.form.awaiting = ""
	m.form.clearMatches()
	return m, nil
}

// renderPaneRow draws one row of the columns, with the rule between them in its
// own colour rather than in whatever the row happens to be.
func (m *MainModel) renderPaneRow(row string, style lipgloss.Style, inner int) string {
	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	parts := strings.Split(" "+row, paneDivider)
	for i, part := range parts {
		parts[i] = style.Render(part)
	}

	drawn := strings.Join(parts, ruleStyle.Render(paneDivider))
	if pad := inner - lipgloss.Width(stripAnsi(drawn)); pad > 0 {
		drawn += strings.Repeat(" ", pad)
	}
	return drawn
}

// Checking what is in the fields before anything is run.
//
// COPY FROM writes to a table and COPY TO writes over a file, and both take a
// while to find out they were pointed somewhere wrong. A name that is not there
// or a count that is not a number is worth saying before rather than after.

// numericOptions take a whole number.
var numericOptions = map[string]bool{
	"PAGESIZE": true, "MAXREQUESTS": true, "MAXROWS": true, "SKIPROWS": true,
	"MAXPARSEERRORS": true, "MAXINSERTERRORS": true,
	"MAXBATCHSIZE": true, "MINBATCHSIZE": true, "CHUNKSIZE": true,
}

// singleCharOptions take one character.
var singleCharOptions = map[string]bool{
	"DELIMITER": true, "QUOTE": true, "ESCAPE": true,
}

// fieldWrong reports whether a field is filled in wrongly.
//
// An optional field left empty is unanswered rather than wrong, which is the
// difference between a form that will not run and one you have not finished.
func (m *MainModel) fieldWrong(i int) bool {
	if i < 0 || i >= len(m.form.fields) {
		return false
	}
	if m.form.fields[i].optional && m.form.fields[i].value() == "" {
		return false
	}
	return m.fieldError(i) != ""
}

// fieldError says what is wrong with a field, or "" when nothing is.
//
// A name is only checked against a list that was actually fetched: with no
// session there is nothing to check against, and refusing everything because
// the cluster could not be asked is worse than letting it through to the
// command, which will say so itself.
func (m *MainModel) fieldError(i int) string {
	if i < 0 || i >= len(m.form.fields) {
		return ""
	}
	field := m.form.fields[i]
	value := field.value()

	switch field.kind {
	case fieldKeyspace:
		switch {
		case value == "":
			return "a keyspace is needed"
		case len(m.form.keyspaces) > 0 && !slices.Contains(m.form.keyspaces, value):
			return "no keyspace of that name"
		}

	case fieldTable:
		tables := m.form.tables[m.form.field("Keyspace")]
		switch {
		case value == "":
			return "a table is needed"
		case len(tables) > 0 && !slices.Contains(tables, value):
			return "no table of that name in this keyspace"
		}

	case fieldPath:
		switch {
		case value == "" || value == pathRoot:
			return "a file is needed"
		case strings.HasSuffix(value, "/"):
			return "that is a directory"
		}

	case fieldColumns:
		if value == "" || len(m.form.columns) == 0 {
			return ""
		}
		for _, name := range strings.Split(value, ",") {
			name = strings.TrimSpace(name)
			if name != "" && !slices.Contains(m.form.columns, name) {
				return "no column called " + name
			}
		}

	case fieldOption:
		return optionError(field.label, value)
	}
	return ""
}

// optionError checks one WITH option's value.
func optionError(name, value string) string {
	if value == "" {
		return "" // not set, which is what an empty option means
	}

	switch {
	case numericOptions[name]:
		if _, err := strconv.Atoi(value); err != nil {
			return "a whole number"
		}
	case singleCharOptions[name]:
		if len([]rune(strings.Trim(value, "'"))) != 1 {
			return "one character"
		}
	case len(copyOptionValues[name]) > 0:
		if !slices.Contains(copyOptionValues[name], strings.ToLower(strings.Trim(value, "'"))) {
			return "one of " + strings.Join(copyOptionValues[name], ", ")
		}
	}
	return ""
}

// formErrors is every field that is wrong, by index.
func (m *MainModel) formErrors() map[int]string {
	wrong := map[int]string{}
	for i := range m.form.fields {
		if m.fieldWrong(i) {
			wrong[i] = m.fieldError(i)
		}
	}
	return wrong
}
