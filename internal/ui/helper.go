package ui

import "strings"

// wrapLongLines wraps lines that exceed the specified width
func wrapLongLines(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		// Don't wrap lines that are part of CREATE TABLE definition
		if strings.Contains(line, "CREATE TABLE") || len(line) <= maxWidth {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Check if this is a CQL property line (starts with AND or has = sign)
		if strings.TrimSpace(line) != "" && (strings.Contains(line, " = ") || strings.HasPrefix(strings.TrimSpace(line), "AND ")) {
			// For CQL WITH clause properties, wrap intelligently
			currentLine := line
			indentLevel := len(line) - len(strings.TrimLeft(line, " "))
			indent := strings.Repeat(" ", indentLevel+4) // Extra indent for continuation

			for len(currentLine) > maxWidth {
				// Find a good break point
				breakPoint := -1

				// First try to break after a comma within a reasonable distance
				for i := maxWidth - 1; i > maxWidth-20 && i >= 0; i-- {
					if i < len(currentLine) && currentLine[i] == ',' {
						breakPoint = i + 1
						break
					}
				}

				// If no comma found, try to break at a space
				if breakPoint == -1 {
					for i := maxWidth - 1; i > maxWidth/2 && i >= 0; i-- {
						if i < len(currentLine) && currentLine[i] == ' ' {
							breakPoint = i + 1
							break
						}
					}
				}

				// If still no good break point, just break at maxWidth
				if breakPoint == -1 || breakPoint >= len(currentLine) {
					breakPoint = maxWidth
				}

				// Add the wrapped part
				wrappedLines = append(wrappedLines, currentLine[:breakPoint])

				// Continue with the rest, adding indentation
				if breakPoint < len(currentLine) {
					currentLine = indent + strings.TrimSpace(currentLine[breakPoint:])
				} else {
					currentLine = ""
					break
				}
			}

			// Add any remaining part
			if len(currentLine) > 0 {
				wrappedLines = append(wrappedLines, currentLine)
			}
		} else {
			// For other lines, just add them as-is
			wrappedLines = append(wrappedLines, line)
		}
	}

	return strings.Join(wrappedLines, "\n")
}
