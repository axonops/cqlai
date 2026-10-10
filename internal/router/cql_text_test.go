package router

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestACommentEndsAtTheEndOfItsLine: what follows on the next line is still
// the statement. A comment used to take the rest of the input with it, so a
// DELETE lost the clustering condition on its second line and deleted the
// whole partition.
func TestACommentEndsAtTheEndOfItsLine(t *testing.T) {
	for in, want := range map[string]string{
		"DELETE FROM t WHERE pk = 1 -- one row\n  AND ck = 2;":        "DELETE FROM t WHERE pk = 1 \n  AND ck = 2;",
		"UPDATE acct SET bal = 5 WHERE id = 1 // guard\nIF bal = 10;": "UPDATE acct SET bal = 5 WHERE id = 1 \nIF bal = 10;",
		"SELECT * FROM t -- note\rWHERE k = 1;":                       "SELECT * FROM t \rWHERE k = 1;",
		"SELECT * FROM t; -- trailing":                                "SELECT * FROM t;",
		"SELECT /* inline */ * FROM t;":                               "SELECT   * FROM t;",
		"SELECT * FROM t /* unclosed":                                 "SELECT * FROM t",
	} {
		assert.Equal(t, want, StripComments(in), "%q", in)
	}
}

// TestACommentMarkInsideTextIsText: in a string, a quoted name or a $$ body
// - a function's Java, with comments of its own - it is not a comment.
func TestACommentMarkInsideTextIsText(t *testing.T) {
	for _, in := range []string{
		"INSERT INTO t (k, v) VALUES (1, 'a--b');",
		"INSERT INTO t (k, v) VALUES (1, 'it''s -- not');",
		`SELECT "odd--name" FROM t;`,
		"CREATE FUNCTION f (x int) RETURNS NULL ON NULL INPUT RETURNS int LANGUAGE java AS $$ // java\n return x; $$;",
		"INSERT INTO t (k, v) VALUES (1, 'http://example.com');",
	} {
		assert.Equal(t, in, StripComments(in), "%q", in)
	}
}

// TestAStatementEndsAtASemicolonThatIsItsOwn.
func TestAStatementEndsAtASemicolonThatIsItsOwn(t *testing.T) {
	for text, complete := range map[string]bool{
		"SELECT * FROM t;":                              true,
		"SELECT * FROM t":                               false,
		"SELECT * FROM t; -- done":                      true,
		"INSERT INTO t (k, v) VALUES (1, 'a;":           false, // the string is still open
		"INSERT INTO t (k, v) VALUES (1, 'a;b');":       true,
		"CREATE FUNCTION f (x int) AS $$ return x;":     false, // the body is still open
		"CREATE FUNCTION f (x int) AS $$ return x; $$;": true,
		"SELECT * FROM t /* ; ":                         false,

		// A BATCH ends at APPLY BATCH, not at the statements inside it.
		"BEGIN BATCH\n  INSERT INTO t (k) VALUES (1);":                                    false,
		"BEGIN BATCH\n  INSERT INTO t (k) VALUES (1);\n  UPDATE t SET v = 1 WHERE k = 2;": false,
		"BEGIN BATCH\n  INSERT INTO t (k) VALUES (1);\nAPPLY BATCH;":                      true,
		"begin unlogged batch insert into t (k) values (1); apply batch;":                 true,
		"BEGIN COUNTER BATCH UPDATE c SET n = n + 1 WHERE k = 1;":                         false,
	} {
		assert.Equal(t, complete, StatementComplete(text), "%q", text)
	}
}

func TestStatementsAreSplitAsTheShellSplitsThem(t *testing.T) {
	text := `-- a comment; with a semicolon
INSERT INTO ks.t (id, s) VALUES (1, 'a;b'); /* c; d */
BEGIN BATCH
  INSERT INTO ks.t (id) VALUES (2);
  INSERT INTO ks.t (id) VALUES (3);
APPLY BATCH;
CREATE FUNCTION ks.f (x int) RETURNS NULL ON NULL INPUT RETURNS int LANGUAGE java AS $$ return x; $$;
SELECT * FROM ks.t
`
	got := SplitStatements(text)
	require.Len(t, got, 4, "%q", got)
	assert.Equal(t, "-- a comment; with a semicolon\nINSERT INTO ks.t (id, s) VALUES (1, 'a;b');", got[0])
	assert.True(t, strings.HasPrefix(got[1], "/* c; d */\nBEGIN BATCH"), got[1])
	assert.True(t, strings.HasSuffix(got[1], "APPLY BATCH;"), got[1])
	assert.True(t, strings.HasSuffix(got[2], "$$ return x; $$;"), got[2])
	assert.Equal(t, "SELECT * FROM ks.t", got[3])

	assert.Empty(t, SplitStatements("  -- only a comment\n /* and another */ ;"))
}

// TestSplittingABigFileTakesNoTime: the splitter before this upper-cased the
// rest of the file at every B and A, so a big file took minutes.
func TestSplittingABigFileTakesNoTime(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 50000; i++ {
		b.WriteString("INSERT INTO ks.a_table (a, b) VALUES ('BATCH', 'APPLY');\n")
	}
	start := time.Now()
	got := SplitStatements(b.String())
	assert.Len(t, got, 50000)
	assert.Less(t, time.Since(start), 2*time.Second)
}
