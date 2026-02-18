package main

import (
	"strings"
	"testing"
)

// Test parseNumberOrFraction helper function

func TestParseNumberOrFraction_Integer(t *testing.T) {
	result, err := parseNumberOrFraction("2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != 2.0 {
		t.Errorf("Expected 2.0, got %.1f", result)
	}
}

func TestParseNumberOrFraction_Decimal(t *testing.T) {
	result, err := parseNumberOrFraction("1.5")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != 1.5 {
		t.Errorf("Expected 1.5, got %.1f", result)
	}
}

func TestParseNumberOrFraction_SimpleFraction(t *testing.T) {
	result, err := parseNumberOrFraction("1/2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != 0.5 {
		t.Errorf("Expected 0.5, got %.1f", result)
	}
}

func TestParseNumberOrFraction_ThreeQuarters(t *testing.T) {
	result, err := parseNumberOrFraction("3/4")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != 0.75 {
		t.Errorf("Expected 0.75, got %.2f", result)
	}
}

func TestParseNumberOrFraction_WithWhitespace(t *testing.T) {
	result, err := parseNumberOrFraction(" 1 / 2 ")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != 0.5 {
		t.Errorf("Expected 0.5, got %.1f", result)
	}
}

func TestParseNumberOrFraction_DecimalInFraction(t *testing.T) {
	result, err := parseNumberOrFraction("1.5/2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != 0.75 {
		t.Errorf("Expected 0.75, got %.2f", result)
	}
}

func TestParseNumberOrFraction_NegativeFraction(t *testing.T) {
	result, err := parseNumberOrFraction("-1/2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != -0.5 {
		t.Errorf("Expected -0.5, got %.1f", result)
	}
}

func TestParseNumberOrFraction_DivisionByZero(t *testing.T) {
	_, err := parseNumberOrFraction("1/0")
	if err == nil {
		t.Fatal("Expected division by zero error")
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Errorf("Expected 'division by zero' error, got: %v", err)
	}
}

func TestParseNumberOrFraction_InvalidFormat_MultipleSlashes(t *testing.T) {
	_, err := parseNumberOrFraction("1/2/3")
	if err == nil {
		t.Fatal("Expected invalid format error")
	}
	if !strings.Contains(err.Error(), "must be 'numerator/denominator'") {
		t.Errorf("Expected format error, got: %v", err)
	}
}

func TestParseNumberOrFraction_InvalidFormat_Incomplete(t *testing.T) {
	_, err := parseNumberOrFraction("1/")
	if err == nil {
		t.Fatal("Expected error for incomplete fraction")
	}
}

func TestParseNumberOrFraction_InvalidFormat_NonNumeric(t *testing.T) {
	_, err := parseNumberOrFraction("abc/def")
	if err == nil {
		t.Fatal("Expected error for non-numeric fraction")
	}
}

func TestParseNumberOrFraction_EmptyString(t *testing.T) {
	_, err := parseNumberOrFraction("")
	if err == nil {
		t.Fatal("Expected error for empty string")
	}
}

// Integration tests

func TestParseDiagramSpec_ValidInput(t *testing.T) {
	text := `
1: 1,2: Sprint Planning
2: 3,2: Daily
3: 4,1: Developing
---
1 -> 2
2 -> 3
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check boxes
	if len(spec.Boxes) != 3 {
		t.Errorf("Expected 3 boxes, got %d", len(spec.Boxes))
	}

	// Check first box
	if spec.Boxes[0].ID != "1" {
		t.Errorf("Expected box ID 1, got %s", spec.Boxes[0].ID)
	}
	if spec.Boxes[0].GridX != 1 {
		t.Errorf("Expected GridX 1, got %d", spec.Boxes[0].GridX)
	}
	if spec.Boxes[0].GridY != 2 {
		t.Errorf("Expected GridY 2, got %d", spec.Boxes[0].GridY)
	}
	if spec.Boxes[0].Label != "Sprint Planning" {
		t.Errorf("Expected label 'Sprint Planning', got %s", spec.Boxes[0].Label)
	}

	// Check arrows
	if len(spec.Arrows) != 2 {
		t.Errorf("Expected 2 arrows, got %d", len(spec.Arrows))
	}

	// Check first arrow
	if spec.Arrows[0].FromID != "1" {
		t.Errorf("Expected FromID 1, got %s", spec.Arrows[0].FromID)
	}
	if spec.Arrows[0].ToID != "2" {
		t.Errorf("Expected ToID 2, got %s", spec.Arrows[0].ToID)
	}
}

func TestParseDiagramSpec_EmptyInput(t *testing.T) {
	text := ""

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 0 {
		t.Errorf("Expected 0 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 0 {
		t.Errorf("Expected 0 arrows, got %d", len(spec.Arrows))
	}
}

func TestParseDiagramSpec_OnlyBoxes(t *testing.T) {
	text := `
1: 1,2: Box One
2: 3,4: Box Two
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 0 {
		t.Errorf("Expected 0 arrows, got %d", len(spec.Arrows))
	}
}

func TestParseDiagramSpec_OnlyArrows(t *testing.T) {
	text := `
---
1 -> 2
2 -> 3
`

	_, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatal("Expected error for arrows without boxes, got nil")
	}

	// Should fail on first arrow that references non-existent box
	expected := "arrow '1 -> 2' references non-existent box label '1'"
	if err.Error() != expected {
		t.Errorf("Error message = %q, want %q", err.Error(), expected)
	}
}

func TestParseDiagramSpec_MalformedBox(t *testing.T) {
	text := `
1: 1,2: Valid Box
invalid line
2: 3,4: Another Valid Box
---
1 -> 2
`

	_, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatal("Expected error for malformed box, got nil")
	}
}

func TestParseDiagramSpec_MalformedArrow(t *testing.T) {
	text := `
1: 1,2: Box One
2: 2,2: Box Two
3: 3,2: Box Three
---
1 -> 2
invalid arrow
2 -> 3
`

	_, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatal("Expected error for malformed arrow, got nil")
	}
}

func TestParseDiagramSpec_WithWhitespace(t *testing.T) {
	text := `

  1: 1,2: Sprint Planning

  2: 3,2: Daily

---

  1 -> 2

`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 arrow, got %d", len(spec.Arrows))
	}

	// Check that whitespace was trimmed from label
	if spec.Boxes[0].Label != "Sprint Planning" {
		t.Errorf("Expected trimmed label 'Sprint Planning', got '%s'", spec.Boxes[0].Label)
	}
}

func TestParseDiagramSpec_ComplexDiagram(t *testing.T) {
	text := `
1: 1,3: Start
2: 3,3: Process
3: 5,2: Decision
4: 7,3: End
5: 5,1: Alternative
---
1 -> 2
2 -> 3
3 -> 4
3 -> 5
5 -> 4
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 5 {
		t.Errorf("Expected 5 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 5 {
		t.Errorf("Expected 5 arrows, got %d", len(spec.Arrows))
	}
}

func TestParseDiagramSpec_AutoArrow_SingleChain(t *testing.T) {
	text := `
1: 1,2: Sprint Planning
2: >3,2: Daily Standup
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 auto-arrow, got %d", len(spec.Arrows))
	}

	// Check that auto-arrow was created correctly
	if spec.Arrows[0].FromID != "1" {
		t.Errorf("Expected auto-arrow FromID 1, got %s", spec.Arrows[0].FromID)
	}
	if spec.Arrows[0].ToID != "2" {
		t.Errorf("Expected auto-arrow ToID 2, got %s", spec.Arrows[0].ToID)
	}

	// Check that box coordinates were parsed correctly (without the >)
	if spec.Boxes[1].GridX != 3 {
		t.Errorf("Expected GridX 3, got %d", spec.Boxes[1].GridX)
	}
	if spec.Boxes[1].GridY != 2 {
		t.Errorf("Expected GridY 2, got %d", spec.Boxes[1].GridY)
	}
}

func TestParseDiagramSpec_AutoArrow_MultipleChain(t *testing.T) {
	text := `
1: 1,2: Sprint Planning
2: >3,2: Daily Standup
3: >4,1: Developing
4: >7,4: Sprint Review
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 4 {
		t.Errorf("Expected 4 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 3 {
		t.Errorf("Expected 3 auto-arrows, got %d", len(spec.Arrows))
	}

	// Check arrow chain: 1->2, 2->3, 3->4
	expectedArrows := []struct{ from, to string }{
		{"1", "2"},
		{"2", "3"},
		{"3", "4"},
	}

	for i, expected := range expectedArrows {
		if spec.Arrows[i].FromID != expected.from {
			t.Errorf("Arrow %d: expected FromID %s, got %s", i, expected.from, spec.Arrows[i].FromID)
		}
		if spec.Arrows[i].ToID != expected.to {
			t.Errorf("Arrow %d: expected ToID %s, got %s", i, expected.to, spec.Arrows[i].ToID)
		}
	}
}

func TestParseDiagramSpec_AutoArrow_FirstBoxError(t *testing.T) {
	text := `
1: >1,2: Sprint Planning
`

	spec, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatalf("Expected error for first box with auto-arrow prefix, got nil")
	}

	if spec != nil {
		t.Errorf("Expected nil spec on error, got %v", spec)
	}

	// Check error message mentions the box ID
	expectedErrSubstring := "first box"
	if !containsSubstring(err.Error(), expectedErrSubstring) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErrSubstring, err)
	}
}

func TestParseDiagramSpec_AutoArrow_MixedWithManual(t *testing.T) {
	text := `
1: 1,2: Sprint Planning
2: >3,2: Daily Standup
3: 4,1: Developing
4: 7,4: Sprint Review
---
1 -> 4
3 -> 4
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 4 {
		t.Errorf("Expected 4 boxes, got %d", len(spec.Boxes))
	}

	// Should have 1 auto-arrow (1->2) and 2 manual arrows (1->4, 3->4)
	if len(spec.Arrows) != 3 {
		t.Errorf("Expected 3 arrows (1 auto + 2 manual), got %d", len(spec.Arrows))
	}

	// First arrow should be the auto-arrow
	if spec.Arrows[0].FromID != "1" || spec.Arrows[0].ToID != "2" {
		t.Errorf("Expected first arrow to be auto-arrow 1->2, got %s->%s",
			spec.Arrows[0].FromID, spec.Arrows[0].ToID)
	}
}

