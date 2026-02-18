package main

import (
	"strings"
	"testing"
)

func TestSvgHeader(t *testing.T) {
	header := svgHeader(800, 600, nil)

	if !strings.Contains(header, `width="800"`) {
		t.Error("Header should contain width=\"800\"")
	}

	if !strings.Contains(header, `height="600"`) {
		t.Error("Header should contain height=\"600\"")
	}

	if !strings.Contains(header, `<svg`) {
		t.Error("Header should contain <svg tag")
	}

	if !strings.Contains(header, `arrowhead`) {
		t.Error("Header should contain arrowhead marker definition")
	}
}

func TestSvgHeader_ArrowheadStyle(t *testing.T) {
	header := svgHeader(800, 600, nil)

	// Should use polyline instead of polygon for two-line arrowhead
	if !strings.Contains(header, `<polyline`) {
		t.Error("Arrowhead should use <polyline> element")
	}

	// Should not have filled polygon
	if strings.Contains(header, `<polygon`) {
		t.Error("Arrowhead should not use <polygon> element (should be polyline)")
	}

	// Should have fill="none" for outline style
	if !strings.Contains(header, `fill="none"`) {
		t.Error("Arrowhead should have fill=\"none\"")
	}

	// Should have stroke color
	if !strings.Contains(header, `stroke="#000"`) {
		t.Error("Arrowhead should have stroke=\"#000\"")
	}

	// Should have stroke-width
	if !strings.Contains(header, `stroke-width="1.5"`) {
		t.Error("Arrowhead should have stroke-width=\"1.5\"")
	}
}

func TestSvgFooter(t *testing.T) {
	footer := svgFooter()

	if footer != `</svg>` {
		t.Errorf("Footer should be </svg>, got %s", footer)
	}
}

func TestStraightArrow(t *testing.T) {
	arrow := straightArrow(10, 20, 30, 40)

	if !strings.Contains(arrow, `<line`) {
		t.Error("Should be a line element")
	}

	if !strings.Contains(arrow, `x1="10"`) {
		t.Error("Should have x1=\"10\"")
	}

	if !strings.Contains(arrow, `y1="20"`) {
		t.Error("Should have y1=\"20\"")
	}

	if !strings.Contains(arrow, `x2="30"`) {
		t.Error("Should have x2=\"30\"")
	}

	if !strings.Contains(arrow, `y2="40"`) {
		t.Error("Should have y2=\"40\"")
	}

	if !strings.Contains(arrow, `marker-end="url(#arrowhead)"`) {
		t.Error("Should have arrowhead marker")
	}
}

func TestOneBentArrow(t *testing.T) {
	// Test vertical-first (vertical then horizontal)
	arrowVertFirst := oneBentArrow(10, 20, 30, 40, true)

	if !strings.Contains(arrowVertFirst, `<polyline`) {
		t.Error("Should be a polyline element")
	}

	// Check for 3 points (vertical then horizontal)
	if !strings.Contains(arrowVertFirst, `points="10,20`) {
		t.Error("Should start at 10,20")
	}

	if !strings.Contains(arrowVertFirst, `10,40`) {
		t.Error("Should have vertical segment to 10,40")
	}

	if !strings.Contains(arrowVertFirst, `30,40"`) {
		t.Error("Should end at 30,40")
	}

	if !strings.Contains(arrowVertFirst, `marker-end="url(#arrowhead)"`) {
		t.Error("Should have arrowhead marker")
	}

	// Test horizontal-first (horizontal then vertical)
	arrowHorizFirst := oneBentArrow(10, 20, 30, 40, false)

	if !strings.Contains(arrowHorizFirst, `<polyline`) {
		t.Error("Should be a polyline element")
	}

	// Check for 3 points (horizontal then vertical)
	if !strings.Contains(arrowHorizFirst, `points="10,20`) {
		t.Error("Should start at 10,20")
	}

	if !strings.Contains(arrowHorizFirst, `30,20`) {
		t.Error("Should have horizontal segment to 30,20")
	}

	if !strings.Contains(arrowHorizFirst, `30,40"`) {
		t.Error("Should end at 30,40")
	}
}

func TestTwoBentArrow(t *testing.T) {
	arrow := twoBentArrow(10, 20, 50, 60)

	if !strings.Contains(arrow, `<polyline`) {
		t.Error("Should be a polyline element")
	}

	// Check for 4 points
	midX := (10 + 50) / 2 // 30

	if !strings.Contains(arrow, `points="10,20`) {
		t.Error("Should start at 10,20")
	}

	if !strings.Contains(arrow, `30,20`) {
		t.Errorf("Should have horizontal segment to %d,20", midX)
	}

	if !strings.Contains(arrow, `30,60`) {
		t.Errorf("Should have vertical segment to %d,60", midX)
	}

	if !strings.Contains(arrow, `50,60"`) {
		t.Error("Should end at 50,60")
	}
}

