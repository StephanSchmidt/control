package main

import "fmt"

// Arrow routing constants
const (
	STROKE_ADJUSTMENT    = 2 // Adjustment for arrow entry/exit to account for box stroke width
	BOX_COLLISION_BUFFER = 3 // Buffer around boxes for collision detection
)

// Dimensions holds calculated diagram dimensions
type Dimensions struct {
	Width             int
	Height            int
	BoxWidth          int
	BoxHeight         int
	LeftMargin        int
	TopMargin         int
	BottomMargin      int
	CellUnits         float64
	VerticalCellUnits float64
}

// CalculateDimensions computes diagram dimensions from spec and config
func CalculateDimensions(maxGridX, maxGridY int, config DiagramConfig) Dimensions {
	cellUnits := (1.0 + config.GapUnits) * config.Stretch
	verticalCellUnits := 1.0 + config.VerticalGapUnits

	boxWidth := int(config.BoxWidthUnits * float64(config.GridUnit))
	boxHeight := config.GridUnit

	leftMargin := 60 + config.AxisOffset
	bottomMargin := 50 + config.AxisOffset
	topMargin := 50

	rightmostBoxEnd := leftMargin + int((float64(maxGridX-1)*cellUnits+config.BoxWidthUnits*config.Stretch)*float64(config.GridUnit))
	width := rightmostBoxEnd + 20

	contentHeight := int(float64(maxGridY) * verticalCellUnits * float64(config.GridUnit))
	height := bottomMargin + contentHeight + topMargin

	return Dimensions{
		Width:             width,
		Height:            height,
		BoxWidth:          boxWidth,
		BoxHeight:         boxHeight,
		LeftMargin:        leftMargin,
		TopMargin:         topMargin,
		BottomMargin:      bottomMargin,
		CellUnits:         cellUnits,
		VerticalCellUnits: verticalCellUnits,
	}
}

// EstimateLegendWidth estimates the pixel width needed for the legend area
// based on the longest label text. Returns 0 if no legend entries.
func EstimateLegendWidth(legend []LegendEntry) int {
	if len(legend) == 0 {
		return 0
	}

	maxLabelLen := 0
	for _, entry := range legend {
		if len(entry.Label) > maxLabelLen {
			maxLabelLen = len(entry.Label)
		}
	}

	// Estimate: ~8px per character at font size 14, plus square + gap + padding
	textWidth := maxLabelLen * 8
	return textWidth + legendSquareSize + legendTextGap + legendPadding*2
}

// BoxData represents box information needed for arrow routing
type BoxData struct {
	ID               string
	GridX, GridY     int
	PixelX, PixelY   int
	CenterX, CenterY int
	Width, Height    int
}

// ArrowRoute represents the calculated arrow path
type ArrowRoute struct {
	FromX         int
	FromY         int
	ToX           int
	ToY           int
	VerticalFirst bool
}

// checkSegmentCollision checks if a line segment (vertical or horizontal) intersects any boxes
func checkSegmentCollision(x1, y1, x2, y2 int, allBoxes []BoxData, excludeIDs map[string]bool) bool {
	// Normalize coordinates
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Check each box
	for _, box := range allBoxes {
		// Skip boxes we should exclude (source and destination)
		if excludeIDs != nil && excludeIDs[box.ID] {
			continue
		}

		// Expand box bounds slightly for stroke width
		boxLeft := box.PixelX - BOX_COLLISION_BUFFER
		boxRight := box.PixelX + box.Width + BOX_COLLISION_BUFFER
		boxTop := box.PixelY - BOX_COLLISION_BUFFER
		boxBottom := box.PixelY + box.Height + BOX_COLLISION_BUFFER

		// Check if line segment intersects this box
		// For a line segment to intersect a box, it must overlap in both dimensions
		xOverlap := x1 <= boxRight && x2 >= boxLeft
		yOverlap := y1 <= boxBottom && y2 >= boxTop

		if xOverlap && yOverlap {
			return true
		}
	}
	return false
}

// checkPathCollision checks if a polyline path collides with any boxes.
// points is a list of (x,y) pairs defining the path segments.
func checkPathCollision(points []int, allBoxes []BoxData, fromID, toID string) bool {
	excludeIDs := map[string]bool{fromID: true, toID: true}
	for i := 0; i < len(points)-2; i += 2 {
		if checkSegmentCollision(points[i], points[i+1], points[i+2], points[i+3], allBoxes, excludeIDs) {
			return true
		}
	}
	return false
}

