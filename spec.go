package main

import (
	"fmt"
	"strconv"
	"strings"
)

// DiagramSpec represents the logical diagram structure (grid-based)
type DiagramSpec struct {
	Boxes  []BoxSpec
	Arrows []ArrowSpec
	Groups []GroupDef
}

// GroupDef represents a visual group that contains boxes
type GroupDef struct {
	Name   string   // Group identifier (e.g., "Team")
	Label  string   // Display label (e.g., "Our Team"); defaults to Name
	BoxIDs []string // IDs of boxes belonging to this group
}

// BoxSpec represents a box in grid coordinates
type BoxSpec struct {
	ID          string
	GridX       int
	GridY       int
	GridWidth   float64 // Width in grid units (default: 2)
	GridHeight  int     // Height in grid units (default: 1)
	Label       string
	Color       string // Optional, uses default if empty
	BorderColor string // Optional border color (default: black)
	BorderWidth int    // Optional border width (default: 2)
	FontSize    int    // Optional font size (default: 24)
	TextColor   string // Optional text color (default: black)
	TouchLeft   bool   // Whether this box touches the previous box ("|" prefix)
	Group       string // Optional group name this box belongs to (e.g., "Team")
}

// ArrowSpec represents logical connection between boxes
type ArrowSpec struct {
	FromID string
	ToID   string
	Flow   string // Optional per-arrow flow hint (e.g., "down")
}

// ParsedCoordinate represents a single parsed coordinate with metadata
type ParsedCoordinate struct {
	IsRelative bool // true if relative (+/- prefix or "0"), false if absolute
	Value      int  // the numeric value (can be negative for relative)
}

// BoxCoordinates represents the complete coordinate pair for a box
type BoxCoordinates struct {
	X         ParsedCoordinate // Parsed X coordinate
	Y         ParsedCoordinate // Parsed Y coordinate
	AutoArrow bool             // Whether this box had the ">" auto-arrow prefix
	GridX     int              // Resolved absolute grid X
	GridY     int              // Resolved absolute grid Y
}

// BoxStyles represents the parsed style attributes for a box
type BoxStyles struct {
	BackgroundColor string
	BorderColor     string
	BorderWidth     int
	FontSize        int
	TextColor       string
}

// parseCoordinate parses a single coordinate value (GridX or GridY)
// Returns: ParsedCoordinate and error
// Examples:
//
//	"5"   -> ParsedCoordinate{IsRelative: false, Value: 5}      // Absolute
//	"+2"  -> ParsedCoordinate{IsRelative: true, Value: 2}       // Relative positive
//	"-1"  -> ParsedCoordinate{IsRelative: true, Value: -1}      // Relative negative
//	"0"   -> ParsedCoordinate{IsRelative: true, Value: 0}       // Shorthand for "+0"
func parseCoordinate(coordStr string) (ParsedCoordinate, error) {
	coordStr = strings.TrimSpace(coordStr)

	// Shorthand: "0" means relative zero "+0"
	// (Absolute 0 is impossible since grid starts at 1)
	if coordStr == "0" {
		return ParsedCoordinate{IsRelative: true, Value: 0}, nil
	}

	// Check for explicit relative prefix
	if strings.HasPrefix(coordStr, "+") {
		// Relative positive
		value, err := strconv.Atoi(coordStr[1:])
		return ParsedCoordinate{IsRelative: true, Value: value}, err
	}

	if strings.HasPrefix(coordStr, "-") {
		// Relative negative
		value, err := strconv.Atoi(coordStr[1:])
		if err != nil {
			return ParsedCoordinate{}, err
		}
		return ParsedCoordinate{IsRelative: true, Value: -value}, nil
	}

	// Absolute coordinate (no prefix)
	value, err := strconv.Atoi(coordStr)
	return ParsedCoordinate{IsRelative: false, Value: value}, err
}

