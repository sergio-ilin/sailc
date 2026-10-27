package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/sergio-ilin/sailc/internal/brand"
	"github.com/sergio-ilin/sailc/internal/domain"
	"github.com/sergio-ilin/sailc/internal/integrations"
)

func (a *App) getAIClick(id string) *widget.Clickable {
	c, ok := a.aiSelectionClicks[id]
	if !ok {
		c = new(widget.Clickable)
		a.aiSelectionClicks[id] = c
	}
	return c
}

func (a *App) getRemoveClick(id string) *widget.Clickable {
	c, ok := a.removeAIClicks[id]
	if !ok {
		c = new(widget.Clickable)
		a.removeAIClicks[id] = c
	}
	return c
}

func (a *App) layoutAITab(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	if a.btnAddAI.Clicked(gtx) {
		a.isAddingAI = true
	}

	if a.btnCloseAdd.Clicked(gtx) {
		a.isAddingAI = false
	}

	if a.isAddingAI {
		return a.layoutAddAI(gtx, pal)
	}

	configured := a.store.ListConfiguredAgents()

	if len(configured) == 0 {
		return a.layoutCenteredAddButton(gtx, pal)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Minimal top bar with compact "+ Add AI" button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, layout.Spacer{}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &a.btnAddAI, func(gtx layout.Context) layout.Dimensions {
						return DrawBadge(gtx, pal.Accent, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return DrawFAIcon(a.theme, gtx, IconPlus, unit.Sp(11), pal.Canvas)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(12), "Add AI")
									lbl.Color = pal.Canvas
									lbl.Font.Weight = 600
									return lbl.Layout(gtx)
								}),
							)
						})
					})
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingNormal)}.Layout),
		// List of configured AI cards
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(a.theme, &a.agentList).Layout(gtx, len(configured), func(gtx layout.Context, i int) layout.Dimensions {
				item := configured[i]
				return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return a.layoutAICard(gtx, item, pal)
				})
			})
		}),
	)
}

// layoutCenteredAddButton displays an elegant, compact button centered on screen.
func (a *App) layoutCenteredAddButton(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Enforce tight compact button dimensions (not full width)
		gtx.Constraints.Min.X = gtx.Dp(unit.Dp(160))
		gtx.Constraints.Max.X = gtx.Dp(unit.Dp(175))
		gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(42))
		gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(42))

		return material.Clickable(gtx, &a.btnAddAI, func(gtx layout.Context) layout.Dimensions {
			bg := pal.Accent
			if a.btnAddAI.Hovered() {
				bg.A = 220
			}

			rect := image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
			paint.FillShape(gtx.Ops, bg, clip.UniformRRect(rect, gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))

			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return DrawFAIcon(a.theme, gtx, IconPlus, unit.Sp(12), pal.Canvas)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(14), "Add AI")
						lbl.Color = pal.Canvas
						lbl.Font.Weight = 600
						return lbl.Layout(gtx)
					}),
				)
			})
		})
	})
}

func (a *App) layoutAICard(gtx layout.Context, item *domain.Agent, pal brand.Palette) layout.Dimensions {
	rmClick := a.getRemoveClick(item.ID)
	if rmClick.Clicked(gtx) {
		a.store.RemoveConfiguredAgent(item.ID)
	}

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
			// Status indicator dot
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				var col = pal.StatusOffline
				switch item.Status {
				case domain.StatusWorking:
					col = pal.StatusWorking
				case domain.StatusOnline:
					col = pal.StatusOnline
				}
				return DrawStatusDot(gtx, unit.Dp(5), col)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
			// Model icon
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				icon := IconMicrochip
				if item.Port > 0 || item.ID == "ollama" || item.ID == "lmstudio" {
					icon = IconServer
				}
				return DrawFAIcon(a.theme, gtx, icon, unit.Sp(16), pal.Accent)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(14)}.Layout),
			// Name and status text
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(14), item.Name)
						lbl.Color = pal.TextPrimary
						lbl.Font.Weight = 600
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						taskText := item.CurrentTask
						if taskText == "" {
							taskText = item.Status.Label()
						}
						lbl := material.Label(a.theme, unit.Sp(12), taskText)
						lbl.Color = pal.TextSecondary
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(14)}.Layout),
			// Remove button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Clickable(gtx, rmClick, func(gtx layout.Context) layout.Dimensions {
					bg := pal.Canvas
					if rmClick.Hovered() {
						bg = pal.SurfaceElevated
					}
					return DrawBadge(gtx, bg, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return DrawFAIcon(a.theme, gtx, IconXmark, unit.Sp(10), pal.TextMuted)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(11), "Remove")
								lbl.Color = pal.TextMuted
								return lbl.Layout(gtx)
							}),
						)
					})
				})
			}),
		)
	})
	call := macro.Stop()

	paint.FillShape(gtx.Ops, pal.Surface, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
	call.Add(gtx.Ops)
	return dims
}

