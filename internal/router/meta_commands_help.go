package router

// handleHelp handles HELP command
func (h *MetaCommandHandler) handleHelp() interface{} {
	return HelpRows()
}

// HelpRows is the help text, as category, command and description.
//
// One source, rendered twice: HELP prints it into the console and the Help
// window on the tab line shows the same rows. Written out twice they would
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
		{"Output", "OUTPUT [format]", "Set output format:"},
		{"", "  TABLE", "Formatted table (default)"},
		{"", "  JSON", "JSON format"},
		{"", "  EXPAND", "Vertical format"},
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
		{"Info", "SHOW VERSION", "Show Cassandra version"},
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

		// Keyboard Shortcuts
		{"─────────", "─────────", "─────────────"},
		{"Keys", "F1 or Alt+H", "Open this help (F1 is taken by some terminals)"},
		{"", "↑/↓ or Ctrl+P/N", "Navigate command history"},
		{"", "Ctrl+R", "Search history"},
		{"", "Tab", "Auto-complete"},
		{"", "Ctrl+L", "Clear screen"},
		{"", "Ctrl+C", "Cancel current command"},
		{"", "Ctrl+D", "Exit (EOF)"},

		// Text Editing
		{"", "Ctrl+A/E", "Jump to start/end of line"},
		{"", "Ctrl+Left/Right", "Jump by word (or 20 chars)"},
		{"", "PgUp/PgDown", "Page left/right in input field"},
		{"", "Alt+B/F", "Move by word"},
		{"", "Ctrl+K/U", "Cut to end/start of line"},
		{"", "Ctrl+W", "Cut previous word"},
		{"", "Alt+D", "Delete next word"},
		{"", "Ctrl+Y", "Paste cut text"},

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

		{"", "", ""},
		{"", "Type 'HELP <topic>' for more details", ""},
	}
}