func TestParseDiagramSpec_AutoArrow_AllowDuplicates(t *testing.T) {
	text := `
1: 1,2: Sprint Planning
2: >3,2: Daily Standup
---
1 -> 2
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should allow both auto-arrow and manual arrow between same boxes
	if len(spec.Arrows) != 2 {
		t.Errorf("Expected 2 arrows (duplicate allowed), got %d", len(spec.Arrows))
	}

	// Both arrows should be 1->2
	for i, arrow := range spec.Arrows {
		if arrow.FromID != "1" || arrow.ToID != "2" {
			t.Errorf("Arrow %d: expected 1->2, got %s->%s", i, arrow.FromID, arrow.ToID)
		}
	}
}

// Helper function to check if string contains substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseDiagramSpec_RelativeCoordinates_RelativeX(t *testing.T) {
	text := `
1: 1,2: Box A
2: +2,4: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	// Check Box B has correct resolved coordinates
	// GridX should be 1 + 2 = 3, GridY should be 4 (absolute)
	if spec.Boxes[1].GridX != 3 {
		t.Errorf("Expected Box B GridX=3 (1+2), got %d", spec.Boxes[1].GridX)
	}
	if spec.Boxes[1].GridY != 4 {
		t.Errorf("Expected Box B GridY=4 (absolute), got %d", spec.Boxes[1].GridY)
	}
}

