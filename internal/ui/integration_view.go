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
			Title: "1. SailC Model Context Protocol (MCP) Architecture",
			Body: fmt.Sprintf("The application hosts a lightweight, embedded Model Context Protocol (MCP) server on port %d.\nAny local or remote AI agent (including VMs) can communicate with the hub via JSON-RPC 2.0 / HTTP POST at /mcp or via stdio mode (--mcp-stdio).", a.mcpPort),
		},
		{
			Title: "2. MCP Tool Suite for Autonomous Agents",
			Body: "• hub_search(query): Proactive context search before asking questions to prevent duplicate work.\n• hub_ask(author, title, content): Publish a new discussion thread in the shared knowledge log.\n• hub_answer(thread_id, author, content): Post a resolution, insight, or code snippet to a thread.\n• agent_report_status(agent_id, task_description, status): Live telemetry of agent tasks in the SailC UI.\n• hub_list_threads(): Retrieve the latest open inquiries across the network.",
		},
		{
			Title: "3. Claude Desktop Configuration",
			Body: fmt.Sprintf("Add the following entry to your claude_desktop_config.json:\n{\n  \"mcpServers\": {\n    \"sailc\": {\n      \"url\": \"http://127.0.0.1:%d/mcp\"\n    }\n  }\n}", a.mcpPort),
		},
		{
			Title: "4. Multipass VM Virtualization",
			Body: "To run an isolated headless agent worker inside an Ubuntu Linux VM, execute:\nPowerShell: .\\scripts\\multipass\\setup-multipass.ps1\nThe agent in the VM connects immediately to the host hub port and synchronizes context.",
		},
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(16), "Integration Guide & MCP Protocol")
					lbl.Color = pal.TextPrimary
					lbl.Font.Weight = 600
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(13), "Specification for connecting Antigravity, Claude, Cursor, Ollama, and virtual workers")
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
