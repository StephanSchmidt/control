package main

import (
	"fmt"
	"strings"
)

// Box represents a rectangular box in the diagram
type Box struct {
	X, Y          int
	Width, Height int
	Text          string
	Color         string
	BorderColor   string
	BorderWidth   int
	FontSize      int
	TextColor     string
	TextLines     []string
}

// Arrow represents a connection between boxes
type Arrow struct {
	FromX, FromY    int
	ToX, ToY        int
	VerticalFirst   bool             // If true, use 2-segment routing (vertical then horizontal)
	NumSegments     int              // 1=straight, 2=L-shape, 3=Z-shape
	FromBoxID       string           // ID of source box (for debug output)
	ToBoxID         string           // ID of destination box (for debug output)
	RoutingStrategy string           // Name of routing strategy used (for debug output)
	Candidates      []RouteCandidate // All routing candidates considered (for debug output)
}

// Group represents a visual grouping rectangle around boxes
type Group struct {
	X, Y          int
	Width, Height int
	Label         string
	BoxIDs        []string // IDs of boxes in this group (for debug output)
}

// Diagram represents the entire workflow diagram
type Diagram struct {
	Width, Height int
	Boxes         []Box
	Arrows        []Arrow
	Groups        []Group
	YAxisLabel    string
	XAxisLabel    string
	ZoneLabel1    string
	ZoneLabel2    string
	ZoneSplit     int
	Font          *FontData         // Optional custom font
	Legend        []LegendEntry     // Optional legend entries
	CustomColors  map[string]string // Custom color definitions (name -> hex)
}

// DiagramConfig holds diagram-wide rendering settings
type DiagramConfig struct {
	GridUnit         int     // Pixels per grid unit
	BoxWidthUnits    float64 // Box width in grid units
	GapUnits         float64 // Gap in grid units (horizontal)
	VerticalGapUnits float64 // Vertical gap in grid units
	AxisOffset       int     // Offset between axes and content
	DefaultColor     string  // Default box color
	Stretch          float64 // X-axis stretch factor (1.0 = normal)
}

// NewDefaultConfig returns default diagram configuration
func NewDefaultConfig() DiagramConfig {
	return DiagramConfig{
		GridUnit:         100,
		BoxWidthUnits:    2.5,
		GapUnits:         0.5,
		VerticalGapUnits: 0.5,
		AxisOffset:       30,
		DefaultColor:     "#FFCE33",
		Stretch:          1.0,
	}
}

// NewDiagram creates a new diagram with default settings
func NewDiagram(width, height int) *Diagram {
	return &Diagram{
		Width:  width,
		Height: height,
		Boxes:  []Box{},
		Arrows: []Arrow{},
	}
}

// AddBox adds a box to the diagram
func (d *Diagram) AddBox(x, y, w, h int, text, color, borderColor string, borderWidth, fontSize int, textColor string) {
	lines := strings.Split(text, "\n")
	d.Boxes = append(d.Boxes, Box{
		X:           x,
		Y:           y,
		Width:       w,
		Height:      h,
		Text:        text,
		Color:       color,
		BorderColor: borderColor,
		BorderWidth: borderWidth,
		FontSize:    fontSize,
		TextColor:   textColor,
		TextLines:   lines,
	})
}

// AddArrow adds an arrow between two points
func (d *Diagram) AddArrow(fromX, fromY, toX, toY int, verticalFirst bool, numSegments int, fromID, toID, routingStrategy string, candidates []RouteCandidate) {
	d.Arrows = append(d.Arrows, Arrow{
		FromX:           fromX,
		FromY:           fromY,
		ToX:             toX,
		ToY:             toY,
		VerticalFirst:   verticalFirst,
		NumSegments:     numSegments,
		FromBoxID:       fromID,
		ToBoxID:         toID,
		RoutingStrategy: routingStrategy,
		Candidates:      candidates,
	})
}

// SetLabels sets axis and zone labels
func (d *Diagram) SetLabels(yAxis, xAxis, zone1, zone2 string, zoneSplit int) {
	d.YAxisLabel = yAxis
	d.XAxisLabel = xAxis
	d.ZoneLabel1 = zone1
	d.ZoneLabel2 = zone2
	d.ZoneSplit = zoneSplit
}

