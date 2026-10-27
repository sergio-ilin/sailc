package ui

import (
	_ "embed"
	"os"
	"path/filepath"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"gioui.org/text"
	"gioui.org/widget/material"
)

//go:embed fonts/Inter-Regular.ttf
var interRegularTTF []byte

//go:embed fonts/Inter-SemiBold.ttf
var interSemiBoldTTF []byte

//go:embed fonts/fa-solid-900.ttf
var fontAwesomeSolidTTF []byte

const (
	FontFamilyInter       = "Inter"
	FontFamilyFontAwesome = "FontAwesome"
)

// FontAwesome icon codepoints
const (
	IconPlus       = "\uf067" // fa-plus
	IconRobot      = "\uf544" // fa-robot
	IconMicrochip  = "\uf2db" // fa-microchip
	IconBolt       = "\uf0e7" // fa-bolt
	IconComments   = "\uf086" // fa-comments
	IconSearch     = "\uf002" // fa-magnifying-glass
	IconCheck      = "\uf00c" // fa-check
	IconXmark      = "\uf00d" // fa-xmark
	IconTerminal   = "\uf120" // fa-terminal
	IconServer     = "\uf233" // fa-server
	IconTrash      = "\uf2ed" // fa-trash-can
	IconLink       = "\uf0c1" // fa-link
	IconCircle     = "\uf111" // fa-circle
	IconLayerGroup = "\uf5fd" // fa-layer-group
)

// createThemeWithSystemFont loads embedded Inter typography and FontAwesome icon fonts,
// with graceful fallback to Segoe UI and gofont.
func createThemeWithSystemFont() *material.Theme {
	th := material.NewTheme()
	faces := make([]font.FontFace, 0)

	// 1. Embedded Inter Typography
	if face, err := opentype.Parse(interRegularTTF); err == nil {
		faces = append(faces, font.FontFace{
			Font: font.Font{Typeface: FontFamilyInter, Weight: font.Normal, Style: font.Regular},
			Face: face,
		})
	}
	if face, err := opentype.Parse(interSemiBoldTTF); err == nil {
		faces = append(faces, font.FontFace{
			Font: font.Font{Typeface: FontFamilyInter, Weight: font.Bold, Style: font.Regular},
			Face: face,
		})
	}

	// 2. Embedded FontAwesome 6 Solid
	if face, err := opentype.Parse(fontAwesomeSolidTTF); err == nil {
		faces = append(faces, font.FontFace{
			Font: font.Font{Typeface: FontFamilyFontAwesome, Weight: font.Normal, Style: font.Regular},
			Face: face,
		})
		faces = append(faces, font.FontFace{
			Font: font.Font{Typeface: FontFamilyFontAwesome, Weight: font.Bold, Style: font.Regular},
			Face: face,
		})
	}

	// 3. System Segoe UI fallback for Cyrillic / system symbols
	windir := os.Getenv("WINDIR")
	if windir == "" {
		windir = `C:\Windows`
	}
	fontsDir := filepath.Join(windir, "Fonts")

	for _, cand := range []struct {
		file   string
		weight font.Weight
	}{
		{"segoeui.ttf", font.Normal},
		{"segoeuib.ttf", font.Bold},
		{"arial.ttf", font.Normal},
	} {
		data, err := os.ReadFile(filepath.Join(fontsDir, cand.file))
		if err == nil {
			if face, err := opentype.Parse(data); err == nil {
				faces = append(faces, font.FontFace{
					Font: font.Font{Typeface: "Segoe UI", Weight: cand.weight},
					Face: face,
				})
			}
		}
	}

	// 4. Default gofont
	faces = append(faces, gofont.Collection()...)

	th.Shaper = text.NewShaper(text.WithCollection(faces))
	th.Face = FontFamilyInter

	return th
}
