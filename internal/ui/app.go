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

// App инкапсулирует состояние и интерфейс окна SailC.
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

	// Список агентов
	agentList widget.List

	// Центральный лог
	logList       widget.List
	selectedThrID string
	searchEditor  widget.Editor
	btnNewAsk     widget.Clickable
	newAskActive  bool
	askTitleEdit  widget.Editor
	askBodyEdit   widget.Editor
	askAuthorEdit widget.Editor
	btnSubmitAsk  widget.Clickable
	btnCancelAsk  widget.Clickable

	// Добавление ответа
	replyBodyEdit   widget.Editor
	replyAuthorEdit widget.Editor
	btnSubmitReply  widget.Clickable

	// Прокрутка деталей
	detailList widget.List
}

// NewApp инициализирует UI приложения.
func NewApp(store *domain.Store, detector *theme.Detector, mcpPort int) *App {
	th := material.NewTheme()
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

	a.askAuthorEdit.SingleLine = true
	a.askAuthorEdit.SetText("Antigravity")
	a.askTitleEdit.SingleLine = true

	a.replyAuthorEdit.SingleLine = true
	a.replyAuthorEdit.SetText("Claude")

	return a
}

// Run запускает главный цикл графического окна.
func (a *App) Run() error {
	a.window = new(app.Window)
	a.window.Option(app.Title("SailC — Хаб координации локальных ИИ"))
	a.window.Option(app.Size(unit.Dp(1080), unit.Dp(720)))
	a.window.Option(app.MinSize(unit.Dp(800), unit.Dp(560)))

	// Подписка на обновление данных в Store для немедленной перерисовки
	storeUpdates := a.store.Subscribe()
	go func() {
		for range storeUpdates {
			if a.window != nil {
				a.window.Invalidate()
			}
		}
	}()

	// Подписка на изменение темы в Windows
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

	// Заливка основного холста
	paint.Fill(gtx.Ops, pal.Canvas)

	// Проверка переключения вкладок
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
		// 1. Верхний заголовок и навигация
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutHeader(gtx, pal)
		}),
		// 2. Основная рабочая область с отступами и воздухом
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
	// Подложка шапки без рамок и без теней
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
			// Логотип SailC
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawSailLogo(gtx, unit.Dp(30), pal.Accent)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
			// Название и статус
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(18), "SailC")
						lbl.Color = pal.TextPrimary
						lbl.Font.Weight = 700
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(11), "Единый контекстный хаб локальных ИИ")
						lbl.Color = pal.TextSecondary
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(28)}.Layout),
			// Вкладки
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutTabButton(gtx, &a.btnAgents, "ИИ на компьютере", a.currentTab == tabAgents, pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutTabButton(gtx, &a.btnLog, "Центральный лог", a.currentTab == tabLog, pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutTabButton(gtx, &a.btnInteg, "Интеграция MCP", a.currentTab == tabIntegration, pal)
					}),
				)
			}),
			layout.Flexed(1, layout.Spacer{}.Layout),
			// Статистика хаба
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				tot, work, tasks := a.store.TotalStats()
				infoText := fmt.Sprintf("MCP: :%d  •  Активно ИИ: %d/%d  •  Задач: %d", a.mcpPort, work, tot, tasks)
				return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(12), infoText)
					lbl.Color = pal.TextSecondary
					return lbl.Layout(gtx)
				})
			}),
		)
	})
	call := macro.Stop()

	// Заливаем подложку шапки
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
