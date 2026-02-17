package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestWrapText_ShortText(t *testing.T) {
	result := WrapText("Hello", 20, 3)

	expected := []string{"Hello"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWrapText_ExactFit(t *testing.T) {
	result := WrapText("Hello World", 11, 3)

	expected := []string{"Hello World"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWrapText_NeedsWrapping(t *testing.T) {
	result := WrapText("Hello World This Is Long", 15, 3)

	// Should break at word boundaries
	if len(result) < 2 {
		t.Errorf("Expected at least 2 lines, got %d", len(result))
	}

	// Each line should be <= 15 chars
	for i, line := range result {
		if len(line) > 15 {
			t.Errorf("Line %d too long (%d chars): %s", i, len(line), line)
		}
	}
}

func TestWrapText_ExistingNewlines(t *testing.T) {
	result := WrapText("Line One\nLine Two", 20, 3)

	expected := []string{"Line One", "Line Two"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWrapText_LongWord(t *testing.T) {
	result := WrapText("Supercalifragilisticexpialidocious", 10, 3)

	// Should break the word
	if len(result) < 2 {
		t.Errorf("Expected word to be broken into multiple lines")
	}

	// Each line should be exactly 10 chars (except possibly last)
	for i := 0; i < len(result)-1; i++ {
		if len(result[i]) != 10 {
			t.Errorf("Line %d should be 10 chars, got %d: %s", i, len(result[i]), result[i])
		}
	}
}

func TestWrapText_EmptyText(t *testing.T) {
	result := WrapText("", 20, 3)

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}
}

func TestWrapText_MaxLinesLimit(t *testing.T) {
	// Text that would wrap to many lines
	longText := "This is a very long text that will definitely need more than three lines to display properly"

	result := WrapText(longText, 15, 3)

	if len(result) > 3 {
		t.Errorf("Expected max 3 lines, got %d", len(result))
	}

	// Should have ellipsis on last line when truncated
	if len(result) == 3 && len(longText) > 45 {
		lastLine := result[2]
		if !containsEllipsis(lastLine) {
			t.Errorf("Expected ellipsis in last line when truncated, got: %s", lastLine)
		}
	}
}

func TestWrapText_MultipleWords(t *testing.T) {
	result := WrapText("Sprint Planning Daily Standup", 15, 3)

	// Check that words are preserved
	combined := ""
	for _, line := range result {
		combined += line + " "
	}

	// Should contain all original words
	if !containsWord(combined, "Sprint") || !containsWord(combined, "Planning") ||
		!containsWord(combined, "Daily") || !containsWord(combined, "Standup") {
		t.Errorf("Not all words preserved in wrapping: %v", result)
	}
}

func TestWrapText_DefaultMaxLines(t *testing.T) {
	longText := "Line1 Line2 Line3 Line4 Line5 Line6 Line7 Line8"

	// maxLines = 0 should default to 3
	result := WrapText(longText, 10, 0)

	if len(result) > 3 {
		t.Errorf("Expected default max 3 lines, got %d", len(result))
	}
}

func TestWrapText_WithLeadingTrailingSpaces(t *testing.T) {
	result := WrapText("  Hello World  ", 20, 3)

	expected := []string{"Hello World"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWrapText_MultipleSpaces(t *testing.T) {
	result := WrapText("Hello    World", 20, 3)

	// strings.Fields collapses multiple spaces, which is expected behavior
	if len(result) != 1 {
		t.Errorf("Expected 1 line, got %d", len(result))
	}

	// Result should contain both words
	if !strings.Contains(result[0], "Hello") || !strings.Contains(result[0], "World") {
		t.Errorf("Expected 'Hello' and 'World' in result, got %v", result)
	}
}

func TestWrapText_RealWorldExample(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxChars int
		maxLines int
		wantMax  int // maximum expected lines
	}{
		{
			name:     "Sprint Planning",
			text:     "Sprint Planning",
			maxChars: 15,
			maxLines: 3,
			wantMax:  2,
		},
		{
			name:     "Daily Standup Meeting",
			text:     "Daily Standup Meeting",
			maxChars: 15,
			maxLines: 3,
			wantMax:  3,
		},
		{
			name:     "Developing",
			text:     "Developing",
			maxChars: 15,
			maxLines: 3,
			wantMax:  1,
		},
		{
			name:     "Sprint Review",
			text:     "Sprint Review",
			maxChars: 15,
			maxLines: 3,
			wantMax:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.text, tt.maxChars, tt.maxLines)

			if len(result) > tt.wantMax {
				t.Errorf("Expected at most %d lines, got %d: %v", tt.wantMax, len(result), result)
			}

			// Verify no line exceeds maxChars
			for i, line := range result {
				if len(line) > tt.maxChars {
					t.Errorf("Line %d exceeds %d chars: %s (%d)", i, tt.maxChars, line, len(line))
				}
			}
		})
	}
}

// Helper functions
func containsEllipsis(s string) bool {
	return len(s) >= 3 && s[len(s)-3:] == "..."
}

func containsWord(s, word string) bool {
	return len(s) > 0 && len(word) > 0 &&
		(s == word || s[:len(word)] == word || s[len(s)-len(word):] == word ||
			strings.Contains(s, " "+word+" ") || strings.Contains(s, " "+word))
}
