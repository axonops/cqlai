package ui

import (
	"regexp"
	"strings"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// scrollLineHorizontally scrolls a line horizontally while preserving ANSI codes
func scrollLineHorizontally(line string, offset, width int) string {
	if offset <= 0 && width <= 0 {
		return line
	}

	// Fast path for lines without ANSI codes
	if !strings.Contains(line, "\x1b[") {
		runes := []rune(line)
		if offset >= len(runes) {
			return ""
		}
		end := offset + width
		if end > len(runes) {
			end = len(runes)
		}
		return string(runes[offset:end])
	}

	// Find all ANSI codes and their positions
	matches := ansiPattern.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		// No ANSI codes found, treat as plain text
		runes := []rune(line)
		if offset >= len(runes) {
			return ""
		}
		end := offset + width
		if end > len(runes) {
			end = len(runes)
		}
		return string(runes[offset:end])
	}

	// Build arrays instead of map for better performance
	// Pre-allocate with reasonable size
	visualPositions := make([]int, 0, len(line)/2)
	actualPos := 0
	visualPos := 0
	matchIdx := 0

	for actualPos < len(line) {
		// Check if we're at an ANSI code
		if matchIdx < len(matches) && actualPos == matches[matchIdx][0] {
			// Skip the ANSI code
			actualPos = matches[matchIdx][1]
			matchIdx++
		} else {
			visualPositions = append(visualPositions, actualPos)
			visualPos++

			// Handle multi-byte characters
			r := []rune(line[actualPos:])
			if len(r) > 0 {
				actualPos += len(string(r[0]))
			} else {
				actualPos++
			}
		}
	}

	// If offset is beyond content, return empty
	if offset >= len(visualPositions) {
		return ""
	}

	// Calculate the visual end position
	endVisual := offset + width
	if endVisual > len(visualPositions) {
		endVisual = len(visualPositions)
	}

	// Find the actual positions for start and end
	startActual := visualPositions[offset]
	endActual := len(line)
	if endVisual < len(visualPositions) {
		endActual = visualPositions[endVisual]
	}

	// Extract the substring
	result := line[startActual:endActual]

	// Add reset at the end if we have any ANSI codes to prevent bleeding to next line
	if strings.Contains(result, "\x1b[") && !strings.HasSuffix(result, "\x1b[0m") {
		result += "\x1b[0m"
	}

	return result
}

// applyHorizontalScrollWithANSI applies horizontal scrolling while preserving ANSI codes
func applyHorizontalScrollWithANSI(lines []string, offset, width int) []string {
	// Quick check - if no offset and width is large, return as-is
	if offset == 0 && width >= 500 {
		return lines
	}

	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = scrollLineHorizontally(line, offset, width)
	}
	return result
}