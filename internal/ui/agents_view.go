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
					lbl := material.Label(a.theme, unit.Sp(13), "Audits local processes and ports with deterministic tracking and live telemetry")
					lbl.Color = pal.TextSecondary
					return lbl.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingMedium)}.Layout),
		// List of cards (stably ordered, no jitter)
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
			// Main agent information
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// Row 1: Agent Name, Type Badge, and Detection Source Badge
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
							// Type Badge
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(11), agent.Type)
									lbl.Color = pal.TextSecondary
									return lbl.Layout(gtx)
								})
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
							// Explicit Detection Source Badge (Where it comes from)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								sourceColor := pal.TextMuted
								if agent.Status == domain.StatusOnline || agent.Status == domain.StatusWorking {
									sourceColor = pal.Accent
								}
								return DrawBadge(gtx, pal.Canvas, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(11), agent.Source)
									lbl.Color = sourceColor
									return lbl.Layout(gtx)
								})
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					// Row 2: Current Task / State Description
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						taskText := agent.CurrentTask
						if taskText == "" {
							taskText = "No active tasks"
						}
						lbl := material.Label(a.theme, unit.Sp(13), "Task: "+taskText)
						if agent.Status == domain.StatusWorking {
							lbl.Color = pal.StatusWorking
						} else if agent.Status == domain.StatusOnline {
							lbl.Color = pal.TextSecondary
						} else {
							lbl.Color = pal.TextMuted
						}
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
			// Right column: Status label and telemetry
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.End,
				}.Layout(gtx,
					// Status Badge
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
					// Telemetry: Active task count
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						var meta string
						if agent.Status == domain.StatusOffline {
							meta = "Service Inactive"
						} else {
							meta = fmt.Sprintf("Active Tasks: %d", agent.TaskCount)
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

	// Fill surface without harsh borders or shadows
	paint.FillShape(gtx.Ops, pal.Surface, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
	call.Add(gtx.Ops)
	return dims
}
