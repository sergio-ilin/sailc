package ui

import (
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

type tabID int

const (
	tabAI tabID = iota
	tabLog
)

// App manages the Gio application lifecycle, state, and rendering loop.
type App struct {
	window   *app.Window
	store    *domain.Store
	detector *theme.Detector
	theme    *material.Theme
	mcpPort  int

	// Navigation & View state
	currentTab tabID
	btnAI      widget.Clickable
	btnLog     widget.Clickable

	// AI View widgets
	btnAddAI          widget.Clickable
	btnCloseAdd       widget.Clickable
	agentList         widget.List
	availableList     widget.List
	aiSelectionClicks map[string]*widget.Clickable
	removeAIClicks    map[string]*widget.Clickable
	isAddingAI        bool

	// Log View widgets
	searchEditor  widget.Editor
	logList       widget.List
	detailList    widget.List
	selectedThrID string
}

// NewApp initializes the SailC desktop interface.
func NewApp(store *domain.Store, detector *theme.Detector, mcpPort int) *App {
	th := createThemeWithSystemFont()

	a := &App{
		store:    store,
		detector: detector,
		theme:    th,
		mcpPort:  mcpPort,
	}

	a.agentList.List.Axis = layout.Vertical
	a.logList.List.Axis = layout.Vertical
	a.detailList.List.Axis = layout.Vertical
	a.availableList.List.Axis = layout.Vertical

	a.searchEditor.SingleLine = true
	a.searchEditor.Submit = true

	a.availableList.Axis = layout.Vertical
	a.aiSelectionClicks = make(map[string]*widget.Clickable)
	a.removeAIClicks = make(map[string]*widget.Clickable)

	return a
}

// Run executes the native window event loop.
func Run(store *domain.Store, detector *theme.Detector, mcpPort int) error {
	a := NewApp(store, detector, mcpPort)
	return a.Run()
}

// Run executes the native window event loop.
func (a *App) Run() error {
	a.window = new(app.Window)
	a.window.Option(app.Title("SailC"))
	a.window.Option(app.Size(unit.Dp(960), unit.Dp(640)))
	a.window.Option(app.MinSize(unit.Dp(720), unit.Dp(480)))

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

	if a.btnAI.Clicked(gtx) {
		a.currentTab = tabAI
		a.isAddingAI = false
	}
	if a.btnLog.Clicked(gtx) {
		a.currentTab = tabLog
		a.isAddingAI = false
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Header bar with vector logo and unified FontAwesome tab items
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutHeader(gtx, pal)
		}),
		// Content area with breathing margins
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(brand.PaddingNormal),
				Bottom: unit.Dp(brand.PaddingNormal),
				Left:   unit.Dp(brand.PaddingMedium),
				Right:  unit.Dp(brand.PaddingMedium),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				switch a.currentTab {
				case tabAI:
					return a.layoutAITab(gtx, pal)
				case tabLog:
					return a.layoutLogTab(gtx, pal)
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
		Top:    unit.Dp(10),
		Bottom: unit.Dp(10),
		Left:   unit.Dp(brand.PaddingMedium),
		Right:  unit.Dp(brand.PaddingMedium),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// Minimalist SailC Geometric Vector Logo
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawSailLogo(gtx, unit.Dp(26), pal.Accent)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(24)}.Layout),
			// Nav tabs: "AI" and "Log" with FontAwesome icons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutTabButton(gtx, &a.btnAI, IconMicrochip, "AI", a.currentTab == tabAI, pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutTabButton(gtx, &a.btnLog, IconComments, "Log", a.currentTab == tabLog, pal)
					}),
				)
			}),
			layout.Flexed(1, layout.Spacer{}.Layout),
		)
	})
	call := macro.Stop()

	paint.FillShape(gtx.Ops, pal.Surface, clip.Rect{Max: dims.Size}.Op())
	call.Add(gtx.Ops)
	return dims
}

func (a *App) layoutTabButton(gtx layout.Context, btn *widget.Clickable, iconCode string, title string, active bool, pal brand.Palette) layout.Dimensions {
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
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return DrawFAIcon(a.theme, gtx, iconCode, unit.Sp(12), textColor)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(13), title)
					lbl.Color = textColor
					if active {
						lbl.Font.Weight = 600
					}
					return lbl.Layout(gtx)
				}),
			)
		})
		call := macro.Stop()

		paint.FillShape(gtx.Ops, bg, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
		call.Add(gtx.Ops)
		return dims
	})
}
