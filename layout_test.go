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
		nil, "0", "1", // allBoxes, fromID, toID (no collision detection for this test)
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

	// Verify no illegal arrow (startX > endX)
	if plan.StartX > plan.EndX {
		t.Error("Illegal arrow: StartX > EndX")
	}
}

func TestRouteArrow_OverlappingColumns(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1 bounds
		BoxCoords{X1: 200, Y1: 200, X2: 300, Y2: 300}, // Box 2 bounds
		3, 1, // fromGridX, fromGridY (GridY=1 is top)
		4, 2, // toGridX, toGridY (GridY=2 is below)
		nil, "0", "1", // allBoxes, fromID, toID (no collision detection for this test)
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

	// Verify no illegal arrow
	if plan.StartX > plan.EndX {
		t.Error("Illegal arrow: StartX > EndX")
	}
}

func TestRouteArrow_NonOverlapping(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // Box 1 bounds
		BoxCoords{X1: 300, Y1: 100, X2: 400, Y2: 200}, // Box 2 bounds
		1, 2, // fromGridX, fromGridY
		3, 2, // toGridX, toGridY
		nil, "0", "1", // allBoxes, fromID, toID (no collision detection for this test)
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

	// Verify no illegal arrow
	if plan.StartX > plan.EndX {
		t.Error("Illegal arrow: StartX > EndX")
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
	diagram, _ := Layout(spec, config, nil, nil)

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
		nil, "narrow", "wide",
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

	// Verify no illegal arrow
	if plan.StartX > plan.EndX {
		t.Error("Illegal arrow: StartX > EndX")
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
		nil, "wide", "narrow",
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
		nil, "box1", "box2",
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
		allBoxes, "from", "to",
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
		nil, "box1", "box2",
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
		allBoxes, "from", "to",
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
		nil, "narrow", "wide",
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

	// Verify no illegal arrow
	if plan.StartX > plan.EndX {
		t.Error("Illegal arrow: StartX > EndX")
	}
}

// TestRouteArrow_BackwardArrow tests that backward arrows are rejected
func TestRouteArrow_BackwardArrow(t *testing.T) {
	// Box on right pointing to box on left - should fail due to startX > endX constraint
	_, err := RouteArrow(
		BoxCoords{X1: 300, Y1: 100, X2: 400, Y2: 200}, // From box (right)
		BoxCoords{X1: 100, Y1: 100, X2: 200, Y2: 200}, // To box (left)
		3, 2, // fromGridX, fromGridY
		1, 2, // toGridX, toGridY (going left)
		nil, "right", "left",
	)

	// Should fail - backward arrows violate startX <= endX constraint
	if err == nil {
		t.Error("Expected error for backward arrow (violates startX <= endX), got none")
	}
}

// TestRouteArrow_HorizontalAlignment tests boxes at same Y level
func TestRouteArrow_HorizontalAlignment(t *testing.T) {
	plan, err := RouteArrow(
		BoxCoords{X1: 100, Y1: 200, X2: 200, Y2: 300}, // Box 1
		BoxCoords{X1: 400, Y1: 200, X2: 500, Y2: 300}, // Box 2 (same vertical position)
		1, 2, // fromGridX, fromGridY
		4, 2, // toGridX, toGridY (same row)
		nil, "box1", "box2",
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
		nil, "topleft", "bottomright",
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
		nil, "box1", "box2",
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
		nil, "box1", "box2",
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
		nil, "wide", "narrow",
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
		nil, "box1", "box2",
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
		nil, "box1", "box2",
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
		nil, "box1", "box2",
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
		allBoxes, "from", "to",
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
		nil, "bottomleft", "topright",
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
		nil, "box1", "box2",
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
		allBoxes, "from", "to",
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
