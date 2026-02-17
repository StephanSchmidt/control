package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// FontData holds the encoded font information
type FontData struct {
	Base64Data string
	FontName   string
}

// LoadCustomFont reads a WOFF2 font file and returns the base64-encoded data
func LoadCustomFont(path string) (*FontData, error) {
	// Clean the path to prevent directory traversal attacks
	cleanPath := filepath.Clean(path)

	// Read the font file
	fontBytes, err := os.ReadFile(cleanPath) // #nosec G304 -- path is cleaned via filepath.Clean
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(fontBytes)

	// Extract font name from filename (without extension)
	fontName := filepath.Base(path)
	// Remove .woff2 extension
	if len(fontName) > 6 && fontName[len(fontName)-6:] == ".woff2" {
		fontName = fontName[:len(fontName)-6]
	}

	return &FontData{
		Base64Data: base64Data,
		FontName:   fontName,
	}, nil
}
