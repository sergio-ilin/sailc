package ui

import (
	"fmt"
	"image"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/sergio-ilin/sailc/internal/brand"
	"github.com/sergio-ilin/sailc/internal/domain"
	"github.com/sergio-ilin/sailc/internal/theme"
)

type activeTab int

const (
	tabAgents activeTab = iota
	tabLog
	tabIntegration
)

// App manages the SailC native desktop user interface.
type App struct {
	store    *domain.Store
	detector *theme.Detector
	theme    *material.Theme
	window   *app.Window
	mcpPort  int

	currentTab activeTab
	btnAgents  widget.Clickable
	btnLog     widget.Clickable
	btnInteg   widget.Clickable

	// Agent monitor list
	agentList widget.List

	// Central Log (Observation & Knowledge Inspector)
	logList       widget.List
	selectedThrID string
	searchEditor  widget.Editor

	detailList widget.List
}

// NewApp initializes the UI controller.
func NewApp(store *domain.Store, detector *theme.Detector, mcpPort int) *App {
	th := createThemeWithSystemFont()
	a := &App{
		store:    store,
		detector: detector,
		theme:    th,
		mcpPort:  mcpPort,
	}

	a.agentList.Axis = layout.Vertical
	a.logList.Axis = layout.Vertical
	a.detailList.Axis = layout.Vertical

	a.searchEditor.SingleLine = true
	a.searchEditor.Submit = true

	return a
}

// Run executes the native window event loop.
func (a *App) Run() error {
	a.window = new(app.Window)
	a.window.Option(app.Title("SailC — Local AI Coordination Hub"))
	a.window.Option(app.Size(unit.Dp(1080), unit.Dp(720)))
	a.window.Option(app.MinSize(unit.Dp(800), unit.Dp(560)))

	// Subscribe to store updates for reactive redraws
	storeUpdates := a.store.Subscribe()
	go func() {
		for range storeUpdates {
			if a.window != nil {
				a.window.Invalidate()
			}
		}
	}()

	// Subscribe to OS theme changes
	a.detector.OnChange(func(_ brand.Palette) {
		if a.window != nil {
			a.window.Invalidate()
		}
	})

	var ops op.Ops
	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (a *App) layout(gtx layout.Context) layout.Dimensions {
	pal := a.detector.GetPalette()

	paint.Fill(gtx.Ops, pal.Canvas)

	if a.btnAgents.Clicked(gtx) {
		a.currentTab = tabAgents
	}
	if a.btnLog.Clicked(gtx) {
		a.currentTab = tabLog
	}
	if a.btnInteg.Clicked(gtx) {
		a.currentTab = tabIntegration
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Header bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutHeader(gtx, pal)
		}),
		// Content area with generous air and padding
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(brand.PaddingNormal),
				Bottom: unit.Dp(brand.PaddingNormal),
				Left:   unit.Dp(brand.PaddingMedium),
				Right:  unit.Dp(brand.PaddingMedium),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				switch a.currentTab {
				case tabAgents:
					return a.layoutAgentsTab(gtx, pal)
				case tabLog:
					return a.layoutLogTab(gtx, pal)
				case tabIntegration:
					return a.layoutIntegrationTab(gtx, pal)
				default:
					return layout.Dimensions{}
				}
			})
		}),
	)
}

func (a *App) layoutHeader(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(14),
		Bottom: unit.Dp(14),
		Left:   unit.Dp(brand.PaddingMedium),
		Right:  unit.Dp(brand.PaddingMedium),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// SailC Vector Brand Logo
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawSailLogo(gtx, unit.Dp(30), pal.Accent)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
			// Title and tagline
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(18), "SailC")
						lbl.Color = pal.TextPrimary
						lbl.Font.Weight = 700
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(11), "Central Knowledge Hub for Local AI Agents")
						lbl.Color = pal.TextSecondary
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(28)}.Layout),
			// Nav tabs
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutTabButton(gtx, &a.btnAgents, "AI Agents", a.currentTab == tabAgents, pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutTabButton(gtx, &a.btnLog, "Central Log", a.currentTab == tabLog, pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutTabButton(gtx, &a.btnInteg, "MCP Protocol", a.currentTab == tabIntegration, pal)
					}),
				)
			}),
			layout.Flexed(1, layout.Spacer{}.Layout),
			// Live statistics badge
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				tot, work, tasks := a.store.TotalStats()
				infoText := fmt.Sprintf("MCP: :%d  •  Active AI: %d/%d  •  Tasks: %d", a.mcpPort, work, tot, tasks)
				return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(12), infoText)
					lbl.Color = pal.TextSecondary
					return lbl.Layout(gtx)
				})
			}),
		)
	})
	call := macro.Stop()

	paint.FillShape(gtx.Ops, pal.Surface, clip.Rect{Max: dims.Size}.Op())
	call.Add(gtx.Ops)
	return dims
}

func (a *App) layoutTabButton(gtx layout.Context, btn *widget.Clickable, title string, active bool, pal brand.Palette) layout.Dimensions {
	bg := pal.Surface
	textColor := pal.TextSecondary
	if active {
		bg = pal.AccentSoft
		textColor = pal.Accent
	} else if btn.Hovered() {
		bg = pal.SurfaceElevated
	}

	return material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		macro := op.Record(gtx.Ops)
		dims := layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   unit.Dp(14),
			Right:  unit.Dp(14),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(a.theme, unit.Sp(13), title)
			lbl.Color = textColor
			if active {
				lbl.Font.Weight = 600
			}
			return lbl.Layout(gtx)
		})
		call := macro.Stop()

		paint.FillShape(gtx.Ops, bg, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
		call.Add(gtx.Ops)
		return dims
	})
}
