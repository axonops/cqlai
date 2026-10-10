package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

// TestWideCharactersKeepTheBordersInLine: a CJK character or an emoji takes
// two screen columns. Widths and padding counted runes, so a row holding
// one was drawn wider than the header and the rows above and below it.
func TestWideCharactersKeepTheBordersInLine(t *testing.T) {
	data := [][]string{{"id", "name"}, {"1", "日本語"}, {"2", "ascii"}, {"3", "🍣 sushi"}}

	sameWidth := func(t *testing.T, what, table string) {
		t.Helper()
		var widths []int
		for _, line := range strings.Split(strings.TrimRight(table, "\n"), "\n") {
			if strings.HasPrefix(line, "(") || strings.TrimSpace(line) == "" {
				continue // the row count
			}
			widths = append(widths, lipgloss.Width(line))
		}
		for _, w := range widths {
			assert.Equal(t, widths[0], w, "%s:\n%s", what, table)
		}
	}

	sameWidth(t, "batch", FormatASCIITable(data))

	m := helpModel()
	sameWidth(t, "results", strings.Join(m.buildFullTable(data, CalculateColumnWidths(data)), "\n"))
}