// parseBoxStyles parses a style string (e.g., "rb-g-rt") into BoxStyles
// Returns: BoxStyles with parsed attributes
// Supported styles:
//   - "rb": Red border (3px width)
//   - "g": Gray background
//   - "p": Purple background
//   - "lp": Light purple background
//   - "nbb": No background, no border
//   - "rt": Red text
//   - "2t": Double text size (48px)
func parseBoxStyles(styleStr string, customColors map[string]string) BoxStyles {
	styles := BoxStyles{
		BackgroundColor: "",
		BorderColor:     "",
		BorderWidth:     0,
		FontSize:        0,
		TextColor:       "",
	}

	if styleStr == "" {
		return styles
	}

	styleParts := strings.Split(styleStr, "-")
	for _, style := range styleParts {
		style = strings.TrimSpace(style)
		switch style {
		case "rb":
			styles.BorderColor = "#FF0000" // Red
			styles.BorderWidth = 3         // Bold (3px instead of default 2px)
		case "g":
			styles.BackgroundColor = "#D3D3D3" // Gray
		case "p":
			styles.BackgroundColor = "#ecbae6" // Purple
		case "lp":
			styles.BackgroundColor = "#f5dbf2" // Light purple
		case "nbb":
			styles.BackgroundColor = "none" // No background
			styles.BorderColor = "none"     // No border
			styles.BorderWidth = 0
		case "rt":
			styles.TextColor = "#FF0000" // Red text
		case "2t":
			styles.FontSize = 48 // 200% of default 24
		default:
			// Check custom colors: "green" → background, "greent" → text color
			if customColors != nil {
				if hex, ok := customColors[style]; ok {
					styles.BackgroundColor = hex
				} else if strings.HasSuffix(style, "t") {
					name := style[:len(style)-1]
					if hex, ok := customColors[name]; ok {
						styles.TextColor = hex
					}
				}
			}
		}
	}

	return styles
}

// parseNumberOrFraction parses a string that can be an integer, decimal, or fraction
// Examples:
//   - Integer: "2" → 2.0
//   - Decimal: "1.5" → 1.5
//   - Fraction: "1/2" → 0.5, "3/4" → 0.75
//
// Returns the float64 value and any parsing error
func parseNumberOrFraction(s string) (float64, error) {
	s = strings.TrimSpace(s)

	// Check if it's a fraction (contains "/")
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid fraction '%s': must be 'numerator/denominator'", s)
		}

		// Parse numerator
		numerator, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid fraction '%s': %w", s, err)
		}

		// Parse denominator
		denominator, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid fraction '%s': %w", s, err)
		}

		// Check for division by zero
		if denominator == 0 {
			return 0, fmt.Errorf("invalid fraction '%s': division by zero", s)
		}

		return numerator / denominator, nil
	}

	// Not a fraction, parse as regular float
	return strconv.ParseFloat(s, 64)
}

// LegendEntry represents a single legend item mapping a style code to a description
type LegendEntry struct {
	Style string // Style code (e.g., "p", "g", "lp")
	Label string // Human-readable description
}

// Frontmatter represents metadata parsed from the top of a diagram file
type Frontmatter struct {
	Font      string            // Path to custom font file (WOFF2 format)
	XLabel    string            // X-axis label (default: "Time"); empty string = no axis
	YLabel    string            // Y-axis label (default: "Control"); empty string = no axis
	Legend    []LegendEntry     // Legend entries mapping style codes to descriptions
	Colors    map[string]string // Custom color definitions (name -> hex)
	ArrowFlow string            // Global arrow flow direction (e.g., "down" for top-down routing)
}

// ParseFrontmatter extracts frontmatter key:value pairs from the top of diagram text.
// Returns the parsed frontmatter and the remaining text with frontmatter lines stripped.
// Supports two formats:
//  1. Delimited: lines between opening and closing "---" markers
//  2. Undelimited: key:value lines at the top (stops at first unrecognized line)
//
// Recognized keys: font
// Comments (#) and blank lines are allowed within frontmatter.
func ParseFrontmatter(text string) (Frontmatter, string) {
	var fm Frontmatter
	lines := strings.Split(text, "\n")
	consumedLines := 0

	// Skip leading blank lines and comments to find potential "---" opener
	for consumedLines < len(lines) {
		trimmed := strings.TrimSpace(lines[consumedLines])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			consumedLines++
			continue
		}
		break
	}

	// Check for delimited frontmatter (--- ... ---)
	if consumedLines < len(lines) && strings.TrimSpace(lines[consumedLines]) == "---" {
		consumedLines++ // consume opening ---
		for consumedLines < len(lines) {
			trimmed := strings.TrimSpace(lines[consumedLines])

			// Closing delimiter ends frontmatter
			if trimmed == "---" {
				consumedLines++ // consume closing ---
				remaining := strings.Join(lines[consumedLines:], "\n")
				return fm, remaining
			}

			consumedLines++
			parseFrontmatterKey(&fm, trimmed)
		}
		// Reached end of input without closing ---; treat entire input as frontmatter
		return fm, ""
	}

	// Undelimited frontmatter: reset and scan from the top
	consumedLines = 0
	for consumedLines < len(lines) {
		trimmed := strings.TrimSpace(lines[consumedLines])

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			consumedLines++
			continue
		}

		if parseFrontmatterKey(&fm, trimmed) {
			consumedLines++
			continue
		}

		// First unrecognized line ends frontmatter
		break
	}

	remaining := strings.Join(lines[consumedLines:], "\n")
	return fm, remaining
}