func TestParseDiagramSpec_RelativeCoordinates_RelativeY(t *testing.T) {
	text := `
1: 3,2: Box A
2: 5,+1: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// GridX should be 5 (absolute), GridY should be 2 + 1 = 3
	if spec.Boxes[1].GridX != 5 {
		t.Errorf("Expected Box B GridX=5 (absolute), got %d", spec.Boxes[1].GridX)
	}
	if spec.Boxes[1].GridY != 3 {
		t.Errorf("Expected Box B GridY=3 (2+1), got %d", spec.Boxes[1].GridY)
	}
}

func TestParseDiagramSpec_RelativeCoordinates_BothRelative(t *testing.T) {
	text := `
1: 1,1: Box A
2: +2,+1: Box B
3: +2,+1: Box C
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 3 {
		t.Errorf("Expected 3 boxes, got %d", len(spec.Boxes))
	}

	// Box B: GridX = 1+2=3, GridY = 1+1=2
	if spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 2 {
		t.Errorf("Expected Box B at (3,2), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}

	// Box C: GridX = 3+2=5, GridY = 2+1=3
	if spec.Boxes[2].GridX != 5 || spec.Boxes[2].GridY != 3 {
		t.Errorf("Expected Box C at (5,3), got (%d,%d)", spec.Boxes[2].GridX, spec.Boxes[2].GridY)
	}
}

func TestParseDiagramSpec_RelativeCoordinates_Negative(t *testing.T) {
	text := `
1: 5,3: Box A
2: +1,-1: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Box B: GridX = 5+1=6, GridY = 3-1=2
	if spec.Boxes[1].GridX != 6 || spec.Boxes[1].GridY != 2 {
		t.Errorf("Expected Box B at (6,2), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}
}

func TestParseDiagramSpec_RelativeCoordinates_FirstBoxError(t *testing.T) {
	text := `
1: +2,3: Box A
`

	spec, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatalf("Expected error for first box with relative coordinates, got nil")
	}

	if spec != nil {
		t.Errorf("Expected nil spec on error, got %v", spec)
	}

	if !containsSubstring(err.Error(), "first box") {
		t.Errorf("Expected error to mention 'first box', got: %v", err)
	}
}

func TestParseDiagramSpec_RelativeCoordinates_NegativeResultError(t *testing.T) {
	text := `
1: 2,3: Box A
2: -5,1: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatalf("Expected error for negative GridX result (2-5=-3), got nil")
	}

	if spec != nil {
		t.Errorf("Expected nil spec on error, got %v", spec)
	}

	if !containsSubstring(err.Error(), "invalid value") {
		t.Errorf("Expected error to mention 'invalid value', got: %v", err)
	}
}

func TestParseDiagramSpec_RelativeCoordinates_WithAutoArrow(t *testing.T) {
	text := `
1: 1,1: Box A
2: >+2,+1: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	// Check relative coordinates resolved correctly
	if spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 2 {
		t.Errorf("Expected Box B at (3,2), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}

	// Check auto-arrow was created
	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 auto-arrow, got %d", len(spec.Arrows))
	}

	if spec.Arrows[0].FromID != "1" || spec.Arrows[0].ToID != "2" {
		t.Errorf("Expected arrow 1->2, got %s->%s", spec.Arrows[0].FromID, spec.Arrows[0].ToID)
	}
}

func TestParseDiagramSpec_RelativeCoordinates_MixedAbsoluteRelative(t *testing.T) {
	text := `
1: 1,1: Box A
2: +2,+1: Box B
3: 7,3: Box C
4: +1,+1: Box D
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 4 {
		t.Errorf("Expected 4 boxes, got %d", len(spec.Boxes))
	}

	// Box A: (1,1) - absolute
	// Box B: (1+2,1+1) = (3,2) - relative to A
	// Box C: (7,3) - absolute
	// Box D: (7+1,3+1) = (8,4) - relative to C

	expectedCoords := []struct{ x, y int }{
		{1, 1},
		{3, 2},
		{7, 3},
		{8, 4},
	}

	for i, expected := range expectedCoords {
		if spec.Boxes[i].GridX != expected.x || spec.Boxes[i].GridY != expected.y {
			t.Errorf("Box %d: expected (%d,%d), got (%d,%d)",
				i+1, expected.x, expected.y, spec.Boxes[i].GridX, spec.Boxes[i].GridY)
		}
	}
}

func TestParseCoordinate_Absolute(t *testing.T) {
	coord, err := parseCoordinate("5")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coord.IsRelative {
		t.Errorf("Expected absolute coordinate, got relative")
	}
	if coord.Value != 5 {
		t.Errorf("Expected value 5, got %d", coord.Value)
	}
}

func TestParseCoordinate_RelativePositive(t *testing.T) {
	coord, err := parseCoordinate("+2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !coord.IsRelative {
		t.Errorf("Expected relative coordinate, got absolute")
	}
	if coord.Value != 2 {
		t.Errorf("Expected value 2, got %d", coord.Value)
	}
}

func TestParseCoordinate_RelativeNegative(t *testing.T) {
	coord, err := parseCoordinate("-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !coord.IsRelative {
		t.Errorf("Expected relative coordinate, got absolute")
	}
	if coord.Value != -1 {
		t.Errorf("Expected value -1, got %d", coord.Value)
	}
}

func TestParseCoordinate_ShorthandZero(t *testing.T) {
	coord, err := parseCoordinate("0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !coord.IsRelative {
		t.Errorf("Expected relative coordinate (shorthand 0 = +0), got absolute")
	}
	if coord.Value != 0 {
		t.Errorf("Expected value 0, got %d", coord.Value)
	}
}

func TestParseCoordinate_ExplicitRelativeZero(t *testing.T) {
	coord, err := parseCoordinate("+0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !coord.IsRelative {
		t.Errorf("Expected relative coordinate, got absolute")
	}
	if coord.Value != 0 {
		t.Errorf("Expected value 0, got %d", coord.Value)
	}
}

func TestParseCoordinate_Invalid(t *testing.T) {
	_, err := parseCoordinate("abc")
	if err == nil {
		t.Errorf("Expected error for invalid coordinate, got nil")
	}
}

func TestParseDiagramSpec_ShorthandZero_XOnly(t *testing.T) {
	text := `
1: 1,1: Box A
2: +2,0: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	// Box B: GridX = 1+2=3, GridY = 1+0=1 (0 is shorthand for +0)
	if spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 1 {
		t.Errorf("Expected Box B at (3,1), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}
}

func TestParseDiagramSpec_ShorthandZero_YOnly(t *testing.T) {
	text := `
1: 3,2: Box A
2: 0,+1: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Box B: GridX = 3+0=3 (0 is shorthand for +0), GridY = 2+1=3
	if spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 3 {
		t.Errorf("Expected Box B at (3,3), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}
}

func TestParseDiagramSpec_ShorthandZero_Both(t *testing.T) {
	text := `
1: 5,3: Box A
2: 0,0: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Box B: GridX = 5+0=5, GridY = 3+0=3 (both 0 are shorthand for +0)
	if spec.Boxes[1].GridX != 5 || spec.Boxes[1].GridY != 3 {
		t.Errorf("Expected Box B at (5,3), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}
}

func TestParseDiagramSpec_ShorthandZero_WithAutoArrow(t *testing.T) {
	text := `
1: 1,1: Box A
2: >+2,0: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	// Box B: GridX = 1+2=3 (relative), GridY = 1+0=1 (0 is shorthand for +0)
	if spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 1 {
		t.Errorf("Expected Box B at (3,1), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}

	// Check auto-arrow was created
	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 auto-arrow, got %d", len(spec.Arrows))
	}

	if spec.Arrows[0].FromID != "1" || spec.Arrows[0].ToID != "2" {
		t.Errorf("Expected arrow 1->2, got %s->%s", spec.Arrows[0].FromID, spec.Arrows[0].ToID)
	}
}

func TestParseDiagramSpec_ShorthandZero_MixedWithRelative(t *testing.T) {
	text := `
1: 1,1: Box A
2: +2,0: Box B
3: 0,+1: Box C
4: +1,+1: Box D
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 4 {
		t.Errorf("Expected 4 boxes, got %d", len(spec.Boxes))
	}

	// Box A: (1,1) - absolute
	// Box B: (1+2,1+0) = (3,1) - using shorthand 0
	// Box C: (3+0,1+1) = (3,2) - using shorthand 0
	// Box D: (3+1,2+1) = (4,3) - regular relative

	expectedCoords := []struct{ x, y int }{
		{1, 1},
		{3, 1},
		{3, 2},
		{4, 3},
	}

	for i, expected := range expectedCoords {
		if spec.Boxes[i].GridX != expected.x || spec.Boxes[i].GridY != expected.y {
			t.Errorf("Box %d: expected (%d,%d), got (%d,%d)",
				i+1, expected.x, expected.y, spec.Boxes[i].GridX, spec.Boxes[i].GridY)
		}
	}
}

func TestParseDiagramSpec_ShorthandZero_FirstBoxError(t *testing.T) {
	text := `
1: 0,2: Box A
`

	spec, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatalf("Expected error for first box with relative coordinate (0), got nil")
	}

	if spec != nil {
		t.Errorf("Expected nil spec on error, got %v", spec)
	}

	if !containsSubstring(err.Error(), "first box") {
		t.Errorf("Expected error to mention 'first box', got: %v", err)
	}
}

func TestParseDiagramSpec_ShorthandZero_BackwardCompatibility(t *testing.T) {
	// Verify that "+0" still works exactly as before
	text := `
1: 1,1: Box A
2: +2,+0: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Box B: GridX = 1+2=3, GridY = 1+0=1
	if spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 1 {
		t.Errorf("Expected Box B at (3,1), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}
}

func TestParseDiagramSpec_Comments_Basic(t *testing.T) {
	text := `
# This is a comment
1: 1,1: Box A
# Another comment
2: 3,2: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes (comments ignored), got %d", len(spec.Boxes))
	}

	if spec.Boxes[0].Label != "Box A" {
		t.Errorf("Expected first box label 'Box A', got '%s'", spec.Boxes[0].Label)
	}
}

func TestParseDiagramSpec_Comments_WithWhitespace(t *testing.T) {
	text := `
  # Comment with leading spaces
	# Comment with leading tab
		  # Comment with mixed whitespace
1: 1,1: Box A
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box (comments ignored), got %d", len(spec.Boxes))
	}
}

func TestParseDiagramSpec_Comments_InArrowSection(t *testing.T) {
	text := `
1: 1,1: Box A
2: 3,2: Box B
---
# Comment in arrow section
1 -> 2
# Another comment
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 arrow (comments ignored), got %d", len(spec.Arrows))
	}
}

func TestParseDiagramSpec_Comments_WithAutoArrow(t *testing.T) {
	text := `
# Start of diagram
1: 1,1: Box A
# Using auto-arrow
2: >+2,0: Box B
# End
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 auto-arrow, got %d", len(spec.Arrows))
	}
}

func TestParseDiagramSpec_Comments_MultipleConsecutive(t *testing.T) {
	text := `
# Comment 1
# Comment 2
# Comment 3
1: 1,1: Box A
# Comment 4
# Comment 5
2: 3,2: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}
}

func TestParseDiagramSpec_Comments_JustHashSymbol(t *testing.T) {
	text := `
#
1: 1,1: Box A
#
2: 3,2: Box B
#
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}
}

func TestParseDiagramSpec_Comments_WithRelativeCoordinates(t *testing.T) {
	text := `
# Start position
1: 1,1: Box A
# Move right, stay in same row
2: +2,0: Box B
# Move down
3: 0,+1: Box C
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 3 {
		t.Errorf("Expected 3 boxes, got %d", len(spec.Boxes))
	}

	// Verify coordinates resolved correctly despite comments
	if spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 1 {
		t.Errorf("Expected Box B at (3,1), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}

	if spec.Boxes[2].GridX != 3 || spec.Boxes[2].GridY != 2 {
		t.Errorf("Expected Box C at (3,2), got (%d,%d)", spec.Boxes[2].GridX, spec.Boxes[2].GridY)
	}
}

func TestParseDiagramSpec_VariableBoxSize_FourValues(t *testing.T) {
	text := `
1: 1,1,3,2: Large Box
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.GridX != 1 {
		t.Errorf("Expected GridX=1, got %d", box.GridX)
	}
	if box.GridY != 1 {
		t.Errorf("Expected GridY=1, got %d", box.GridY)
	}
	if box.GridWidth != 3.0 {
		t.Errorf("Expected GridWidth=3, got %.1f", box.GridWidth)
	}
	if box.GridHeight != 2 {
		t.Errorf("Expected GridHeight=2, got %d", box.GridHeight)
	}
}

func TestParseDiagramSpec_VariableBoxSize_DefaultValues(t *testing.T) {
	text := `
1: 1,1: Default Size Box
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.GridWidth != 2.0 {
		t.Errorf("Expected default GridWidth=2, got %.1f", box.GridWidth)
	}
	if box.GridHeight != 1 {
		t.Errorf("Expected default GridHeight=1, got %d", box.GridHeight)
	}
}

func TestParseDiagramSpec_VariableBoxSize_WithRelativePosition(t *testing.T) {
	text := `
1: 1,1: Box A
2: +2,+1,3,2: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[1]
	// Position should be relative: (1+2, 1+1) = (3, 2)
	if box.GridX != 3 || box.GridY != 2 {
		t.Errorf("Expected position (3,2), got (%d,%d)", box.GridX, box.GridY)
	}
	// Size should be absolute
	if box.GridWidth != 3.0 {
		t.Errorf("Expected GridWidth=3, got %.1f", box.GridWidth)
	}
	if box.GridHeight != 2 {
		t.Errorf("Expected GridHeight=2, got %d", box.GridHeight)
	}
}

func TestParseDiagramSpec_VariableBoxSize_InvalidWidth(t *testing.T) {
	text := `
1: 1,1,0,1: Invalid Width
`

	spec, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatalf("Expected error for GridWidth=0, got nil")
	}

	if spec != nil {
		t.Errorf("Expected nil spec on error, got %v", spec)
	}

	if !containsSubstring(err.Error(), "GridWidth") {
		t.Errorf("Expected error to mention GridWidth, got: %v", err)
	}
}

func TestParseDiagramSpec_VariableBoxSize_InvalidHeight(t *testing.T) {
	text := `
1: 1,1,2,-1: Invalid Height
`

	spec, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatalf("Expected error for GridHeight=-1, got nil")
	}

	if spec != nil {
		t.Errorf("Expected nil spec on error, got %v", spec)
	}

	if !containsSubstring(err.Error(), "GridHeight") {
		t.Errorf("Expected error to mention GridHeight, got: %v", err)
	}
}

func TestParseDiagramSpec_VariableBoxSize_WithAutoArrow(t *testing.T) {
	text := `
1: 1,1,2,1: Box A
2: >3,1,3,2: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 auto-arrow, got %d", len(spec.Arrows))
	}

	// Verify Box B has custom size
	box := spec.Boxes[1]
	if box.GridWidth != 3.0 || box.GridHeight != 2 {
		t.Errorf("Expected Box B size (3,2), got (%.1f,%d)", box.GridWidth, box.GridHeight)
	}
}

func TestParseDiagramSpec_VariableBoxSize_Mixed(t *testing.T) {
	text := `
1: 1,1: Default
2: 3,1,3,1: Custom Width
3: 1,3,2,2: Custom Height
4: 4,3,4,3: Both Custom
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 4 {
		t.Errorf("Expected 4 boxes, got %d", len(spec.Boxes))
	}

	// Box 1: default (2,1)
	if spec.Boxes[0].GridWidth != 2.0 || spec.Boxes[0].GridHeight != 1 {
		t.Errorf("Box 1: expected (2,1), got (%.1f,%d)", spec.Boxes[0].GridWidth, spec.Boxes[0].GridHeight)
	}

	// Box 2: custom width (3,1)
	if spec.Boxes[1].GridWidth != 3.0 || spec.Boxes[1].GridHeight != 1 {
		t.Errorf("Box 2: expected (3,1), got (%.1f,%d)", spec.Boxes[1].GridWidth, spec.Boxes[1].GridHeight)
	}

	// Box 3: custom height (2,2)
	if spec.Boxes[2].GridWidth != 2.0 || spec.Boxes[2].GridHeight != 2 {
		t.Errorf("Box 3: expected (2,2), got (%.1f,%d)", spec.Boxes[2].GridWidth, spec.Boxes[2].GridHeight)
	}

	// Box 4: both custom (4,3)
	if spec.Boxes[3].GridWidth != 4.0 || spec.Boxes[3].GridHeight != 3 {
		t.Errorf("Box 4: expected (4,3), got (%.1f,%d)", spec.Boxes[3].GridWidth, spec.Boxes[3].GridHeight)
	}
}

func TestParseDiagramSpec_ThreeValueFormat_Basic(t *testing.T) {
	text := `
1: 1,1,3: Wide Box
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.GridX != 1 {
		t.Errorf("Expected GridX=1, got %d", box.GridX)
	}
	if box.GridY != 1 {
		t.Errorf("Expected GridY=1, got %d", box.GridY)
	}
	if box.GridWidth != 3.0 {
		t.Errorf("Expected GridWidth=3, got %.1f", box.GridWidth)
	}
	if box.GridHeight != 1 {
		t.Errorf("Expected default GridHeight=1, got %d", box.GridHeight)
	}
}

func TestParseDiagramSpec_ThreeValueFormat_WithRelativeCoordinates(t *testing.T) {
	text := `
1: 1,1: Box A
2: +2,+1,3: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[1]
	// Position should be relative: (1+2, 1+1) = (3, 2)
	if box.GridX != 3 || box.GridY != 2 {
		t.Errorf("Expected position (3,2), got (%d,%d)", box.GridX, box.GridY)
	}
	// Width should be 3, height should default to 1
	if box.GridWidth != 3.0 {
		t.Errorf("Expected GridWidth=3, got %.1f", box.GridWidth)
	}
	if box.GridHeight != 1 {
		t.Errorf("Expected default GridHeight=1, got %d", box.GridHeight)
	}
}

func TestParseDiagramSpec_ThreeValueFormat_WithAutoArrow(t *testing.T) {
	text := `
1: 1,1: Box A
2: >+2,0,4: Box B
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 auto-arrow, got %d", len(spec.Arrows))
	}

	// Verify Box B has custom width and default height
	box := spec.Boxes[1]
	if box.GridWidth != 4.0 {
		t.Errorf("Expected GridWidth=4, got %.1f", box.GridWidth)
	}
	if box.GridHeight != 1 {
		t.Errorf("Expected default GridHeight=1, got %d", box.GridHeight)
	}

	// Verify auto-arrow was created
	if spec.Arrows[0].FromID != "1" || spec.Arrows[0].ToID != "2" {
		t.Errorf("Expected arrow 1->2, got %s->%s", spec.Arrows[0].FromID, spec.Arrows[0].ToID)
	}
}

func TestParseDiagramSpec_ThreeValueFormat_Mixed(t *testing.T) {
	text := `
1: 1,1: Default (2,1)
2: 3,1,3: Three values (3,1)
3: 1,3,2,2: Four values (2,2)
4: 7,1,4,3: Four values (4,3)
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 4 {
		t.Errorf("Expected 4 boxes, got %d", len(spec.Boxes))
	}

	// Box 1: default (2,1)
	if spec.Boxes[0].GridWidth != 2.0 || spec.Boxes[0].GridHeight != 1 {
		t.Errorf("Box 1: expected (2,1), got (%.1f,%d)", spec.Boxes[0].GridWidth, spec.Boxes[0].GridHeight)
	}

	// Box 2: three-value format (3,1)
	if spec.Boxes[1].GridWidth != 3.0 || spec.Boxes[1].GridHeight != 1 {
		t.Errorf("Box 2: expected (3,1), got (%.1f,%d)", spec.Boxes[1].GridWidth, spec.Boxes[1].GridHeight)
	}

	// Box 3: four-value format (2,2)
	if spec.Boxes[2].GridWidth != 2.0 || spec.Boxes[2].GridHeight != 2 {
		t.Errorf("Box 3: expected (2,2), got (%.1f,%d)", spec.Boxes[2].GridWidth, spec.Boxes[2].GridHeight)
	}

	// Box 4: four-value format (4,3)
	if spec.Boxes[3].GridWidth != 4.0 || spec.Boxes[3].GridHeight != 3 {
		t.Errorf("Box 4: expected (4,3), got (%.1f,%d)", spec.Boxes[3].GridWidth, spec.Boxes[3].GridHeight)
	}
}

func TestParseDiagramSpec_ThreeValueFormat_InvalidWidth(t *testing.T) {
	text := `
1: 1,1,0: Invalid Width
`

	spec, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatalf("Expected error for GridWidth=0, got nil")
	}

	if spec != nil {
		t.Errorf("Expected nil spec on error, got %v", spec)
	}

	if !containsSubstring(err.Error(), "GridWidth") {
		t.Errorf("Expected error to mention GridWidth, got: %v", err)
	}
}

