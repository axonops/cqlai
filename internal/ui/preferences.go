package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
)

// The PREFERENCES window: every setting in cqlai.json, edited in the shell.
//
// Until now the only way to change one was to leave cqlai, find whichever of
// the three files it reads is yours, and know the key names. LoadConfig has
// always read them and nothing has ever written one back.
//
// What this window edits is the file - what cqlai starts with. It does not
// change the session running now: consistency, paging and output format have
// controls on the status line for that, and those are the values in use. Two
// places setting the same thing at the same time is the mistake this project
// keeps finding; the window says which one it is instead.
//
// It shows the file rather than the configuration cqlai is running on, which is
// that file plus a cqlshrc, plus the environment, plus the command line. Shown
// the merged thing, Save would write all of it back: a password that was in
// $CASSANDRA_PASSWORD and nowhere else would land in the file, put there by
// someone who opened the window to change the port. What is drawn is what will
// be written, and opening the window and saving it changes nothing.

// prefKind is how a setting is edited.
type prefKind int

const (
	prefText   prefKind = iota // typed
	prefNumber                 // typed, and has to be a whole number
	prefSecret                 // typed, drawn as dots
	prefPath                   // typed, Tab completes it against the filesystem
	prefYesNo                  // space or a click changes it
	prefChoice                 // Tab lists what it can be
)

// prefSpec describes one setting.
//
// path names the field in config.Config - "Host", or "SSL.CertPath" for one in
// a nested struct. Reading and writing go through it by reflection rather than
// through a getter and a setter per field: forty pairs of one-line closures is
// forty chances for one to read one field and write another.
type prefSpec struct {
	section string // set on the first setting of a section, and drawn above it
	path    string
	label   string
	kind    prefKind
	hint    string
	choices func() []string
}

// prefSpecs is every setting in the file, in the order the window shows them.
//
// The order is how often you touch them rather than the order of the struct:
// the connection first, then what the shell does with results, then the parts
// most configurations leave alone.
func prefSpecs() []prefSpec {
	specs := []prefSpec{
		{section: "CONNECTION", path: "Host", label: "Host", kind: prefText, hint: "the node cqlai connects to"},
		{path: "Port", label: "Port", kind: prefNumber, hint: "9042 unless the cluster was moved"},
		{path: "Keyspace", label: "Keyspace", kind: prefText, hint: "the keyspace to start in"},
		{path: "Username", label: "Username", kind: prefText, hint: "left empty when the cluster has no authentication"},
		{path: "Password", label: "Password", kind: prefSecret, hint: "kept in the file, which is written readable only by you"},
		{path: "ConnectTimeout", label: "Connect timeout", kind: prefNumber, hint: "seconds to wait for the connection"},
		{path: "RequestTimeout", label: "Request timeout", kind: prefNumber, hint: "seconds to wait for a query"},

		{section: "RESULTS", path: "Consistency", label: "Consistency", kind: prefChoice, choices: db.ConsistencyLevels, hint: "how many replicas have to answer"},
		{path: "PageSize", label: "Page size", kind: prefNumber, hint: "rows fetched from Cassandra at a time"},
		{path: "MaxMemoryMB", label: "Max memory (MB)", kind: prefNumber, hint: "how much of a result is held in memory"},
		{path: "OutputFormat", label: "Output format", kind: prefChoice, choices: config.OutputFormats, hint: "how results are drawn"},
		{path: "RequireConfirmation", label: "Confirm changes", kind: prefYesNo, hint: "ask before a statement that alters data or schema"},
		{path: "Debug", label: "Debug", kind: prefYesNo, hint: "write the debug log"},

		{section: "FILES", path: "HistoryFile", label: "Command history", kind: prefPath, hint: "~/.cqlai_history unless set"},
		{path: "AIHistoryFile", label: "Chat history", kind: prefPath, hint: "~/.cqlai_ai_history unless set"},

		{section: "SSL", path: "SSL.Enabled", label: "Enabled", kind: prefYesNo, hint: "connect over TLS"},
		{path: "SSL.CertPath", label: "Certificate", kind: prefPath, hint: "this client's certificate"},
		{path: "SSL.KeyPath", label: "Key", kind: prefPath, hint: "this client's private key"},
		{path: "SSL.CAPath", label: "CA certificate", kind: prefPath, hint: "what the cluster's certificate is checked against"},
		{path: "SSL.HostVerification", label: "Verify hostname", kind: prefYesNo, hint: "the certificate has to name the host connected to"},
		{path: "SSL.InsecureSkipVerify", label: "Skip verification", kind: prefYesNo, hint: "accept any certificate - not for production"},
		{path: "SSL.AllowLegacyCN", label: "Allow legacy CN", kind: prefYesNo, hint: "certificates with no subject alternative name"},

		{section: "CHAT", path: "AI.Provider", label: "Provider", kind: prefChoice, choices: ai.Providers, hint: "the service the CHAT tab asks"},
		{path: "AI.APIKey", label: "API key", kind: prefSecret, hint: "used when the provider below has none of its own"},
		{path: "AI.Model", label: "Model", kind: prefText, hint: "used when the provider below has none of its own"},
		{path: "AI.URL", label: "URL", kind: prefText, hint: "used when the provider below has none of its own"},
	}

	// A block per provider, so the one in use can be switched without its key
	// and model being retyped. They are the same three settings every time,
	// which is a reason to generate them rather than to write the same three
	// lines five times over.
	for _, provider := range []struct{ section, path string }{
		{"CHAT - OPENAI", "AI.OpenAI"},
		{"CHAT - ANTHROPIC", "AI.Anthropic"},
		{"CHAT - GEMINI", "AI.Gemini"},
		{"CHAT - OLLAMA", "AI.Ollama"},
		{"CHAT - OPENROUTER", "AI.OpenRouter"},
	} {
		specs = append(specs,
			prefSpec{section: provider.section, path: provider.path + ".APIKey", label: "API key", kind: prefSecret},
			prefSpec{path: provider.path + ".Model", label: "Model", kind: prefText},
			prefSpec{path: provider.path + ".URL", label: "URL", kind: prefText, hint: "where the provider is, for one that runs locally"},
		)
	}

	return append(specs,
		prefSpec{section: "AUTH PROVIDER", path: "AuthProvider.Module", label: "Module", kind: prefText, hint: "e.g. cassandra.auth"},
		prefSpec{path: "AuthProvider.ClassName", label: "Class", kind: prefText, hint: "e.g. PlainTextAuthProvider"},
	)
}

