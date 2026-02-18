package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
)

var version = "dev"

type CLI struct {
	Version     kong.VersionFlag `help:"Print version and exit"`
	Diagram     string           `help:"Input diagram file" type:"path" default:"examples/diagram.txt"`
	Out         string           `help:"Output SVG file" type:"path" default:"examples/diagram.svg"`
	Stretch     float64          `help:"Horizontal stretch factor (1.0 = normal, 0.8 = 80% width)" default:"1.0"`
	VerticalGap float64          `help:"Vertical gap between boxes in grid units" default:"0.5"`
	Font        string           `help:"Custom font file (WOFF2 format) to embed in SVG" type:"path" optional:""`
	Debug       string           `help:"Output debug information to JSON file" type:"path" optional:""`
}

func printHelp() {
	help := `control - Generate SVG diagrams from text specifications

USAGE:
  control --diagram <file> --out <output.svg> [options]

OPTIONS:
  --diagram <file>    Input diagram file (default: examples/diagram.txt)
  --out <file>        Output SVG file (default: examples/diagram.svg)
  --stretch <float>   Horizontal stretch factor, 1.0 = normal (default: 1.0)
  --vertical-gap <f>  Vertical gap between boxes in grid units (default: 0.5)
  --font <file>       Custom font file (WOFF2 format) to embed in SVG
  --debug <file>      Output debug information to JSON file

FRONTMATTER:
  Diagram files can include optional metadata at the top of the file,
  enclosed between "---" delimiters.

  Recognized keys:
    font: <path>           Path to a custom font file (WOFF2 format)
    x-label: <text>        X-axis label (omit to hide axes)
    y-label: <text>        Y-axis label (omit to hide axes)
    legend: <style> = <text>  Legend entry (repeatable)
    color: <name> = <hex>  Custom color definition (repeatable)

  If neither x-label nor y-label is set, axes are not drawn.
  Legend entries map style codes to descriptions, rendered top-right.
  Comments (#) and blank lines are allowed within frontmatter.
  CLI flags (e.g. --font) take precedence over frontmatter values.

  Example:
    ---
    font: fonts/BerkeleyMono-Condensed.woff2
    x-label: Time
    y-label: Control
    legend: p = In Progress
    legend: g = Completed
    color: green = #00FF00
    ---

  Custom colors can be used as style codes:
    green       Use as background color
    greent      Use as text color (append "t")
    nbb-greent  Combine with other styles

    1,1: OKRs
    >+1,+1: Sprint Planning

DIAGRAM FORMAT:
  A diagram file contains box definitions and arrow definitions.
  Arrow lines (containing "->") can appear anywhere in the file.
  An optional "---" separator can be used for visual clarity.

  Lines starting with "#" are comments.

BOX SYNTAX:
  [id:] x,y[,width[,height]]: Label[, style] [@Group]

  - id        Optional identifier for referencing in arrows (alphanumeric, _, -)
  - x,y       Grid coordinates (starting from 1)
  - width     Grid width (default: 2, supports fractions like 1/2)
  - height    Grid height (default: 1)
  - Label     Display text inside the box
  - style     Optional comma-separated style codes
  - @Group    Optional group assignment (see GROUPS below)

RELATIVE COORDINATES:
  Use +N or -N for coordinates relative to the previous box.
  Use 0 as shorthand for +0 (same position as previous box).

  Example: if previous box is at 3,2:
    +1,+1  -> resolves to 4,3
    +2,-1  -> resolves to 5,1
    0,+1   -> resolves to 3,3

AUTO-ARROWS (> prefix):
  Prefix coordinates with ">" to automatically create an arrow
  from the previous box to this one.

    1,1: Start
    >+1,+1: Next       # creates arrow Start -> Next

TOUCH-LEFT (| prefix):
  Prefix coordinates with "|" to visually attach this box to the
  previous one (no gap). Requires same row (Y=0) and positive
  relative X.

    1,1: First
    |+2,0: Second      # Second touches First on its left side

STYLES:
  g     Gray background (#D3D3D3)
  p     Purple background (#ecbae6)
  lp    Light purple background (#f5dbf2)
  rb    Red border (3px)
  rt    Red text
  nbb   No background, no border (for overlay text)
  2t    Double text size (48px)

  Combine with dashes:
    rb-g      Red border + gray background (highlighted task)
    rb-p      Red border + purple background (urgent in-progress)
    lp-rb     Light purple + red border
    nbb-2t    Invisible box, large text (section header)
    nbb-rt    Invisible box, red text (warning label)
    nbb-rt-2t Invisible box, red large text (floating overlay)

GROUPS (@prefix):
  Visually group boxes with a surrounding dashed rectangle.

  Assign a box to a group by appending @GroupName to the box line:
    1,2: Stephan @Team
    >3,2: Stefanie, p @Team

  Optionally define a group label:
    @Team: Our Team

  If no label line is defined, the group name is used as the label.

ARROW SYNTAX:
  from_id -> to_id

  Arrow lines can appear anywhere in the file (no separator needed).
  Arrows route automatically using orthogonal segments (left-to-right).

EXAMPLES:

  Simple flow:
    a: 1,1: Planning
    b: >+2,0: Development
    c: >+2,0: Review
    ---
    c -> a

  Scrum board:
    1,1: OKRs
    >+1,+1: Business Initiatives
    >+1,+1: Backlog Planning
    >+1,+1: Sprint Planning
    >+1,+1,1: Daily
    >+2,+1: Developing, p
    >+2,-1,1/2:
    >0,+1: Developing, p
    >+2,-1,2: Sprint Review
    9,1: SCRUM, nbb-rt-2t

  Boxes with IDs and manual arrows:
    plan: 1,1: Plan
    dev: 3,2: Develop
    test: 5,2: Test
    ship: 7,1: Ship
    ---
    plan -> dev
    dev -> test
    test -> ship
    test -> dev

  Visual groups:
    1,2: Stephan @Team
    >+2,0: Stefanie, p @Team
    >+2,0: Me, g @Team
    1,1: AI, nbb-2t
    @Team: Our Team
`
	fmt.Print(help)
}

