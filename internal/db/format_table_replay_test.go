package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestADescriptionRecreatesTheSameTable: the description built for clusters
// too old to describe themselves left out static and the clustering order,
// and pasted a comment in unescaped - a quote in it ended the string, and
// what followed ran as CQL when the description was replayed.
func TestADescriptionRecreatesTheSameTable(t *testing.T) {
	info := &TableInfo{
		KeyspaceName: "shop",
		TableName:    "Events",
		Columns: []ColumnInfo{
			{Name: "p", DataType: "int", Kind: "partition_key", Position: 0},
			{Name: "c", DataType: "timestamp", Kind: "clustering", Position: 0, ClusteringOrder: "desc"},
			{Name: "d", DataType: "int", Kind: "clustering", Position: 1, ClusteringOrder: "asc"},
			{Name: "s", DataType: "text", Kind: "static", Position: -1},
			{Name: "v", DataType: "text", Kind: "regular", Position: -1},
		},
		PartitionKeys:  []string{"p"},
		ClusteringKeys: []string{"c", "d"},
		TableProps:     map[string]interface{}{"comment": "it's'; DROP TABLE shop.x; --"},
	}
	got := FormatTableCreateStatement(info, false)

	assert.Contains(t, got, `CREATE TABLE shop."Events" (`)
	assert.Contains(t, got, "    s text static,\n")
	assert.Contains(t, got, "PRIMARY KEY (p, c, d)")
	assert.Contains(t, got, ") WITH CLUSTERING ORDER BY (c DESC, d ASC)\n    AND comment = 'it''s''; DROP TABLE shop.x; --'")
}
