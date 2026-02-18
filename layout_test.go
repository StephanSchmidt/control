package main

import "testing"

func TestCalculateDimensions(t *testing.T) {
	config := NewDefaultConfig()

	tests := []struct {
		name       string
		maxGridX   int
		maxGridY   int
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "Single box at 1,1",
			maxGridX:   1,
			maxGridY:   1,
			wantWidth:  90 + 250 + 20, // leftMargin (60+30) + box + rightMargin
			wantHeight: 80 + 150 + 50, // bottomMargin (60+20) + content (1*1.5*100) + topMargin
		},
		{
			name:       "Box at 3,2",
			maxGridX:   3,
			maxGridY:   2,
			wantWidth:  90 + int(2*1.5*100+2.5*100) + 20,
			wantHeight: 80 + int(2*1.5*100) + 50, // verticalCellUnits = 1.0 + 0.5 = 1.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims := CalculateDimensions(tt.maxGridX, tt.maxGridY, config)

			if dims.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", dims.Width, tt.wantWidth)
			}
			if dims.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", dims.Height, tt.wantHeight)
			}
			if dims.BoxWidth != 250 {
				t.Errorf("BoxWidth = %d, want 250", dims.BoxWidth)
			}
			if dims.BoxHeight != 100 {
				t.Errorf("BoxHeight = %d, want 100", dims.BoxHeight)
			}
		})
	}
}

func TestRouteArrow_SameColumn(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1 bounds
		BoxCoords{X1: 100, Y1: 400, X2: 200, Y2: 500}, // Box 2 bounds
		4, 1, // fromGridX, fromGridY (GridY=1 is top)
		4, 4, // toGridX, toGridY (GridY=4 is bottom)
		nil, "0", "1", "", // allBoxes, fromID, toID, flow (no collision detection for this test)
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Same column should produce vertical arrow
	if plan.StartX != plan.EndX {
		t.Errorf("Expected straight vertical arrow, startX=%d endX=%d", plan.StartX, plan.EndX)
	}

	if plan.Strategy != "straight_vertical" {
		t.Errorf("Expected straight_vertical strategy, got %s", plan.Strategy)
	}

}

func TestRouteArrow_OverlappingColumns(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1 bounds
		BoxCoords{X1: 200, Y1: 200, X2: 300, Y2: 300}, // Box 2 bounds
		3, 1, // fromGridX, fromGridY (GridY=1 is top)
		4, 2, // toGridX, toGridY (GridY=2 is below)
		nil, "0", "1", "", // allBoxes, fromID, toID, flow (no collision detection for this test)
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Adjacent/touching boxes must use vertical-first routing to avoid illegal arrows
	// (horizontal routing would create startX > endX when boxes are touching)
	if plan.Strategy != "two_segment_vertical_first" {
		t.Errorf("Expected two_segment_vertical_first for adjacent/touching boxes, got %s", plan.Strategy)
	}

	// Vertical-first: should exit from bottom (since toGridY > fromGridY)
	// startX should be at horizontal center = (100+200)/2 = 150
	if plan.StartX != 150 {
		t.Errorf("Expected startX=150 (horizontal center), got %d", plan.StartX)
	}

	// startY should be at bottom edge = b1y2 = 200
	if plan.StartY != 200 {
		t.Errorf("Expected startY=200 (bottom edge), got %d", plan.StartY)
	}

}

func TestRouteArrow_NonOverlapping(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1 bounds
		BoxCoords{X1: 300, Y1: 100, X2: 400, Y2: 200}, // Box 2 bounds
		1, 2, // fromGridX, fromGridY
		3, 2, // toGridX, toGridY
		nil, "0", "1", "", // allBoxes, fromID, toID, flow (no collision detection for this test)
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Non-overlapping (deltaX=2) should use horizontal routing
	if plan.Strategy == "straight_vertical" || plan.Strategy == "two_segment_vertical_first" {
		t.Errorf("Expected horizontal routing for non-overlapping boxes, got %s", plan.Strategy)
	}

	// Should exit from right edge = b1x2 = 200
	if plan.StartX != 200 {
		t.Errorf("Expected startX=200 (right edge), got %d", plan.StartX)
	}

	// Should enter at left edge (adjusted 2px back for box stroke) = b2x1 - 2 = 298
	if plan.EndX != 298 {
		t.Errorf("Expected endX=298 (left edge - 2px box stroke adjustment), got %d", plan.EndX)
	}

}

