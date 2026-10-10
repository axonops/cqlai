package parquet

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoPartitionFileIsWrittenOver: more partitions than files open at once,
// rows arriving for each in turn, so each partition's file is closed to make
// room and opened again. Each time is a new file: every row is on disk.
func TestNoPartitionFileIsWrittenOver(t *testing.T) {
	for _, maxOpen := range []int{0, 2} { // 0 is the default
		t.Run(fmt.Sprintf("open=%d", maxOpen), func(t *testing.T) {
			base := t.TempDir()
			opts := DefaultPartitionedOptions()
			opts.PartitionColumns = []string{"region"}
			opts.MaxOpenFiles = maxOpen
			w, err := NewPartitionedParquetWriter(base, []string{"region", "id"}, []string{"text", "int"}, opts)
			require.NoError(t, err)

			const regions, rounds = 12, 3
			for round := 0; round < rounds; round++ {
				for r := 0; r < regions; r++ {
					require.NoError(t, w.WriteRows([]map[string]interface{}{
						{"region": fmt.Sprintf("r%02d", r), "id": round*regions + r},
					}))
				}
			}
			require.NoError(t, w.Close())

			files, err := filepath.Glob(filepath.Join(base, "*", "*.parquet"))
			require.NoError(t, err)
			ids := map[int32]bool{}
			for _, f := range files {
				rd, err := NewParquetReader(f)
				require.NoError(t, err, f)
				rows, err := rd.ReadAll()
				require.NoError(t, err, f)
				_ = rd.Close()
				for _, row := range rows {
					ids[row["id"].(int32)] = true
				}
			}
			assert.Len(t, ids, regions*rounds, "every row is in a file")

			entries, err := os.ReadDir(filepath.Join(base, "region=r00"))
			require.NoError(t, err)
			assert.NotEmpty(t, entries)
		})
	}
}