// parseFrontmatterKey parses a single frontmatter line into the Frontmatter struct.
// Returns true if the line was a recognized key.
func parseFrontmatterKey(fm *Frontmatter, trimmed string) bool {
	// Skip blank lines and comments
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return true
	}

	if strings.HasPrefix(trimmed, "font:") {
		fm.Font = strings.TrimSpace(strings.TrimPrefix(trimmed, "font:"))
		return true
	}

	if strings.HasPrefix(trimmed, "x-label:") {
		fm.XLabel = strings.TrimSpace(strings.TrimPrefix(trimmed, "x-label:"))
		return true
	}

	if strings.HasPrefix(trimmed, "y-label:") {
		fm.YLabel = strings.TrimSpace(strings.TrimPrefix(trimmed, "y-label:"))
		return true
	}

	if strings.HasPrefix(trimmed, "legend:") {
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "legend:"))
		parts := strings.SplitN(value, "=", 2)
		if len(parts) == 2 {
			entry := LegendEntry{
				Style: strings.TrimSpace(parts[0]),
				Label: strings.TrimSpace(parts[1]),
			}
			fm.Legend = append(fm.Legend, entry)
		}
		return true
	}

	if strings.HasPrefix(trimmed, "arrow-flow:") {
		fm.ArrowFlow = strings.TrimSpace(strings.TrimPrefix(trimmed, "arrow-flow:"))
		return true
	}

	if strings.HasPrefix(trimmed, "color:") {
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "color:"))
		parts := strings.SplitN(value, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			hex := strings.TrimSpace(parts[1])
			if fm.Colors == nil {
				fm.Colors = make(map[string]string)
			}
			fm.Colors[name] = hex
		}
		return true
	}

	return false
}

