package textmode

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	readline "github.com/chzyer/readline"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/ui"
)

// printer holds output configuration used by printResult and streaming helpers.
type printer struct {
	opts Options
	// rl is the active readline instance used for the --MORE-- prompt so that
	// raw-mode transitions are handled by readline's own machinery rather than
	// duplicating them with golang.org/x/term.  May be nil in batch/test mode.
	rl *readline.Instance
}

// newPrinter constructs a printer from Options.
func newPrinter(opts Options) *printer {
	return &printer{opts: opts}
}

// printResult dispatches on the dynamic type returned by router.ProcessCommand
// and writes formatted output to w. Errors are written to stderr and do not
// cause a non-nil return (the REPL continues). Returns true if the caller
// should exit (ExitSignal).
func (p *printer) printResult(w io.Writer, result interface{}) (exit bool, err error) {
	switch v := result.(type) {
	case nil:
		// nothing to print
	case string:
		if v != "" {
			fmt.Fprintln(w, v)
		}
	case error:
		fmt.Fprintln(os.Stderr, v.Error())
	case db.QueryResult:
		if err := p.printQueryResult(w, v); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	case db.StreamingQueryResult:
		if err := p.printStreamingResult(w, v); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	case [][]string:
		if len(v) > 0 {
			fmt.Fprint(w, ui.FormatASCIITable(v))
		}
	default:
		// Covers SaveCommand, MetaCommand, etc. — render as string.
		s := fmt.Sprintf("%v", v)
		if s != "" && s != "<nil>" {
			fmt.Fprintln(w, s)
		}
	}
	return false, nil
}

// printQueryResult formats and prints a db.QueryResult.
func (p *printer) printQueryResult(w io.Writer, r db.QueryResult) error {
	switch strings.ToLower(p.opts.Format) {
	case "json":
		return printQueryResultJSON(w, r)
	case "csv":
		return printQueryResultCSV(w, r, p.opts.NoHeader, p.opts.FieldSep)
	default:
		if len(r.Data) == 0 {
			fmt.Fprintln(w, "(0 rows)")
			return nil
		}
		fmt.Fprint(w, ui.FormatASCIITable(r.Data))
		return nil
	}
}

// printStreamingResult streams rows from a StreamingQueryResult, pausing with
// a --MORE-- prompt every pageSize rows unless format is json/csv.
func (p *printer) printStreamingResult(w io.Writer, r db.StreamingQueryResult) error {
	format := strings.ToLower(p.opts.Format)

	switch format {
	case "json":
		return p.printStreamingJSON(w, r)
	case "csv":
		return p.printStreamingCSV(w, r)
	default:
		return p.printStreamingASCII(w, r)
	}
}

// printStreamingASCII streams rows in ASCII table format with --MORE-- paging.
func (p *printer) printStreamingASCII(w io.Writer, r db.StreamingQueryResult) error {
	pageSize := p.opts.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}

	rowCount := 0
	var rows [][]string
	isFirstBatch := true
	var columnWidths []int

	flush := func(final bool) {
		if len(rows) == 0 {
			return
		}

		allData := append([][]string{r.Headers}, rows...)
		if isFirstBatch {
			columnWidths = ui.CalculateColumnWidths(allData)
			output := ui.FormatASCIITable(allData)
			// Strip row-count footer and bottom border so we can append more rows.
			if !final {
				output = stripTableFooter(output, len(rows))
			}
			fmt.Fprint(w, output)
			isFirstBatch = false
		} else {
			output := ui.FormatASCIITableRowsOnlyWithWidths(allData, columnWidths)
			fmt.Fprint(w, output)
			if final {
				bottom := ui.FormatASCIITableBottomWithWidths(allData, columnWidths)
				fmt.Fprint(w, bottom)
			}
		}
	}

	for {
		rowMap := make(map[string]interface{})
		if !r.Iterator.MapScan(rowMap) {
			break
		}

		row := make([]string, len(r.ColumnNames))
		for i, colName := range r.ColumnNames {
			if val, exists := rowMap[colName]; exists {
				row[i] = db.FormatValue(val)
			} else {
				row[i] = "null"
			}
		}
		rows = append(rows, row)
		rowCount++

		if len(rows) >= pageSize {
			flush(false)
			rows = nil

			// --MORE-- prompt: read one raw byte
			abort := p.morePrompt()
			if abort {
				// Drain and close the iterator
				for r.Iterator.MapScan(make(map[string]interface{})) {
				}
				fmt.Fprintf(w, "\n(%d rows)\n", rowCount)
				return r.Iterator.Close()
			}
		}
	}

	if err := r.Iterator.Close(); err != nil {
		return fmt.Errorf("iterator error: %w", err)
	}

	flush(true)

	// If we never had any rows at all, print empty header table.
	if rowCount == 0 && isFirstBatch {
		fmt.Fprint(w, ui.FormatASCIITable([][]string{r.Headers}))
	}

	fmt.Fprintf(w, "\n(%d rows)\n", rowCount)
	return nil
}