func TestParseDiagramSpec_ThreeValueFormat_NegativeWidth(t *testing.T) {
	text := `
1: 1,1,-2: Negative Width
`

	spec, err := ParseDiagramSpec(text, nil)
	if err == nil {
		t.Fatalf("Expected error for negative GridWidth, got nil")
	}

	if spec != nil {
		t.Errorf("Expected nil spec on error, got %v", spec)
	}

	if !containsSubstring(err.Error(), "GridWidth") {
		t.Errorf("Expected error to mention GridWidth, got: %v", err)
	}
}

func TestParseDiagramSpec_ThreeValueFormat_AllSizes(t *testing.T) {
	text := `
1: 1,1,1: Width 1
2: 3,1,2: Width 2
3: 6,1,3: Width 3
4: 10,1,5: Width 5
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 4 {
		t.Errorf("Expected 4 boxes, got %d", len(spec.Boxes))
	}

	expectedWidths := []float64{1.0, 2.0, 3.0, 5.0}
	for i, expectedWidth := range expectedWidths {
		if spec.Boxes[i].GridWidth != expectedWidth {
			t.Errorf("Box %d: expected width %.1f, got %.1f", i+1, expectedWidth, spec.Boxes[i].GridWidth)
		}
		if spec.Boxes[i].GridHeight != 1 {
			t.Errorf("Box %d: expected height 1, got %d", i+1, spec.Boxes[i].GridHeight)
		}
	}
}

func TestParseDiagramSpec_Styles_RedBorder(t *testing.T) {
	text := `
1: 1,1: Task, rb
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.Label != "Task" {
		t.Errorf("Expected label 'Task', got '%s'", box.Label)
	}
	if box.BorderColor != "#FF0000" {
		t.Errorf("Expected red border color #FF0000, got '%s'", box.BorderColor)
	}
	if box.BorderWidth != 3 {
		t.Errorf("Expected border width 3, got %d", box.BorderWidth)
	}
	if box.Color != "" {
		t.Errorf("Expected no background color, got '%s'", box.Color)
	}
}

func TestParseDiagramSpec_Styles_GrayBackground(t *testing.T) {
	text := `
1: 1,1: Task, g
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.Label != "Task" {
		t.Errorf("Expected label 'Task', got '%s'", box.Label)
	}
	if box.Color != "#D3D3D3" {
		t.Errorf("Expected gray background #D3D3D3, got '%s'", box.Color)
	}
	if box.BorderColor != "" {
		t.Errorf("Expected no border color, got '%s'", box.BorderColor)
	}
	if box.BorderWidth != 0 {
		t.Errorf("Expected border width 0, got %d", box.BorderWidth)
	}
}

func TestParseDiagramSpec_Styles_PurpleBackground(t *testing.T) {
	text := `
1: 1,1: Task, p
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.Label != "Task" {
		t.Errorf("Expected label 'Task', got '%s'", box.Label)
	}
	if box.Color != "#ecbae6" {
		t.Errorf("Expected purple background #ecbae6, got '%s'", box.Color)
	}
}

func TestParseDiagramSpec_Styles_Combined(t *testing.T) {
	text := `
1: 1,1: Task, rb-g
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.Label != "Task" {
		t.Errorf("Expected label 'Task', got '%s'", box.Label)
	}
	if box.BorderColor != "#FF0000" {
		t.Errorf("Expected red border color #FF0000, got '%s'", box.BorderColor)
	}
	if box.BorderWidth != 3 {
		t.Errorf("Expected border width 3, got %d", box.BorderWidth)
	}
	if box.Color != "#D3D3D3" {
		t.Errorf("Expected gray background #D3D3D3, got '%s'", box.Color)
	}
}

func TestParseDiagramSpec_Styles_NoStyle(t *testing.T) {
	text := `
1: 1,1: Task
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.Label != "Task" {
		t.Errorf("Expected label 'Task', got '%s'", box.Label)
	}
	if box.Color != "" {
		t.Errorf("Expected no background color, got '%s'", box.Color)
	}
	if box.BorderColor != "" {
		t.Errorf("Expected no border color, got '%s'", box.BorderColor)
	}
	if box.BorderWidth != 0 {
		t.Errorf("Expected border width 0, got %d", box.BorderWidth)
	}
}

