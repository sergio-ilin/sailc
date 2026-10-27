package brand

import (
	"image/color"
)

// Palette defines the modern, sleek color scheme for SailC.
// Directives: zero heavy borders, zero shadows, spacious breathing layout,
// high-contrast obsidian & electric cyan/mint aesthetic.
type Palette struct {
	IsDark          bool
	Canvas          color.NRGBA // Deep obsidian / clean canvas background
	Surface         color.NRGBA // Card & container surfaces
	SurfaceElevated color.NRGBA // Hover / active elevated surfaces
	TextPrimary     color.NRGBA // Primary high-contrast typography
	TextSecondary   color.NRGBA // Secondary subtitle / metadata typography
	TextMuted       color.NRGBA // Subdued / disabled text
	Accent          color.NRGBA // Distinctive Electric Cyan / Aqua accent
	AccentSoft      color.NRGBA // Subtle accent background tint
	StatusWorking   color.NRGBA // Warm amber for active processing
	StatusOnline    color.NRGBA // Emerald / Mint for ready state
	StatusOffline   color.NRGBA // Neutral slate for inactive state
	StatusPillBg    color.NRGBA // Pill background
}

// DarkTheme returns the obsidian & electric cyan dark palette.
func DarkTheme() Palette {
	return Palette{
		IsDark:          true,
		Canvas:          color.NRGBA{R: 0x0A, G: 0x0D, B: 0x12, A: 0xFF}, // Obsidian
		Surface:         color.NRGBA{R: 0x13, G: 0x17, B: 0x22, A: 0xFF}, // Deep matte slate
		SurfaceElevated: color.NRGBA{R: 0x1C, G: 0x23, B: 0x33, A: 0xFF}, // Interactive slate
		TextPrimary:     color.NRGBA{R: 0xF8, G: 0xFA, B: 0xFC, A: 0xFF}, // Crisp white
		TextSecondary:   color.NRGBA{R: 0x94, G: 0xA3, B: 0xB8, A: 0xFF}, // Cool slate
		TextMuted:       color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xFF}, // Deep slate
		Accent:          color.NRGBA{R: 0x06, G: 0xB6, B: 0xD4, A: 0xFF}, // Electric Cyan
		AccentSoft:      color.NRGBA{R: 0x08, G: 0x33, B: 0x44, A: 0x66}, // Cyan glow
		StatusWorking:   color.NRGBA{R: 0xF5, G: 0x9E, B: 0x0B, A: 0xFF}, // Amber
		StatusOnline:    color.NRGBA{R: 0x10, G: 0xB9, B: 0x81, A: 0xFF}, // Emerald
		StatusOffline:   color.NRGBA{R: 0x64, G: 0x74, B: 0x8B, A: 0xFF}, // Cool gray
		StatusPillBg:    color.NRGBA{R: 0x0E, G: 0x13, B: 0x1D, A: 0xEE},
	}
}

// LightTheme returns the clean refined light palette.
func LightTheme() Palette {
	return Palette{
		IsDark:          false,
		Canvas:          color.NRGBA{R: 0xF4, G: 0xF6, B: 0xF9, A: 0xFF}, // Soft slate pearl
		Surface:         color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, // Pure white
		SurfaceElevated: color.NRGBA{R: 0xEA, G: 0xEF, B: 0xF5, A: 0xFF}, // Soft hover
		TextPrimary:     color.NRGBA{R: 0x0B, G: 0x0F, B: 0x19, A: 0xFF}, // Deep obsidian ink
		TextSecondary:   color.NRGBA{R: 0x52, G: 0x60, B: 0x71, A: 0xFF}, // Slate
		TextMuted:       color.NRGBA{R: 0x8C, G: 0x9B, B: 0xAE, A: 0xFF}, // Light slate
		Accent:          color.NRGBA{R: 0x02, G: 0x84, B: 0xC7, A: 0xFF}, // Vivid ocean cyan
		AccentSoft:      color.NRGBA{R: 0xE0, G: 0xF2, B: 0xFE, A: 0xCC}, // Soft cyan tint
		StatusWorking:   color.NRGBA{R: 0xD9, G: 0x77, B: 0x06, A: 0xFF},
		StatusOnline:    color.NRGBA{R: 0x05, G: 0x96, B: 0x69, A: 0xFF},
		StatusOffline:   color.NRGBA{R: 0x94, G: 0xA3, B: 0xB8, A: 0xFF},
		StatusPillBg:    color.NRGBA{R: 0xF1, G: 0xF5, B: 0xF9, A: 0xCC},
	}
}

// Spatial rhythm and border radii constants
const (
	PaddingSmall  = 8
	PaddingNormal = 16
	PaddingMedium = 20
	PaddingLarge  = 28

	RadiusSmall  = 6
	RadiusNormal = 10
	RadiusLarge  = 14
)
