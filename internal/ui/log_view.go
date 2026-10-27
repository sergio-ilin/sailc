package ui

import (
	"fmt"
	"image"
	"strings"

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

func getThreadClick(id string) *widget.Clickable {
	c, ok := threadClicks[id]
	if !ok {
		c = new(widget.Clickable)
		threadClicks[id] = c
	}
	return c
}

func (a *App) layoutLogTab(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	searchQuery := a.searchEditor.Text()
	threads := a.store.SearchThreads(searchQuery)

	if a.btnNewAsk.Clicked(gtx) {
		a.newAskActive = !a.newAskActive
	}

	if a.btnCancelAsk.Clicked(gtx) {
		a.newAskActive = false
	}

	if a.btnSubmitAsk.Clicked(gtx) {
		title := strings.TrimSpace(a.askTitleEdit.Text())
		content := strings.TrimSpace(a.askBodyEdit.Text())
		author := strings.TrimSpace(a.askAuthorEdit.Text())
		if author == "" {
			author = "Operator"
		}
		if title != "" && content != "" {
			thr := a.store.CreateThread(author, title, content, []string{"general-context"})
			a.selectedThrID = thr.ID
			a.newAskActive = false
			a.askTitleEdit.SetText("")
			a.askBodyEdit.SetText("")
		}
	}

	if a.btnSubmitReply.Clicked(gtx) {
		replyText := strings.TrimSpace(a.replyBodyEdit.Text())
		author := strings.TrimSpace(a.replyAuthorEdit.Text())
		if author == "" {
			author = "Operator"
		}
		if replyText != "" && a.selectedThrID != "" {
			_, _ = a.store.AddResponse(a.selectedThrID, author, replyText)
			a.replyBodyEdit.SetText("")
		}
	}

	if a.selectedThrID == "" && len(threads) > 0 {
		a.selectedThrID = threads[0].ID
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Control & search bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutLogControlBar(gtx, pal)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingNormal)}.Layout),
		// New question form (collapsible)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !a.newAskActive {
				return layout.Dimensions{}
			}
			return a.layoutNewAskForm(gtx, pal)
		}),
		// Two-column master-detail layout
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Left pane: threads list
				layout.Flexed(0.45, func(gtx layout.Context) layout.Dimensions {
					return a.layoutThreadsList(gtx, threads, pal)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(brand.PaddingNormal)}.Layout),
				// Right pane: thread detail & replies
				layout.Flexed(0.55, func(gtx layout.Context) layout.Dimensions {
					return a.layoutThreadDetail(gtx, pal)
				}),
			)
		}),
	)
}

func (a *App) layoutLogControlBar(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		// Search input
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dims := layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(10),
				Left:   unit.Dp(14),
				Right:  unit.Dp(14),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(a.theme, &a.searchEditor, "Search context & knowledge base before asking...")
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
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		// New question button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btnText := "+ Ask Question in Hub"
			if a.newAskActive {
				btnText = "Hide Form"
			}
			return material.Clickable(gtx, &a.btnNewAsk, func(gtx layout.Context) layout.Dimensions {
				macro := op.Record(gtx.Ops)
				dims := layout.Inset{
					Top:    unit.Dp(10),
					Bottom: unit.Dp(10),
					Left:   unit.Dp(16),
					Right:  unit.Dp(16),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(13), btnText)
					lbl.Color = pal.TextPrimary
					lbl.Font.Weight = 600
					return lbl.Layout(gtx)
				})
				call := macro.Stop()

				bg := pal.SurfaceElevated
				if a.btnNewAsk.Hovered() {
					bg = pal.AccentSoft
				}
				paint.FillShape(gtx.Ops, bg, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
				call.Add(gtx.Ops)
				return dims
			})
		}),
	)
}

func (a *App) layoutNewAskForm(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(14),
		Bottom: unit.Dp(14),
		Left:   unit.Dp(16),
		Right:  unit.Dp(16),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.theme, unit.Sp(14), "Publish New Inquiry to Shared Log")
				lbl.Color = pal.TextPrimary
				lbl.Font.Weight = 600
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Author and Topic
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
						return a.layoutSimpleInput(gtx, &a.askAuthorEdit, "AI Agent / Author name", pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
						return a.layoutSimpleInput(gtx, &a.askTitleEdit, "Task summary / Question topic", pal)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Body
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutSimpleInput(gtx, &a.askBodyEdit, "Detailed context, code snippet, or exact problem definition...", pal)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			// Action buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnSubmitAsk, func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.Accent, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(13), "Publish to Log")
								lbl.Color = pal.Canvas
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
							})
						})
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnCancelAsk, func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.SurfaceElevated, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(13), "Cancel")
								lbl.Color = pal.TextSecondary
								return lbl.Layout(gtx)
							})
						})
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

func (a *App) layoutSimpleInput(gtx layout.Context, ed *widget.Editor, placeholder string, pal brand.Palette) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(8),
		Bottom: unit.Dp(8),
		Left:   unit.Dp(12),
		Right:  unit.Dp(12),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		mEd := material.Editor(a.theme, ed, placeholder)
		mEd.Color = pal.TextPrimary
		mEd.HintColor = pal.TextMuted
		mEd.TextSize = unit.Sp(12)
		return mEd.Layout(gtx)
	})
	call := macro.Stop()

	paint.FillShape(gtx.Ops, pal.Canvas, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusSmall))).Op(gtx.Ops))
	call.Add(gtx.Ops)
	return dims
}

func (a *App) layoutThreadsList(gtx layout.Context, threads []*domain.Thread, pal brand.Palette) layout.Dimensions {
	if len(threads) == 0 {
		return layout.Inset{Top: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(a.theme, unit.Sp(13), "No discussions found. Click '+ Ask Question in Hub' to begin.")
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
						// Title and status
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
						// Content snippet
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							snippet := thr.Content
							if len(snippet) > 100 {
								snippet = snippet[:100] + "..."
							}
							lbl := material.Label(a.theme, unit.Sp(12), snippet)
							lbl.Color = pal.TextSecondary
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						// Meta (responses, status, time)
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

func (a *App) layoutThreadDetail(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	thr, ok := a.store.GetThreadByID(a.selectedThrID)
	if !ok || thr == nil {
		return layout.Inset{Top: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(a.theme, unit.Sp(13), "Select a discussion thread from the left to view details.")
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
			// Author & timestamp
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				info := fmt.Sprintf("Author: %s  •  Published: %s", thr.AuthorAgent, thr.CreatedAt.Format("Jan 02 15:04:05"))
				lbl := material.Label(a.theme, unit.Sp(12), info)
				lbl.Color = pal.TextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			// Question / Context box
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				macro := op.Record(gtx.Ops)
				dims := layout.Inset{
					Top:    unit.Dp(10),
					Bottom: unit.Dp(10),
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
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
				lbl := material.Label(a.theme, unit.Sp(13), fmt.Sprintf("Responses & Solutions from other AI agents (%d):", len(thr.Responses)))
				lbl.Color = pal.TextSecondary
				lbl.Font.Weight = 600
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Responses list
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if len(thr.Responses) == 0 {
					return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(12), "No responses yet. Other AI agents can reply using the hub_answer MCP tool.")
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
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			// Submit reply input bar
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(0.25, func(gtx layout.Context) layout.Dimensions {
						return a.layoutSimpleInput(gtx, &a.replyAuthorEdit, "Author name", pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
						return a.layoutSimpleInput(gtx, &a.replyBodyEdit, "Write solution or context hint...", pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnSubmitReply, func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.Accent, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(12), "Reply")
								lbl.Color = pal.Canvas
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
							})
						})
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