func TestParseDiagramSpec_Styles_MultipleBoxes(t *testing.T) {
	text := `
1: 1,1: Task A, rb
2: 2,1: Task B, g
3: 3,1: Task C, p
4: 4,1: Task D, rb-g
5: 5,1: Task E
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 5 {
		t.Errorf("Expected 5 boxes, got %d", len(spec.Boxes))
	}

	// Box 1: rb
	if spec.Boxes[0].BorderColor != "#FF0000" || spec.Boxes[0].BorderWidth != 3 {
		t.Errorf("Box 1 should have red border")
	}

	// Box 2: g
	if spec.Boxes[1].Color != "#D3D3D3" {
		t.Errorf("Box 2 should have gray background")
	}

	// Box 3: p
	if spec.Boxes[2].Color != "#ecbae6" {
		t.Errorf("Box 3 should have purple background")
	}

	// Box 4: rb-g
	if spec.Boxes[3].BorderColor != "#FF0000" || spec.Boxes[3].BorderWidth != 3 || spec.Boxes[3].Color != "#D3D3D3" {
		t.Errorf("Box 4 should have red border and gray background")
	}

	// Box 5: no style
	if spec.Boxes[4].Color != "" || spec.Boxes[4].BorderColor != "" {
		t.Errorf("Box 5 should have no styles")
	}
}

func TestParseDiagramSpec_Styles_LightPurple(t *testing.T) {
	text := `
1: 1,1: Task, lp
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.Label != "Task" {
		t.Errorf("Expected label 'Task', got '%s'", box.Label)
	}
	if box.Color != "#f5dbf2" {
		t.Errorf("Expected light purple background #f5dbf2, got '%s'", box.Color)
	}
}

func TestParseDiagramSpec_Styles_LightPurpleCombined(t *testing.T) {
	text := `
1: 1,1: Task, lp-rb
`

	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.Color != "#f5dbf2" {
		t.Errorf("Expected light purple background #f5dbf2, got '%s'", box.Color)
	}
	if box.BorderColor != "#FF0000" || box.BorderWidth != 3 {
		t.Errorf("Expected red border, got color=%s width=%d", box.BorderColor, box.BorderWidth)
	}
}

func TestTouchLeftConnector_Valid(t *testing.T) {
	input := `1: 1,1: Box 1
2: |+2,0: Box 2`

	spec, err := ParseDiagramSpec(input, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Fatalf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	// Box 1 should not have TouchLeft
	if spec.Boxes[0].TouchLeft {
		t.Error("Box 1 should not have TouchLeft=true")
	}

	// Box 2 should have TouchLeft
	if !spec.Boxes[1].TouchLeft {
		t.Error("Box 2 should have TouchLeft=true")
	}

	// Box 2 should be at GridX=3, GridY=1 (1+2, 1+0)
	if spec.Boxes[1].GridX != 3 {
		t.Errorf("Box 2 GridX = %d, want 3", spec.Boxes[1].GridX)
	}
	if spec.Boxes[1].GridY != 1 {
		t.Errorf("Box 2 GridY = %d, want 1", spec.Boxes[1].GridY)
	}
}

func TestTouchLeftConnector_InvalidY(t *testing.T) {
	input := `1: 1,1: Box 1
2: |+2,+1: Box 2`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for Y coordinate != 0, got nil")
	}

	expected := "touch-left prefix '|' requires Y coordinate to be 0"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Error message should contain %q, got: %v", expected, err)
	}
}

func TestTouchLeftConnector_InvalidY_Negative(t *testing.T) {
	input := `1: 1,2: Box 1
2: |+2,-1: Box 2`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for Y coordinate != 0, got nil")
	}

	expected := "touch-left prefix '|' requires Y coordinate to be 0"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Error message should contain %q, got: %v", expected, err)
	}
}

func TestTouchLeftConnector_InvalidX_Absolute(t *testing.T) {
	input := `1: 1,1: Box 1
2: |2,0: Box 2`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for absolute X coordinate, got nil")
	}

	expected := "requires X coordinate to be relative with '+' prefix"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Error message should contain %q, got: %v", expected, err)
	}
}

func TestTouchLeftConnector_InvalidX_Zero(t *testing.T) {
	input := `1: 1,1: Box 1
2: |0,0: Box 2`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for X coordinate '0', got nil")
	}

	expected := "requires X coordinate to be relative with '+' prefix"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Error message should contain %q, got: %v", expected, err)
	}
}

func TestTouchLeftConnector_InvalidX_Negative(t *testing.T) {
	input := `1: 1,1: Box 1
2: |-2,0: Box 2`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for negative X coordinate, got nil")
	}

	expected := "requires X coordinate to be relative with '+' prefix"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Error message should contain %q, got: %v", expected, err)
	}
}

func TestTouchLeftConnector_FirstBox(t *testing.T) {
	input := `1: |1,1: Box 1`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for first box with touch-left prefix, got nil")
	}

	expected := "first box (label '1') cannot have touch-left prefix '|'"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Error message should contain %q, got: %v", expected, err)
	}
}

func TestTouchLeftConnector_MultipleBoxes(t *testing.T) {
	input := `1: 1,1: Box 1
2: |+2,0: Box 2
3: |+2,0: Box 3`

	spec, err := ParseDiagramSpec(input, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(spec.Boxes) != 3 {
		t.Fatalf("Expected 3 boxes, got %d", len(spec.Boxes))
	}

	// Box 2 and 3 should both have TouchLeft
	if !spec.Boxes[1].TouchLeft {
		t.Error("Box 2 should have TouchLeft=true")
	}
	if !spec.Boxes[2].TouchLeft {
		t.Error("Box 3 should have TouchLeft=true")
	}

	// Check positions
	if spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 1 {
		t.Errorf("Box 2 position = (%d, %d), want (3, 1)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}
	if spec.Boxes[2].GridX != 5 || spec.Boxes[2].GridY != 1 {
		t.Errorf("Box 3 position = (%d, %d), want (5, 1)", spec.Boxes[2].GridX, spec.Boxes[2].GridY)
	}
}

func TestArrowValidation_InvalidFromID(t *testing.T) {
	input := `1: 1,1: Box 1
2: 2,1: Box 2
---
3 -> 2`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for arrow with non-existent FromID, got nil")
	}

	expected := "arrow '3 -> 2' references non-existent box label '3'"
	if err.Error() != expected {
		t.Errorf("Error message = %q, want %q", err.Error(), expected)
	}
}

func TestArrowValidation_InvalidToID(t *testing.T) {
	input := `1: 1,1: Box 1
2: 2,1: Box 2
---
1 -> 5`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for arrow with non-existent ToID, got nil")
	}

	expected := "arrow '1 -> 5' references non-existent box label '5'"
	if err.Error() != expected {
		t.Errorf("Error message = %q, want %q", err.Error(), expected)
	}
}

func TestArrowValidation_BothInvalid(t *testing.T) {
	input := `1: 1,1: Box 1
2: 2,1: Box 2
---
10 -> 20`

	_, err := ParseDiagramSpec(input, nil)
	if err == nil {
		t.Fatal("Expected error for arrow with both IDs non-existent, got nil")
	}

	// Should fail on FromID first
	expected := "arrow '10 -> 20' references non-existent box label '10'"
	if err.Error() != expected {
		t.Errorf("Error message = %q, want %q", err.Error(), expected)
	}
}

func TestArrowValidation_ValidReferences(t *testing.T) {
	input := `1: 1,1: Box 1
2: 2,1: Box 2
3: 3,1: Box 3
---
1 -> 2
2 -> 3
1 -> 3`

	spec, err := ParseDiagramSpec(input, nil)
	if err != nil {
		t.Fatalf("Expected no error for valid arrows, got: %v", err)
	}

	if len(spec.Arrows) != 3 {
		t.Errorf("Expected 3 arrows, got %d", len(spec.Arrows))
	}
}

func TestArrowValidation_AutoArrowValid(t *testing.T) {
	input := `1: 1,1: Box 1
2: >2,1: Box 2
3: >3,1: Box 3`

	spec, err := ParseDiagramSpec(input, nil)
	if err != nil {
		t.Fatalf("Expected no error for valid auto-arrows, got: %v", err)
	}

	if len(spec.Arrows) != 2 {
		t.Errorf("Expected 2 auto-arrows, got %d", len(spec.Arrows))
	}

	// Verify arrow references are valid
	if spec.Arrows[0].FromID != "1" || spec.Arrows[0].ToID != "2" {
		t.Errorf("Arrow 0 = %s -> %s, want 1 -> 2", spec.Arrows[0].FromID, spec.Arrows[0].ToID)
	}
	if spec.Arrows[1].FromID != "2" || spec.Arrows[1].ToID != "3" {
		t.Errorf("Arrow 1 = %s -> %s, want 2 -> 3", spec.Arrows[1].FromID, spec.Arrows[1].ToID)
	}
}

func TestArrowValidation_MixedManualAndAutoArrows(t *testing.T) {
	input := `1: 1,1: Box 1
2: >2,1: Box 2
3: 3,1: Box 3
---
1 -> 3`

	spec, err := ParseDiagramSpec(input, nil)
	if err != nil {
		t.Fatalf("Expected no error for mixed arrows, got: %v", err)
	}

	if len(spec.Arrows) != 2 {
		t.Errorf("Expected 2 arrows (1 auto + 1 manual), got %d", len(spec.Arrows))
	}
}

// Tests for parseBoxStyles function

func TestParseBoxStyles_Empty(t *testing.T) {
	styles := parseBoxStyles("", nil)

	if styles.BackgroundColor != "" {
		t.Errorf("Expected empty BackgroundColor, got %s", styles.BackgroundColor)
	}
	if styles.BorderColor != "" {
		t.Errorf("Expected empty BorderColor, got %s", styles.BorderColor)
	}
	if styles.BorderWidth != 0 {
		t.Errorf("Expected BorderWidth 0, got %d", styles.BorderWidth)
	}
	if styles.FontSize != 0 {
		t.Errorf("Expected FontSize 0, got %d", styles.FontSize)
	}
	if styles.TextColor != "" {
		t.Errorf("Expected empty TextColor, got %s", styles.TextColor)
	}
}

func TestParseBoxStyles_RedBorder(t *testing.T) {
	styles := parseBoxStyles("rb", nil)

	if styles.BorderColor != "#FF0000" {
		t.Errorf("Expected BorderColor #FF0000, got %s", styles.BorderColor)
	}
	if styles.BorderWidth != 3 {
		t.Errorf("Expected BorderWidth 3, got %d", styles.BorderWidth)
	}
}

func TestParseBoxStyles_GrayBackground(t *testing.T) {
	styles := parseBoxStyles("g", nil)

	if styles.BackgroundColor != "#D3D3D3" {
		t.Errorf("Expected BackgroundColor #D3D3D3, got %s", styles.BackgroundColor)
	}
}

func TestParseBoxStyles_PurpleBackground(t *testing.T) {
	styles := parseBoxStyles("p", nil)

	if styles.BackgroundColor != "#ecbae6" {
		t.Errorf("Expected BackgroundColor #ecbae6, got %s", styles.BackgroundColor)
	}
}

// Tests for ParseFrontmatter

func TestParseFrontmatter_FontKey(t *testing.T) {
	text := `font: fonts/BerkeleyMono.woff2

1,1: OKRs
>+1,+1: Sprint Planning
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/BerkeleyMono.woff2" {
		t.Errorf("Expected font 'fonts/BerkeleyMono.woff2', got '%s'", fm.Font)
	}

	// Remaining text should start with box definitions
	if !strings.Contains(remaining, "1,1: OKRs") {
		t.Errorf("Expected remaining to contain box definitions, got: %s", remaining)
	}

	// Remaining text should NOT contain frontmatter
	if strings.Contains(remaining, "font:") {
		t.Errorf("Expected remaining to not contain frontmatter, got: %s", remaining)
	}
}

func TestParseFrontmatter_StrippedBeforeParsing(t *testing.T) {
	text := `font: fonts/test.woff2

1,1: OKRs
>+1,+1: Sprint Planning
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Fatalf("Expected font path, got '%s'", fm.Font)
	}

	// The remaining text should parse correctly as a diagram
	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error parsing remaining text, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if spec.Boxes[0].Label != "OKRs" {
		t.Errorf("Expected first box label 'OKRs', got '%s'", spec.Boxes[0].Label)
	}
}

