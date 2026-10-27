package theme

import (
	"runtime"
	"sync"
	"time"

	"github.com/sergio-ilin/sailc/internal/brand"
	"golang.org/x/sys/windows/registry"
)

// Detector отслеживает системную тему оформления ОС Windows.
type Detector struct {
	mu           sync.RWMutex
	isDark       bool
	palette      brand.Palette
	onChangeList []func(brand.Palette)
}

// NewDetector инициализирует детектор темы.
func NewDetector() *Detector {
	d := &Detector{}
	d.refresh()

	// Запуск фонового мониторинга реестра Windows
	go d.watchLoop()
	return d
}

// GetPalette возвращает текущую палитру.
func (d *Detector) GetPalette() brand.Palette {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.palette
}

// IsDark возвращает true, если активна тёмная тема.
func (d *Detector) IsDark() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.isDark
}

// OnChange регистрирует колбэк на смену темы.
func (d *Detector) OnChange(fn func(brand.Palette)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onChangeList = append(d.onChangeList, fn)
}

func (d *Detector) refresh() bool {
	isDark := true // По умолчанию тёмная тема

	if runtime.GOOS == "windows" {
		k, err := registry.OpenKey(
			registry.CURRENT_USER,
			`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
			registry.QUERY_VALUE,
		)
		if err == nil {
			val, _, err := k.GetIntegerValue("AppsUseLightTheme")
			k.Close()
			if err == nil {
				isDark = (val == 0)
			}
		}
	}

	d.mu.Lock()
	changed := (d.isDark != isDark || d.palette.Canvas.A == 0)
	d.isDark = isDark
	if isDark {
		d.palette = brand.DarkTheme()
	} else {
		d.palette = brand.LightTheme()
	}
	callbacks := append([]func(brand.Palette){}, d.onChangeList...)
	currentPal := d.palette
	d.mu.Unlock()

	if changed {
		for _, cb := range callbacks {
			cb(currentPal)
		}
	}
	return changed
}

func (d *Detector) watchLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		d.refresh()
	}
}