// prefField is one setting as the window holds it while it is open.
//
// Every field has an input, including the yes/no ones that never draw it.
// Focusing a zero textinput dereferences a cursor that is not there, which is
// what crashed the COPY FROM form (#165); one field without an input is enough.
type prefField struct {
	spec  prefSpec
	input textinput.Model
	yes   bool
}

// preferences is the open window.
type preferences struct {
	active bool
	fields []prefField

	focus    int // the field being edited
	onButton buttonFocus

	scroll int // the first line of the settings shown
	rows   int // how many lines of settings the window has room for

	matches   []string // what Tab offered for the focused field
	match     int
	scrollTop int // the first candidate shown
	matchRows int

	// cfg is the file as it was read, and what Save writes back. Not the
	// configuration cqlai is running on: see the note at the top.
	cfg  *config.Config
	path string // where Save will write
}

// prefLine is one drawn line of the settings area: a section heading, or a
// setting. Drawing and clicking both come from this list, so a click cannot
// land on a different setting from the one drawn there.
type prefLine struct {
	heading string
	field   int // -1 on a heading
}

func (p preferences) lines() []prefLine {
	lines := make([]prefLine, 0, len(p.fields)+8)
	for i, field := range p.fields {
		if field.spec.section != "" {
			if i > 0 {
				lines = append(lines, prefLine{field: -1})
			}
			lines = append(lines, prefLine{heading: field.spec.section, field: -1})
		}
		lines = append(lines, prefLine{field: i})
	}
	return lines
}

// lineOf is the line a field is drawn on.
func (p preferences) lineOf(field int) int {
	for i, line := range p.lines() {
		if line.field == field {
			return i
		}
	}
	return 0
}

// openPreferences reads the file into the window.
func (m *MainModel) openPreferences() (*MainModel, tea.Cmd) {
	cfg := m.prefsFile()

	specs := prefSpecs()
	fields := make([]prefField, 0, len(specs))
	for _, spec := range specs {
		value := prefValue(cfg, spec.path)
		field := prefField{spec: spec, input: newPrefInput(spec, value)}
		if spec.kind == prefYesNo {
			field.yes = value == "true"
		}
		fields = append(fields, field)
	}

	m.preferences = preferences{
		active: true,
		fields: fields,
		cfg:    cfg,
		path:   cfg.SavePath(),
	}
	m.preferences.rows, m.preferences.matchRows = m.fitPrefHeights(m.windowHeight)
	m.preferences.focusField(0)

	// A window that does not fit is not drawn, and an undrawn window that still
	// takes every key press is a shell that has stopped responding. Say so
	// instead.
	if _, ok := m.prefGeometry(m.windowWidth, m.windowHeight); !ok {
		m.closePreferences()
		return m.report("The terminal is too small for the preferences window.")
	}
	return m, nil
}