func TestParseFrontmatter_WithComments(t *testing.T) {
	text := `# Diagram configuration
font: fonts/test.woff2
# End of frontmatter

1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Errorf("Expected font 'fonts/test.woff2', got '%s'", fm.Font)
	}

	if strings.Contains(remaining, "font:") {
		t.Errorf("Expected remaining to not contain frontmatter")
	}

	// Should still parse the box
	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_WithBlankLines(t *testing.T) {
	text := `
font: fonts/test.woff2

1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Errorf("Expected font 'fonts/test.woff2', got '%s'", fm.Font)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_UnknownKeyStopsParsing(t *testing.T) {
	text := `font: fonts/test.woff2
unknown: value
1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Errorf("Expected font 'fonts/test.woff2', got '%s'", fm.Font)
	}

	// "unknown: value" should be in remaining since it stops frontmatter
	if !strings.Contains(remaining, "unknown: value") {
		t.Errorf("Expected remaining to contain 'unknown: value', got: %s", remaining)
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	text := `1,1: OKRs
>+1,+1: Sprint Planning
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "" {
		t.Errorf("Expected empty font, got '%s'", fm.Font)
	}

	// All text should be in remaining
	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_EmptyInput(t *testing.T) {
	fm, remaining := ParseFrontmatter("")

	if fm.Font != "" {
		t.Errorf("Expected empty font, got '%s'", fm.Font)
	}

	if remaining != "" {
		t.Errorf("Expected empty remaining, got '%s'", remaining)
	}
}

func TestParseFrontmatter_FontWithExtraSpaces(t *testing.T) {
	text := `font:   fonts/test.woff2

1,1: Box A
`

	fm, _ := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Errorf("Expected trimmed font path 'fonts/test.woff2', got '%s'", fm.Font)
	}
}

func TestParseFrontmatter_OnlyComments(t *testing.T) {
	text := `# Just a comment
# Another comment
1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "" {
		t.Errorf("Expected empty font, got '%s'", fm.Font)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_Delimited(t *testing.T) {
	text := `---
font: fonts/BerkeleyMono.woff2
---
1,1: OKRs
>+1,+1: Sprint Planning
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/BerkeleyMono.woff2" {
		t.Errorf("Expected font 'fonts/BerkeleyMono.woff2', got '%s'", fm.Font)
	}

	if strings.Contains(remaining, "---") {
		t.Errorf("Expected remaining to not contain '---' delimiters, got: %s", remaining)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 auto-arrow, got %d", len(spec.Arrows))
	}
}

func TestParseFrontmatter_DelimitedWithComments(t *testing.T) {
	text := `---
# Config
font: fonts/test.woff2
# End
---
1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Errorf("Expected font 'fonts/test.woff2', got '%s'", fm.Font)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_DelimitedDoesNotConflictWithArrowSeparator(t *testing.T) {
	text := `---
font: fonts/test.woff2
---
a: 1,1: Box A
b: 3,1: Box B
---
a -> b
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Errorf("Expected font 'fonts/test.woff2', got '%s'", fm.Font)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	if len(spec.Arrows) != 1 {
		t.Errorf("Expected 1 arrow, got %d", len(spec.Arrows))
	}
}

