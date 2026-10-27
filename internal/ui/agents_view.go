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
		// Section title
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(16), "Installed & Detected AI Agents")
					lbl.Color = pal.TextPrimary
					lbl.Font.Weight = 600
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(13), "Real-time process audit, local LLM endpoints, and background task tracker")
					lbl.Color = pal.TextSecondary
					return lbl.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingMedium)}.Layout),
		// List of cards
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
			// Status dot
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
			// Main agent info
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							// Name
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(15), agent.Name)
								lbl.Color = pal.TextPrimary
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
							// Type badge
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
					// Current task
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						taskText := agent.CurrentTask
						if taskText == "" {
							taskText = "No active tasks"
						}
						lbl := material.Label(a.theme, unit.Sp(13), "Task: "+taskText)
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
			// Right column: status badge and telemetry
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
							lbl := material.Label(a.theme, unit.Sp(12), agent.Status.Label())
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
							meta = fmt.Sprintf("Port: %d", agent.Port)
						} else {
							meta = "Service offline"
						}

						if agent.TaskCount > 0 {
							meta += fmt.Sprintf(" • Tasks: %d", agent.TaskCount)
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

	paint.FillShape(gtx.Ops, pal.Surface, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
	call.Add(gtx.Ops)
	return dims
}