// prefsFile is the JSON configuration file, read fresh.
//
// Empty when there is not one yet, which is what an untouched file holds: the
// window then writes whatever is typed into it and nothing else.
func (m *MainModel) prefsFile() *config.Config {
	path := ""
	if m.config != nil {
		path = m.config.SavePath()
	}
	if path == "" {
		path = (&config.Config{}).SavePath()
	}

	cfg := &config.Config{SourcePath: path}
	data, err := os.ReadFile(path) // #nosec G304 - the file cqlai reads its configuration from
	if err != nil {
		return cfg
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		// A file that will not parse is not one to write over from here. The
		// window opens empty and Save says what went wrong.
		return &config.Config{SourcePath: path}
	}
	cfg.SourcePath = path
	return cfg
}

// newPrefInput is a setting's box.
func newPrefInput(spec prefSpec, value string) textinput.Model {
	in := newFormInput(value, prefPlaceholder(spec))
	if spec.kind == prefSecret {
		in.EchoMode = textinput.EchoPassword
		in.EchoCharacter = '•'
	}
	return in
}

// prefPlaceholder is what an empty setting says, which is either what it takes
// or how to fill it in.
func prefPlaceholder(spec prefSpec) string {
	switch spec.kind {
	case prefPath:
		return "Tab to complete a path"
	case prefChoice:
		return "Tab for the values"
	case prefNumber:
		return "a number"
	}
	return ""
}

// closePreferences puts the window away, discarding anything not saved.
func (m *MainModel) closePreferences() {
	m.preferences = preferences{}
}

// focusField moves the cursor to a setting and brings it into view.
func (p *preferences) focusField(i int) {
	if i < 0 || i >= len(p.fields) {
		return
	}
	for n := range p.fields {
		p.fields[n].input.Blur()
	}
	p.focus = i
	p.onButton = onNoButton
	p.fields[i].input.Focus()
	p.clearMatches()
	p.show(i)
}

// show scrolls the settings area only as far as it takes to put a setting on
// it, so moving up and down does not jump the list about.
func (p *preferences) show(field int) {
	line := p.lineOf(field)

	// A heading above the setting comes into view with it: the first field of a
	// section on its own says nothing about which section it is.
	top := line
	lines := p.lines()
	for top > 0 && lines[top-1].field == -1 {
		top--
	}

	p.scroll = min(p.scroll, top)
	p.scroll = max(p.scroll, line-p.rows+1)
	p.scroll = max(p.scroll, 0)
}

func (p *preferences) clearMatches() {
	p.matches = nil
	p.match = 0
	p.scrollTop = 0
}

// current is the setting being edited, or nil when the cursor is on a button.
func (p *preferences) current() *prefField {
	if p.onButton != onNoButton || p.focus < 0 || p.focus >= len(p.fields) {
		return nil
	}
	return &p.fields[p.focus]
}

// value is what a field holds now, as it goes into the file.
func (f prefField) value() string {
	if f.spec.kind == prefYesNo {
		return strconv.FormatBool(f.yes)
	}
	return strings.TrimSpace(f.input.Value())
}

// display is how a field is drawn.
func (f prefField) display() string {
	if f.spec.kind == prefYesNo {
		if f.yes {
			return "yes"
		}
		return "no"
	}
	return f.input.View()
}

// savePreferences writes the file.
func (m *MainModel) savePreferences() (*MainModel, tea.Cmd) {
	if !m.preferencesReady() {
		return m, nil
	}

	cfg := m.preferences.cfg
	for _, field := range m.preferences.fields {
		if err := setPrefValue(cfg, field.spec.path, field.value()); err != nil {
			// Nothing has been written yet, so the window stays open on what
			// it could not write rather than closing over a half-saved file.
			return m.report("Preferences not saved: " + err.Error())
		}
	}
	prunePrefs(cfg)

	written, err := cfg.Save()
	if err != nil {
		return m.report("Preferences not saved: " + err.Error())
	}

	m.closePreferences()
	return m.report("Preferences saved to " + written + prefRestartNote)
}

// prefRestartNote goes after what was saved, because none of it is in use yet.
//
// Everything in the window is read when cqlai starts: the connection is made
// once, and the rest is settled from the file before the first prompt. Saying
// so beats a window that looks like it changed something and did not.
const prefRestartNote = ". They take effect when cqlai next starts."