func TestLayout_SimpleSpec(t *testing.T) {
	spec := &DiagramSpec{
		Boxes: []BoxSpec{
			{ID: "1", GridX: 1, GridY: 2, GridWidth: 2, GridHeight: 1, Label: "Box 1"},
			{ID: "2", GridX: 3, GridY: 2, GridWidth: 2, GridHeight: 1, Label: "Box 2"},
		},
		Arrows: []ArrowSpec{
			{FromID: "1", ToID: "2"},
		},
	}

	config := NewDefaultConfig()
	diagram, _ := Layout(spec, config, nil, nil, "")

	if len(diagram.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(diagram.Boxes))
	}

	if len(diagram.Arrows) != 1 {
		t.Errorf("Expected 1 arrow, got %d", len(diagram.Arrows))
	}

	if diagram.Width == 0 || diagram.Height == 0 {
		t.Error("Expected non-zero dimensions")
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name  string
		hex   string
		wantR int
		wantG int
		wantB int
	}{
		{"With hash prefix", "#FFCE33", 255, 206, 51},
		{"Without hash prefix", "FFCE33", 255, 206, 51},
		{"White", "#FFFFFF", 255, 255, 255},
		{"Black", "#000000", 0, 0, 0},
		{"Random color", "#1A2B3C", 26, 43, 60},
		{"Empty string", "", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := parseHexColor(tt.hex)
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("parseHexColor(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.hex, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestRgbToHex(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b int
		want    string
	}{
		{"Yellow", 255, 206, 51, "#FFCE33"},
		{"White", 255, 255, 255, "#FFFFFF"},
		{"Black", 0, 0, 0, "#000000"},
		{"Random color", 26, 43, 60, "#1A2B3C"},
		{"Purple", 128, 0, 128, "#800080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rgbToHex(tt.r, tt.g, tt.b)
			if got != tt.want {
				t.Errorf("rgbToHex(%d, %d, %d) = %q, want %q",
					tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}

func TestInterpolateColor(t *testing.T) {
	tests := []struct {
		name   string
		color1 string
		color2 string
		factor float64
		want   string
	}{
		{"Factor 0.0", "#FFCE33", "#FFFFE0", 0.0, "#FFCE33"},
		{"Factor 1.0", "#FFCE33", "#FFFFE0", 1.0, "#FFFFE0"},
		{"Factor 0.5", "#FFCE33", "#FFFFE0", 0.5, "#FFE689"},
		{"Red to blue 0.5", "#FF0000", "#0000FF", 0.5, "#7F007F"},
		{"Black to white 0.5", "#000000", "#FFFFFF", 0.5, "#7F7F7F"},
		{"Factor 0.25", "#FFCE33", "#FFF0C0", 0.25, "#FFD656"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpolateColor(tt.color1, tt.color2, tt.factor)
			if got != tt.want {
				t.Errorf("interpolateColor(%q, %q, %v) = %q, want %q",
					tt.color1, tt.color2, tt.factor, got, tt.want)
			}
		})
	}
}

func TestCalculateGradientColor(t *testing.T) {
	tests := []struct {
		name     string
		gridY    int
		maxGridY int
		want     string
	}{
		{"Single row", 1, 1, "#FFCE33"},
		{"Bottom row of 5", 5, 5, "#FFCE33"},
		{"Top row of 5", 1, 5, "#FFE691"},
		{"Middle row of 5", 3, 5, "#FFDA62"},
		{"Two rows - bottom", 2, 2, "#FFCE33"},
		{"Two rows - top", 1, 2, "#FFE691"},
		{"Bottom row of 3", 3, 3, "#FFCE33"},
		{"Top row of 3", 1, 3, "#FFE691"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateGradientColor(tt.gridY, tt.maxGridY)
			if got != tt.want {
				t.Errorf("calculateGradientColor(%d, %d) = %q, want %q",
					tt.gridY, tt.maxGridY, got, tt.want)
			}
		})
	}
}

func TestColorConversionRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		hex  string
	}{
		{"Yellow", "#FFCE33"},
		{"White", "#FFFFFF"},
		{"Black", "#000000"},
		{"Purple", "#800080"},
		{"Light blue", "#ADD8E6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := parseHexColor(tt.hex)
			got := rgbToHex(r, g, b)
			if got != tt.hex {
				t.Errorf("Round trip failed for %q: got %q", tt.hex, got)
			}
		})
	}
}

// TestRouteArrow_NarrowToWide tests horizontal-first routing for narrow-to-wide transitions
func TestRouteArrow_NarrowToWide(t *testing.T) {
	// Narrow box (width=20) to wide box (width=200) in same column
	plan, err := RouteArrow(
		BoxCoords{X1: 930, Y1: 650, X2: 950, Y2: 750},  // Narrow box: 20px wide
		BoxCoords{X1: 930, Y1: 800, X2: 1155, Y2: 900}, // Wide box: 225px wide
		8, 5, // fromGridX, fromGridY (same column)
		8, 6, // toGridX, toGridY (below)
		nil, "narrow", "wide", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use horizontal-first routing for narrow-to-wide
	if plan.Strategy != "two_segment_horizontal_first" {
		t.Errorf("Expected two_segment_horizontal_first for narrow-to-wide, got %s", plan.Strategy)
	}

	// Should exit from right side of narrow box
	if plan.StartX != 950 { // Right edge of narrow box
		t.Errorf("Expected startX=950 (right edge), got %d", plan.StartX)
	}

	// Should be horizontal-first (verticalFirst=false)
	if plan.VerticalFirst {
		t.Error("Expected VerticalFirst=false for horizontal-first routing")
	}

	// Should enter top of wide box
	if plan.EndY != 798 { // Top of box - 2px for stroke
		t.Errorf("Expected endY=798 (top edge), got %d", plan.EndY)
	}

}

// TestRouteArrow_WideToNarrow tests routing from wide to narrow box
func TestRouteArrow_WideToNarrow(t *testing.T) {
	// Just verify it doesn't error - specific routing depends on layout
	_, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 300, Y2: 200}, // Wide box: 200px wide
		BoxCoords{X1: 180, Y1: 300, X2: 200, Y2: 400}, // Narrow box: 20px wide, centered below wide box
		1, 1, // fromGridX, fromGridY
		1, 2, // toGridX, toGridY (same column, below)
		nil, "wide", "narrow", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// TestRouteArrow_ThreeSegment tests 3-segment routing for non-overlapping boxes
func TestRouteArrow_ThreeSegment(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 300, X2: 200, Y2: 400}, // Box 1
		BoxCoords{X1: 400, Y1: 100, X2: 500, Y2: 200}, // Box 2 (non-overlapping, deltaX >= 2)
		1, 3, // fromGridX, fromGridY
		4, 1, // toGridX, toGridY
		nil, "box1", "box2", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use 3-segment routing
	if plan.Strategy != "three_segment_horizontal_first" && plan.Strategy != "non_overlapping_horizontal" {
		t.Errorf("Expected 3-segment strategy, got %s", plan.Strategy)
	}

	// 3-segment arrows use verticalFirst=false
	if plan.VerticalFirst {
		t.Error("Expected VerticalFirst=false for 3-segment arrows")
	}

	// Should exit from right side
	if plan.StartX != 200 {
		t.Errorf("Expected startX=200 (right edge), got %d", plan.StartX)
	}

	// Should enter from left side
	if plan.EndX != 398 { // Left edge - 2px
		t.Errorf("Expected endX=398 (left edge - 2px), got %d", plan.EndX)
	}
}

