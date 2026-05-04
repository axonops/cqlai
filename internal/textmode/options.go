package textmode

// Options holds configuration for the text-mode REPL.
// All fields are populated from existing CLI flags so behaviour is consistent
// with batch mode.
type Options struct {
	// Format controls the output format: "ascii" (default), "table", "json", "csv".
	Format string

	// PageSize is the number of rows to print before showing the --MORE-- prompt.
	// A value ≤ 0 means "no paging" (print all rows without pausing).
	PageSize int

	// NoHeader suppresses column headers. Used only for CSV format.
	NoHeader bool

	// FieldSep is the CSV field separator (default ",").
	FieldSep string

	// HistoryFile is the path to the readline history file.
	// If empty, defaults to ~/.cqlai_history, overridable via CQLAI_HISTORY_FILE.
	HistoryFile string

	// Version is the application version string, shown in the banner.
	Version string
}
