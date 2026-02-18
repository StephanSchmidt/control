package main

import (
	"fmt"
	"strings"
)

// Layout dimension constants
const (
	CHAR_WIDTH_PIXELS  = 16.0 // Approximate width per character for text wrapping
	MIN_CHARS_PER_LINE = 1    // Minimum characters per line in wrapped text
	MAX_TEXT_LINES     = 3    // Maximum lines for wrapped text
)

// Coordinate conversion helpers

// gridToPixelX converts grid X coordinate to pixel X coordinate
func gridToPixelX(gridX int, dims Dimensions, config DiagramConfig) int {
	return dims.LeftMargin + int(float64(gridX-1)*dims.CellUnits*float64(config.GridUnit))
}

// gridToPixelY converts grid Y coordinate to pixel Y coordinate
func gridToPixelY(gridY int, dims Dimensions, config DiagramConfig) int {
	return dims.TopMargin + int(float64(gridY-1)*dims.VerticalCellUnits*float64(config.GridUnit))
}

// calculateBoxWidth calculates the pixel width of a box based on its grid width
func calculateBoxWidth(gridWidth float64, dims Dimensions, config DiagramConfig) int {
	return int((gridWidth*dims.CellUnits - config.GapUnits*config.Stretch) * float64(config.GridUnit))
}

// calculateBoxHeight calculates the pixel height of a box based on its grid height
func calculateBoxHeight(gridHeight int, config DiagramConfig) int {
	return gridHeight * config.GridUnit
}

// calculateTouchExtension calculates the extension amount for TouchLeft boxes
func calculateTouchExtension(config DiagramConfig) int {
	return int((config.GapUnits / 2.0) * float64(config.GridUnit))
}

// groupPadding is the padding (in pixels) around boxes within a group rectangle
const groupPadding = 15

