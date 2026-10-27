package ui

import (
	"fmt"
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/sergio-ilin/sailc/internal/brand"
	"github.com/sergio-ilin/sailc/internal/domain"
)

func (a *App) layoutAgentsTab(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	agents := a.store.ListAgents()

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Заголовок раздела
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(16), "Установленные и активные ИИ")
					lbl.Color = pal.TextPrimary
					lbl.Font.Weight = 600
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(13), "Аудит запущенных процессов, локальных нейросетевых движков и отслеживание задач")
					lbl.Color = pal.TextSecondary
					return lbl.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingMedium)}.Layout),
		// Список карточек
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(a.theme, &a.agentList).Layout(gtx, len(agents), func(gtx layout.Context, i int) layout.Dimensions {
				agent := agents[i]
				return layout.Inset{
					Bottom: unit.Dp(12),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return a.layoutAgentCard(gtx, agent, pal)
				})
			})
		}),
	)
}

func (a *App) layoutAgentCard(gtx layout.Context, agent *domain.Agent, pal brand.Palette) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(14),
		Bottom: unit.Dp(14),
		Left:   unit.Dp(18),
		Right:  unit.Dp(18),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// Индикатор статуса (точка)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				var col = pal.StatusOffline
				switch agent.Status {
				case domain.StatusWorking:
					col = pal.StatusWorking
				case domain.StatusOnline:
					col = pal.StatusOnline
				}
				return DrawStatusDot(gtx, unit.Dp(5), col)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(14)}.Layout),
			// Основная информация об агенте
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							// Имя
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(15), agent.Name)
								lbl.Color = pal.TextPrimary
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
							// Тип
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(11), agent.Type)
									lbl.Color = pal.TextSecondary
									return lbl.Layout(gtx)
								})
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					// Текущая задача
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						taskText := agent.CurrentTask
						if taskText == "" {
							taskText = "Нет активных задач"
						}
						lbl := material.Label(a.theme, unit.Sp(13), "Задача: "+taskText)
						if agent.Status == domain.StatusWorking {
							lbl.Color = pal.StatusWorking
						} else {
							lbl.Color = pal.TextSecondary
						}
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
			// Правая колонка: статус, счётчик задач и технические детали
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.End,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						var stColor = pal.StatusOffline
						switch agent.Status {
						case domain.StatusWorking:
							stColor = pal.StatusWorking
						case domain.StatusOnline:
							stColor = pal.StatusOnline
						}

						return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(12), agent.Status.RussianLabel())
							lbl.Color = stColor
							lbl.Font.Weight = 600
							return lbl.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						var meta string
						if agent.PID > 0 {
							meta = fmt.Sprintf("PID: %d", agent.PID)
						} else if agent.Port > 0 {
							meta = fmt.Sprintf("Порт: %d", agent.Port)
						} else {
							meta = "Служба не активна"
						}

						if agent.TaskCount > 0 {
							meta += fmt.Sprintf(" • Задач: %d", agent.TaskCount)
						}

						lbl := material.Label(a.theme, unit.Sp(11), meta)
						lbl.Color = pal.TextMuted
						return lbl.Layout(gtx)
					}),
				)
			}),
		)
	})
	call := macro.Stop()

	// Заливаем поверхность карточки (без рамок и без теней)
	paint.FillShape(gtx.Ops, pal.Surface, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
	call.Add(gtx.Ops)
	return dims
}
