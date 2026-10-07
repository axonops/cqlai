package db

import (
	"context"
	"fmt"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"

	"github.com/axonops/cqlai/internal/logger"
)

// Reading rows: one row decoder for every place that turns what the driver
// scanned into values, and a way to read one page of a query at a time.

// RowDecoder scans rows of a result and turns each into values by column
// name. A user-defined type is decoded from its bytes against its definition,
// which the driver cannot do alone; everything else is what ScanValue makes of
// it.
type RowDecoder struct {
	columns  []gocql.ColumnInfo
	dest     []interface{}
	udt      map[int]*CQLTypeInfo
	decoder  *BinaryDecoder
	keyspace string
}

// NewRowDecoder prepares to read rows with these columns. types are their CQL
// types as columnTypesOf gives them, and keyspace is the one to look a
// user-defined type up in when its type does not say.
func (s *Session) NewRowDecoder(columns []gocql.ColumnInfo, types []string, keyspace string) *RowDecoder {
	d := &RowDecoder{
		columns:  columns,
		dest:     make([]interface{}, len(columns)),
		udt:      map[int]*CQLTypeInfo{},
		keyspace: keyspace,
	}
	if s != nil {
		d.decoder = NewBinaryDecoder(s.GetUDTRegistry())
	}

	for i, col := range columns {
		if col.TypeInfo != nil && col.TypeInfo.Type() == gocql.TypeUDT {
			d.dest[i] = new(RawBytes)
			if i < len(types) && types[i] != "" && types[i] != "udt" {
				if parsed, err := ParseCQLType(types[i]); err == nil && parsed != nil {
					d.udt[i] = parsed
				} else if err != nil {
					logger.DebugfToFile("RowDecoder", "Failed to parse type for %s: %v", col.Name, err)
				}
			}
			continue
		}
		// Not *interface{}: gocql panics on a NULL there, and silently
		// stores the previous row's zero value once it has one.
		d.dest[i] = NewScanDest(col.TypeInfo)
	}
	return d
}

// Dest is what to pass to Iter.Scan.
func (d *RowDecoder) Dest() []interface{} { return d.dest }

// Row is the row last scanned, by column name.
func (d *RowDecoder) Row() map[string]interface{} {
	row := make(map[string]interface{}, len(d.columns))
	for i, col := range d.columns {
		raw, isUDT := d.dest[i].(*RawBytes)
		if !isUDT {
			row[col.Name] = ScanValue(d.dest[i])
			continue
		}
		if raw == nil || *raw == nil || d.decoder == nil {
			row[col.Name] = nil
			continue
		}
		info := d.udt[i]
		if info == nil {
			row[col.Name] = fmt.Sprintf("0x%x", *raw)
			continue
		}
		keyspace := info.Keyspace
		if keyspace == "" {
			keyspace = d.keyspace
		}
		value, err := d.decoder.Decode([]byte(*raw), info, keyspace)
		if err != nil {
			logger.DebugfToFile("RowDecoder", "Failed to decode UDT %s: %v", col.Name, err)
			row[col.Name] = fmt.Sprintf("0x%x", *raw)
			continue
		}
		row[col.Name] = value
	}
	return row
}

// ConsistencyNamed is the level of this name, ignoring case.
func ConsistencyNamed(name string) (gocql.Consistency, bool) {
	for _, c := range consistencyLevels {
		if strings.EqualFold(c.name, strings.TrimSpace(name)) {
			return c.level, true
		}
	}
	return 0, false
}

// PageRequest is one page of a query.
type PageRequest struct {
	Statement   string
	Consistency string // empty for the session's
	PageSize    int
	PageState   []byte // where the page before left off; nil for the first
	Trace       bool
}

// PageColumn is a column of a page: its name and CQL type.
type PageColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// QueryPageResult is one page of a query's rows.
type QueryPageResult struct {
	Columns   []PageColumn
	Rows      []map[string]interface{} // values as the driver and RowDecoder give them
	PageState []byte                   // nil when there is nothing after this page
	Warnings  []string
	Keyspace  string
}

// QueryPage runs a statement and reads one page of it: no more rows than the
// page size, and never the next page, which the caller asks for with the page
// state if it wants it.
func (s *Session) QueryPage(ctx context.Context, req PageRequest) (QueryPageResult, error) {
	if s == nil || s.Session == nil {
		return QueryPageResult{}, fmt.Errorf("not connected")
	}
	if req.PageSize <= 0 {
		return QueryPageResult{}, fmt.Errorf("a page needs a size")
	}

	q := s.Query(req.Statement).PageSize(req.PageSize)
	if req.PageState != nil {
		q = q.PageState(req.PageState)
	}
	if req.Consistency != "" {
		level, ok := ConsistencyNamed(req.Consistency)
		if !ok {
			return QueryPageResult{}, fmt.Errorf("%s is not a consistency level", req.Consistency)
		}
		q = q.Consistency(level)
	}
	if req.Trace {
		s.startTracing()
		q = q.Trace(&captureTracer{session: s})
	}

	iter := q.IterContext(ctx)
	columns := iter.Columns()
	types, keyspace := s.columnTypesOf(req.Statement, columns)

	page := QueryPageResult{Keyspace: keyspace}
	for i, col := range columns {
		page.Columns = append(page.Columns, PageColumn{Name: col.Name, Type: types[i]})
	}

	decoder := s.NewRowDecoder(columns, types, keyspace)
	for len(page.Rows) < req.PageSize {
		// Stop at the end of the page rather than let the driver fetch the
		// next one: the next page is the caller's to ask for.
		if len(page.Rows) > 0 && iter.WillSwitchPage() {
			break
		}
		if !iter.Scan(decoder.Dest()...) {
			break
		}
		page.Rows = append(page.Rows, decoder.Row())
	}

	page.Warnings = iter.Warnings()
	if state := iter.PageState(); len(state) > 0 {
		page.PageState = append([]byte{}, state...)
	}
	if err := iter.Close(); err != nil {
		return QueryPageResult{}, err
	}
	return page, nil
}
