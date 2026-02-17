package main

import "strings"

// WrapText splits text into multiple lines based on max characters per line.
// It tries to break at word boundaries when possible and respects existing newlines.
// Returns at most maxLines lines (default 3 if maxLines <= 0).
func WrapText(text string, maxCharsPerLine int, maxLines int) []string {
	if maxLines <= 0 {
		maxLines = 3
	}

	// Handle existing newlines first
	existingLines := strings.Split(text, "\n")
	var result = []string{}

	for _, line := range existingLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// If line fits, add it as-is
		if len(line) <= maxCharsPerLine {
			result = append(result, line)
			continue
		}

		// Need to wrap this line
		wrapped := wrapLine(line, maxCharsPerLine)
		result = append(result, wrapped...)
	}

	// Truncate to maxLines if needed
	if len(result) > maxLines {
		result = result[:maxLines]
		// Add ellipsis to last line if truncated
		if maxCharsPerLine > 3 && len(result[maxLines-1]) > maxCharsPerLine-3 {
			result[maxLines-1] = result[maxLines-1][:maxCharsPerLine-3] + "..."
		} else if maxCharsPerLine > 0 && len(result[maxLines-1]) > maxCharsPerLine {
			// For very small widths, just truncate to max chars
			result[maxLines-1] = result[maxLines-1][:maxCharsPerLine]
		}
	}

	return result
}

// wrapLine wraps a single line at word boundaries
func wrapLine(line string, maxChars int) []string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{}
	}

	var lines = []string{}
	currentLine := ""

	for _, word := range words {
		// If word itself is longer than maxChars, break it
		if len(word) > maxChars {
			// Finish current line if any
			if currentLine != "" {
				lines = append(lines, strings.TrimSpace(currentLine))
				currentLine = ""
			}
			// Break long word into chunks
			for len(word) > maxChars {
				lines = append(lines, word[:maxChars])
				word = word[maxChars:]
			}
			if word != "" {
				currentLine = word
			}
			continue
		}

		// Try adding word to current line
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= maxChars {
			currentLine = testLine
		} else {
			// Word doesn't fit, start new line
			if currentLine != "" {
				lines = append(lines, strings.TrimSpace(currentLine))
			}
			currentLine = word
		}
	}

	// Add remaining text
	if currentLine != "" {
		lines = append(lines, strings.TrimSpace(currentLine))
	}

	return lines
}