func TestTwoBentArrowVertical(t *testing.T) {
	arrow := twoBentArrowVertical(10, 20, 50, 60)

	if !strings.Contains(arrow, `<polyline`) {
		t.Error("Should be a polyline element")
	}

	// Check for 4 points (V-H-V path)
	midY := (20 + 60) / 2 // 40

	if !strings.Contains(arrow, `points="10,20`) {
		t.Error("Should start at 10,20")
	}

	if !strings.Contains(arrow, `10,40`) {
		t.Errorf("Should have vertical segment to 10,%d", midY)
	}

	if !strings.Contains(arrow, `50,40`) {
		t.Errorf("Should have horizontal segment to 50,%d", midY)
	}

	if !strings.Contains(arrow, `50,60"`) {
		t.Error("Should end at 50,60")
	}

	if !strings.Contains(arrow, `marker-end="url(#arrowhead)"`) {
		t.Error("Should have arrowhead marker")
	}
}

func TestDrawBox(t *testing.T) {
	box := drawBox(10, 20, 100, 50, "#FFCE33", "", 0)

	if !strings.Contains(box, `<rect`) {
		t.Error("Should be a rect element")
	}

	if !strings.Contains(box, `x="10"`) {
		t.Error("Should have x=\"10\"")
	}

	if !strings.Contains(box, `y="20"`) {
		t.Error("Should have y=\"20\"")
	}

	if !strings.Contains(box, `width="100"`) {
		t.Error("Should have width=\"100\"")
	}

	if !strings.Contains(box, `height="50"`) {
		t.Error("Should have height=\"50\"")
	}

	if !strings.Contains(box, `fill="#FFCE33"`) {
		t.Error("Should have fill=\"#FFCE33\"")
	}
}

func TestDrawLine(t *testing.T) {
	tests := []struct {
		name          string
		dashed        bool
		withArrow     bool
		shouldContain []string
	}{
		{
			name:          "Plain line",
			dashed:        false,
			withArrow:     false,
			shouldContain: []string{`<line`, `x1="10"`, `y1="20"`, `x2="30"`, `y2="40"`},
		},
		{
			name:          "Dashed line",
			dashed:        true,
			withArrow:     false,
			shouldContain: []string{`stroke-dasharray="5,5"`},
		},
		{
			name:          "Line with arrow",
			dashed:        false,
			withArrow:     true,
			shouldContain: []string{`marker-end="url(#arrowhead)"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := drawLine(10, 20, 30, 40, tt.dashed, tt.withArrow)

			for _, s := range tt.shouldContain {
				if !strings.Contains(line, s) {
					t.Errorf("Line should contain %s", s)
				}
			}
		})
	}
}

func TestDrawText(t *testing.T) {
	attrs := map[string]string{
		"font-weight": "normal",
		"text-anchor": "middle",
	}

	text := drawText(100, 200, "Hello", 24, attrs, nil)

	if !strings.Contains(text, `<text`) {
		t.Error("Should be a text element")
	}

	if !strings.Contains(text, `x="100"`) {
		t.Error("Should have x=\"100\"")
	}

	if !strings.Contains(text, `y="200"`) {
		t.Error("Should have y=\"200\"")
	}

	if !strings.Contains(text, `font-size="24"`) {
		t.Error("Should have font-size=\"24\"")
	}

	if !strings.Contains(text, `Hello`) {
		t.Error("Should contain text content")
	}

	if !strings.Contains(text, `font-weight="normal"`) {
		t.Error("Should have font-weight attribute")
	}

	if !strings.Contains(text, `text-anchor="middle"`) {
		t.Error("Should have text-anchor attribute")
	}
}

func TestDrawBoxText(t *testing.T) {
	lines := []string{"Line 1", "Line 2"}
	text := drawBoxText(100, 100, 200, 100, 0, "", lines, nil) // 0 means use default font size, "" means default color

	// Should contain two text elements
	count := strings.Count(text, "<text")
	if count != 2 {
		t.Errorf("Expected 2 text elements, got %d", count)
	}

	if !strings.Contains(text, "Line 1") {
		t.Error("Should contain 'Line 1'")
	}

	if !strings.Contains(text, "Line 2") {
		t.Error("Should contain 'Line 2'")
	}

	if !strings.Contains(text, `font-size="24"`) {
		t.Error("Should have font-size=\"24\"")
	}

	if !strings.Contains(text, `font-weight="normal"`) {
		t.Error("Should have bold font weight")
	}

	if !strings.Contains(text, `dominant-baseline="middle"`) {
		t.Error("Should have dominant-baseline=\"middle\" for vertical centering")
	}
}