func (a *App) layoutAddAI(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	available := domain.GetAvailableAIs()

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header: "Add AI" + Close button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(16), "Add AI")
					lbl.Color = pal.TextPrimary
					lbl.Font.Weight = 600
					return lbl.Layout(gtx)
				}),
				layout.Flexed(1, layout.Spacer{}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &a.btnCloseAdd, func(gtx layout.Context) layout.Dimensions {
						return DrawBadge(gtx, pal.SurfaceElevated, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return DrawFAIcon(a.theme, gtx, IconXmark, unit.Sp(11), pal.TextSecondary)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(12), "Close")
									lbl.Color = pal.TextSecondary
									return lbl.Layout(gtx)
								}),
							)
						})
					})
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingNormal)}.Layout),
		// List of available AIs (1-click to auto configure and add)
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(a.theme, &a.availableList).Layout(gtx, len(available), func(gtx layout.Context, i int) layout.Dimensions {
				target := available[i]
				click := a.getAIClick(target.ID)

				if click.Clicked(gtx) {
					switch target.ID {
					case "claude":
						_, _ = integrations.ConfigureClaudeDesktop()
					case "antigravity":
						_, _ = integrations.ConfigureAntigravity()
					case "cursor":
						_, _ = integrations.ConfigureCursor()
					case "ollama":
						_, _ = integrations.ProbeOllama()
					case "lmstudio":
						_, _ = integrations.ProbeLMStudio()
					}
					a.store.AddConfiguredAgent(target.ID)
					a.isAddingAI = false
				}

				return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, click, func(gtx layout.Context) layout.Dimensions {
						macro := op.Record(gtx.Ops)
						dims := layout.Inset{
							Top:    unit.Dp(12),
							Bottom: unit.Dp(12),
							Left:   unit.Dp(16),
							Right:  unit.Dp(16),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									icon := IconMicrochip
									if target.ID == "ollama" || target.ID == "lmstudio" {
										icon = IconServer
									}
									return DrawFAIcon(a.theme, gtx, icon, unit.Sp(15), pal.Accent)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(14), target.Name)
									lbl.Color = pal.TextPrimary
									lbl.Font.Weight = 600
									return lbl.Layout(gtx)
								}),
								layout.Flexed(1, layout.Spacer{}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return DrawBadge(gtx, pal.AccentSoft, func(gtx layout.Context) layout.Dimensions {
										return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
											layout.Rigid(func(gtx layout.Context) layout.Dimensions {
												return DrawFAIcon(a.theme, gtx, IconPlus, unit.Sp(10), pal.Accent)
											}),
											layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
											layout.Rigid(func(gtx layout.Context) layout.Dimensions {
												lbl := material.Label(a.theme, unit.Sp(12), "Connect")
												lbl.Color = pal.Accent
												lbl.Font.Weight = 600
												return lbl.Layout(gtx)
											}),
										)
									})
								}),
							)
						})
						call := macro.Stop()

						bg := pal.Surface
						if click.Hovered() {
							bg = pal.SurfaceElevated
						}

						paint.FillShape(gtx.Ops, bg, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
						call.Add(gtx.Ops)
						return dims
					})
				})
			})
		}),
	)
}
