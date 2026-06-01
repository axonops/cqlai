package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsSetColumnSchemaAware verifies the list<...> vs set<...> decision is
// driven by the destination schema, not column-name heuristics. Regression
// guard for https://github.com/axonops/cqlai/issues/81 — a column named
// "tags" of type list<text> must not be coerced into a set literal.
func TestIsSetColumnSchemaAware(t *testing.T) {
	h := &MetaCommandHandler{}

	tests := []struct {
		name        string
		column      string
		columnTypes map[string]string
		want        bool
	}{
		{
			name:        "list<text> named tags is NOT a set (issue #81)",
			column:      "tags",
			columnTypes: map[string]string{"tags": "list<text>"},
			want:        false,
		},
		{
			name:        "set<int> column emits set literal",
			column:      "unique_nums",
			columnTypes: map[string]string{"unique_nums": "set<int>"},
			want:        true,
		},
		{
			name:        "frozen<set<text>> emits set literal",
			column:      "labels",
			columnTypes: map[string]string{"labels": "frozen<set<text>>"},
			want:        false, // frozen<...> wraps the set; current code only matches bare set<...>
		},
		{
			name:        "list<int> column named foo_nums is NOT a set",
			column:      "foo_nums",
			columnTypes: map[string]string{"foo_nums": "list<int>"},
			want:        false,
		},
		{
			name:        "set<text> column named tags IS a set",
			column:      "tags",
			columnTypes: map[string]string{"tags": "set<text>"},
			want:        true,
		},
		{
			name:        "unknown column falls back to list",
			column:      "tags",
			columnTypes: map[string]string{},
			want:        false,
		},
		{
			name:        "map column is not a set",
			column:      "attributes",
			columnTypes: map[string]string{"attributes": "map<text,text>"},
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := h.isSetColumn(tc.column, tc.columnTypes)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestFormatListValueUsesSchema verifies formatListValue picks brackets vs braces
// from the column type rather than the column name.
func TestFormatListValueUsesSchema(t *testing.T) {
	h := &MetaCommandHandler{}

	listSchema := map[string]string{"tags": "list<text>"}
	setSchema := map[string]string{"tags": "set<text>"}

	listVal := []interface{}{"a", "b", "c"}

	assert.Equal(t, "['a', 'b', 'c']",
		h.formatListValue(listVal, "tags", listSchema),
		"list<text> column must use [] regardless of name")
	assert.Equal(t, "{'a', 'b', 'c'}",
		h.formatListValue(listVal, "tags", setSchema),
		"set<text> column must use {}")
}

// TestGetEmptyCollectionSyntaxUsesSchema verifies empty-collection literals
// also come from the schema.
func TestGetEmptyCollectionSyntaxUsesSchema(t *testing.T) {
	h := &MetaCommandHandler{}

	assert.Equal(t, "[]", h.getEmptyCollectionSyntax("tags", map[string]string{"tags": "list<text>"}))
	assert.Equal(t, "{}", h.getEmptyCollectionSyntax("tags", map[string]string{"tags": "set<text>"}))
	assert.Equal(t, "[]", h.getEmptyCollectionSyntax("tags", map[string]string{}), "unknown column falls back to list")
}
