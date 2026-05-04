package textmode

import (
	"fmt"
	"io"

	"github.com/axonops/cqlai/internal/db"
)

// PrintBanner writes a cqlsh-style startup banner to w.
// It attempts to read the Cassandra release_version from system.local; if that
// fails the version segment is omitted silently.
func PrintBanner(w io.Writer, sess *db.Session, version string) {
	cassVersion := cassandraVersion(sess)

	if cassVersion != "" {
		fmt.Fprintf(w, "cqlai %s | Connected to Cassandra %s | Use HELP for help.\n\n",
			version, cassVersion)
	} else {
		fmt.Fprintf(w, "cqlai %s | Use HELP for help.\n\n", version)
	}
}

// cassandraVersion attempts to read the release_version from system.local.
// Returns an empty string if the query fails for any reason.
func cassandraVersion(sess *db.Session) string {
	if sess == nil {
		return ""
	}
	iter := sess.Query("SELECT release_version FROM system.local").Iter()
	var v string
	if iter.Scan(&v) {
		_ = iter.Close()
		return v
	}
	_ = iter.Close()
	return ""
}
