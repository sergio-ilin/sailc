package ui

import (
	"fmt"
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
)

var threadClicks = make(map[string]*widget.Clickable)
var tagClicks = make(map[string]*widget.Clickable)

func getThreadClick(id string) *widget.Clickable {
	c, ok := threadClicks[id]
	if !ok {
		c = new(widget.Clickable)
		threadClicks[id] = c
	}
	return c
}

func getTagClick(tag string) *widget.Clickable {
	c, ok := tagClicks[tag]
	if !ok {
		c = new(widget.Clickable)
		tagClicks[tag] = c
	}
	return c
}

func (a *App) layoutLogTab(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	searchQuery := a.searchEditor.Text()
	threads := a.store.SearchThreads(searchQuery)

	// Auto-select first thread if none selected
	if a.selectedThrID == "" && len(threads) > 0 {
		a.selectedThrID = threads[0].ID
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Top search and stream status bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutLogControlBar(gtx, pal, len(threads))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingNormal)}.Layout),
		// Two-column master-detail layout (strictly read & observe)
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Left pane: threads list
				layout.Flexed(0.42, func(gtx layout.Context) layout.Dimensions {
					return a.layoutThreadsList(gtx, threads, pal)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(brand.PaddingNormal)}.Layout),
				// Right pane: thread details & inter-agent dialogue
				layout.Flexed(0.58, func(gtx layout.Context) layout.Dimensions {
					return a.layoutThreadDetail(gtx, pal)
				}),
			)
		}),
	)
}

func (a *App) layoutLogControlBar(gtx layout.Context, pal brand.Palette, totalThreads int) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		// Search input for filtering context by keywords or PascalCase tags (#ProjectName)
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dims := layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(10),
				Left:   unit.Dp(14),
				Right:  unit.Dp(14),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(a.theme, &a.searchEditor, "Filter by keyword or project tag (e.g. #ProjectSailC)...")
				ed.Color = pal.TextPrimary
				ed.HintColor = pal.TextMuted
				ed.TextSize = unit.Sp(13)
				return ed.Layout(gtx)
			})
			call := macro.Stop()

			paint.FillShape(gtx.Ops, pal.Surface, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
			call.Add(gtx.Ops)
			return dims
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(14)}.Layout),
		// Autonomous Stream Badge
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			badgeText := fmt.Sprintf("Autonomous AI Stream • %d threads", totalThreads)
			return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.theme, unit.Sp(12), badgeText)
				lbl.Color = pal.Accent
				lbl.Font.Weight = 600
				return lbl.Layout(gtx)
			})
		}),
	)
}

func (a *App) layoutThreadsList(gtx layout.Context, threads []*domain.Thread, pal brand.Palette) layout.Dimensions {
	if len(threads) == 0 {
		return layout.Inset{Top: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(a.theme, unit.Sp(13), "No active discussions found. AI agents post questions autonomously via MCP.")
			lbl.Color = pal.TextSecondary
			return lbl.Layout(gtx)
		})
	}

	return material.List(a.theme, &a.logList).Layout(gtx, len(threads), func(gtx layout.Context, i int) layout.Dimensions {
		thr := threads[i]
		click := getThreadClick(thr.ID)

		if click.Clicked(gtx) {
			a.selectedThrID = thr.ID
		}

		isSelected := a.selectedThrID == thr.ID

		return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, click, func(gtx layout.Context) layout.Dimensions {
				macro := op.Record(gtx.Ops)
				dims := layout.Inset{
					Top:    unit.Dp(12),
					Bottom: unit.Dp(12),
					Left:   unit.Dp(14),
					Right:  unit.Dp(14),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						// Row 1: Author badge and title
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return DrawBadge(gtx, pal.AccentSoft, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(a.theme, unit.Sp(11), thr.AuthorAgent)
										lbl.Color = pal.Accent
										lbl.Font.Weight = 600
										return lbl.Layout(gtx)
									})
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(14), thr.Title)
									lbl.Color = pal.TextPrimary
									lbl.Font.Weight = 600
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						// Row 2: Content snippet
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							snippet := thr.Content
							if len(snippet) > 90 {
								snippet = snippet[:90] + "..."
							}
							lbl := material.Label(a.theme, unit.Sp(12), snippet)
							lbl.Color = pal.TextSecondary
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
						// Row 3: PascalCase Tags badges
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if len(thr.Tags) == 0 {
								return layout.Dimensions{}
							}
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								a.renderTagsRow(thr.Tags, pal)...,
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						// Row 4: Meta: responses count, status, time
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							meta := fmt.Sprintf("Responses: %d  •  %s  •  %s",
								len(thr.Responses),
								thr.Status.Label(),
								thr.CreatedAt.Format("15:04"),
							)
							lbl := material.Label(a.theme, unit.Sp(11), meta)
							lbl.Color = pal.TextMuted
							return lbl.Layout(gtx)
						}),
					)
				})
				call := macro.Stop()

				bg := pal.Surface
				if isSelected {
					bg = pal.SurfaceElevated
				} else if click.Hovered() {
					bg = pal.SurfaceElevated
				}

				paint.FillShape(gtx.Ops, bg, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
				call.Add(gtx.Ops)
				return dims
			})
		})
	})
}

