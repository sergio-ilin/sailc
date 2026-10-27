package ui

import (
	"os"
	"path/filepath"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"gioui.org/text"
	"gioui.org/widget/material"
)

// createThemedShaper loads Windows system fonts (Segoe UI or Arial)
// to guarantee full Unicode support (Latin, Cyrillic, Symbols) without missing glyphs or question marks.
func createThemeWithSystemFont() *material.Theme {
	th := material.NewTheme()

	faces := make([]font.FontFace, 0)
	fontLoaded := false

	// Attempt to load Segoe UI (Windows native modern UI font)
	windir := os.Getenv("WINDIR")
	if windir == "" {
		windir = `C:\Windows`
	}
	fontsDir := filepath.Join(windir, "Fonts")

	fontCandidates := []struct {
		file     string
		typeface font.Typeface
		weight   font.Weight
		style    font.Style
	}{
		// Segoe UI
		{"segoeui.ttf", "Segoe UI", font.Normal, font.Regular},
		{"segoeuib.ttf", "Segoe UI", font.Bold, font.Regular},
		{"segoeuii.ttf", "Segoe UI", font.Normal, font.Italic},
		{"segoeuisb.ttf", "Segoe UI", font.Medium, font.Regular},
		// Arial fallback
		{"arial.ttf", "Arial", font.Normal, font.Regular},
		{"arialbd.ttf", "Arial", font.Bold, font.Regular},
	}

	for _, cand := range fontCandidates {
		fullPath := filepath.Join(fontsDir, cand.file)
		data, err := os.ReadFile(fullPath)
		if err == nil {
			face, err := opentype.Parse(data)
			if err == nil {
				faces = append(faces, font.FontFace{
					Font: font.Font{
						Typeface: cand.typeface,
						Weight:   cand.weight,
						Style:    cand.style,
					},
					Face: face,
				})
				fontLoaded = true
			}
		}
	}

	// Also append gofont fallback faces
	faces = append(faces, gofont.Collection()...)

	if fontLoaded {
		th.Shaper = text.NewShaper(text.WithCollection(faces))
		th.Face = "Segoe UI"
	}

	return th
}