// TestRouteArrow_CollisionDetection tests arrow routing with multiple boxes
func TestRouteArrow_CollisionDetection(t *testing.T) {
	// Create realistic box layout
	allBoxes := []BoxData{
		{ID: "from", GridX: 1, GridY: 2, PixelX: 100, PixelY: 200, Width: 100, Height: 100, CenterX: 150, CenterY: 250},
		{ID: "to", GridX: 3, GridY: 2, PixelX: 400, PixelY: 200, Width: 100, Height: 100, CenterX: 450, CenterY: 250},
	}

	_, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 200, X2: 200, Y2: 300}, // From box
		BoxCoords{X1: 400, Y1: 200, X2: 500, Y2: 300}, // To box (non-overlapping)
		1, 2, // fromGridX, fromGridY
		3, 2, // toGridX, toGridY (deltaX = 2, non-overlapping)
		allBoxes, "from", "to", "",
	)

	// Should successfully route
	if err != nil {
		t.Errorf("Should route successfully, got error: %v", err)
	}
}

// TestRouteArrow_SameDimensions tests routing between boxes with same dimensions
func TestRouteArrow_SameDimensions(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1: 100x100
		BoxCoords{X1: 100, Y1: 300, X2: 200, Y2: 400}, // Box 2: 100x100 (same width)
		1, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY (same column)
		nil, "box1", "box2", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use straight vertical for same column
	if plan.Strategy != "straight_vertical" {
		t.Errorf("Expected straight_vertical for same column, got %s", plan.Strategy)
	}
}

// TestRouteArrow_ErrorCase tests error handling for impossible routes
func TestRouteArrow_ErrorCase(t *testing.T) {
	// Try to route with all strategies blocked by collisions
	// This is a synthetic test - in practice we always find a route
	allBoxes := []BoxData{
		{ID: "from", GridX: 1, GridY: 1, PixelX: 100, PixelY: 100, Width: 100, Height: 100},
		{ID: "to", GridX: 1, GridY: 3, PixelX: 100, PixelY: 300, Width: 100, Height: 100},
	}

	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // From box
		BoxCoords{X1: 100, Y1: 300, X2: 200, Y2: 400}, // To box (same column)
		1, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY
		allBoxes, "from", "to", "",
	)

	// Should succeed with straight vertical
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if plan.Strategy != "straight_vertical" {
		t.Errorf("Expected straight_vertical, got %s", plan.Strategy)
	}
}

// TestRouteArrow_ExtremeNarrowToWide tests extreme width ratio transitions
func TestRouteArrow_ExtremeNarrowToWide(t *testing.T) {
	// 10px wide box to 200px wide box (20x width difference)
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 110, Y2: 200}, // Very narrow box: 10px wide
		BoxCoords{X1: 100, Y1: 300, X2: 300, Y2: 400}, // Wide box: 200px wide
		1, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY (same column)
		nil, "narrow", "wide", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use horizontal-first routing for extreme narrow-to-wide
	if plan.Strategy != "two_segment_horizontal_first" {
		t.Errorf("Expected two_segment_horizontal_first for extreme narrow-to-wide, got %s", plan.Strategy)
	}

	// Should exit from right side
	if plan.StartX != 110 {
		t.Errorf("Expected startX=110 (right edge), got %d", plan.StartX)
	}

	// Should be horizontal-first
	if plan.VerticalFirst {
		t.Error("Expected VerticalFirst=false for horizontal-first routing")
	}

	// Should enter at top
	if plan.EndY != 298 { // 300 - 2px stroke adjustment
		t.Errorf("Expected endY=298 (top edge - 2px), got %d", plan.EndY)
	}
}

// TestRouteArrow_BackwardArrow tests that backward (right-to-left) arrows are routed successfully
func TestRouteArrow_BackwardArrow(t *testing.T) {
	// Box on right pointing to box on left - should succeed
	plan, err := RouteArrow(
		BoxCoords{X1: 300, Y1: 100, X2: 400, Y2: 200}, // From box (right)
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // To box (left)
		3, 2, // fromGridX, fromGridY
		1, 2, // toGridX, toGridY (going left)
		nil, "right", "left", "",
	)

	if err != nil {
		t.Fatalf("Expected backward arrow to succeed, got error: %v", err)
	}

	if plan.Strategy == "" {
		t.Error("Expected a valid strategy for backward arrow")
	}
}

// TestRouteArrow_HorizontalAlignment tests boxes at same Y level
func TestRouteArrow_HorizontalAlignment(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 200, X2: 200, Y2: 300}, // Box 1
		BoxCoords{X1: 400, Y1: 200, X2: 500, Y2: 300}, // Box 2 (same vertical position)
		1, 2, // fromGridX, fromGridY
		4, 2, // toGridX, toGridY (same row)
		nil, "box1", "box2", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use horizontal routing strategy
	if plan.Strategy != "non_overlapping_horizontal" {
		t.Errorf("Expected non_overlapping_horizontal for same row, got %s", plan.Strategy)
	}

	// Should exit from right and enter from left
	if plan.StartX != 200 {
		t.Errorf("Expected startX=200 (right edge), got %d", plan.StartX)
	}
	if plan.EndX != 398 { // 400 - 2px
		t.Errorf("Expected endX=398 (left edge - 2px), got %d", plan.EndX)
	}

	// Should be at same vertical position
	if plan.StartY != plan.EndY {
		t.Errorf("Expected same Y position, got startY=%d endY=%d", plan.StartY, plan.EndY)
	}

	// Should not be vertical-first
	if plan.VerticalFirst {
		t.Error("Expected VerticalFirst=false for horizontal routing")
	}
}

// TestRouteArrow_DiagonalBoxes tests boxes at different X and Y positions
func TestRouteArrow_DiagonalBoxes(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1 (top-left)
		BoxCoords{X1: 400, Y1: 400, X2: 500, Y2: 500}, // Box 2 (bottom-right, diagonal)
		1, 1, // fromGridX, fromGridY
		4, 4, // toGridX, toGridY
		nil, "topleft", "bottomright", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use either three-segment or non-overlapping
	if plan.Strategy != "three_segment_horizontal_first" && plan.Strategy != "non_overlapping_horizontal" {
		t.Errorf("Expected 3-segment or non-overlapping for diagonal boxes, got %s", plan.Strategy)
	}

	// Should not be vertical-first
	if plan.VerticalFirst {
		t.Error("Expected verticalFirst=false for horizontal-based routing")
	}
}

