package router

import "strings"

// handleHelp is HELP in batch mode: the list of everything, or with a command
// after it, how that command is written. The shell opens its Help window
// instead.
func (h *MetaCommandHandler) handleHelp(command string) interface{} {
	topic := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(command), ";"))
	if len(topic) >= 4 {
		topic = strings.TrimSpace(topic[4:]) // after HELP
	}
	if topic == "" {
		return HelpRows()
	}
	return strings.Join(HelpForTopic(topic), "\n")
}

// HelpRows is the help text, as category, command and description.
//
// One source, rendered twice: the Help window - opened by F1, the tab line, or
// HELP at the prompt - and HELP in batch mode, which prints the rows. Written out twice they would
// drift, and the one nobody looks at would be the one that goes stale.
func HelpRows() [][]string {
	return [][]string{
		{"Category", "Command", "Description"},
		{"━━━━━━━━━", "━━━━━━━━", "━━━━━━━━━━━"},

		// CQL Operations
		{"CQL", "SELECT ...", "Query data from tables"},
		{"", "INSERT ...", "Insert data into tables"},
		{"", "UPDATE ...", "Update existing data"},
		{"", "DELETE ...", "Delete data from tables"},
		{"", "TRUNCATE ...", "Remove all data from table"},
		{"", "CREATE ...", "Create keyspace/table/index/etc"},
		{"", "ALTER ...", "Modify keyspace/table structure"},
		{"", "DROP ...", "Remove keyspace/table/index/etc"},
		{"", "USE <keyspace>", "Switch to specified keyspace"},

		// AI Features
		{"─────────", "─────────", "─────────────"},
		{"AI", ".AI <request>", "Generate CQL from natural language"},
		{"", ".AI show users", "Example: generate SELECT query"},
		{"", ".AI create user table", "Example: generate CREATE TABLE"},

		// Schema Commands
		{"─────────", "─────────", "─────────────"},
		{"Schema", "DESCRIBE KEYSPACES", "List all keyspaces"},
		{"", "DESCRIBE TABLES", "List tables in current keyspace"},
		{"", "DESCRIBE TABLE <name>", "Show table schema details"},
		{"", "DESCRIBE TYPE <name>", "Show user-defined type"},
		{"", "DESCRIBE TYPES", "List all UDTs"},
		{"", "DESCRIBE CLUSTER", "Show cluster information"},
		{"", "DESC ...", "Short form of DESCRIBE"},

		// Session Settings
		{"─────────", "─────────", "─────────────"},
		{"Session", "CONSISTENCY [level]", "Show/set consistency level"},
		{"", "  ONE, QUORUM, ALL", "Common consistency levels"},
		{"", "  LOCAL_ONE, LOCAL_QUORUM", "Datacenter-aware levels"},
		{"", "TRACING ON|OFF", "Enable/disable query tracing"},
		{"", "PAGING [size]", "Set result page size"},
		{"", "AUTOFETCH ON|OFF", "Auto-fetch all pages without scroll pauses"},
		{"", "EXPAND ON|OFF", "Toggle vertical output format"},

		// Output Control
		{"─────────", "─────────", "─────────────"},
		{"Output", "OUTPUT [format]", "Set output format, and redraw what is on screen:"},
		{"", "  TABLE", "Formatted table (default)"},
		{"", "  JSON", "JSON format"},
		{"", "  EXPAND", "Vertical format (EXPAND ON does the same)"},
		{"", "  ASCII", "ASCII table"},
		{"", "AUTOSAVE 'dir'", "Save each query's output into a directory"},
		{"", "AUTOSAVE JSON 'dir'", "One JSON file per query"},
		{"", "AUTOSAVE PARQUET 'dir'", "One Parquet file per query"},
		{"", "AUTOSAVE OFF", "Stop saving"},
		{"", "SAVE", "Interactive save dialog for last results"},
		{"", "SAVE 'file.csv'", "Save last results to CSV file"},
		{"", "SAVE 'file.json'", "Save last results to JSON file"},
		{"", "SAVE 'file.txt' ASCII", "Save as ASCII table format"},

		// Information
		{"─────────", "─────────", "─────────────"},
		{"Info", "HELP <command>", "How a command is written: HELP INSERT, HELP CREATE TABLE, HELP COPY"},
		{"", "SHOW VERSION", "Show Cassandra version"},
		{"", "SHOW HOST", "Show connection details"},
		{"", "SHOW SESSION", "Display session settings"},

		// File Operations
		{"─────────", "─────────", "─────────────"},
		{"Files", "SOURCE 'file'", "Execute CQL from file"},
		{"", "COPY <table> TO 'file'", "Export table data to CSV/Parquet"},
		{"", "COPY <table> FROM 'file'", "Import CSV/Parquet to table"},
		{"", "  WITH FORMAT='parquet'", "Use Apache Parquet format"},
		{"", "  WITH HEADER=true", "First row has column names"},
		{"", "  WITH DELIMITER=','", "Field separator (CSV only)"},
		{"", "  WITH MAXROWS=n", "Max rows to import (-1=all)"},
		{"", "  WITH SKIPROWS=n", "Skip first n rows (CSV only)"},
		{"", "FILE menu (Alt+F)", "The same operations, as forms to fill in"},

		// Configuration
		{"─────────", "─────────", "─────────────"},
		{"Settings", "FILE > PREFERENCES", "Edit the settings cqlai starts with"},
		{"", "  Saved to", "cqlai.json, or ~/.cassandra/cqlai.json when there is none"},
		{"", "  This session", "The status line, which the file does not change"},

		// The MCP server
		{"─────────", "─────────", "─────────────"},
		{"MCP", "cqlai mcp", "This shell, serving MCP for the connection picked in it"},
		{"", "  Changes", "Never run by MCP: proposed into the prompt, confirmed before they run"},
		{"", "  Refusals", "What the settings refuse comes back as the statement, for you to run"},
		{"", "  Client setup", "FILE > MCP SERVER: the address, and the token to send"},
		{"", "  Token", "Off unless Require token is on; then ~/.cassandra/cqlai_mcp_token"},
		{"", "  Address", "127.0.0.1:7845 unless Listen on and Port say otherwise"},
		{"", "  TLS", "TLS certificate and key serve HTTPS; TLS client CA asks for client certificates"},
		{"", "  --headless", "No shell: MCP on stdin and stdout, for a client that starts cqlai"},
		{"", "  What it may do", "PREFERENCES and CONNECT, under MCP SERVER"},
		{"", "  Pages", "Page size, unless the client asks for another; Auto fetch returns every row"},
		{"", "  System keyspaces", "Visible with no keyspace list when ticked; system_auth's data never"},
		{"", "  Audit log", "~/.cassandra/cqlai_mcp_audit.log: every call, refusals too"},

		// Keyboard Shortcuts
		{"─────────", "─────────", "─────────────"},
		{"Keys", "F1 or Alt+H", "Open this help (F1 is taken by some terminals)"},
		{"", "F2", "Console: what you typed and what came back"},
		{"", "F3", "Schema: the keyspaces and tables, with their definitions"},
		{"", "  Groups", "In Schema, each keyspace holds its tables, views, indexes, types, functions, aggregates, triggers"},
		{"", "  / or ↑ at top", "In Schema, filter the tree by name; Esc clears it"},
		{"", "F4", "Results: the last query, in whatever OUTPUT is set to"},
		{"", "F5", "Trace: the second tab of that view, when tracing is on"},
		{"", "F6", "Chat: the AI conversation"},
		{"", "↑/↓ or Ctrl+P/N", "Navigate command history"},
		{"", "Ctrl+R", "Search history"},
		{"", "Tab", "Auto-complete"},
		{"", "  Click", "Use the completion clicked; the wheel moves through the list"},
		{"", "Ctrl+L", "Clear screen"},
		{"", "Ctrl+C", "Cancel current command"},
		{"", "Ctrl+D", "Exit (EOF)"},
		{"", "Ctrl+Q", "Quit, asking first - the same as FILE > QUIT"},
		{"", "Alt+A", "Read what is on screen with the AI: a trace, or a definition"},
		{"", "Option on a Mac", "Option+F/H/D/B work as Alt; for the rest, set Option to Esc+ or Meta"},

		// Text Editing
		{"", "Ctrl+A/E", "Jump to start/end of line"},
		{"", "Ctrl+Left/Right", "Jump by word (or 20 chars)"},
		{"", "PgUp/PgDown", "Page left/right in input field"},
		{"", "Alt+B/F", "Move by word"},
		{"", "Ctrl+K/U", "Cut to end/start of line"},
		{"", "Ctrl+W", "Cut previous word"},
		{"", "Alt+D", "Delete next word"},
		{"", "Ctrl+Y", "Paste cut text"},
		{"", "Right click", "Paste into the prompt, a form, or preferences"},
		{"", "Terminal paste", "The same, from the terminal's own paste"},

		// Navigation
		{"─────────", "─────────", "─────────────"},
		{"Navigate", "PgUp/PgDown", "Scroll results/Page input"},
		{"", "Alt+↑/↓", "Scroll line by line"},
		{"", "↑/↓", "Navigate command history"},
		{"", "Alt+←/→", "Scroll horizontally (wide tables)"},

		// Exit
		{"─────────", "─────────", "─────────────"},
		{"Exit", "EXIT or QUIT", "Exit cqlai"},
		{"", "Ctrl+D", "Exit via EOF"},
		{"", "FILE > QUIT", "The same from the menu, which asks first"},

		{"", "", ""},
		{"", "Type 'HELP <topic>' for more details", ""},
	}
}
