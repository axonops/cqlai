package ui

import (
	"strings"
	"unicode"
)

// wrapHistoryContent wraps the full history content to fit the viewport width
func (m *MainModel) wrapHistoryContent(width int) string {
	if m.fullHistoryContent == "" || width <= 0 {
		return ""
	}
	
	// Split content into lines first
	lines := strings.Split(m.fullHistoryContent, "\n")
	var wrappedLines []string
	
	for _, line := range lines {
		// Skip empty lines
		if line == "" {
			wrappedLines = append(wrappedLines, "")
			continue
		}
		
		// Wrap long lines
		wrapped := wrapHistoryLine(line, width)
		wrappedLines = append(wrappedLines, wrapped...)
	}
	
	return strings.Join(wrappedLines, "\n")
}

// wrapHistoryLine wraps a single line to fit within the given width
func wrapHistoryLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	
	// Handle ANSI escape codes by measuring visible length
	visibleLen := visibleLength(line)
	if visibleLen <= width {
		return []string{line}
	}
	
	var result []string
	currentLine := ""
	currentVisibleLen := 0
	inAnsiCode := false
	ansiBuffer := ""
	
	for _, r := range line {
		if r == '\x1b' {
			inAnsiCode = true
			ansiBuffer = string(r)
			continue
		}
		
		if inAnsiCode {
			ansiBuffer += string(r)
			if r == 'm' {
				// End of ANSI code
				currentLine += ansiBuffer
				ansiBuffer = ""
				inAnsiCode = false
			}
			continue
		}
		
		// Check if adding this character would exceed the width
		if currentVisibleLen >= width {
			// Try to break at a word boundary
			lastSpace := strings.LastIndexFunc(currentLine, unicode.IsSpace)
			if lastSpace > 0 && currentVisibleLen - visibleLength(currentLine[:lastSpace]) < width/2 {
				// Break at the last space if it's not too far back
				result = append(result, currentLine[:lastSpace])
				currentLine = strings.TrimSpace(currentLine[lastSpace:]) + string(r)
				currentVisibleLen = visibleLength(currentLine)
			} else {
				// No good break point, just break at the width
				result = append(result, currentLine)
				currentLine = string(r)
				currentVisibleLen = 1
			}
		} else {
			currentLine += string(r)
			currentVisibleLen++
		}
	}
	
	if currentLine != "" {
		result = append(result, currentLine)
	}
	
	return result
}

// visibleLength calculates the visible length of a string, ignoring ANSI escape codes
func visibleLength(s string) int {
	length := 0
	inAnsiCode := false
	
	for _, r := range s {
		if r == '\x1b' {
			inAnsiCode = true
			continue
		}
		
		if inAnsiCode {
			if r == 'm' {
				inAnsiCode = false
			}
			continue
		}
		
		length++
	}
	
	return length
}

// updateHistoryWrapping re-wraps the history content for the current viewport width
func (m *MainModel) updateHistoryWrapping() {
	if m.historyViewport.Width > 0 {
		wrapped := m.wrapHistoryContent(m.historyViewport.Width)
		m.historyViewport.SetContent(wrapped)
	}
}