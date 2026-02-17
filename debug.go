package main

import (
	"encoding/json"
	"os"
)

// DebugOutput represents the complete debug information for a diagram
type DebugOutput struct {
	Diagram DiagramInfo  `json:"diagram"`
	Boxes   []BoxDebug   `json:"boxes"`
	Arrows  []ArrowDebug `json:"arrows"`
	Groups  []GroupDebug `json:"groups,omitempty"`
}

// DiagramInfo contains overall diagram dimensions
type DiagramInfo struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// BoxDebug contains debug information for a single box
type BoxDebug struct {
	ID     string `json:"id"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	GridX  int    `json:"gridX"`
	GridY  int    `json:"gridY"`
	Label  string `json:"label"`
	Color  string `json:"color"`
}

// GroupDebug contains debug information for a single group
type GroupDebug struct {
	Label  string   `json:"label"`
	X      int      `json:"x"`
	Y      int      `json:"y"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
	BoxIDs []string `json:"boxIds"`
}

// ArrowDebug contains debug information for a single arrow
type ArrowDebug struct {
	FromBox              string           `json:"fromBox"`
	ToBox                string           `json:"toBox"`
	StartX               int              `json:"startX"`
	StartY               int              `json:"startY"`
	EndX                 int              `json:"endX"`
	EndY                 int              `json:"endY"`
	RoutingStrategy      string           `json:"routingStrategy"`
	ArrowType            string           `json:"arrowType"`
	ArrowheadOrientation string           `json:"arrowheadOrientation"`
	VerticalFirst        bool             `json:"verticalFirst"`
	Candidates           []CandidateDebug `json:"candidates,omitempty"`
}

// CandidateDebug contains debug information for a candidate routing strategy
type CandidateDebug struct {
	Strategy      string `json:"strategy"`
	Score         int    `json:"score"`
	StartX        int    `json:"startX"`
	StartY        int    `json:"startY"`
	EndX          int    `json:"endX"`
	EndY          int    `json:"endY"`
	VerticalFirst bool   `json:"verticalFirst"`
	Selected      bool   `json:"selected"`
	Rejected      bool   `json:"rejected,omitempty"`
	RejectReason  string `json:"rejectReason,omitempty"`
}