// BoxCoords represents the bounding box coordinates of a box
type BoxCoords struct {
	X1, Y1, X2, Y2 int
}

// RoutingPlan represents the final routing plan for an arrow
type RoutingPlan struct {
	StartX, StartY, EndX, EndY int
	Strategy                   string
	VerticalFirst              bool
	NumSegments                int // 1=straight, 2=L-shape, 3=Z-shape
	AllCandidates              []RouteCandidate
}

// RouteCandidate represents a potential arrow route
type RouteCandidate struct {
	startX, startY, endX, endY int
	strategy                   string
	verticalFirst              bool
	score                      int
	boxWidth1, boxWidth2       int // Store box widths for scoring
	rejected                   bool
	rejectReason               string // "collision_detected", or empty if not rejected
	flow                       string // Flow direction hint (e.g., "down")
	segments                   []int  // Polyline points [x0,y0,x1,y1,...] for collision detection
}

// scoreRoute assigns a quality score to a route (higher is better)
func scoreRoute(candidate RouteCandidate) int {
	score := 0

	// Calculate width difference for narrow-to-wide transition detection
	widthDiff := candidate.boxWidth2 - candidate.boxWidth1
	if widthDiff < 0 {
		widthDiff = -widthDiff
	}
	isNarrowToWide := candidate.boxWidth1 > 0 && widthDiff > candidate.boxWidth1*2

	// Base score by strategy type
	switch candidate.strategy {
	case "straight_vertical":
		score += 100
		if isNarrowToWide {
			score -= 50
		}
	case "non_overlapping_horizontal":
		score += 90
	case "two_segment_vertical_first":
		score += 80
	case "two_segment_horizontal_first":
		if isNarrowToWide {
			score += 95
		} else {
			score += 75
		}
	case "three_segment_horizontal_first", "three_segment_vertical_first":
		score += 70
	}

	// Distance penalty (Manhattan distance, max 50)
	dx := candidate.endX - candidate.startX
	dy := candidate.endY - candidate.startY
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	distance := dx + dy
	if distance > 500 {
		score -= 50
	} else {
		score -= distance / 10
	}

	// Flow direction bonuses
	if candidate.flow == "down" {
		switch candidate.strategy {
		case "straight_vertical":
			score += 30
		case "three_segment_vertical_first":
			score += 40
		case "two_segment_vertical_first":
			score += 15
		}
	}

	return score
}

// validateAndAddCandidate validates a candidate route and adds it to the appropriate lists
// Returns true if the candidate is valid (not rejected)
func validateAndAddCandidate(
	candidate RouteCandidate,
	allBoxes []BoxData,
	fromID, toID string,
	candidates *[]RouteCandidate,
	allCandidates *[]RouteCandidate,
) bool {
	if checkPathCollision(candidate.segments, allBoxes, fromID, toID) {
		candidate.rejected = true
		candidate.rejectReason = "collision_detected"
	}

	// Add to allCandidates for debugging
	*allCandidates = append(*allCandidates, candidate)

	// If valid, add to candidates for scoring
	if !candidate.rejected {
		*candidates = append(*candidates, candidate)
		return true
	}
	return false
}

