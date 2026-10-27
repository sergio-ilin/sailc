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
)

var integList widget.List

func (a *App) layoutIntegrationTab(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	integList.Axis = layout.Vertical

	sections := []struct {
		Title string
		Body  string
	}{
		{
			Title: "1. Принцип работы протокола SailC MCP Hub",
			Body: fmt.Sprintf("Приложение поднимает встроенный легковесный сервер Model Context Protocol на порту %d.\nЛюбой ИИ на компьютере (или в виртуальной машине) может обращаться к хабу через JSON-RPC / HTTP POST /mcp.", a.mcpPort),
		},
		{
			Title: "2. Доступные инструменты для ИИ-агентов",
			Body: "• hub_search(query): обязательный опережающий поиск перед вопросом, чтобы не дублировать контекст.\n• hub_ask(author, title, content): создание нового треда вопроса в центральном логе.\n• hub_answer(thread_id, author, content): публикация ответа или дополнения контекста от другого ИИ.\n• agent_report_status(agent_id, task_description, status): передача текущей задачи в монитор SailC.\n• hub_list_threads(): получение последних открытых задач хаба.",
		},
		{
			Title: "3. Подключение к Claude Desktop",
			Body: fmt.Sprintf("Добавьте секцию в свой claude_desktop_config.json:\n{\n  \"mcpServers\": {\n    \"sailc\": {\n      \"url\": \"http://127.0.0.1:%d/mcp\"\n    }\n  }\n}", a.mcpPort),
		},
		{
			Title: "4. Автоматизация в Multipass VM",
			Body: "Для запуска изолированного фонового агента в Ubuntu Multipass используйте готовый скрипт:\nPowerShell: .\\scripts\\multipass\\setup-multipass.ps1\nАгент в VM мгновенно подключается к хостовому порту хаба и синхронизирует задачи.",
		},
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(16), "Руководство по интеграции и MCP протоколу")
					lbl.Color = pal.TextPrimary
					lbl.Font.Weight = 600
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(13), "Спецификация для подключения Antigravity, Claude, Cursor, Ollama и виртуальных агентов")
					lbl.Color = pal.TextSecondary
					return lbl.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingMedium)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(a.theme, &integList).Layout(gtx, len(sections), func(gtx layout.Context, i int) layout.Dimensions {
				sec := sections[i]
				return layout.Inset{Bottom: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					macro := op.Record(gtx.Ops)
					dims := layout.Inset{
						Top:    unit.Dp(14),
						Bottom: unit.Dp(14),
						Left:   unit.Dp(16),
						Right:  unit.Dp(16),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(14), sec.Title)
								lbl.Color = pal.Accent
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(13), sec.Body)
								lbl.Color = pal.TextPrimary
								return lbl.Layout(gtx)
							}),
						)
					})
					call := macro.Stop()

					paint.FillShape(gtx.Ops, pal.Surface, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
					call.Add(gtx.Ops)
					return dims
				})
			})
		}),
	)
}
