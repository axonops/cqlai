package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/db"
)

// DESCRIBE TABLES as rows.
//
// It was written out twice, and a virtual table listed through either said
// its compaction was "Unknown" and its gc_grace_seconds 0. Neither is unknown
// or zero: a virtual table is computed when it is asked, so there is nothing
// on disk for compaction, compression or a grace period to apply to.

// TestAVirtualTableSaysItIsVirtual rather than making up a compaction.
func TestAVirtualTableSaysItIsVirtual(t *testing.T) {
	p := &CommandParser{}

	rows := p.tableListRows([]db.TableListInfo{{
		Name:           "clients",
		Virtual:        true,
		PartitionKeys:  []string{"address"},
		ClusteringKeys: []string{"port"},
	}})

	require.Len(t, rows, 2)
	assert.Equal(t, []string{"clients", "address, port", "virtual", "-", "-"}, rows[1])
}

// TestAStoredTableIsListedAsBefore: the change must not touch the tables that
// were right already.
func TestAStoredTableIsListedAsBefore(t *testing.T) {
	p := &CommandParser{}

	rows := p.tableListRows([]db.TableListInfo{{
		Name:          "users",
		PartitionKeys: []string{"id"},
		GcGrace:       864000,
		Compaction:    map[string]string{"class": "org.apache.cassandra.db.compaction.SizeTieredCompactionStrategy"},
	}})

	require.Len(t, rows, 2)
	assert.Equal(t, "users", rows[1][0])
	assert.Equal(t, "id", rows[1][1])
	assert.NotEqual(t, "virtual", rows[1][2])
	assert.Equal(t, "10d", rows[1][4])
}

// TestBothListsShareTheHeader: one function, so the two ways of asking cannot
// grow apart again.
func TestBothListsShareTheHeader(t *testing.T) {
	p := &CommandParser{}

	assert.Equal(t, []string{"Table", "Primary Key", "Compaction", "Compression", "GC Grace"},
		p.tableListRows(nil)[0])
}