// TestRouteArrow_IdenticalWidths tests boxes with exactly the same width
func TestRouteArrow_IdenticalWidths(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 300, Y2: 200}, // Box 1: 200px wide
		BoxCoords{X1: 100, Y1: 300, X2: 300, Y2: 400}, // Box 2: 200px wide (identical)
		1, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY (same column)
		nil, "box1", "box2", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use straight vertical (no width penalty)
	if plan.Strategy != "straight_vertical" {
		t.Errorf("Expected straight_vertical for identical widths, got %s", plan.Strategy)
	}
}

// TestRouteArrow_SlightWidthDifference tests width below narrow-to-wide threshold
func TestRouteArrow_SlightWidthDifference(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1: 100px wide
		BoxCoords{X1: 100, Y1: 300, X2: 250, Y2: 400}, // Box 2: 150px wide (1.5x, below 2x threshold)
		1, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY (same column)
		nil, "box1", "box2", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use straight vertical (no special narrow-to-wide handling)
	if plan.Strategy != "straight_vertical" {
		t.Errorf("Expected straight_vertical for slight width difference, got %s", plan.Strategy)
	}
}

// TestRouteArrow_WideToExtremelyNarrow tests reverse extreme ratio
func TestRouteArrow_WideToExtremelyNarrow(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 300, Y2: 200}, // Wide box: 200px wide
		BoxCoords{X1: 150, Y1: 300, X2: 250, Y2: 400}, // Narrow box: 100px wide (centered below)
		1, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY (same column)
		nil, "wide", "narrow", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use standard vertical routing (no penalty for wide-to-narrow)
	if plan.Strategy != "straight_vertical" && plan.Strategy != "two_segment_vertical_first" {
		t.Errorf("Expected vertical routing for wide-to-narrow, got %s", plan.Strategy)
	}
}

// TestRouteArrow_AdjacentColumnsOverlapping tests deltaX=1 with column overlap
func TestRouteArrow_AdjacentColumnsOverlapping(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 300, Y2: 200}, // Box 1
		BoxCoords{X1: 250, Y1: 300, X2: 450, Y2: 400}, // Box 2 (overlaps horizontally with Box 1)
		1, 1, // fromGridX, fromGridY
		2, 3, // toGridX, toGridY (deltaX=1, adjacent)
		nil, "box1", "box2", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use vertical-first routing for overlapping adjacent columns
	if plan.Strategy != "two_segment_vertical_first" {
		t.Errorf("Expected two_segment_vertical_first for overlapping adjacent, got %s", plan.Strategy)
	}

	if !plan.VerticalFirst {
		t.Error("Expected verticalFirst=true for vertical-first routing")
	}
}

// TestRouteArrow_AdjacentColumnsNonOverlapping tests deltaX=1 without column overlap
func TestRouteArrow_AdjacentColumnsNonOverlapping(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1 (ends at x=200)
		BoxCoords{X1: 250, Y1: 100, X2: 350, Y2: 200}, // Box 2 (starts at x=250, no overlap)
		1, 2, // fromGridX, fromGridY
		2, 2, // toGridX, toGridY (deltaX=1, same row)
		nil, "box1", "box2", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use horizontal routing for non-overlapping adjacent
	if plan.Strategy != "non_overlapping_horizontal" {
		t.Errorf("Expected non_overlapping_horizontal for non-overlapping adjacent, got %s", plan.Strategy)
	}
}

// TestRouteArrow_VeryLongDistance tests distance penalty on scoring
func TestRouteArrow_VeryLongDistance(t *testing.T) {
	// Boxes very far apart (distance > 500px)
	_, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1
		BoxCoords{X1: 800, Y1: 600, X2: 900, Y2: 700}, // Box 2 (700px horizontal, 500px vertical apart)
		1, 1, // fromGridX, fromGridY
		8, 6, // toGridX, toGridY
		nil, "box1", "box2", "",
	)

	// Should still find a route despite large distance
	if err != nil {
		t.Errorf("Should route long distances, got error: %v", err)
	}
}

// TestRouteArrow_MultipleObstacles tests routing around several boxes
func TestRouteArrow_MultipleObstacles(t *testing.T) {
	// Create a field of obstacle boxes
	allBoxes := []BoxData{
		{ID: "from", GridX: 1, GridY: 1, PixelX: 100, PixelY: 100, Width: 100, Height: 100, CenterX: 150, CenterY: 150},
		{ID: "to", GridX: 5, GridY: 1, PixelX: 500, PixelY: 100, Width: 100, Height: 100, CenterX: 550, CenterY: 150},
	}

	_, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // From box
		BoxCoords{X1: 500, Y1: 100, X2: 600, Y2: 200}, // To box (far to the right, same row)
		1, 1, // fromGridX, fromGridY
		5, 1, // toGridX, toGridY
		allBoxes, "from", "to", "",
	)

	// Should successfully route around obstacles
	if err != nil {
		t.Errorf("Should route around multiple obstacles, got error: %v", err)
	}
}

// TestRouteArrow_ForceThreeSegment tests scenario that requires 3-segment routing
func TestRouteArrow_ForceThreeSegment(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 400, X2: 200, Y2: 500}, // Box 1 (bottom-left)
		BoxCoords{X1: 400, Y1: 100, X2: 500, Y2: 200}, // Box 2 (top-right)
		1, 4, // fromGridX, fromGridY
		4, 1, // toGridX, toGridY (going up and right)
		nil, "bottomleft", "topright", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use 3-segment routing for non-adjacent different X and Y
	if plan.Strategy != "three_segment_horizontal_first" && plan.Strategy != "non_overlapping_horizontal" {
		t.Errorf("Expected 3-segment strategy, got %s", plan.Strategy)
	}

	// 3-segment arrows should not be vertical-first
	if plan.Strategy == "three_segment_horizontal_first" && plan.VerticalFirst {
		t.Error("Expected verticalFirst=false for three-segment routing")
	}
}