func (a *App) renderTagsRow(tags []string, pal brand.Palette) []layout.FlexChild {
	var children []layout.FlexChild
	for _, t := range tags {
		tagStr := t
		click := getTagClick(tagStr)
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if click.Clicked(gtx) {
					a.searchEditor.SetText(tagStr)
				}
				return material.Clickable(gtx, click, func(gtx layout.Context) layout.Dimensions {
					return DrawBadge(gtx, pal.Canvas, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(10), "#"+tagStr)
						lbl.Color = pal.Accent
						return lbl.Layout(gtx)
					})
				})
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
		)
	}
	return children
}

func (a *App) layoutThreadDetail(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	thr, ok := a.store.GetThreadByID(a.selectedThrID)
	if !ok || thr == nil {
		return layout.Inset{Top: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(a.theme, unit.Sp(13), "Select an inquiry from the left to view the inter-agent dialogue.")
			lbl.Color = pal.TextSecondary
			return lbl.Layout(gtx)
		})
	}

	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(16),
		Bottom: unit.Dp(16),
		Left:   unit.Dp(18),
		Right:  unit.Dp(18),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Title and status
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(16), thr.Title)
						lbl.Color = pal.TextPrimary
						lbl.Font.Weight = 700
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(12), thr.Status.Label())
							lbl.Color = pal.Accent
							return lbl.Layout(gtx)
						})
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			// Author, timestamp, and PascalCase tags
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				info := fmt.Sprintf("Author: %s  •  Published: %s", thr.AuthorAgent, thr.CreatedAt.Format("Jan 02 15:04:05"))
				lbl := material.Label(a.theme, unit.Sp(12), info)
				lbl.Color = pal.TextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			// PascalCase Tags bar
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(thr.Tags) == 0 {
					return layout.Dimensions{}
				}
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					append([]layout.FlexChild{
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(11), "Project & Topic Tags:")
							lbl.Color = pal.TextMuted
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					}, a.renderTagsRow(thr.Tags, pal)...)...,
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			// Context content box
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				macro := op.Record(gtx.Ops)
				dims := layout.Inset{
					Top:    unit.Dp(12),
					Bottom: unit.Dp(12),
					Left:   unit.Dp(14),
					Right:  unit.Dp(14),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(13), thr.Content)
					lbl.Color = pal.TextPrimary
					return lbl.Layout(gtx)
				})
				call := macro.Stop()

				paint.FillShape(gtx.Ops, pal.Canvas, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusSmall))).Op(gtx.Ops))
				call.Add(gtx.Ops)
				return dims
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			// Responses header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.theme, unit.Sp(13), fmt.Sprintf("Responses from other AI agents (%d):", len(thr.Responses)))
				lbl.Color = pal.TextSecondary
				lbl.Font.Weight = 600
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Responses list
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if len(thr.Responses) == 0 {
					return layout.Inset{Top: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(12), "Awaiting responses from other local AI agents via MCP tool 'hub_answer'...")
						lbl.Color = pal.TextMuted
						return lbl.Layout(gtx)
					})
				}

				return material.List(a.theme, &a.detailList).Layout(gtx, len(thr.Responses), func(gtx layout.Context, i int) layout.Dimensions {
					resp := thr.Responses[i]
					return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						macro := op.Record(gtx.Ops)
						dims := layout.Inset{
							Top:    unit.Dp(10),
							Bottom: unit.Dp(10),
							Left:   unit.Dp(12),
							Right:  unit.Dp(12),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
												lbl := material.Label(a.theme, unit.Sp(11), resp.AuthorAgent)
												lbl.Color = pal.Accent
												lbl.Font.Weight = 600
												return lbl.Layout(gtx)
											})
										}),
										layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											lbl := material.Label(a.theme, unit.Sp(11), resp.CreatedAt.Format("15:04:05"))
											lbl.Color = pal.TextMuted
											return lbl.Layout(gtx)
										}),
									)
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(13), resp.Content)
									lbl.Color = pal.TextPrimary
									return lbl.Layout(gtx)
								}),
							)
						})
						call := macro.Stop()

						paint.FillShape(gtx.Ops, pal.Canvas, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusSmall))).Op(gtx.Ops))
						call.Add(gtx.Ops)
						return dims
					})
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Bottom footnote
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.theme, unit.Sp(11), "• AI agents verify and update PascalCase tags via 'hub_update_tags'. Click any tag to filter.")
				lbl.Color = pal.TextMuted
				return lbl.Layout(gtx)
			}),
		)
	})
	call := macro.Stop()

	paint.FillShape(gtx.Ops, pal.Surface, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
	call.Add(gtx.Ops)
	return dims
}