// Layout converts a DiagramSpec into a concrete Diagram with pixel coordinates
func Layout(spec *DiagramSpec, config DiagramConfig, legend []LegendEntry, groups []GroupDef, arrowFlow string) (*Diagram, map[string]BoxData) {
	// Find maximum grid positions
	maxGridX := 0
	maxGridY := 0
	for _, box := range spec.Boxes {
		if box.GridX > maxGridX {
			maxGridX = box.GridX
		}
		if box.GridY > maxGridY {
			maxGridY = box.GridY
		}
	}

	// Calculate dimensions
	dims := CalculateDimensions(maxGridX, maxGridY, config)

	// Extend width for legend area if needed
	legendWidth := EstimateLegendWidth(legend)
	dims.Width += legendWidth

	// Create diagram
	diagram := NewDiagram(dims.Width, dims.Height)

	// Map box IDs to their data for arrow routing
	boxData := make(map[string]BoxData)

	// Create boxes
	for _, boxSpec := range spec.Boxes {
		// Convert grid coordinates to pixel coordinates
		pixelX := gridToPixelX(boxSpec.GridX, dims, config)
		pixelY := gridToPixelY(boxSpec.GridY, dims, config)

		// Calculate per-box dimensions based on GridWidth and GridHeight
		boxWidth := calculateBoxWidth(boxSpec.GridWidth, dims, config)
		boxHeight := calculateBoxHeight(boxSpec.GridHeight, config)

		// Handle touch-left connector: extend both boxes into the gap
		if boxSpec.TouchLeft {
			touchExtension := calculateTouchExtension(config)
			boxWidth += touchExtension // Extend current box to the left
			pixelX -= touchExtension   // Shift position left by extension amount

			// Extend previous box to the right
			if len(diagram.Boxes) > 0 {
				prevIdx := len(diagram.Boxes) - 1
				diagram.Boxes[prevIdx].Width += touchExtension
			}
		}

		color := boxSpec.Color
		if color == "" {
			// Apply gradient: bottom row uses #FFCE33, top rows lighten to #FFFFE0
			color = calculateGradientColor(boxSpec.GridY, maxGridY)
		}

		borderColor := boxSpec.BorderColor
		borderWidth := boxSpec.BorderWidth
		fontSize := boxSpec.FontSize
		textColor := boxSpec.TextColor

		// Wrap text to fit in box
		maxCharsPerLine := int(float64(boxWidth) / CHAR_WIDTH_PIXELS)
		if maxCharsPerLine < MIN_CHARS_PER_LINE {
			maxCharsPerLine = MIN_CHARS_PER_LINE
		}
		wrappedLines := WrapText(boxSpec.Label, maxCharsPerLine, MAX_TEXT_LINES)
		wrappedLabel := strings.Join(wrappedLines, "\n")

		diagram.AddBox(pixelX, pixelY, boxWidth, boxHeight, wrappedLabel, color, borderColor, borderWidth, fontSize, textColor)

		// Store box data for arrow routing
		boxData[boxSpec.ID] = BoxData{
			ID:      boxSpec.ID,
			GridX:   boxSpec.GridX,
			GridY:   boxSpec.GridY,
			PixelX:  pixelX,
			PixelY:  pixelY,
			CenterX: pixelX + boxWidth/2,
			CenterY: pixelY + boxHeight/2,
			Width:   boxWidth,
			Height:  boxHeight,
		}
	}

	// Create arrows
	for _, arrowSpec := range spec.Arrows {
		fromBox := boxData[arrowSpec.FromID]
		toBox := boxData[arrowSpec.ToID]

		// Convert boxData map to slice for collision detection
		allBoxes := make([]BoxData, 0, len(boxData))
		for _, box := range boxData {
			allBoxes = append(allBoxes, box)
		}

		// Create box coordinates
		box1 := BoxCoords{
			X1: fromBox.PixelX,
			Y1: fromBox.PixelY,
			X2: fromBox.PixelX + fromBox.Width,
			Y2: fromBox.PixelY + fromBox.Height,
		}
		box2 := BoxCoords{
			X1: toBox.PixelX,
			Y1: toBox.PixelY,
			X2: toBox.PixelX + toBox.Width,
			Y2: toBox.PixelY + toBox.Height,
		}

		// Resolve per-arrow flow vs global flow
		flow := arrowFlow
		if arrowSpec.Flow != "" {
			flow = arrowSpec.Flow
		}

		plan, err := RouteArrow(
			box1, box2,
			fromBox.GridX, fromBox.GridY,
			toBox.GridX, toBox.GridY,
			allBoxes,
			arrowSpec.FromID, arrowSpec.ToID,
			flow,
		)
		if err != nil {
			fmt.Printf("Error routing arrow from %s to %s: %v\n", arrowSpec.FromID, arrowSpec.ToID, err)
			continue
		}

		diagram.AddArrow(plan.StartX, plan.StartY, plan.EndX, plan.EndY, plan.VerticalFirst, plan.NumSegments, arrowSpec.FromID, arrowSpec.ToID, plan.Strategy, plan.AllCandidates)
	}

	// Resolve groups to pixel coordinates
	for _, g := range groups {
		if len(g.BoxIDs) == 0 {
			continue
		}
		// Compute bounding box of all member boxes
		first := true
		var minX, minY, maxX, maxY int
		for _, boxID := range g.BoxIDs {
			bd, ok := boxData[boxID]
			if !ok {
				continue
			}
			bx1 := bd.PixelX
			by1 := bd.PixelY
			bx2 := bd.PixelX + bd.Width
			by2 := bd.PixelY + bd.Height
			if first {
				minX, minY, maxX, maxY = bx1, by1, bx2, by2
				first = false
			} else {
				if bx1 < minX {
					minX = bx1
				}
				if by1 < minY {
					minY = by1
				}
				if bx2 > maxX {
					maxX = bx2
				}
				if by2 > maxY {
					maxY = by2
				}
			}
		}
		if first {
			continue // no valid boxes found
		}
		diagram.Groups = append(diagram.Groups, Group{
			X:      minX - groupPadding,
			Y:      minY - groupPadding - 20, // extra space for label
			Width:  (maxX - minX) + 2*groupPadding,
			Height: (maxY - minY) + 2*groupPadding + 20, // extra space for label
			Label:  g.Label,
			BoxIDs: g.BoxIDs,
		})
	}

	return diagram, boxData
}

// parseHexColor converts hex color string to RGB values (0-255 range)
func parseHexColor(hex string) (r, g, b int) {
	// Remove # if present
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	// Parse hex values
	if len(hex) == 6 {
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return 0, 0, 0 // Return black on invalid hex
		}
	}
	return
}

// rgbToHex converts RGB values (0-255) to hex color string
func rgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// interpolateColor blends two hex colors based on factor (0.0 to 1.0)
// factor=0.0 returns color1, factor=1.0 returns color2
func interpolateColor(color1, color2 string, factor float64) string {
	r1, g1, b1 := parseHexColor(color1)
	r2, g2, b2 := parseHexColor(color2)

	r := int(float64(r1) + float64(r2-r1)*factor)
	g := int(float64(g1) + float64(g2-g1)*factor)
	b := int(float64(b1) + float64(b2-b1)*factor)

	return rgbToHex(r, g, b)
}

// calculateGradientColor returns a yellow gradient color based on row position
// Bottom row (highest GridY) gets #FFCE33, top rows get progressively lighter
func calculateGradientColor(gridY, maxGridY int) string {
	if maxGridY <= 1 {
		return "#FFCE33" // Single row, use default yellow
	}

	// Calculate factor: 0.0 for bottom row, increasing towards 1.0 for top
	// Apply 0.5 multiplier for subtle gradient (use only 50% of color range)
	factor := float64(maxGridY-gridY) / float64(maxGridY-1) * 0.5

	// Interpolate between dark yellow (#FFCE33) and very light yellow (#FFFEF0)
	return interpolateColor("#FFCE33", "#FFFEF0", factor)
}
