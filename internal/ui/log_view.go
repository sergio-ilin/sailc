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
	// Обработка поискового запроса
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
			author = "Оператор"
		}
		if title != "" && content != "" {
			thr := a.store.CreateThread(author, title, content, []string{"общий-контекст"})
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
			author = "Оператор"
		}
		if replyText != "" && a.selectedThrID != "" {
			_, _ = a.store.AddResponse(a.selectedThrID, author, replyText)
			a.replyBodyEdit.SetText("")
		}
	}

	// Если тред не выбран, выберем первый доступный
	if a.selectedThrID == "" && len(threads) > 0 {
		a.selectedThrID = threads[0].ID
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Панель управления поиском и созданием вопроса
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutLogControlBar(gtx, pal)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingNormal)}.Layout),
		// Форма добавления нового вопроса (если развёрнута)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !a.newAskActive {
				return layout.Dimensions{}
			}
			return a.layoutNewAskForm(gtx, pal)
		}),
		// Двухколоночный интерфейс: Список тредов слева, Детали и ответы справа
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Левая колонка: список тредов
				layout.Flexed(0.45, func(gtx layout.Context) layout.Dimensions {
					return a.layoutThreadsList(gtx, threads, pal)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(brand.PaddingNormal)}.Layout),
				// Правая колонка: детали выбранного треда и переписка ИИ
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
		// Поле поиска по контексту
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dims := layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(10),
				Left:   unit.Dp(14),
				Right:  unit.Dp(14),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(a.theme, &a.searchEditor, "Поиск по контексту и логам (агенты ищут здесь перед вопросом)...")
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
		// Кнопка "Задать вопрос"
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btnText := "+ Задать вопрос в хаб"
			if a.newAskActive {
				btnText = "Скрыть форму"
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
				lbl := material.Label(a.theme, unit.Sp(14), "Публикация нового вопроса в центральный лог")
				lbl.Color = pal.TextPrimary
				lbl.Font.Weight = 600
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Автор и тема
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
						return a.layoutSimpleInput(gtx, &a.askAuthorEdit, "Имя ИИ / автора", pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
						return a.layoutSimpleInput(gtx, &a.askTitleEdit, "Краткая суть задачи / вопроса", pal)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Подробный контекст
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutSimpleInput(gtx, &a.askBodyEdit, "Подробный контекст, фрагменты кода или точная постановка проблемы...", pal)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			// Кнопки
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnSubmitAsk, func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.Accent, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(13), "Опубликовать в лог")
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
								lbl := material.Label(a.theme, unit.Sp(13), "Отмена")
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
			lbl := material.Label(a.theme, unit.Sp(13), "Обсуждений не найдено. Нажмите «Задать вопрос», чтобы начать.")
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
						// Заголовок вопроса и статус
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
						// Сниппет содержания
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
						// Мета-информация (кол-во ответов и время)
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							meta := fmt.Sprintf("Ответов: %d  •  %s  •  %s",
								len(thr.Responses),
								thr.Status.RussianLabel(),
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
			lbl := material.Label(a.theme, unit.Sp(13), "Выберите обсуждение слева для просмотра деталей.")
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
			// Заголовок треда
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
							lbl := material.Label(a.theme, unit.Sp(12), thr.Status.RussianLabel())
							lbl.Color = pal.Accent
							return lbl.Layout(gtx)
						})
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			// Автор и время
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				info := fmt.Sprintf("Автор: %s  •  Опубликовано: %s", thr.AuthorAgent, thr.CreatedAt.Format("02.01 15:04:05"))
				lbl := material.Label(a.theme, unit.Sp(12), info)
				lbl.Color = pal.TextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			// Текст вопроса / контекст
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
			// Заголовок блока ответов
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.theme, unit.Sp(13), fmt.Sprintf("Ответы и решения от других ИИ (%d):", len(thr.Responses)))
				lbl.Color = pal.TextSecondary
				lbl.Font.Weight = 600
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Список ответов
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if len(thr.Responses) == 0 {
					return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(12), "Другие ИИ ещё не предоставили ответ. Они могут ответить через MCP метод hub_answer.")
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
			// Поле отправки ответа
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(0.25, func(gtx layout.Context) layout.Dimensions {
						return a.layoutSimpleInput(gtx, &a.replyAuthorEdit, "Автор ответа", pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
						return a.layoutSimpleInput(gtx, &a.replyBodyEdit, "Написать ответ или подсказку...", pal)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnSubmitReply, func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.Accent, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(12), "Ответить")
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
