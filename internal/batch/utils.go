package batch

import (
	"github.com/axonops/cqlai/internal/router"
)

// stripComments removes the comments from CQL, as the shell does: one reading
// of where a comment ends, for every way a statement arrives.
func stripComments(input string) string { return router.StripComments(input) }

// splitStatements splits CQL into statements, as the shell does.
func splitStatements(content string) []string { return router.SplitStatements(content) }