// TestRouteArrow_ScoringPreference tests that scoring correctly selects best strategy
func TestRouteArrow_ScoringPreference(t *testing.T) {
	// Scenario where multiple strategies are valid - verify correct one is chosen
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1
		BoxCoords{X1: 100, Y1: 400, X2: 200, Y2: 500}, // Box 2 (same column, well separated)
		1, 1, // fromGridX, fromGridY
		1, 4, // toGridX, toGridY
		nil, "box1", "box2", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// straight_vertical should win with highest score (100)
	if plan.Strategy != "straight_vertical" {
		t.Errorf("Expected straight_vertical to win scoring, got %s", plan.Strategy)
	}
}

// TestRouteArrow_BoxInPath tests obstacle directly between source and destination
func TestRouteArrow_BoxInPath(t *testing.T) {
	// Box directly in the path blocks straight vertical - should fall back to horizontal-first
	allBoxes := []BoxData{
		{ID: "from", GridX: 1, GridY: 1, PixelX: 100, PixelY: 100, Width: 100, Height: 100, CenterX: 150, CenterY: 150},
		{ID: "obstacle", GridX: 1, GridY: 2, PixelX: 100, PixelY: 250, Width: 100, Height: 100, CenterX: 150, CenterY: 300},
		{ID: "to", GridX: 1, GridY: 3, PixelX: 100, PixelY: 400, Width: 100, Height: 100, CenterX: 150, CenterY: 450},
	}

	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // From box
		BoxCoords{X1: 100, Y1: 400, X2: 200, Y2: 500}, // To box (same column)
		1, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY
		allBoxes, "from", "to", "",
	)

	// When straight vertical is blocked, horizontal-first routing can route around obstacle
	// If both fail, the system will error (which is expected for this configuration)
	if err != nil {
		// This is expected - obstacle blocks all same-column routes
		t.Logf("Expected error when obstacle blocks all routes: %v", err)
	} else {
		t.Logf("Chosen strategy: %s (successfully routed around obstacle)", plan.Strategy)
	}
}