// GenerateDebugOutput creates a DebugOutput from a diagram and box data
func GenerateDebugOutput(diagram *Diagram, boxData map[string]BoxData) DebugOutput {
	output := DebugOutput{
		Diagram: DiagramInfo{
			Width:  diagram.Width,
			Height: diagram.Height,
		},
		Boxes:  make([]BoxDebug, 0, len(diagram.Boxes)),
		Arrows: make([]ArrowDebug, 0, len(diagram.Arrows)),
	}

	// Collect box debug information
	for _, box := range diagram.Boxes {
		// Find corresponding BoxData for grid coordinates
		var gridX, gridY int
		for id, data := range boxData {
			// Match by position (best effort - boxes don't store their ID)
			if data.PixelX == box.X && data.PixelY == box.Y {
				gridX = data.GridX
				gridY = data.GridY
				_ = id // boxID
				break
			}
		}

		output.Boxes = append(output.Boxes, BoxDebug{
			ID:     "", // Will be filled from arrow references
			X:      box.X,
			Y:      box.Y,
			Width:  box.Width,
			Height: box.Height,
			GridX:  gridX,
			GridY:  gridY,
			Label:  box.Text,
			Color:  box.Color,
		})
	}

	// Collect arrow debug information
	for _, arrow := range diagram.Arrows {
		arrowType := classifyArrowType(arrow.FromX, arrow.FromY, arrow.ToX, arrow.ToY, arrow.VerticalFirst, arrow.RoutingStrategy)
		orientation := calculateArrowheadOrientation(arrow.FromX, arrow.FromY, arrow.ToX, arrow.ToY, arrow.VerticalFirst)

		// Convert RouteCandidate to CandidateDebug
		candidatesDebug := make([]CandidateDebug, 0, len(arrow.Candidates))
		for _, candidate := range arrow.Candidates {
			// Determine if this candidate was selected (matches the arrow's final route)
			selected := candidate.startX == arrow.FromX &&
				candidate.startY == arrow.FromY &&
				candidate.endX == arrow.ToX &&
				candidate.endY == arrow.ToY &&
				candidate.strategy == arrow.RoutingStrategy

			candidatesDebug = append(candidatesDebug, CandidateDebug{
				Strategy:      candidate.strategy,
				Score:         candidate.score,
				StartX:        candidate.startX,
				StartY:        candidate.startY,
				EndX:          candidate.endX,
				EndY:          candidate.endY,
				VerticalFirst: candidate.verticalFirst,
				Selected:      selected,
				Rejected:      candidate.rejected,
				RejectReason:  candidate.rejectReason,
			})
		}

		output.Arrows = append(output.Arrows, ArrowDebug{
			FromBox:              arrow.FromBoxID,
			ToBox:                arrow.ToBoxID,
			StartX:               arrow.FromX,
			StartY:               arrow.FromY,
			EndX:                 arrow.ToX,
			EndY:                 arrow.ToY,
			RoutingStrategy:      arrow.RoutingStrategy,
			ArrowType:            arrowType,
			ArrowheadOrientation: orientation,
			VerticalFirst:        arrow.VerticalFirst,
			Candidates:           candidatesDebug,
		})
	}

	// Fill in box IDs from arrow references
	boxIDs := make(map[string]bool)
	for _, arrow := range output.Arrows {
		if arrow.FromBox != "" {
			boxIDs[arrow.FromBox] = true
		}
		if arrow.ToBox != "" {
			boxIDs[arrow.ToBox] = true
		}
	}

	// Update box IDs (match by grid position from boxData)
	for i := range output.Boxes {
		for id, data := range boxData {
			if data.PixelX == output.Boxes[i].X && data.PixelY == output.Boxes[i].Y {
				output.Boxes[i].ID = id
				break
			}
		}
	}

	// Collect group debug information
	for _, group := range diagram.Groups {
		output.Groups = append(output.Groups, GroupDebug{
			Label:  group.Label,
			X:      group.X,
			Y:      group.Y,
			Width:  group.Width,
			Height: group.Height,
			BoxIDs: group.BoxIDs,
		})
	}

	return output
}

// classifyArrowType determines the arrow rendering type based on coordinates and strategy
func classifyArrowType(fromX, fromY, toX, toY int, verticalFirst bool, strategy string) string {
	if fromX == toX {
		return "straight_vertical"
	}
	if fromY == toY {
		return "straight_horizontal"
	}
	// Two-segment arrows (one bend) have strategies: two_segment_vertical_first, two_segment_horizontal_first
	// Three-segment arrows (two bends) have strategies: three_segment_horizontal_first, non_overlapping_horizontal
	if strategy == "two_segment_vertical_first" || strategy == "two_segment_horizontal_first" {
		return "one_bent"
	}
	return "two_bent"
}

// calculateArrowheadOrientation determines which direction the arrowhead points
func calculateArrowheadOrientation(fromX, fromY, toX, toY int, verticalFirst bool) string {
	// For straight arrows, direction is obvious
	if fromX == toX {
		if toY > fromY {
			return "down"
		}
		return "up"
	}
	if fromY == toY {
		if toX > fromX {
			return "right"
		}
		return "left"
	}

	// For bent arrows, orientation depends on final segment
	if verticalFirst {
		// Final segment is horizontal
		if toX > fromX {
			return "right"
		}
		return "left"
	} else {
		// For two-bent arrows (horizontal-first), need to determine final segment
		// The final segment in twoBentArrow is horizontal: (midX,toY) -> (toX,toY)
		if toX > fromX {
			return "right"
		}
		return "left"
	}
}

// WriteDebugJSON writes debug output to a JSON file
func WriteDebugJSON(filename string, output DebugOutput) error {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}