func main() {
	// Print help when called without arguments
	if len(os.Args) == 1 {
		printHelp()
		return
	}

	var cli CLI
	kong.Parse(&cli, kong.Vars{"version": version})

	// Security check: prevent running as root
	if err := checkNotRoot(); err != nil {
		fmt.Println("Security error:", err)
		os.Exit(1)
	}

	// Open output file early (before dropping capabilities)
	// This ensures we have write permission and get the file handle
	outFile, err := os.Create(cli.Out)
	if err != nil {
		fmt.Printf("Error creating output file '%s': %v\n", cli.Out, err)
		os.Exit(1)
	}
	defer func() { _ = outFile.Close() }()

	// Read diagram from file
	diagramBytes, err := os.ReadFile(cli.Diagram)
	if err != nil {
		fmt.Printf("Error reading file '%s': %v\n", cli.Diagram, err)
		os.Exit(1)
	}

	// Extract frontmatter (font path, etc.) before dropping capabilities
	frontmatter, diagramText := ParseFrontmatter(string(diagramBytes))

	// Determine font path: CLI flag takes precedence over frontmatter
	fontPath := cli.Font
	if fontPath == "" && frontmatter.Font != "" {
		fontPath = frontmatter.Font
	}

	// Load custom font if specified (must happen before dropping capabilities)
	var fontData *FontData
	if fontPath != "" {
		fontData, err = LoadCustomFont(fontPath)
		if err != nil {
			fmt.Printf("Error loading font '%s': %v\n", fontPath, err)
			os.Exit(1)
		}
	}

	// Drop capabilities now that we have file handles secured
	if err := dropCapabilities(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: failed to drop capabilities:", err)
		os.Exit(1)
	}

	// Parse text into internal representation (pure logical structure)
	spec, err := ParseDiagramSpec(diagramText, frontmatter.Colors)
	if err != nil {
		fmt.Println("Error parsing diagram:", err)
		os.Exit(1)
	}

	// Create layout configuration
	config := NewDefaultConfig()
	config.Stretch = cli.Stretch
	config.VerticalGapUnits = cli.VerticalGap

	// Layout: convert logical spec to concrete diagram with pixel coordinates
	diagram, boxData := Layout(spec, config, frontmatter.Legend, spec.Groups, frontmatter.ArrowFlow)

	// Set presentation details
	diagram.YAxisLabel = frontmatter.YLabel
	diagram.XAxisLabel = frontmatter.XLabel
	diagram.Font = fontData
	diagram.Legend = frontmatter.Legend
	diagram.CustomColors = frontmatter.Colors

	// Render: generate SVG
	svg := diagram.GenerateSVG()

	// Write to the already-opened file
	_, err = io.WriteString(outFile, svg)
	if err != nil {
		fmt.Println("Error writing file:", err)
		os.Exit(1)
	}

	fmt.Printf("Diagram generated successfully: %s\n", cli.Out)

	// Write debug output if requested
	if cli.Debug != "" {
		debugOutput := GenerateDebugOutput(diagram, boxData)
		if err := WriteDebugJSON(cli.Debug, debugOutput); err != nil {
			fmt.Printf("Error writing debug file '%s': %v\n", cli.Debug, err)
			os.Exit(1)
		}
		fmt.Printf("Debug output written: %s\n", cli.Debug)
	}
}
