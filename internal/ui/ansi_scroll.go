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
	
	// Find all ANSI codes and their positions
	matches := ansiPattern.FindAllStringIndex(line, -1)
	
	// Build a map of visual position to actual position
	visualToActual := make(map[int]int)
	actualPos := 0
	visualPos := 0
	
	for actualPos < len(line) {
		// Check if we're at an ANSI code
		isAnsi := false
		for _, match := range matches {
			if actualPos == match[0] {
				// Skip the ANSI code
				actualPos = match[1]
				isAnsi = true
				break
			}
		}
		
		if !isAnsi && actualPos < len(line) {
			visualToActual[visualPos] = actualPos
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
	if offset >= visualPos {
		return ""
	}
	
	// Calculate the visual end position
	endVisual := offset + width
	if endVisual > visualPos {
		endVisual = visualPos
	}
	
	// Find the actual positions for start and end
	startActual := 0
	if pos, ok := visualToActual[offset]; ok {
		startActual = pos
	}
	
	endActual := len(line)
	if pos, ok := visualToActual[endVisual]; ok {
		endActual = pos
	}
	
	// Extract the substring without including prior ANSI codes
	// This prevents color bleeding between lines
	result := line[startActual:endActual]
	
	// Check if we need to add a reset at the beginning if we're starting mid-style
	hasAnsiInResult := ansiPattern.MatchString(result)
	
	// Add reset at the end if we have any ANSI codes to prevent bleeding to next line
	if hasAnsiInResult && !strings.HasSuffix(result, "\x1b[0m") {
		result += "\x1b[0m"
	}
	
	return result
}

// applyHorizontalScrollWithANSI applies horizontal scrolling while preserving ANSI codes
func applyHorizontalScrollWithANSI(lines []string, offset, width int) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = scrollLineHorizontally(line, offset, width)
	}
	return result
}