// stripTableFooter removes the bottom border + row-count footer from FormatASCIITable output.
func stripTableFooter(output string, rowLen int) string {
	// FormatASCIITable appends "\n(%d rows)\n" or "\n(%d row)\n"
	suffix1 := fmt.Sprintf("\n(%d rows)\n", rowLen)
	suffix2 := fmt.Sprintf("\n(%d row)\n", rowLen)
	output = strings.TrimSuffix(output, suffix1)
	output = strings.TrimSuffix(output, suffix2)

	// Remove the bottom border line (starts with + and contains -)
	lines := strings.Split(output, "\n")
	done := false
	for i := len(lines) - 1; i >= 0 && !done; i-- {
		switch {
		case lines[i] == "":
			lines = lines[:i]
		case strings.HasPrefix(lines[i], "+") && strings.Contains(lines[i], "-"):
			lines = lines[:i]
			done = true
		default:
			done = true
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// morePrompt prints "--MORE--", reads a single raw byte via readline's owned
// terminal machinery and returns true if the user pressed q/Q.
//
// By routing through readOneRawViaReadline we avoid double raw-mode toggling:
// readline manages the terminal state; we simply borrow its FuncMakeRaw /
// FuncExitRaw hooks and its already-installed Stdin reader.  rl.Clean() is
// called first so readline's prompt area is erased, and rl.Refresh() is called
// after to restore the prompt after the user's key is consumed.
func (p *printer) morePrompt() bool {
	if p.rl != nil {
		p.rl.Clean()
	}
	fmt.Print("--MORE-- (press any key to continue, q to quit) ")
	ch, ok := readOneRawViaReadline(p.rl)
	fmt.Println()
	if p.rl != nil {
		p.rl.Refresh()
	}
	if !ok {
		return false
	}
	return ch == 'q' || ch == 'Q'
}

// printStreamingJSON collects all rows then emits a JSON document.
func (p *printer) printStreamingJSON(w io.Writer, r db.StreamingQueryResult) error {
	type doc struct {
		Columns []string                 `json:"columns"`
		Rows    []map[string]interface{} `json:"rows"`
		Count   int                      `json:"row_count"`
	}
	d := doc{Columns: r.Headers}
	for {
		rowMap := make(map[string]interface{})
		if !r.Iterator.MapScan(rowMap) {
			break
		}
		d.Rows = append(d.Rows, rowMap)
	}
	if err := r.Iterator.Close(); err != nil {
		return fmt.Errorf("iterator error: %w", err)
	}
	d.Count = len(d.Rows)
	if d.Rows == nil {
		d.Rows = []map[string]interface{}{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(d)
}

// printStreamingCSV streams rows as CSV.
func (p *printer) printStreamingCSV(w io.Writer, r db.StreamingQueryResult) error {
	sep := ','
	if len(p.opts.FieldSep) > 0 {
		sep = rune(p.opts.FieldSep[0])
	}
	cw := csv.NewWriter(w)
	cw.Comma = sep
	if !p.opts.NoHeader {
		if err := cw.Write(r.Headers); err != nil {
			return err
		}
	}
	for {
		rowMap := make(map[string]interface{})
		if !r.Iterator.MapScan(rowMap) {
			break
		}
		row := make([]string, len(r.ColumnNames))
		for i, colName := range r.ColumnNames {
			if val, exists := rowMap[colName]; exists {
				row[i] = db.FormatValue(val)
			} else {
				row[i] = "null"
			}
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	if err := r.Iterator.Close(); err != nil {
		return fmt.Errorf("iterator error: %w", err)
	}
	return cw.Error()
}

// printQueryResultJSON writes a QueryResult as JSON.
func printQueryResultJSON(w io.Writer, r db.QueryResult) error {
	type doc struct {
		Columns []string                 `json:"columns"`
		Rows    []map[string]interface{} `json:"rows"`
		Count   int                      `json:"row_count"`
	}

	if len(r.Data) == 0 {
		d := doc{Columns: []string{}, Rows: []map[string]interface{}{}, Count: 0}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(d)
	}

	headers := r.Data[0]
	rows := r.Data[1:]

	var rowMaps []map[string]interface{}
	if len(r.RawData) == len(rows) {
		rowMaps = r.RawData
	} else {
		rowMaps = make([]map[string]interface{}, len(rows))
		for i, row := range rows {
			m := make(map[string]interface{}, len(headers))
			for j, h := range headers {
				if j < len(row) {
					m[h] = row[j]
				}
			}
			rowMaps[i] = m
		}
	}

	d := doc{Columns: headers, Rows: rowMaps, Count: len(rows)}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(d)
}

// printQueryResultCSV writes a QueryResult as CSV.
func printQueryResultCSV(w io.Writer, r db.QueryResult, noHeader bool, fieldSep string) error {
	sep := ','
	if len(fieldSep) > 0 {
		sep = rune(fieldSep[0])
	}
	cw := csv.NewWriter(w)
	cw.Comma = sep
	if len(r.Data) == 0 {
		cw.Flush()
		return cw.Error()
	}
	start := 0
	if !noHeader {
		if err := cw.Write(r.Data[0]); err != nil {
			return err
		}
		start = 1
	}
	for _, row := range r.Data[start:] {
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