func TestParseFrontmatter_AxisLabels(t *testing.T) {
	text := `---
x-label: Time
y-label: Control
---
1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if fm.XLabel != "Time" {
		t.Errorf("Expected x-label 'Time', got '%s'", fm.XLabel)
	}
	if fm.YLabel != "Control" {
		t.Errorf("Expected y-label 'Control', got '%s'", fm.YLabel)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_AxisLabelsUndelimited(t *testing.T) {
	text := `x-label: Sprints
y-label: Ownership

1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if fm.XLabel != "Sprints" {
		t.Errorf("Expected x-label 'Sprints', got '%s'", fm.XLabel)
	}
	if fm.YLabel != "Ownership" {
		t.Errorf("Expected y-label 'Ownership', got '%s'", fm.YLabel)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_NoLabelsNoAxes(t *testing.T) {
	text := `1,1: Box A
`

	fm, _ := ParseFrontmatter(text)

	if fm.XLabel != "" {
		t.Errorf("Expected empty x-label, got '%s'", fm.XLabel)
	}
	if fm.YLabel != "" {
		t.Errorf("Expected empty y-label, got '%s'", fm.YLabel)
	}
}

func TestParseFrontmatter_AllKeys(t *testing.T) {
	text := `---
font: fonts/test.woff2
x-label: Time
y-label: Control
---
1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Errorf("Expected font 'fonts/test.woff2', got '%s'", fm.Font)
	}
	if fm.XLabel != "Time" {
		t.Errorf("Expected x-label 'Time', got '%s'", fm.XLabel)
	}
	if fm.YLabel != "Control" {
		t.Errorf("Expected y-label 'Control', got '%s'", fm.YLabel)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_SingleLegend(t *testing.T) {
	text := `---
legend: p = In Progress
---
1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if len(fm.Legend) != 1 {
		t.Fatalf("Expected 1 legend entry, got %d", len(fm.Legend))
	}
	if fm.Legend[0].Style != "p" {
		t.Errorf("Expected style 'p', got '%s'", fm.Legend[0].Style)
	}
	if fm.Legend[0].Label != "In Progress" {
		t.Errorf("Expected label 'In Progress', got '%s'", fm.Legend[0].Label)
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_MultipleLegends(t *testing.T) {
	text := `---
legend: p = In Progress
legend: g = Completed
legend: lp = Planned
---
1,1: Box A
`

	fm, _ := ParseFrontmatter(text)

	if len(fm.Legend) != 3 {
		t.Fatalf("Expected 3 legend entries, got %d", len(fm.Legend))
	}

	expected := []struct{ style, label string }{
		{"p", "In Progress"},
		{"g", "Completed"},
		{"lp", "Planned"},
	}

	for i, exp := range expected {
		if fm.Legend[i].Style != exp.style {
			t.Errorf("Legend %d: expected style '%s', got '%s'", i, exp.style, fm.Legend[i].Style)
		}
		if fm.Legend[i].Label != exp.label {
			t.Errorf("Legend %d: expected label '%s', got '%s'", i, exp.label, fm.Legend[i].Label)
		}
	}
}

func TestParseFrontmatter_NoLegend(t *testing.T) {
	text := `---
font: fonts/test.woff2
---
1,1: Box A
`

	fm, _ := ParseFrontmatter(text)

	if len(fm.Legend) != 0 {
		t.Errorf("Expected 0 legend entries, got %d", len(fm.Legend))
	}
}

func TestParseFrontmatter_LegendStyleResolvesToColor(t *testing.T) {
	text := `---
legend: p = Purple items
legend: g = Gray items
legend: rb = Red border items
---
1,1: Box A
`

	fm, _ := ParseFrontmatter(text)

	if len(fm.Legend) != 3 {
		t.Fatalf("Expected 3 legend entries, got %d", len(fm.Legend))
	}

	// Verify style codes resolve to correct colors via parseBoxStyles
	styles := parseBoxStyles(fm.Legend[0].Style, nil)
	if styles.BackgroundColor != "#ecbae6" {
		t.Errorf("Style 'p' should resolve to purple #ecbae6, got '%s'", styles.BackgroundColor)
	}

	styles = parseBoxStyles(fm.Legend[1].Style, nil)
	if styles.BackgroundColor != "#D3D3D3" {
		t.Errorf("Style 'g' should resolve to gray #D3D3D3, got '%s'", styles.BackgroundColor)
	}

	styles = parseBoxStyles(fm.Legend[2].Style, nil)
	if styles.BorderColor != "#FF0000" {
		t.Errorf("Style 'rb' should resolve to red border #FF0000, got '%s'", styles.BorderColor)
	}
}

func TestParseFrontmatter_LegendUndelimited(t *testing.T) {
	text := `legend: p = In Progress
legend: g = Completed

1,1: Box A
`

	fm, remaining := ParseFrontmatter(text)

	if len(fm.Legend) != 2 {
		t.Fatalf("Expected 2 legend entries, got %d", len(fm.Legend))
	}

	spec, err := ParseDiagramSpec(remaining, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseFrontmatter_LegendWithAllKeys(t *testing.T) {
	text := `---
font: fonts/test.woff2
x-label: Time
y-label: Control
legend: p = In Progress
legend: g = Completed
---
1,1: Box A
`

	fm, _ := ParseFrontmatter(text)

	if fm.Font != "fonts/test.woff2" {
		t.Errorf("Expected font 'fonts/test.woff2', got '%s'", fm.Font)
	}
	if fm.XLabel != "Time" {
		t.Errorf("Expected x-label 'Time', got '%s'", fm.XLabel)
	}
	if fm.YLabel != "Control" {
		t.Errorf("Expected y-label 'Control', got '%s'", fm.YLabel)
	}
	if len(fm.Legend) != 2 {
		t.Fatalf("Expected 2 legend entries, got %d", len(fm.Legend))
	}
}

// Custom color tests

func TestParseBoxStyles_CustomColorBackground(t *testing.T) {
	colors := map[string]string{"green": "#00FF00"}
	styles := parseBoxStyles("green", colors)
	if styles.BackgroundColor != "#00FF00" {
		t.Errorf("Expected background '#00FF00', got '%s'", styles.BackgroundColor)
	}
}

func TestParseBoxStyles_CustomColorText(t *testing.T) {
	colors := map[string]string{"green": "#00FF00"}
	styles := parseBoxStyles("greent", colors)
	if styles.TextColor != "#00FF00" {
		t.Errorf("Expected text color '#00FF00', got '%s'", styles.TextColor)
	}
}

func TestParseBoxStyles_CustomColorCombinedWithBuiltin(t *testing.T) {
	colors := map[string]string{"green": "#00FF00"}
	styles := parseBoxStyles("nbb-greent", colors)
	if styles.BackgroundColor != "none" {
		t.Errorf("Expected background 'none', got '%s'", styles.BackgroundColor)
	}
	if styles.TextColor != "#00FF00" {
		t.Errorf("Expected text color '#00FF00', got '%s'", styles.TextColor)
	}
}

func TestParseBoxStyles_CustomColorNilMap(t *testing.T) {
	styles := parseBoxStyles("green", nil)
	if styles.BackgroundColor != "" {
		t.Errorf("Expected empty background with nil colors, got '%s'", styles.BackgroundColor)
	}
}

func TestParseFrontmatter_CustomColors(t *testing.T) {
	input := `---
color: green = #00FF00
color: blue = #0000FF
---
1,1: Task A
`
	fm, remaining := ParseFrontmatter(input)
	if len(fm.Colors) != 2 {
		t.Fatalf("Expected 2 custom colors, got %d", len(fm.Colors))
	}
	if fm.Colors["green"] != "#00FF00" {
		t.Errorf("Expected green=#00FF00, got '%s'", fm.Colors["green"])
	}
	if fm.Colors["blue"] != "#0000FF" {
		t.Errorf("Expected blue=#0000FF, got '%s'", fm.Colors["blue"])
	}

	spec, err := ParseDiagramSpec(remaining, fm.Colors)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(spec.Boxes) != 1 {
		t.Errorf("Expected 1 box, got %d", len(spec.Boxes))
	}
}

func TestParseDiagramSpec_CustomColorInBox(t *testing.T) {
	colors := map[string]string{"green": "#00FF00"}
	text := `1,1: Task A, green`
	spec, err := ParseDiagramSpec(text, colors)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if spec.Boxes[0].Color != "#00FF00" {
		t.Errorf("Expected box color '#00FF00', got '%s'", spec.Boxes[0].Color)
	}
}

func TestParseDiagramSpec_CustomColorTextInBox(t *testing.T) {
	colors := map[string]string{"green": "#00FF00"}
	text := `1,1: Warning, nbb-greent`
	spec, err := ParseDiagramSpec(text, colors)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if spec.Boxes[0].Color != "none" {
		t.Errorf("Expected box color 'none', got '%s'", spec.Boxes[0].Color)
	}
	if spec.Boxes[0].TextColor != "#00FF00" {
		t.Errorf("Expected text color '#00FF00', got '%s'", spec.Boxes[0].TextColor)
	}
}

// Group parsing tests

func TestParseDiagramSpec_Group_BasicAssignment(t *testing.T) {
	text := `
1,2: Stephan @Team
>+2,0: Stefanie @Team
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 2 {
		t.Fatalf("Expected 2 boxes, got %d", len(spec.Boxes))
	}

	// Both boxes should be in group "Team"
	if spec.Boxes[0].Group != "Team" {
		t.Errorf("Expected box 0 group 'Team', got '%s'", spec.Boxes[0].Group)
	}
	if spec.Boxes[1].Group != "Team" {
		t.Errorf("Expected box 1 group 'Team', got '%s'", spec.Boxes[1].Group)
	}

	// Labels should not include @Team
	if spec.Boxes[0].Label != "Stephan" {
		t.Errorf("Expected label 'Stephan', got '%s'", spec.Boxes[0].Label)
	}
	if spec.Boxes[1].Label != "Stefanie" {
		t.Errorf("Expected label 'Stefanie', got '%s'", spec.Boxes[1].Label)
	}

	// Group should be auto-created
	if len(spec.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(spec.Groups))
	}
	if spec.Groups[0].Name != "Team" {
		t.Errorf("Expected group name 'Team', got '%s'", spec.Groups[0].Name)
	}
	// Without explicit definition, label defaults to name
	if spec.Groups[0].Label != "Team" {
		t.Errorf("Expected group label 'Team', got '%s'", spec.Groups[0].Label)
	}
	if len(spec.Groups[0].BoxIDs) != 2 {
		t.Errorf("Expected 2 box IDs in group, got %d", len(spec.Groups[0].BoxIDs))
	}
}

func TestParseDiagramSpec_Group_WithDefinitionLine(t *testing.T) {
	text := `
1,2: Stephan @Team
>+2,0: Stefanie @Team
@Team: Our Team
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(spec.Groups))
	}

	// Label should come from the definition line
	if spec.Groups[0].Label != "Our Team" {
		t.Errorf("Expected group label 'Our Team', got '%s'", spec.Groups[0].Label)
	}
}

func TestParseDiagramSpec_Group_WithStyle(t *testing.T) {
	text := `
1,2: Stefanie, p @Team
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Fatalf("Expected 1 box, got %d", len(spec.Boxes))
	}

	box := spec.Boxes[0]
	if box.Label != "Stefanie" {
		t.Errorf("Expected label 'Stefanie', got '%s'", box.Label)
	}
	if box.Color != "#ecbae6" {
		t.Errorf("Expected purple background, got '%s'", box.Color)
	}
	if box.Group != "Team" {
		t.Errorf("Expected group 'Team', got '%s'", box.Group)
	}
}

func TestParseDiagramSpec_Group_NoGroup(t *testing.T) {
	text := `
1,2: Stephan
3,2: Stefanie
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(spec.Groups))
	}

	for _, box := range spec.Boxes {
		if box.Group != "" {
			t.Errorf("Expected no group, got '%s'", box.Group)
		}
	}
}

func TestParseDiagramSpec_Group_MultipleGroups(t *testing.T) {
	text := `
1,1: Alice @Dev
2,1: Bob @Ops
3,1: Charlie @Dev
4,1: Dave @Ops
@Dev: Development
@Ops: Operations
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(spec.Groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(spec.Groups))
	}

	// Find each group
	groupMap := make(map[string]GroupDef)
	for _, g := range spec.Groups {
		groupMap[g.Name] = g
	}

	dev := groupMap["Dev"]
	if dev.Label != "Development" {
		t.Errorf("Expected Dev label 'Development', got '%s'", dev.Label)
	}
	if len(dev.BoxIDs) != 2 {
		t.Errorf("Expected 2 boxes in Dev, got %d", len(dev.BoxIDs))
	}

	ops := groupMap["Ops"]
	if ops.Label != "Operations" {
		t.Errorf("Expected Ops label 'Operations', got '%s'", ops.Label)
	}
	if len(ops.BoxIDs) != 2 {
		t.Errorf("Expected 2 boxes in Ops, got %d", len(ops.BoxIDs))
	}
}

func TestParseDiagramSpec_Group_DefinitionWithoutBoxes(t *testing.T) {
	text := `
1,1: Alice
@EmptyGroup: No boxes
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Group definition without any boxes assigned should result in no groups
	if len(spec.Groups) != 0 {
		t.Errorf("Expected 0 groups (no boxes assigned), got %d", len(spec.Groups))
	}
}

func TestParseFrontmatter_ArrowFlow(t *testing.T) {
	text := `---
arrow-flow: down
---
a: 1,1: Box A
b: 3,2: Box B
`
	fm, remaining := ParseFrontmatter(text)

	if fm.ArrowFlow != "down" {
		t.Errorf("Expected ArrowFlow='down', got %q", fm.ArrowFlow)
	}

	if !strings.Contains(remaining, "a: 1,1: Box A") {
		t.Error("Remaining text should contain box definitions")
	}
}

func TestParseFrontmatter_ArrowFlowUndelimited(t *testing.T) {
	text := `arrow-flow: down
a: 1,1: Box A
`
	fm, remaining := ParseFrontmatter(text)

	if fm.ArrowFlow != "down" {
		t.Errorf("Expected ArrowFlow='down', got %q", fm.ArrowFlow)
	}

	if !strings.Contains(remaining, "a: 1,1: Box A") {
		t.Error("Remaining text should contain box definitions")
	}
}

func TestParseDiagramSpec_ArrowFlowPerArrow(t *testing.T) {
	text := `
a: 1,1: Box A
b: 3,2: Box B
c: 5,1: Box C
---
a -> b | down
b -> c
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(spec.Arrows) != 2 {
		t.Fatalf("Expected 2 arrows, got %d", len(spec.Arrows))
	}

	// First arrow should have flow="down"
	if spec.Arrows[0].Flow != "down" {
		t.Errorf("Expected first arrow Flow='down', got %q", spec.Arrows[0].Flow)
	}

	// Second arrow should have no flow
	if spec.Arrows[1].Flow != "" {
		t.Errorf("Expected second arrow Flow='', got %q", spec.Arrows[1].Flow)
	}
}

func TestParseDiagramSpec_ArrowFlowPerArrowWithSpaces(t *testing.T) {
	text := `
a: 1,1: Box A
b: 3,2: Box B
---
a -> b | down
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if spec.Arrows[0].FromID != "a" {
		t.Errorf("Expected FromID='a', got %q", spec.Arrows[0].FromID)
	}
	if spec.Arrows[0].ToID != "b" {
		t.Errorf("Expected ToID='b', got %q", spec.Arrows[0].ToID)
	}
	if spec.Arrows[0].Flow != "down" {
		t.Errorf("Expected Flow='down', got %q", spec.Arrows[0].Flow)
	}
}

// Container tests

func TestParseDiagramSpec_ContainerBasic(t *testing.T) {
	text := `
G: 1,1 [
    X: 0,0: A
    +1,0: B
    Y: 4,4: C
    X -> Y
]
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(spec.Boxes) != 3 {
		t.Fatalf("Expected 3 boxes, got %d", len(spec.Boxes))
	}

	// X at 0,0 inside container at 1,1 → grid (1,1)
	if spec.Boxes[0].ID != "X" {
		t.Errorf("Expected box 0 ID 'X', got '%s'", spec.Boxes[0].ID)
	}
	if spec.Boxes[0].GridX != 1 || spec.Boxes[0].GridY != 1 {
		t.Errorf("Expected X at (1,1), got (%d,%d)", spec.Boxes[0].GridX, spec.Boxes[0].GridY)
	}
	if spec.Boxes[0].Label != "A" {
		t.Errorf("Expected label 'A', got '%s'", spec.Boxes[0].Label)
	}

	// B at +1,0 relative to X(1,1) → grid (2,1)
	if spec.Boxes[1].GridX != 2 || spec.Boxes[1].GridY != 1 {
		t.Errorf("Expected B at (2,1), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}

	// Y at 4,4 inside container at 1,1 → grid (5,5)
	if spec.Boxes[2].ID != "Y" {
		t.Errorf("Expected box 2 ID 'Y', got '%s'", spec.Boxes[2].ID)
	}
	if spec.Boxes[2].GridX != 5 || spec.Boxes[2].GridY != 5 {
		t.Errorf("Expected Y at (5,5), got (%d,%d)", spec.Boxes[2].GridX, spec.Boxes[2].GridY)
	}

	// Arrow X → Y
	if len(spec.Arrows) != 1 {
		t.Fatalf("Expected 1 arrow, got %d", len(spec.Arrows))
	}
	if spec.Arrows[0].FromID != "X" || spec.Arrows[0].ToID != "Y" {
		t.Errorf("Expected arrow X->Y, got %s->%s", spec.Arrows[0].FromID, spec.Arrows[0].ToID)
	}

	// Containers are purely organizational — no group created
	if len(spec.Groups) != 0 {
		t.Errorf("Expected 0 groups (containers don't create groups), got %d", len(spec.Groups))
	}
}

func TestParseDiagramSpec_ContainerWithThirdField(t *testing.T) {
	// Third field after ID and coords is accepted (ignored — containers have no label)
	text := `
G: 1,1: My Group [
    X: 0,0: Hello
]
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(spec.Boxes) != 1 {
		t.Fatalf("Expected 1 box, got %d", len(spec.Boxes))
	}
	if spec.Boxes[0].GridX != 1 || spec.Boxes[0].GridY != 1 {
		t.Errorf("Expected X at (1,1), got (%d,%d)", spec.Boxes[0].GridX, spec.Boxes[0].GridY)
	}
	if len(spec.Groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(spec.Groups))
	}
}

func TestParseDiagramSpec_ContainerMoved(t *testing.T) {
	text := `
G: 3,2 [
    0,0: A
    +1,0: B
    4,4: C
]
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(spec.Boxes) != 3 {
		t.Fatalf("Expected 3 boxes, got %d", len(spec.Boxes))
	}

	// A at 0,0 inside container at 3,2 → grid (3,2)
	if spec.Boxes[0].GridX != 3 || spec.Boxes[0].GridY != 2 {
		t.Errorf("Expected A at (3,2), got (%d,%d)", spec.Boxes[0].GridX, spec.Boxes[0].GridY)
	}

	// B at +1,0 relative to A(3,2) → grid (4,2)
	if spec.Boxes[1].GridX != 4 || spec.Boxes[1].GridY != 2 {
		t.Errorf("Expected B at (4,2), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}

	// C at 4,4 inside container at 3,2 → grid (7,6)
	if spec.Boxes[2].GridX != 7 || spec.Boxes[2].GridY != 6 {
		t.Errorf("Expected C at (7,6), got (%d,%d)", spec.Boxes[2].GridX, spec.Boxes[2].GridY)
	}
}

func TestParseDiagramSpec_ContainerWithBoxesBeforeAfter(t *testing.T) {
	text := `
A: 1,1: Before
G: 3,3 [
    X: 0,0: Inside
]
B: 5,5: After
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(spec.Boxes) != 3 {
		t.Fatalf("Expected 3 boxes, got %d", len(spec.Boxes))
	}

	// A at (1,1)
	if spec.Boxes[0].ID != "A" || spec.Boxes[0].GridX != 1 || spec.Boxes[0].GridY != 1 {
		t.Errorf("Expected A at (1,1), got (%d,%d)", spec.Boxes[0].GridX, spec.Boxes[0].GridY)
	}

	// X at 0,0 inside container at 3,3 → grid (3,3)
	if spec.Boxes[1].ID != "X" || spec.Boxes[1].GridX != 3 || spec.Boxes[1].GridY != 3 {
		t.Errorf("Expected X at (3,3), got (%d,%d)", spec.Boxes[1].GridX, spec.Boxes[1].GridY)
	}

	// B at (5,5) - absolute, outside container
	if spec.Boxes[2].ID != "B" || spec.Boxes[2].GridX != 5 || spec.Boxes[2].GridY != 5 {
		t.Errorf("Expected B at (5,5), got (%d,%d)", spec.Boxes[2].GridX, spec.Boxes[2].GridY)
	}

	// Containers don't create groups
	if len(spec.Groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(spec.Groups))
	}
}