// GenerateSVG creates the SVG output
func (d *Diagram) GenerateSVG() string {
	var svg strings.Builder

	// SVG header with arrowhead marker definition
	svg.WriteString(svgHeader(d.Width, d.Height, d.Font))

	// Draw Y-axis if y-label is set
	if d.YAxisLabel != "" {
		svg.WriteString(drawLine(60, d.Height-50, 60, 50, false, true))
		attrs := map[string]string{
			"transform":   "rotate(-90 30 30)",
			"text-anchor": "end",
		}
		svg.WriteString(drawText(30, 30, d.YAxisLabel, 18, attrs, d.Font))
	}

	// Draw X-axis if x-label is set
	if d.XAxisLabel != "" {
		svg.WriteString(drawLine(60, d.Height-50, d.Width-20, d.Height-50, false, true))
		attrs := map[string]string{
			"text-anchor": "end",
		}
		svg.WriteString(drawText(d.Width-30, d.Height-20, d.XAxisLabel, 18, attrs, d.Font))
	}

	// Draw zone split line if specified
	if d.ZoneSplit > 0 {
		svg.WriteString(drawLine(60, d.ZoneSplit, d.Width-20, d.ZoneSplit, true, false))

		if d.ZoneLabel1 != "" {
			attrs := map[string]string{"font-style": "italic"}
			svg.WriteString(drawText(70, d.ZoneSplit-20, d.ZoneLabel1, 12, attrs, d.Font))
		}
		if d.ZoneLabel2 != "" {
			attrs := map[string]string{"font-style": "italic"}
			svg.WriteString(drawText(70, d.ZoneSplit+30, d.ZoneLabel2, 12, attrs, d.Font))
		}
	}

	// Draw groups (behind boxes and arrows)
	for _, group := range d.Groups {
		svg.WriteString(drawGroup(group.X, group.Y, group.Width, group.Height, group.Label, d.Font))
	}

	// Draw arrows
	for _, arrow := range d.Arrows {
		switch {
		case arrow.FromX == arrow.ToX || arrow.FromY == arrow.ToY:
			svg.WriteString(straightArrow(arrow.FromX, arrow.FromY, arrow.ToX, arrow.ToY))
		case arrow.NumSegments == 3 && arrow.VerticalFirst:
			svg.WriteString(twoBentArrowVertical(arrow.FromX, arrow.FromY, arrow.ToX, arrow.ToY))
		case arrow.NumSegments == 3:
			svg.WriteString(twoBentArrow(arrow.FromX, arrow.FromY, arrow.ToX, arrow.ToY))
		default:
			svg.WriteString(oneBentArrow(arrow.FromX, arrow.FromY, arrow.ToX, arrow.ToY, arrow.VerticalFirst))
		}
	}

	// Draw boxes
	for _, box := range d.Boxes {
		svg.WriteString(drawBox(box.X, box.Y, box.Width, box.Height, box.Color, box.BorderColor, box.BorderWidth))
		svg.WriteString(drawBoxText(box.X, box.Y, box.Width, box.Height, box.FontSize, box.TextColor, box.TextLines, d.Font))
	}

	// Draw legend if entries exist
	if len(d.Legend) > 0 {
		svg.WriteString(d.renderLegend())
	}

	svg.WriteString(svgFooter())
	return svg.String()
}

// Legend rendering constants
const (
	legendSquareSize = 30 // Size of the colored square
	legendFontSize   = 28 // Font size for legend text
	legendLineHeight = 44 // Vertical spacing between entries
	legendPadding    = 10 // Padding inside the legend area
	legendTextGap    = 12 // Gap between square and text
	legendTopMargin  = 50 // Top margin (same as diagram top margin)
)

// renderLegend renders the legend entries in the top-right corner of the SVG
func (d *Diagram) renderLegend() string {
	var svg strings.Builder

	// Position legend in top-right corner
	startX := d.Width - legendPadding
	startY := legendTopMargin

	for i, entry := range d.Legend {
		y := startY + i*legendLineHeight

		// Resolve style code to color
		styles := parseBoxStyles(entry.Style, d.CustomColors)
		color := styles.BackgroundColor
		if color == "" || color == "none" {
			color = "#FFCE33" // Use default box color
		}

		// Draw colored square (right-aligned)
		squareX := startX - legendSquareSize
		fmt.Fprintf(&svg, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" stroke="#000" stroke-width="1"/>`,
			squareX, y, legendSquareSize, legendSquareSize, color)

		// Draw label text to the left of the square
		textX := squareX - legendTextGap
		textY := y + legendSquareSize/2
		attrs := map[string]string{
			"text-anchor":       "end",
			"dominant-baseline": "middle",
		}
		svg.WriteString(drawText(textX, textY, entry.Label, legendFontSize, attrs, d.Font))
	}

	return svg.String()
}