// RouteArrow calculates arrow routing between two boxes
// This function generates ALL legal arrow routes from all strategies,
// then chooses the best one based on quality scoring
// Returns an error if no valid routing can be found
// Also returns all candidates (both valid and rejected) for debug purposes
func RouteArrow(
	box1, box2 BoxCoords,
	fromGridX, fromGridY, toGridX, toGridY int,
	allBoxes []BoxData,
	fromID, toID string,
	flow string,
) (*RoutingPlan, error) {
	var candidates []RouteCandidate
	allCandidates := make([]RouteCandidate, 0)

	// Calculate box widths for scoring
	boxWidth1 := box1.X2 - box1.X1
	boxWidth2 := box2.X2 - box2.X1

	// Strategy 1: Same column - straight vertical arrow
	if fromGridX == toGridX && fromGridY != toGridY {
		fromCenterX := (box1.X1 + box1.X2) / 2
		toCenterX := (box2.X1 + box2.X2) / 2

		var sx, sy, ex, ey int
		sx = fromCenterX
		ex = toCenterX

		if toGridY > fromGridY {
			// Going down
			sy = box1.Y2
			ey = box2.Y1 - STROKE_ADJUSTMENT
		} else {
			// Going up
			sy = box1.Y1
			ey = box2.Y2 + STROKE_ADJUSTMENT
		}

		candidate := RouteCandidate{
			startX: sx, startY: sy, endX: ex, endY: ey,
			strategy: "straight_vertical", verticalFirst: true,
			boxWidth1: boxWidth1, boxWidth2: boxWidth2, flow: flow,
			segments: []int{sx, sy, ex, ey},
		}
		validateAndAddCandidate(candidate, allBoxes, fromID, toID, &candidates, &allCandidates)
	}

	// Strategy 1b: Same column - horizontal-first 2-segment routing (horizontal then vertical)
	// This is useful for narrow-to-wide transitions where exiting horizontally is preferred
	if fromGridX == toGridX && fromGridY != toGridY {
		fromCenterY := (box1.Y1 + box1.Y2) / 2
		toCenterX := (box2.X1 + box2.X2) / 2

		var sx, sy, ex, ey int

		// Exit horizontally from source
		if toGridX >= fromGridX {
			sx = box1.X2 // Exit from right (same column, so exit right)
		} else {
			sx = box1.X1 // Exit from left
		}
		sy = fromCenterY

		// Enter vertically into destination
		if toGridY > fromGridY {
			ey = box2.Y1 - STROKE_ADJUSTMENT // Enter from top
		} else {
			ey = box2.Y2 + STROKE_ADJUSTMENT // Enter from bottom
		}
		ex = toCenterX

		candidate := RouteCandidate{
			startX: sx, startY: sy, endX: ex, endY: ey,
			strategy: "two_segment_horizontal_first", verticalFirst: false,
			boxWidth1: boxWidth1, boxWidth2: boxWidth2, flow: flow,
			segments: []int{sx, sy, ex, sy, ex, ey},
		}
		validateAndAddCandidate(candidate, allBoxes, fromID, toID, &candidates, &allCandidates)
	}

	// Strategy 2: Vertical-first 2-segment routing (vertical then horizontal)
	if fromGridY != toGridY {
		fromCenterX := (box1.X1 + box1.X2) / 2
		toCenterY := (box2.Y1 + box2.Y2) / 2

		var sx, sy, ex, ey int

		// Exit vertically from source
		if toGridY > fromGridY {
			// Going down
			sx = fromCenterX
			sy = box1.Y2 // Exit from bottom
		} else {
			// Going up
			sx = fromCenterX
			sy = box1.Y1 // Exit from top
		}

		// Enter horizontally into destination
		if toGridX > fromGridX {
			ex = box2.X1 - STROKE_ADJUSTMENT // Enter from left
		} else {
			ex = box2.X2 + STROKE_ADJUSTMENT // Enter from right
		}
		ey = toCenterY

		candidate := RouteCandidate{
			startX: sx, startY: sy, endX: ex, endY: ey,
			strategy: "two_segment_vertical_first", verticalFirst: true,
			boxWidth1: boxWidth1, boxWidth2: boxWidth2, flow: flow,
			segments: []int{sx, sy, sx, ey, ex, ey},
		}
		validateAndAddCandidate(candidate, allBoxes, fromID, toID, &candidates, &allCandidates)
	}

	// Strategy 3: Horizontal-first 3-segment routing (horizontal, vertical, horizontal)
	if fromGridX != toGridX && fromGridY != toGridY {
		fromCenterY := (box1.Y1 + box1.Y2) / 2
		toCenterY := (box2.Y1 + box2.Y2) / 2

		var sx, ex int
		if toGridX > fromGridX {
			// Forward arrow (left to right)
			sx = box1.X2
			ex = box2.X1 - STROKE_ADJUSTMENT
		} else {
			// Backward arrow (right to left)
			sx = box1.X1
			ex = box2.X2 + STROKE_ADJUSTMENT
		}

		midX := (sx + ex) / 2
		candidate := RouteCandidate{
			startX: sx, startY: fromCenterY, endX: ex, endY: toCenterY,
			strategy: "three_segment_horizontal_first", verticalFirst: false,
			boxWidth1: boxWidth1, boxWidth2: boxWidth2, flow: flow,
			segments: []int{sx, fromCenterY, midX, fromCenterY, midX, toCenterY, ex, toCenterY},
		}
		validateAndAddCandidate(candidate, allBoxes, fromID, toID, &candidates, &allCandidates)
	}

	// Strategy 4: Non-overlapping horizontal routing (straight or 3-segment)
	if fromGridX != toGridX {
		fromCenterY := (box1.Y1 + box1.Y2) / 2
		toCenterY := (box2.Y1 + box2.Y2) / 2

		var sx, ex int
		if toGridX > fromGridX {
			// Forward arrow (left to right)
			sx = box1.X2
			ex = box2.X1 - STROKE_ADJUSTMENT
		} else {
			// Backward arrow (right to left)
			sx = box1.X1
			ex = box2.X2 + STROKE_ADJUSTMENT
		}

		// Only generate candidate if there's actual horizontal space between boxes
		// (boxes don't overlap in the direction of travel)
		if (toGridX > fromGridX && sx < ex) || (toGridX < fromGridX && sx > ex) {
			midX := (sx + ex) / 2
			candidate := RouteCandidate{
				startX: sx, startY: fromCenterY, endX: ex, endY: toCenterY,
				strategy: "non_overlapping_horizontal", verticalFirst: false,
				boxWidth1: boxWidth1, boxWidth2: boxWidth2, flow: flow,
				segments: []int{sx, fromCenterY, midX, fromCenterY, midX, toCenterY, ex, toCenterY},
			}
			validateAndAddCandidate(candidate, allBoxes, fromID, toID, &candidates, &allCandidates)
		}
	}

	// Strategy 5: Vertical-first 3-segment routing (vertical, horizontal, vertical)
	// Only generated when flow == "down" for top-down org chart style arrows
	if fromGridX != toGridX && fromGridY != toGridY && flow == "down" {
		fromCenterX := (box1.X1 + box1.X2) / 2
		toCenterX := (box2.X1 + box2.X2) / 2

		var sy, ey int
		if toGridY > fromGridY {
			// Going down: exit bottom, enter top
			sy = box1.Y2
			ey = box2.Y1 - STROKE_ADJUSTMENT
		} else {
			// Going up: exit top, enter bottom
			sy = box1.Y1
			ey = box2.Y2 + STROKE_ADJUSTMENT
		}

		midY := (sy + ey) / 2
		candidate := RouteCandidate{
			startX: fromCenterX, startY: sy, endX: toCenterX, endY: ey,
			strategy: "three_segment_vertical_first", verticalFirst: true,
			boxWidth1: boxWidth1, boxWidth2: boxWidth2, flow: flow,
			segments: []int{fromCenterX, sy, fromCenterX, midY, toCenterX, midY, toCenterX, ey},
		}
		validateAndAddCandidate(candidate, allBoxes, fromID, toID, &candidates, &allCandidates)
	}

	// If no valid candidates, return error
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no valid arrow routing found from %s to %s (all strategies failed or produced illegal arrows)", fromID, toID)
	}

	// Score all candidates and find the best one
	bestIdx := 0
	candidates[0].score = scoreRoute(candidates[0])
	bestScore := candidates[0].score

	for i := 1; i < len(candidates); i++ {
		candidates[i].score = scoreRoute(candidates[i])
		if candidates[i].score > bestScore {
			bestScore = candidates[i].score
			bestIdx = i
		}
	}

	// Update allCandidates with scores for valid candidates
	for i := range allCandidates {
		if !allCandidates[i].rejected {
			allCandidates[i].score = scoreRoute(allCandidates[i])
		}
	}

	// Return the best routing plan
	best := candidates[bestIdx]
	return &RoutingPlan{
		StartX:        best.startX,
		StartY:        best.startY,
		EndX:          best.endX,
		EndY:          best.endY,
		Strategy:      best.strategy,
		VerticalFirst: best.verticalFirst,
		NumSegments:   len(best.segments)/2 - 1,
		AllCandidates: allCandidates,
	}, nil
}
