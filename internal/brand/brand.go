package brand

import (
	"image/color"
)

// Palette определяет набор цветов для темы интерфейса SailC.
// Директива: отсутствие теней и рамок, мягкие контрастные подложки.
type Palette struct {
	IsDark          bool
	Canvas          color.NRGBA // Основной фон окна
	Surface         color.NRGBA // Фон карточек и панелей
	SurfaceElevated color.NRGBA // Фон при наведении / выделении
	TextPrimary     color.NRGBA // Основной цвет текста
	TextSecondary   color.NRGBA // Вторичный текст и подсказки
	TextMuted       color.NRGBA // Неактивный текст
	Accent          color.NRGBA // Фирменный лазурный акцент
	AccentSoft      color.NRGBA // Мягкий фон акцента для бейджей
	StatusWorking   color.NRGBA // Статус «В работе» (янтарный)
	StatusOnline    color.NRGBA // Статус «Готов» (изумрудный)
	StatusOffline   color.NRGBA // Статус «Не запущен» (серый)
	StatusPillBg    color.NRGBA // Фон статусных плашек
}

// DarkTheme возвращает фирменную тёмную палитру SailC.
func DarkTheme() Palette {
	return Palette{
		IsDark:          true,
		Canvas:          color.NRGBA{R: 0x0E, G: 0x11, B: 0x17, A: 0xFF},
		Surface:         color.NRGBA{R: 0x16, G: 0x1B, B: 0x26, A: 0xFF},
		SurfaceElevated: color.NRGBA{R: 0x1F, G: 0x26, B: 0x36, A: 0xFF},
		TextPrimary:     color.NRGBA{R: 0xE6, G: 0xED, B: 0xF3, A: 0xFF},
		TextSecondary:   color.NRGBA{R: 0x8B, G: 0x94, B: 0x9E, A: 0xFF},
		TextMuted:       color.NRGBA{R: 0x56, G: 0x5D, B: 0x68, A: 0xFF},
		Accent:          color.NRGBA{R: 0x3B, G: 0x82, B: 0xF6, A: 0xFF},
		AccentSoft:      color.NRGBA{R: 0x1E, G: 0x3A, B: 0x8A, A: 0x66},
		StatusWorking:   color.NRGBA{R: 0xF5, G: 0x9E, B: 0x0B, A: 0xFF},
		StatusOnline:    color.NRGBA{R: 0x10, G: 0xB9, B: 0x81, A: 0xFF},
		StatusOffline:   color.NRGBA{R: 0x6B, G: 0x72, B: 0x80, A: 0xFF},
		StatusPillBg:    color.NRGBA{R: 0x11, G: 0x16, B: 0x22, A: 0xCC},
	}
}

// LightTheme возвращает фирменную светлую палитру SailC.
func LightTheme() Palette {
	return Palette{
		IsDark:          false,
		Canvas:          color.NRGBA{R: 0xF8, G: 0xFA, B: 0xFC, A: 0xFF},
		Surface:         color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		SurfaceElevated: color.NRGBA{R: 0xED, G: 0xF2, B: 0xF7, A: 0xFF},
		TextPrimary:     color.NRGBA{R: 0x0F, G: 0x17, B: 0x2A, A: 0xFF},
		TextSecondary:   color.NRGBA{R: 0x64, G: 0x74, B: 0x8B, A: 0xFF},
		TextMuted:       color.NRGBA{R: 0x94, G: 0xA3, B: 0xB8, A: 0xFF},
		Accent:          color.NRGBA{R: 0x25, G: 0x63, B: 0xEB, A: 0xFF},
		AccentSoft:      color.NRGBA{R: 0xDB, G: 0xEA, B: 0xFE, A: 0xCC},
		StatusWorking:   color.NRGBA{R: 0xD9, G: 0x77, B: 0x06, A: 0xFF},
		StatusOnline:    color.NRGBA{R: 0x05, G: 0x96, B: 0x69, A: 0xFF},
		StatusOffline:   color.NRGBA{R: 0x94, G: 0xA3, B: 0xB8, A: 0xFF},
		StatusPillBg:    color.NRGBA{R: 0xF1, G: 0xF5, B: 0xF9, A: 0xCC},
	}
}

// Константы отступов и пространственного ритма (воздух)
const (
	PaddingSmall  = 8
	PaddingNormal = 16
	PaddingMedium = 20
	PaddingLarge  = 28

	RadiusSmall  = 6
	RadiusNormal = 10
	RadiusLarge  = 16
)