func TestParseDiagramSpec_ContainerArrowsCrossBoundary(t *testing.T) {
	text := `
A: 1,1: Outside
G: 3,3 [
    X: 0,0: Inside
]
---
A -> X
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(spec.Arrows) != 1 {
		t.Fatalf("Expected 1 arrow, got %d", len(spec.Arrows))
	}
	if spec.Arrows[0].FromID != "A" || spec.Arrows[0].ToID != "X" {
		t.Errorf("Expected arrow A->X, got %s->%s", spec.Arrows[0].FromID, spec.Arrows[0].ToID)
	}
}

func TestParseDiagramSpec_ContainerErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name: "nested container",
			input: `
G1: 1,1 [
    G2: 2,2 [
        0,0: A
    ]
]
`,
			wantErr: "nested containers not supported",
		},
		{
			name: "unclosed container",
			input: `
G: 1,1 [
    0,0: A
`,
			wantErr: "unclosed container 'G'",
		},
		{
			name: "closing bracket outside container",
			input: `
1,1: A
]
`,
			wantErr: "unexpected ']' outside container",
		},
		{
			name: "separator inside container",
			input: `
G: 1,1 [
    X: 0,0: A
    ---
    X -> X
]
`,
			wantErr: "section separator not allowed inside container",
		},
		{
			name: "container in arrow section",
			input: `
A: 1,1: Box
---
G: 2,2 [
    0,0: Inside
]
`,
			wantErr: "container not allowed in arrow section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDiagramSpec(tt.input, nil)
			if err == nil {
				t.Fatalf("Expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