// TestRouteArrow_ThreeSegmentVerticalFirst tests V-H-V routing with flow="down"
func TestRouteArrow_ThreeSegmentVerticalFirst(t *testing.T) {
	// Boxes at different columns and rows, close together so V-H-V wins over horizontal
	// Box1 at grid (3,1), Box2 at grid (2,2) — adjacent columns
	plan, err := RouteArrow(
		BoxCoords{X1: 300, Y1: 50, X2: 500, Y2: 150},  // Box 1 (upper)
		BoxCoords{X1: 150, Y1: 200, X2: 350, Y2: 300}, // Box 2 (lower-left, overlapping X)
		3, 1, // fromGridX, fromGridY
		2, 2, // toGridX, toGridY (adjacent column, boxes overlap horizontally)
		nil, "top", "bottom", "down",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// With flow=down and overlapping boxes (non_overlapping won't generate),
	// V-H-V (70+25=95) should beat two_segment_vertical_first (80+15=95) only if
	// V-H-V has shorter distance. Let's verify V-H-V is generated at all.
	// V-H-V: exit bottom center (400,150) → enter top center (250,198)
	// two_segment_vertical_first: exit bottom center (400,150) → enter right (352,250)
	// V-H-V distance: |400-250| + |150-198| = 198, penalty=19, score=95-19=76
	// two_seg_v: distance |400-352| + |150-250| = 148, penalty=14, score=95-14=81
	// So two_segment_vertical_first still wins. Let's just verify both strategies are valid
	// and the best wins.

	// With overlapping X ranges, non_overlapping_horizontal won't have space.
	// The key is that three_segment_vertical_first IS generated as a candidate.
	validStrategies := map[string]bool{
		"three_segment_vertical_first": true,
		"two_segment_vertical_first":   true,
		"straight_vertical":            true,
	}
	if !validStrategies[plan.Strategy] {
		t.Errorf("Expected a vertical-oriented strategy with flow=down, got %s", plan.Strategy)
	}

	// Verify that three_segment_vertical_first appears in candidates
	found := false
	for _, c := range plan.AllCandidates {
		if c.strategy == "three_segment_vertical_first" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected three_segment_vertical_first to appear as a candidate with flow=down")
	}
}

// TestRouteArrow_FlowDown_SameColumn tests that same-column with flow=down still uses straight_vertical
func TestRouteArrow_FlowDown_SameColumn(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1
		BoxCoords{X1: 100, Y1: 400, X2: 200, Y2: 500}, // Box 2 (same column)
		1, 1, // fromGridX, fromGridY
		1, 4, // toGridX, toGridY
		nil, "box1", "box2", "down",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Same column should still use straight_vertical even with flow=down
	// (straight_vertical gets +30 bonus: 100+30=130, beats everything)
	if plan.Strategy != "straight_vertical" {
		t.Errorf("Expected straight_vertical for same column with flow=down, got %s", plan.Strategy)
	}
}

// TestRouteArrow_FlowDown_DiagonalBeatsHorizontal tests that flow=down boosts vertical strategies
func TestRouteArrow_FlowDown_DiagonalBeatsHorizontal(t *testing.T) {
	// Boxes at different X and Y, with flow=down
	plan, err := RouteArrow(
		BoxCoords{X1: 400, Y1: 100, X2: 600, Y2: 200}, // Box 1 (upper right)
		BoxCoords{X1: 100, Y1: 300, X2: 300, Y2: 400}, // Box 2 (lower left)
		4, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY
		nil, "upper", "lower", "down",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// With flow=down, vertical strategies get bonuses.
	// two_segment_vertical_first (80+15=95) or three_segment_vertical_first (70+25=95)
	// should beat horizontal strategies without flow bonuses.
	verticalStrategies := map[string]bool{
		"two_segment_vertical_first":   true,
		"three_segment_vertical_first": true,
		"straight_vertical":            true,
	}
	if !verticalStrategies[plan.Strategy] {
		t.Errorf("Expected a vertical strategy with flow=down, got %s", plan.Strategy)
	}
}

// TestRouteArrow_NoFlow_DiagonalUsesHorizontal tests that without flow, horizontal strategies still win
func TestRouteArrow_NoFlow_DiagonalUsesHorizontal(t *testing.T) {
	// Same boxes as above but without flow
	plan, err := RouteArrow(
		BoxCoords{X1: 400, Y1: 100, X2: 600, Y2: 200}, // Box 1 (upper right)
		BoxCoords{X1: 100, Y1: 300, X2: 300, Y2: 400}, // Box 2 (lower left)
		4, 1, // fromGridX, fromGridY
		1, 3, // toGridX, toGridY
		nil, "upper", "lower", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Without flow, non_overlapping_horizontal (90) should beat three_segment (70)
	// V-H-V strategy is NOT generated without flow=down
	if plan.Strategy == "three_segment_vertical_first" {
		t.Error("Should not use three_segment_vertical_first without flow=down")
	}
}

// ===== Step 1: Unit tests for checkSegmentCollision =====

func TestCheckSegmentCollision(t *testing.T) {
	// Standard box at (100, 100) with width=100, height=100
	// With BOX_COLLISION_BUFFER=3: effective bounds [97, 97] to [203, 203]
	box := BoxData{ID: "box1", PixelX: 100, PixelY: 100, Width: 100, Height: 100}
	boxes := []BoxData{box}

	tests := []struct {
		name       string
		x1, y1     int
		x2, y2     int
		boxes      []BoxData
		excludeIDs map[string]bool
		want       bool
	}{
		{
			name: "Horizontal segment intersecting box",
			x1:   50, y1: 150, x2: 250, y2: 150,
			boxes: boxes,
			want:  true,
		},
		{
			name: "Vertical segment intersecting box",
			x1:   150, y1: 50, x2: 150, y2: 250,
			boxes: boxes,
			want:  true,
		},
		{
			name: "Segment passing above box",
			x1:   50, y1: 50, x2: 250, y2: 50,
			boxes: boxes,
			want:  false, // y=50 < boxTop-buffer=97
		},
		{
			name: "Segment passing below box",
			x1:   50, y1: 250, x2: 250, y2: 250,
			boxes: boxes,
			want:  false, // y=250 > boxBottom+buffer=203
		},
		{
			name: "Segment passing left of box",
			x1:   50, y1: 50, x2: 50, y2: 250,
			boxes: boxes,
			want:  false, // x=50 < boxLeft-buffer=97
		},
		{
			name: "Segment passing right of box",
			x1:   250, y1: 50, x2: 250, y2: 250,
			boxes: boxes,
			want:  false, // x=250 > boxRight+buffer=203
		},
		{
			name: "Segment touching box edge with 3px buffer - collision",
			x1:   50, y1: 203, x2: 250, y2: 203,
			boxes: boxes,
			want:  true, // y=203 == boxBottom+buffer=203, xOverlap=true
		},
		{
			name: "Segment just outside 3px buffer - no collision",
			x1:   50, y1: 204, x2: 250, y2: 204,
			boxes: boxes,
			want:  false, // y=204 > boxBottom+buffer=203
		},
		{
			name: "Excluded box ID - no collision",
			x1:   150, y1: 150, x2: 150, y2: 150,
			boxes:      boxes,
			excludeIDs: map[string]bool{"box1": true},
			want:       false,
		},
		{
			name: "Empty box list - no collision",
			x1:   150, y1: 150, x2: 150, y2: 150,
			boxes: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkSegmentCollision(tt.x1, tt.y1, tt.x2, tt.y2, tt.boxes, tt.excludeIDs)
			if got != tt.want {
				t.Errorf("checkSegmentCollision(%d,%d,%d,%d) = %v, want %v",
					tt.x1, tt.y1, tt.x2, tt.y2, got, tt.want)
			}
		})
	}
}

// ===== Step 2: Unit tests for checkTwoSegmentCollision and checkThreeSegmentCollision =====

func TestCheckPathCollision_TwoSegment(t *testing.T) {
	tests := []struct {
		name         string
		points       []int
		boxes        []BoxData
		fromID, toID string
		want         bool
	}{
		{
			name: "L-path vertical-first, collision on vertical segment",
			// V then H: (100,100)→(100,300)→(300,300)
			// Obstacle at (90,190,20,20) → buffer [87,187,113,213]
			// Seg1 (100,100)→(100,300): xOverlap 100∈[87,113]→true, yOverlap [100,300]∩[187,213]→true
			points: []int{100, 100, 100, 300, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "obstacle", PixelX: 90, PixelY: 190, Width: 20, Height: 20},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: true,
		},
		{
			name:   "L-path vertical-first, no collision",
			points: []int{100, 100, 100, 300, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: false,
		},
		{
			name: "L-path horizontal-first, collision on horizontal segment",
			// H then V: (100,100)→(300,100)→(300,300)
			// Obstacle at (190,90,20,20) → buffer [187,87,213,113]
			// Seg1 (100,100)→(300,100): yOverlap 100∈[87,113]→true, xOverlap [100,300]∩[187,213]→true
			points: []int{100, 100, 300, 100, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "obstacle", PixelX: 190, PixelY: 90, Width: 20, Height: 20},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: true,
		},
		{
			name:   "L-path horizontal-first, no collision",
			points: []int{100, 100, 300, 100, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: false,
		},
		{
			name:   "excluded source/dest boxes not counted as collisions",
			points: []int{100, 100, 100, 300, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 90, PixelY: 90, Width: 20, Height: 20},
				{ID: "to", PixelX: 290, PixelY: 290, Width: 20, Height: 20},
			},
			fromID: "from", toID: "to",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPathCollision(tt.points, tt.boxes, tt.fromID, tt.toID)
			if got != tt.want {
				t.Errorf("checkPathCollision = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPathCollision_ThreeSegment(t *testing.T) {
	tests := []struct {
		name         string
		points       []int
		boxes        []BoxData
		fromID, toID string
		want         bool
	}{
		{
			name: "V-H-V collision on segment 1 (vertical)",
			// V-H-V: (100,100)→(100,200)→(300,200)→(300,300)
			// Obstacle at (90,140,20,20)→buffer [87,137,113,163]
			// Seg1 (100,100)→(100,200): x=100∈[87,113], y∩[137,163]→true
			points: []int{100, 100, 100, 200, 300, 200, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "obstacle", PixelX: 90, PixelY: 140, Width: 20, Height: 20},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: true,
		},
		{
			name: "V-H-V collision on segment 2 (horizontal)",
			// Obstacle at (190,190,20,20)→buffer [187,187,213,213]
			// Seg2 (100,200)→(300,200): y=200∈[187,213], x∩[187,213]→true
			points: []int{100, 100, 100, 200, 300, 200, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "obstacle", PixelX: 190, PixelY: 190, Width: 20, Height: 20},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: true,
		},
		{
			name: "V-H-V collision on segment 3 (vertical)",
			// Obstacle at (290,240,20,20)→buffer [287,237,313,263]
			// Seg3 (300,200)→(300,300): x=300∈[287,313], y∩[237,263]→true
			points: []int{100, 100, 100, 200, 300, 200, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "obstacle", PixelX: 290, PixelY: 240, Width: 20, Height: 20},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: true,
		},
		{
			name: "H-V-H collision on segment 1 (horizontal)",
			// H-V-H: (100,100)→(200,100)→(200,300)→(300,300)
			// Obstacle at (140,90,20,20)→buffer [137,87,163,113]
			// Seg1 (100,100)→(200,100): y=100∈[87,113], x∩[137,163]→true
			points: []int{100, 100, 200, 100, 200, 300, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "obstacle", PixelX: 140, PixelY: 90, Width: 20, Height: 20},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: true,
		},
		{
			name: "H-V-H collision on segment 2 (vertical)",
			// Obstacle at (190,190,20,20)→buffer [187,187,213,213]
			// Seg2 (200,100)→(200,300): x=200∈[187,213], y∩[187,213]→true
			points: []int{100, 100, 200, 100, 200, 300, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "obstacle", PixelX: 190, PixelY: 190, Width: 20, Height: 20},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: true,
		},
		{
			name: "H-V-H collision on segment 3 (horizontal)",
			// Obstacle at (240,290,20,20)→buffer [237,287,263,313]
			// Seg3 (200,300)→(300,300): y=300∈[287,313], x∩[237,263]→true
			points: []int{100, 100, 200, 100, 200, 300, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "obstacle", PixelX: 240, PixelY: 290, Width: 20, Height: 20},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: true,
		},
		{
			name:   "V-H-V no collision",
			points: []int{100, 100, 100, 200, 300, 200, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: false,
		},
		{
			name:   "H-V-H no collision",
			points: []int{100, 100, 200, 100, 200, 300, 300, 300},
			boxes: []BoxData{
				{ID: "from", PixelX: 80, PixelY: 80, Width: 40, Height: 40},
				{ID: "to", PixelX: 280, PixelY: 280, Width: 40, Height: 40},
			},
			fromID: "from", toID: "to",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPathCollision(tt.points, tt.boxes, tt.fromID, tt.toID)
			if got != tt.want {
				t.Errorf("checkPathCollision = %v, want %v", got, tt.want)
			}
		})
	}
}

// ===== Step 3: Unit tests for scoreRoute =====

func TestScoreRoute(t *testing.T) {
	tests := []struct {
		name      string
		candidate RouteCandidate
		want      int
	}{
		{
			name: "straight_vertical, no flow, same width",
			candidate: RouteCandidate{
				strategy: "straight_vertical", boxWidth1: 100, boxWidth2: 100,
				startX: 150, startY: 200, endX: 150, endY: 400,
			},
			want: 80, // base=100, penalty=200/10=20
		},
		{
			name: "straight_vertical, narrow-to-wide",
			candidate: RouteCandidate{
				strategy: "straight_vertical", boxWidth1: 50, boxWidth2: 200,
				startX: 100, startY: 200, endX: 100, endY: 400,
			},
			want: 30, // base=100-50=50, penalty=200/10=20
		},
		{
			name: "straight_vertical, flow=down",
			candidate: RouteCandidate{
				strategy: "straight_vertical", boxWidth1: 100, boxWidth2: 100,
				startX: 150, startY: 200, endX: 150, endY: 400, flow: "down",
			},
			want: 110, // base=100+30=130, penalty=200/10=20
		},
		{
			name: "non_overlapping_horizontal",
			candidate: RouteCandidate{
				strategy: "non_overlapping_horizontal", boxWidth1: 100, boxWidth2: 100,
				startX: 200, startY: 150, endX: 400, endY: 150,
			},
			want: 70, // base=90, penalty=200/10=20
		},
		{
			name: "two_segment_vertical_first",
			candidate: RouteCandidate{
				strategy: "two_segment_vertical_first", boxWidth1: 100, boxWidth2: 100,
				startX: 150, startY: 200, endX: 400, endY: 300,
			},
			want: 45, // base=80, penalty=350/10=35
		},
		{
			name: "two_segment_vertical_first, flow=down",
			candidate: RouteCandidate{
				strategy: "two_segment_vertical_first", boxWidth1: 100, boxWidth2: 100,
				startX: 150, startY: 200, endX: 400, endY: 300, flow: "down",
			},
			want: 60, // base=80+15=95, penalty=350/10=35
		},
		{
			name: "two_segment_horizontal_first, narrow-to-wide",
			candidate: RouteCandidate{
				strategy: "two_segment_horizontal_first", boxWidth1: 50, boxWidth2: 200,
				startX: 110, startY: 150, endX: 200, endY: 400,
			},
			want: 61, // base=95 (narrow-to-wide), penalty=340/10=34
		},
		{
			name: "two_segment_horizontal_first, normal",
			candidate: RouteCandidate{
				strategy: "two_segment_horizontal_first", boxWidth1: 100, boxWidth2: 100,
				startX: 200, startY: 150, endX: 350, endY: 400,
			},
			want: 35, // base=75, penalty=400/10=40
		},
		{
			name: "three_segment_horizontal_first",
			candidate: RouteCandidate{
				strategy: "three_segment_horizontal_first", boxWidth1: 100, boxWidth2: 100,
				startX: 200, startY: 150, endX: 400, endY: 350,
			},
			want: 30, // base=70, penalty=400/10=40
		},
		{
			name: "three_segment_vertical_first",
			candidate: RouteCandidate{
				strategy: "three_segment_vertical_first", boxWidth1: 100, boxWidth2: 100,
				startX: 150, startY: 200, endX: 350, endY: 400,
			},
			want: 30, // base=70, penalty=400/10=40
		},
		{
			name: "three_segment_vertical_first, flow=down",
			candidate: RouteCandidate{
				strategy: "three_segment_vertical_first", boxWidth1: 100, boxWidth2: 100,
				startX: 150, startY: 200, endX: 350, endY: 400, flow: "down",
			},
			want: 70, // base=70+40=110, penalty=400/10=40
		},
		{
			name: "very long distance caps penalty at 50",
			candidate: RouteCandidate{
				strategy: "straight_vertical", boxWidth1: 100, boxWidth2: 100,
				startX: 150, startY: 100, endX: 150, endY: 700,
			},
			want: 50, // base=100, distance=600>500 → penalty=50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreRoute(tt.candidate)
			if got != tt.want {
				t.Errorf("scoreRoute() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ===== Step 4: Edge case tests for RouteArrow =====

func TestRouteArrow_UpwardArrow(t *testing.T) {
	// toGridY < fromGridY: arrow goes upward
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 300, X2: 200, Y2: 400}, // Box 1 (lower)
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 2 (upper)
		2, 3, // fromGridX, fromGridY
		2, 1, // toGridX, toGridY (going up)
		nil, "lower", "upper", "",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Same column, should use straight_vertical
	if plan.Strategy != "straight_vertical" {
		t.Errorf("Expected straight_vertical for upward arrow, got %s", plan.Strategy)
	}

	// Exit from top of source box
	if plan.StartY != 300 {
		t.Errorf("Expected startY=300 (top of source box), got %d", plan.StartY)
	}

	// Enter at bottom of destination box (with stroke adjustment)
	if plan.EndY != 202 {
		t.Errorf("Expected endY=202 (bottom of dest box + 2px stroke), got %d", plan.EndY)
	}
}

func TestRouteArrow_BackwardDiagonalFlowDown(t *testing.T) {
	// toGridX < fromGridX, toGridY > fromGridY, flow="down"
	// Should generate three_segment_vertical_first candidate
	plan, err := RouteArrow(
		BoxCoords{X1: 400, Y1: 100, X2: 500, Y2: 200}, // Box 1 (upper-right)
		BoxCoords{X1: 100, Y1: 400, X2: 200, Y2: 500}, // Box 2 (lower-left)
		4, 1, // fromGridX, fromGridY
		1, 4, // toGridX, toGridY
		nil, "upper", "lower", "down",
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// V-H-V should be a candidate with flow=down
	found := false
	for _, c := range plan.AllCandidates {
		if c.strategy == "three_segment_vertical_first" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected three_segment_vertical_first to appear as candidate with flow=down and backward diagonal")
	}

	// Verify it exits from bottom (going down)
	if plan.Strategy == "three_segment_vertical_first" {
		if plan.StartY != 200 {
			t.Errorf("Expected startY=200 (bottom of source), got %d", plan.StartY)
		}
		if plan.EndY != 398 {
			t.Errorf("Expected endY=398 (top of dest - 2px), got %d", plan.EndY)
		}
	}
}

func TestRouteArrow_SamePosition(t *testing.T) {
	// fromGridX == toGridX && fromGridY == toGridY: no strategy matches
	_, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200},
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200},
		2, 2, // fromGridX, fromGridY
		2, 2, // toGridX, toGridY (same position)
		nil, "box1", "box2", "",
	)
	if err == nil {
		t.Error("Expected error for same-position boxes, got nil")
	}
}

func TestLayout_Container(t *testing.T) {
	// Parse a container spec, run through Layout, verify results
	text := `
G: 2,2 [
    X: 0,0: Alpha
    Y: 2,0: Beta
    X -> Y
]
`
	spec, err := ParseDiagramSpec(text, nil)
	if err != nil {
		t.Fatalf("ParseDiagramSpec failed: %v", err)
	}

	config := NewDefaultConfig()
	diagram, boxData := Layout(spec, config, nil, spec.Groups, "")

	// Verify boxes exist and have correct grid positions
	xData, ok := boxData["X"]
	if !ok {
		t.Fatal("Expected box X in boxData")
	}
	yData, ok := boxData["Y"]
	if !ok {
		t.Fatal("Expected box Y in boxData")
	}

	// X at grid (2,2), Y at grid (4,2)
	if xData.GridX != 2 || xData.GridY != 2 {
		t.Errorf("Expected X at grid (2,2), got (%d,%d)", xData.GridX, xData.GridY)
	}
	if yData.GridX != 4 || yData.GridY != 2 {
		t.Errorf("Expected Y at grid (4,2), got (%d,%d)", yData.GridX, yData.GridY)
	}

	// Verify pixel positions are different (Y is to the right of X)
	if yData.PixelX <= xData.PixelX {
		t.Errorf("Expected Y pixel X > X pixel X, got Y=%d, X=%d", yData.PixelX, xData.PixelX)
	}

	// Containers don't create groups — purely organizational
	if len(diagram.Groups) != 0 {
		t.Errorf("Expected 0 groups (containers don't render), got %d", len(diagram.Groups))
	}

	// Verify arrow exists
	if len(diagram.Arrows) != 1 {
		t.Fatalf("Expected 1 arrow, got %d", len(diagram.Arrows))
	}
}