// report writes a line to the Console and shows it.
func (m *MainModel) report(message string) (*MainModel, tea.Cmd) {
	m.viewMode = "history"
	m.fullHistoryContent += "\n" + m.styles.AccentText.Render(message)
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
	return m, nil
}

// preferencesReady reports whether everything in the window can be written.
func (m *MainModel) preferencesReady() bool {
	return len(m.preferenceErrors()) == 0
}

// preferenceErrors is what is wrong with the window, by field.
//
// A setting left empty is not wrong: empty means unset, and unset is what most
// of these are. What is wrong is a value that cannot be what the setting is -
// a port that is not a number, a consistency level that does not exist.
func (m *MainModel) preferenceErrors() map[int]string {
	errors := map[int]string{}
	for i, field := range m.preferences.fields {
		value := field.value()
		if value == "" {
			continue
		}

		switch field.spec.kind {
		case prefNumber:
			n, err := strconv.Atoi(value)
			switch {
			case err != nil:
				errors[i] = field.spec.label + " has to be a number"
			case n < 0:
				errors[i] = field.spec.label + " cannot be negative"
			}

		case prefChoice:
			if !hasChoice(field.spec, value) {
				errors[i] = value + " is not one of the values for " + field.spec.label
			}
		}
	}
	return errors
}

// hasChoice reports whether a value is one the setting offers, ignoring case:
// the consistency levels are written in capitals and the AI providers are not,
// and neither is worth being strict about.
func hasChoice(spec prefSpec, value string) bool {
	if spec.choices == nil {
		return true
	}
	for _, choice := range spec.choices() {
		if strings.EqualFold(choice, value) {
			return true
		}
	}
	return false
}

// fieldWrong reports whether a setting is filled in wrongly.
func (m *MainModel) prefWrong(i int) bool {
	_, wrong := m.preferenceErrors()[i]
	return wrong
}

// Reading and writing config.Config by the paths in the table above.
//
// The alternative was a getter and a setter beside every spec. This is one
// walk, written once, and a path that does not name a field is caught by the
// test that walks all of them rather than by a field that quietly never saves.

// prefValue reads a setting, and returns "" for one inside a struct that has
// not been configured at all.
func prefValue(cfg *config.Config, path string) string {
	value := reflect.ValueOf(cfg)
	for _, name := range strings.Split(path, ".") {
		value = value.Elem().FieldByName(name)
		if !value.IsValid() {
			return ""
		}
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return ""
			}
			continue
		}

		switch value.Kind() {
		case reflect.String:
			return value.String()
		case reflect.Int:
			if value.Int() == 0 {
				return "" // unset rather than zero: every number here is a size or a timeout
			}
			return strconv.FormatInt(value.Int(), 10)
		case reflect.Bool:
			return strconv.FormatBool(value.Bool())
		default:
			return ""
		}
	}
	return ""
}

// setPrefValue writes a setting, making the structs on the way to it.
func setPrefValue(cfg *config.Config, path, value string) error {
	target := reflect.ValueOf(cfg)
	names := strings.Split(path, ".")

	for i, name := range names {
		field := target.Elem().FieldByName(name)
		if !field.IsValid() {
			return fmt.Errorf("no setting called %s", path)
		}

		if i < len(names)-1 {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			target = field
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Int:
			if value == "" {
				field.SetInt(0)
				break
			}
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("%s has to be a number", path)
			}
			field.SetInt(int64(n))
		case reflect.Bool:
			field.SetBool(value == "true")
		default:
			return fmt.Errorf("cannot write %s", path)
		}
	}
	return nil
}

// prunePrefs drops the nested structs nothing was put in.
//
// Opening the window makes an SSL block and five chat provider blocks whether
// or not they are wanted, since every setting needs somewhere to go. Written
// out as they are, a file that had none of them gains six empty objects for
// having been looked at.
func prunePrefs(cfg *config.Config) {
	pruneStruct(reflect.ValueOf(cfg).Elem())
}

func pruneStruct(value reflect.Value) {
	for i := range value.NumField() {
		field := value.Field(i)
		if field.Kind() != reflect.Pointer || field.IsNil() || field.Type().Elem().Kind() != reflect.Struct {
			continue
		}

		// Innermost first: the chat providers have to go before the block
		// holding them can count as empty.
		pruneStruct(field.Elem())
		if field.Elem().IsZero() {
			field.Set(reflect.Zero(field.Type()))
		}
	}
}
