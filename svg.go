package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// svgHeader returns the SVG opening tag with optional embedded font and arrowhead marker definition
func svgHeader(width, height int, font *FontData) string {
	header := fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height) +
		`<defs>`

	// Add @font-face only if custom font is provided
	if font != nil {
		safeName := sanitizeFontName(font.FontName)
		header += `<style type="text/css">` +
			`@font-face {` +
			`font-family: '` + safeName + `';` +
			`src: url(data:font/woff2;base64,` + font.Base64Data + `);` +
			`}` +
			`</style>`
	}

	// Add arrowhead marker definition
	header += `<marker id="arrowhead" markerWidth="12" markerHeight="13" refX="8" refY="5.5" orient="auto" viewBox="-1 -1 14 13"><polyline points="0 0.5, 8 5.5, 0 10.4" fill="none" stroke="#000" stroke-width="1.5" stroke-linejoin="miter"/></marker>` +
		`</defs>` +
		fmt.Sprintf(`<rect width="%d" height="%d" fill="#fff"/>`, width, height)

	return header
}

// svgFooter returns the closing SVG tag
func svgFooter() string {
	return `</svg>`
}

// sanitizeFontName removes characters that could break CSS font-family declarations
func sanitizeFontName(name string) string {
	// Remove characters that could escape CSS string context or inject SVG/XML
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, "<", "")
	name = strings.ReplaceAll(name, ">", "")
	name = strings.ReplaceAll(name, ";", "")
	name = strings.ReplaceAll(name, "{", "")
	name = strings.ReplaceAll(name, "}", "")
	return name
}

// getFontFamily returns the font-family CSS value with fallback stack
func getFontFamily(customFont *FontData) string {
	fallbacks := `'Arial Narrow', 'Helvetica Neue Condensed', 'Ubuntu Condensed', 'Liberation Sans Narrow', Impact, sans-serif`
	if customFont != nil {
		return `'` + sanitizeFontName(customFont.FontName) + `', ` + fallbacks
	}
	return fallbacks
}

// straightArrow generates a single-line arrow
func straightArrow(fromX, fromY, toX, toY int) string {
	return fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#000" stroke-width="2" marker-end="url(#arrowhead)"/>`,
		fromX, fromY, toX, toY)
}

// oneBentArrow generates a 2-segment L-shaped arrow
// If verticalFirst=true: vertical then horizontal
// If verticalFirst=false: horizontal then vertical
func oneBentArrow(fromX, fromY, toX, toY int, verticalFirst bool) string {
	if verticalFirst {
		// Vertical then horizontal
		return fmt.Sprintf(`<polyline points="%d,%d %d,%d %d,%d" fill="none" stroke="#000" stroke-width="2" marker-end="url(#arrowhead)"/>`,
			fromX, fromY, // Start point
			fromX, toY, // Vertical to target Y
			toX, toY) // Horizontal to end point
	}
	// Horizontal then vertical
	return fmt.Sprintf(`<polyline points="%d,%d %d,%d %d,%d" fill="none" stroke="#000" stroke-width="2" marker-end="url(#arrowhead)"/>`,
		fromX, fromY, // Start point
		toX, fromY, // Horizontal to target X
		toX, toY) // Vertical to end point
}

// twoBentArrow generates a 3-segment arrow (horizontal, vertical, horizontal)
func twoBentArrow(fromX, fromY, toX, toY int) string {
	midX := (fromX + toX) / 2
	return fmt.Sprintf(`<polyline points="%d,%d %d,%d %d,%d %d,%d" fill="none" stroke="#000" stroke-width="2" marker-end="url(#arrowhead)"/>`,
		fromX, fromY, // Start point
		midX, fromY, // Horizontal to midpoint
		midX, toY, // Vertical to target Y
		toX, toY) // Horizontal to end point
}

// twoBentArrowVertical generates a 3-segment arrow (vertical, horizontal, vertical)
func twoBentArrowVertical(fromX, fromY, toX, toY int) string {
	midY := (fromY + toY) / 2
	return fmt.Sprintf(`<polyline points="%d,%d %d,%d %d,%d %d,%d" fill="none" stroke="#000" stroke-width="2" marker-end="url(#arrowhead)"/>`,
		fromX, fromY, // Start point
		fromX, midY, // Vertical to midpoint
		toX, midY, // Horizontal to target X
		toX, toY) // Vertical to end point
}

// drawGroup generates a dashed rounded rectangle with an optional label
func drawGroup(x, y, width, height int, label string, font *FontData) string {
	result := fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="#000" stroke-width="1.5" stroke-dasharray="6,4" rx="8"/>`,
		x, y, width, height)
	if label != "" {
		attrs := map[string]string{}
		result += drawText(x+10, y+24, label, 24, attrs, font)
	}
	return result
}

// drawBox generates an SVG rectangle
func drawBox(x, y, width, height int, fillColor, strokeColor string, strokeWidth int) string {
	// Use defaults if not specified
	if strokeColor == "" {
		strokeColor = "#000"
	}
	if strokeWidth == 0 {
		strokeWidth = 2
	}
	return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" stroke="%s" stroke-width="%d"/>`,
		x, y, width, height, fillColor, strokeColor, strokeWidth)
}

// drawLine generates an SVG line, optionally dashed
func drawLine(x1, y1, x2, y2 int, dashed bool, withArrow bool) string {
	dashAttr := ""
	if dashed {
		dashAttr = ` stroke-dasharray="5,5"`
	}
	arrowAttr := ""
	if withArrow {
		arrowAttr = ` marker-end="url(#arrowhead)"`
	}
	return fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#000" stroke-width="2"%s%s/>`,
		x1, y1, x2, y2, dashAttr, arrowAttr)
}

// drawText generates SVG text with optional attributes
func drawText(x, y int, text string, fontSize int, attrs map[string]string, font *FontData) string {
	attrStr := ""
	for key, val := range attrs {
		attrStr += fmt.Sprintf(` %s="%s"`, key, val)
	}

	fontFamily := getFontFamily(font)

	// Sanitize text to prevent XSS
	buffer := &bytes.Buffer{}
	if err := xml.EscapeText(buffer, []byte(text)); err != nil {
		// This should not happen with valid UTF-8 strings, but handle it defensively
		// Replace with a placeholder to avoid injecting potentially broken XML
		return fmt.Sprintf(`<text x="%d" y="%d" font-family="%s" font-size="%d"%s>[error]</text>`,
			x, y, fontFamily, fontSize, attrStr)
	}
	sanitizedText := buffer.String()

	return fmt.Sprintf(`<text x="%d" y="%d" font-family="%s" font-size="%d"%s>%s</text>`,
		x, y, fontFamily, fontSize, attrStr, sanitizedText)
}

// drawBoxText generates multi-line centered text for a box
func drawBoxText(x, y, width, height, fontSize int, textColor string, lines []string, font *FontData) string {
	// Use default font size if not specified
	if fontSize == 0 {
		fontSize = 24
	}
	// Line height is proportional to font size (default: 28 for font size 24)
	lineHeight := fontSize * 28 / 24
	startY := y + height/2 - (len(lines)-1)*lineHeight/2
	result := ""
	for i, line := range lines {
		attrs := map[string]string{
			"font-weight":       "normal",
			"text-anchor":       "middle",
			"dominant-baseline": "middle",
		}
		// Add text color if specified
		if textColor != "" {
			attrs["fill"] = textColor
		}
		result += drawText(x+width/2, startY+i*lineHeight, line, fontSize, attrs, font)
	}
	return result
}