// ParseDiagramSpec parses the text format into a DiagramSpec
func ParseDiagramSpec(text string, customColors map[string]string) (*DiagramSpec, error) {
	spec := &DiagramSpec{
		Boxes:  []BoxSpec{},
		Arrows: []ArrowSpec{},
	}

	inArrowSection := false
	lines := strings.Split(text, "\n")
	var previousBoxID string             // Track previous box for auto-arrows
	var previousGridX int                // Track previous box GridX for relative coordinates
	var previousGridY int                // Track previous box GridY for relative coordinates
	internalIDCounter := 0               // Counter for generating internal IDs for unlabeled boxes
	groupDefs := make(map[string]string) // Group name -> label (from @Group: Label lines)
	boxGroups := make(map[string]string) // Box ID -> group name (from @Group suffix on box lines)

	// Container state (purely organizational, no visual rendering)
	var inContainer bool
	var containerID string
	var containerBaseX, containerBaseY int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Detect container closing "]"
		if line == "]" {
			if !inContainer {
				return nil, fmt.Errorf("unexpected ']' outside container")
			}
			// Containers are purely organizational (coordinate grouping).
			// No GroupDef is created — use @Group for visual borders.
			inContainer = false
			containerID = ""
			continue
		}

		// Detect container header: line ends with "["
		if strings.HasSuffix(line, "[") {
			if inContainer {
				return nil, fmt.Errorf("nested containers not supported")
			}
			if inArrowSection {
				return nil, fmt.Errorf("container not allowed in arrow section")
			}
			// Strip "[" and trim
			headerStr := strings.TrimSpace(strings.TrimSuffix(line, "["))

			// Parse: "ID: x,y [" or "ID: x,y: Label ["
			headerParts := strings.SplitN(headerStr, ":", 3)
			if len(headerParts) < 2 {
				return nil, fmt.Errorf("invalid container definition: '%s'", line)
			}
			containerID = strings.TrimSpace(headerParts[0])
			coordsStr := strings.TrimSpace(headerParts[1])

			coords := strings.Split(coordsStr, ",")
			if len(coords) != 2 {
				return nil, fmt.Errorf("invalid container coordinates: '%s'", line)
			}

			baseX, err := strconv.Atoi(strings.TrimSpace(coords[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid container X coordinate in line: '%s'", line)
			}
			baseY, err := strconv.Atoi(strings.TrimSpace(coords[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid container Y coordinate in line: '%s'", line)
			}

			containerBaseX = baseX
			containerBaseY = baseY

			// Set previous position to container base so relative coords inside work
			previousGridX = containerBaseX
			previousGridY = containerBaseY

			inContainer = true
			continue
		}

		if line == "---" {
			if inContainer {
				return nil, fmt.Errorf("section separator not allowed inside container")
			}
			inArrowSection = true
			continue
		}

		// Arrow lines (containing "->") are recognized in any section
		if strings.Contains(line, "->") {
			parts := strings.Split(line, "->")
			if len(parts) == 2 {
				from := strings.TrimSpace(parts[0])
				toAndFlow := strings.TrimSpace(parts[1])

				// Parse optional "| flow" suffix (e.g., "HE | down")
				var to, arrowFlow string
				if pipeIdx := strings.Index(toAndFlow, "|"); pipeIdx >= 0 {
					to = strings.TrimSpace(toAndFlow[:pipeIdx])
					arrowFlow = strings.TrimSpace(toAndFlow[pipeIdx+1:])
				} else {
					to = toAndFlow
				}

				if from != "" && to != "" {
					// Manual arrows cannot reference internal IDs
					if strings.HasPrefix(from, "_box_") {
						return nil, fmt.Errorf("arrow '%s -> %s' references box without explicit label (internal ID: %s)", from, to, from)
					}
					if strings.HasPrefix(to, "_box_") {
						return nil, fmt.Errorf("arrow '%s -> %s' references box without explicit label (internal ID: %s)", from, to, to)
					}
					spec.Arrows = append(spec.Arrows, ArrowSpec{
						FromID: from,
						ToID:   to,
						Flow:   arrowFlow,
					})
					continue
				}
			}
		}

		if !inArrowSection {
			// Check for group definition line: @GroupName: Label
			if strings.HasPrefix(line, "@") {
				groupLine := line[1:] // Strip "@"
				groupParts := strings.SplitN(groupLine, ":", 2)
				groupName := strings.TrimSpace(groupParts[0])
				groupLabel := groupName // Default label is the group name
				if len(groupParts) == 2 {
					groupLabel = strings.TrimSpace(groupParts[1])
				}
				groupDefs[groupName] = groupLabel
				continue
			}

			// Parse box: "dev: 1,2: Sprint Planning" or ">3,2: Daily Standup" (no ID)
			parts := strings.SplitN(line, ":", 3)

			var id string
			var coordsAndLabelParts []string

			if len(parts) == 3 {
				// Format: "ID: coords: label"
				id = strings.TrimSpace(parts[0])
				// Validate ID: alphanumeric + underscore + hyphen only
				if id == "" {
					return nil, fmt.Errorf("invalid box definition: empty ID in line '%s'", line)
				}
				for _, ch := range id {
					if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') &&
						(ch < '0' || ch > '9') && ch != '_' && ch != '-' {
						return nil, fmt.Errorf("invalid ID '%s': must contain only alphanumeric characters, underscore, or hyphen", id)
					}
				}
				coordsAndLabelParts = parts[1:3]
			} else if len(parts) == 2 {
				// Format: "coords: label" (no ID)
				id = "" // Will be assigned internal ID if needed
				coordsAndLabelParts = parts
			} else {
				return nil, fmt.Errorf("invalid box definition: '%s'", line)
			}

			// Generate internal ID for boxes without explicit IDs
			if id == "" {
				id = fmt.Sprintf("_box_%d", internalIDCounter)
				internalIDCounter++
			}

			// Check for auto-arrow prefix ">" or touch-left prefix "|"
			coordsStr := strings.TrimSpace(coordsAndLabelParts[0])
			autoArrow := strings.HasPrefix(coordsStr, ">")
			touchLeft := strings.HasPrefix(coordsStr, "|")

			if autoArrow {
				// Check if this is the first box
				if previousBoxID == "" {
					return nil, fmt.Errorf("first box (label '%s') cannot have auto-arrow prefix '>'", id)
				}
				// Strip the ">" prefix
				coordsStr = strings.TrimPrefix(coordsStr, ">")
			} else if touchLeft {
				// Check if this is the first box
				if previousBoxID == "" {
					idStr := id
					if idStr == "" {
						idStr = "(unlabeled)"
					}
					return nil, fmt.Errorf("first box (label '%s') cannot have touch-left prefix '|'", idStr)
				}
				// Strip the "|" prefix
				coordsStr = strings.TrimPrefix(coordsStr, "|")
			}

			coords := strings.Split(coordsStr, ",")
			if len(coords) != 2 && len(coords) != 3 && len(coords) != 4 {
				return nil, fmt.Errorf("invalid coordinate definition: '%s'", line)
			}

			// Parse GridX coordinate (may be relative or absolute)
			coordX, err := parseCoordinate(coords[0])
			if err != nil {
				return nil, fmt.Errorf("invalid X coordinate in line: '%s'", line)
			}

			// Parse GridY coordinate (may be relative or absolute)
			coordY, err := parseCoordinate(coords[1])
			if err != nil {
				return nil, fmt.Errorf("invalid Y coordinate in line: '%s'", line)
			}

			// Parse GridWidth and GridHeight (absolute only)
			var gridWidth float64
			var gridHeight int
			if len(coords) == 4 {
				gridWidth, err = parseNumberOrFraction(coords[2])
				if err != nil {
					return nil, fmt.Errorf("invalid width in line: '%s'", line)
				}
				gridHeight, err = strconv.Atoi(strings.TrimSpace(coords[3]))
				if err != nil {
					return nil, fmt.Errorf("invalid height in line: '%s'", line)
				}
				// Validate dimensions
				if gridWidth < 0.2 {
					idStr := id
					if idStr == "" {
						idStr = "(unlabeled)"
					}
					return nil, fmt.Errorf("box '%s': GridWidth must be >= 0.2, got %.1f", idStr, gridWidth)
				}
				if gridHeight < 1 {
					idStr := id
					if idStr == "" {
						idStr = "(unlabeled)"
					}
					return nil, fmt.Errorf("box '%s': GridHeight must be >= 1, got %d", idStr, gridHeight)
				}
			} else if len(coords) == 3 {
				// Custom width, default height
				gridWidth, err = parseNumberOrFraction(coords[2])
				if err != nil {
					return nil, fmt.Errorf("invalid width in line: '%s'", line)
				}
				gridHeight = 1 // Default height
				// Validate width
				if gridWidth < 0.2 {
					idStr := id
					if idStr == "" {
						idStr = "(unlabeled)"
					}
					return nil, fmt.Errorf("box '%s': GridWidth must be >= 0.2, got %.1f", idStr, gridWidth)
				}
			} else {
				// Use defaults (len == 2)
				gridWidth = 2.0
				gridHeight = 1
			}

			// Check if first box tries to use relative coordinates
			// Inside a container, previousGridX/Y are set to container base, so relative coords are OK
			if previousBoxID == "" && !inContainer && (coordX.IsRelative || coordY.IsRelative) {
				idStr := id
				if idStr == "" {
					idStr = "(unlabeled)"
				}
				return nil, fmt.Errorf("first box (label '%s') cannot use relative coordinates", idStr)
			}

			// Validate touch-left requirements
			if touchLeft {
				idStr := id
				if idStr == "" {
					idStr = "(unlabeled)"
				}
				// Y coordinate must be 0 (relative, same row)
				if !coordY.IsRelative || coordY.Value != 0 {
					return nil, fmt.Errorf("box '%s': touch-left prefix '|' requires Y coordinate to be 0 (same row as previous box)", idStr)
				}
				// X coordinate must be relative with "+" prefix (positive relative)
				if !coordX.IsRelative || coordX.Value <= 0 {
					return nil, fmt.Errorf("box '%s': touch-left prefix '|' requires X coordinate to be relative with '+' prefix (e.g. '+2'), got relative=%v value=%d", idStr, coordX.IsRelative, coordX.Value)
				}
			}

			// Resolve coordinates
			var gridX, gridY int
			if coordX.IsRelative {
				gridX = previousGridX + coordX.Value
			} else if inContainer {
				gridX = containerBaseX + coordX.Value
			} else {
				gridX = coordX.Value
			}

			if coordY.IsRelative {
				gridY = previousGridY + coordY.Value
			} else if inContainer {
				gridY = containerBaseY + coordY.Value
			} else {
				gridY = coordY.Value
			}

			// Validate that resulting coordinates are positive
			if gridX < 1 {
				idStr := id
				if idStr == "" {
					idStr = "(unlabeled)"
				}
				return nil, fmt.Errorf("box '%s': relative GridX coordinate resulted in invalid value %d (must be >= 1)", idStr, gridX)
			}
			if gridY < 1 {
				idStr := id
				if idStr == "" {
					idStr = "(unlabeled)"
				}
				return nil, fmt.Errorf("box '%s': relative GridY coordinate resulted in invalid value %d (must be >= 1)", idStr, gridY)
			}
			// Parse label and optional style attributes
			labelAndStyle := strings.TrimSpace(coordsAndLabelParts[1])

			// Extract @GroupName suffix (e.g., "Stefanie, p @Team" -> group="Team")
			var groupName string
			if atIdx := strings.LastIndex(labelAndStyle, " @"); atIdx >= 0 {
				groupName = strings.TrimSpace(labelAndStyle[atIdx+2:])
				labelAndStyle = strings.TrimSpace(labelAndStyle[:atIdx])
			}

			labelParts := strings.SplitN(labelAndStyle, ",", 2)
			if len(labelParts) == 0 {
				return nil, fmt.Errorf("invalid label format: %s", labelAndStyle)
			}
			label := strings.TrimSpace(labelParts[0])

			// Parse optional styles (e.g., "rb-g" -> red border + gray background)
			var styleStr string
			if len(labelParts) == 2 {
				styleStr = strings.TrimSpace(labelParts[1])
			}
			parsedStyles := parseBoxStyles(styleStr, customColors)

			backgroundColor := parsedStyles.BackgroundColor
			borderColor := parsedStyles.BorderColor
			borderWidth := parsedStyles.BorderWidth
			fontSize := parsedStyles.FontSize
			textColor := parsedStyles.TextColor

			spec.Boxes = append(spec.Boxes, BoxSpec{
				ID:          id,
				GridX:       gridX,
				GridY:       gridY,
				GridWidth:   gridWidth,
				GridHeight:  gridHeight,
				Label:       label,
				Color:       backgroundColor,
				BorderColor: borderColor,
				BorderWidth: borderWidth,
				FontSize:    fontSize,
				TextColor:   textColor,
				TouchLeft:   touchLeft,
				Group:       groupName,
			})

			// Track box-to-group mapping
			if groupName != "" {
				boxGroups[id] = groupName
			}

			// Create auto-arrow if prefix was present
			if autoArrow {
				spec.Arrows = append(spec.Arrows, ArrowSpec{
					FromID: previousBoxID,
					ToID:   id,
				})
			}

			// Update previous box tracking
			previousBoxID = id
			previousGridX = gridX
			previousGridY = gridY
		} else {
			// In arrow section, non-arrow lines are invalid
			return nil, fmt.Errorf("invalid arrow definition: '%s'", line)
		}
	}

	// Check for unclosed container
	if inContainer {
		return nil, fmt.Errorf("unclosed container '%s'", containerID)
	}

	// Validate that all arrows reference existing boxes
	validBoxIDs := make(map[string]bool)
	for _, box := range spec.Boxes {
		// All boxes now have IDs (either explicit or internal)
		validBoxIDs[box.ID] = true
	}

	for _, arrow := range spec.Arrows {
		if !validBoxIDs[arrow.FromID] {
			return nil, fmt.Errorf("arrow '%s -> %s' references non-existent box label '%s'", arrow.FromID, arrow.ToID, arrow.FromID)
		}
		if !validBoxIDs[arrow.ToID] {
			return nil, fmt.Errorf("arrow '%s -> %s' references non-existent box label '%s'", arrow.FromID, arrow.ToID, arrow.ToID)
		}
	}

	// Build groups from box assignments and group definitions
	groupBoxIDs := make(map[string][]string) // group name -> list of box IDs
	var groupOrder []string                  // preserve first-seen order
	for boxID, gName := range boxGroups {
		if _, seen := groupBoxIDs[gName]; !seen {
			groupOrder = append(groupOrder, gName)
		}
		groupBoxIDs[gName] = append(groupBoxIDs[gName], boxID)
	}

	for _, gName := range groupOrder {
		label := gName // default label is the group name
		if defLabel, ok := groupDefs[gName]; ok {
			label = defLabel
		}
		spec.Groups = append(spec.Groups, GroupDef{
			Name:   gName,
			Label:  label,
			BoxIDs: groupBoxIDs[gName],
		})
	}

	return spec, nil
}
