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
	"github.com/sergio-ilin/sailc/internal/integrations"
)

func (a *App) getSelectionClick(id string) *widget.Clickable {
	c, ok := a.aiSelectionClicks[id]
	if !ok {
		c = new(widget.Clickable)
		a.aiSelectionClicks[id] = c
	}
	return c
}

func (a *App) getRemoveClick(id string) *widget.Clickable {
	c, ok := a.removeAgentClicks[id]
	if !ok {
		c = new(widget.Clickable)
		a.removeAgentClicks[id] = c
	}
	return c
}

func (a *App) layoutAgentsTab(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	if a.btnAddAgent.Clicked(gtx) {
		a.isAddingAgent = true
		a.wizardStep = 0
		a.wizardResultMsg = ""
		a.wizardSuccess = false
	}

	if a.btnCancelWizard.Clicked(gtx) {
		a.isAddingAgent = false
		a.wizardStep = 0
	}

	if a.isAddingAgent {
		return a.layoutWizard(gtx, pal)
	}

	configuredAgents := a.store.ListConfiguredAgents()

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header row with Title and "+ Connect AI Agent" button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(16), "Connected AI Agents")
							lbl.Color = pal.TextPrimary
							lbl.Font.Weight = 600
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(13), "Active assistants sharing context and collaborating through the SailC Knowledge Hub")
							lbl.Color = pal.TextSecondary
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &a.btnAddAgent, func(gtx layout.Context) layout.Dimensions {
						return DrawBadge(gtx, pal.Accent, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(13), "+ Connect AI Agent")
							lbl.Color = pal.Canvas
							lbl.Font.Weight = 600
							return lbl.Layout(gtx)
						})
					})
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingMedium)}.Layout),
		// List or Empty State
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if len(configuredAgents) == 0 {
				return a.layoutEmptyAgentsState(gtx, pal)
			}
			return material.List(a.theme, &a.agentList).Layout(gtx, len(configuredAgents), func(gtx layout.Context, i int) layout.Dimensions {
				agent := configuredAgents[i]
				return layout.Inset{
					Bottom: unit.Dp(12),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return a.layoutConfiguredAgentCard(gtx, agent, pal)
				})
			})
		}),
	)
}

func (a *App) layoutEmptyAgentsState(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(36),
		Bottom: unit.Dp(36),
		Left:   unit.Dp(24),
		Right:  unit.Dp(24),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawSailLogo(gtx, unit.Dp(48), pal.Accent)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.theme, unit.Sp(18), "No AI Assistants Connected Yet")
				lbl.Color = pal.TextPrimary
				lbl.Font.Weight = 700
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.theme, unit.Sp(13), "Connect Claude Desktop, Cursor, Antigravity, or local LLMs to enable autonomous inter-agent context sharing.")
				lbl.Color = pal.TextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Clickable(gtx, &a.btnAddAgent, func(gtx layout.Context) layout.Dimensions {
					macroBtn := op.Record(gtx.Ops)
					dimsBtn := layout.Inset{
						Top:    unit.Dp(10),
						Bottom: unit.Dp(10),
						Left:   unit.Dp(20),
						Right:  unit.Dp(20),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(14), "+ Connect Your First AI Agent")
						lbl.Color = pal.Canvas
						lbl.Font.Weight = 600
						return lbl.Layout(gtx)
					})
					callBtn := macroBtn.Stop()

					paint.FillShape(gtx.Ops, pal.Accent, clip.UniformRRect(image.Rect(0, 0, dimsBtn.Size.X, dimsBtn.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
					callBtn.Add(gtx.Ops)
					return dimsBtn
				})
			}),
		)
	})
	call := macro.Stop()

	paint.FillShape(gtx.Ops, pal.Surface, clip.UniformRRect(image.Rect(0, 0, dims.Size.X, dims.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
	call.Add(gtx.Ops)
	return dims
}

func (a *App) layoutConfiguredAgentCard(gtx layout.Context, agent *domain.Agent, pal brand.Palette) layout.Dimensions {
	rmClick := a.getRemoveClick(agent.ID)
	if rmClick.Clicked(gtx) {
		a.store.RemoveConfiguredAgent(agent.ID)
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
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(15), agent.Name)
								lbl.Color = pal.TextPrimary
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(11), agent.Type)
									lbl.Color = pal.TextSecondary
									return lbl.Layout(gtx)
								})
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
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
			// Right column: Status label, telemetry, and Disconnect button
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
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, rmClick, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(11), "Disconnect")
							lbl.Color = pal.TextMuted
							return lbl.Layout(gtx)
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

// -------------------------------------------------------------
// Interactive Step-by-Step Connection Wizard
// -------------------------------------------------------------
func (a *App) layoutWizard(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Wizard Top Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					stepTitle := "Step 1: Select AI Assistant to Connect"
					if a.wizardStep == 1 {
						stepTitle = fmt.Sprintf("Step 2: Configure %s", a.selectedAI.Name)
					} else if a.wizardStep == 2 {
						stepTitle = fmt.Sprintf("Step 3: Verify & Activate %s", a.selectedAI.Name)
					}

					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(16), stepTitle)
							lbl.Color = pal.TextPrimary
							lbl.Font.Weight = 600
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(13), "Step-by-step setup with 1-click automatic config injection")
							lbl.Color = pal.TextSecondary
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &a.btnCancelWizard, func(gtx layout.Context) layout.Dimensions {
						return DrawBadge(gtx, pal.SurfaceElevated, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(12), "Close Wizard")
							lbl.Color = pal.TextSecondary
							return lbl.Layout(gtx)
						})
					})
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(brand.PaddingMedium)}.Layout),
		// Wizard Step Content
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			switch a.wizardStep {
			case 0:
				return a.layoutWizardStep0_SelectAI(gtx, pal)
			case 1:
				return a.layoutWizardStep1_Configure(gtx, pal)
			case 2:
				return a.layoutWizardStep2_Verify(gtx, pal)
			default:
				return layout.Dimensions{}
			}
		}),
	)
}

// Step 0: Choose an AI Assistant
func (a *App) layoutWizardStep0_SelectAI(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	available := domain.GetAvailableAIs()

	return material.List(a.theme, &a.availableList).Layout(gtx, len(available), func(gtx layout.Context, i int) layout.Dimensions {
		target := available[i]
		click := a.getSelectionClick(target.ID)

		if click.Clicked(gtx) {
			a.selectedAI = target
			a.wizardStep = 1
			a.wizardResultMsg = ""
			a.wizardSuccess = false
		}

		return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, click, func(gtx layout.Context) layout.Dimensions {
				macro := op.Record(gtx.Ops)
				dims := layout.Inset{
					Top:    unit.Dp(14),
					Bottom: unit.Dp(14),
					Left:   unit.Dp(16),
					Right:  unit.Dp(16),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											lbl := material.Label(a.theme, unit.Sp(15), target.Name)
											lbl.Color = pal.TextPrimary
											lbl.Font.Weight = 600
											return lbl.Layout(gtx)
										}),
										layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return DrawBadge(gtx, pal.StatusPillBg, func(gtx layout.Context) layout.Dimensions {
												lbl := material.Label(a.theme, unit.Sp(11), target.Provider)
												lbl.Color = pal.TextSecondary
												return lbl.Layout(gtx)
											})
										}),
										layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											badgeText := "Manual Config"
											badgeBg := pal.StatusPillBg
											badgeColor := pal.TextMuted
											if target.AutoSupport {
												badgeText = "1-Click Auto Config"
												badgeBg = pal.AccentSoft
												badgeColor = pal.Accent
											}
											return DrawBadge(gtx, badgeBg, func(gtx layout.Context) layout.Dimensions {
												lbl := material.Label(a.theme, unit.Sp(11), badgeText)
												lbl.Color = badgeColor
												lbl.Font.Weight = 600
												return lbl.Layout(gtx)
											})
										}),
									)
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(a.theme, unit.Sp(12), target.Description)
									lbl.Color = pal.TextSecondary
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.SurfaceElevated, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(12), "Select →")
								lbl.Color = pal.Accent
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
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
}

// Step 1: Configuration & Auto-Integration
func (a *App) layoutWizardStep1_Configure(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	if a.btnAutoIntegrate.Clicked(gtx) {
		var msg string
		var err error

		switch a.selectedAI.ID {
		case "claude":
			msg, err = integrations.ConfigureClaudeDesktop()
		case "antigravity":
			msg, err = integrations.ConfigureAntigravity()
		case "cursor":
			msg, err = integrations.ConfigureCursor()
		case "ollama":
			msg, err = integrations.ProbeOllama()
		case "lmstudio":
			msg, err = integrations.ProbeLMStudio()
		default:
			msg = "Manual configuration required for this assistant."
		}

		if err != nil {
			a.wizardResultMsg = fmt.Sprintf("Error: %v", err)
			a.wizardSuccess = false
		} else {
			a.wizardResultMsg = msg
			a.wizardSuccess = true
		}
	}

	if a.btnNextStep.Clicked(gtx) {
		a.wizardStep = 2
	}

	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(20),
		Bottom: unit.Dp(20),
		Left:   unit.Dp(22),
		Right:  unit.Dp(22),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Target AI header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(16), a.selectedAI.Name)
						lbl.Color = pal.TextPrimary
						lbl.Font.Weight = 700
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return DrawBadge(gtx, pal.AccentSoft, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.theme, unit.Sp(11), a.selectedAI.Provider)
							lbl.Color = pal.Accent
							lbl.Font.Weight = 600
							return lbl.Layout(gtx)
						})
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
			// Explanation
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				desc := fmt.Sprintf("SailC will inject the required Model Context Protocol (MCP) server configuration directly into %s.\nOnce connected, %s will autonomously query SailC for context (hub_search) and is obligated to answer open inquiries (hub_answer).", a.selectedAI.Name, a.selectedAI.Name)
				lbl := material.Label(a.theme, unit.Sp(13), desc)
				lbl.Color = pal.TextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(18)}.Layout),
			// 1-Click Auto Configure Button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btnText := "⚡ 1-Click Auto-Configure " + a.selectedAI.Name
				if !a.selectedAI.AutoSupport {
					btnText = "🔍 Verify & Probe " + a.selectedAI.Name
				}

				return material.Clickable(gtx, &a.btnAutoIntegrate, func(gtx layout.Context) layout.Dimensions {
					macroBtn := op.Record(gtx.Ops)
					dimsBtn := layout.Inset{
						Top:    unit.Dp(10),
						Bottom: unit.Dp(10),
						Left:   unit.Dp(18),
						Right:  unit.Dp(18),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(13), btnText)
						lbl.Color = pal.Canvas
						lbl.Font.Weight = 600
						return lbl.Layout(gtx)
					})
					callBtn := macroBtn.Stop()

					paint.FillShape(gtx.Ops, pal.Accent, clip.UniformRRect(image.Rect(0, 0, dimsBtn.Size.X, dimsBtn.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
					callBtn.Add(gtx.Ops)
					return dimsBtn
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
			// Status / Result notification box
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if a.wizardResultMsg == "" {
					return layout.Dimensions{}
				}
				macroBox := op.Record(gtx.Ops)
				dimsBox := layout.Inset{
					Top:    unit.Dp(10),
					Bottom: unit.Dp(10),
					Left:   unit.Dp(14),
					Right:  unit.Dp(14),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(12), a.wizardResultMsg)
					if a.wizardSuccess {
						lbl.Color = pal.StatusOnline
					} else {
						lbl.Color = pal.StatusWorking
					}
					return lbl.Layout(gtx)
				})
				callBox := macroBox.Stop()

				paint.FillShape(gtx.Ops, pal.Canvas, clip.UniformRRect(image.Rect(0, 0, dimsBox.Size.X, dimsBox.Size.Y), gtx.Dp(unit.Dp(brand.RadiusSmall))).Op(gtx.Ops))
				callBox.Add(gtx.Ops)
				return dimsBox
			}),
			layout.Flexed(1, layout.Spacer{}.Layout),
			// Bottom navigation
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnCancelWizard, func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.SurfaceElevated, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(12), "← Back to AI List")
								lbl.Color = pal.TextSecondary
								return lbl.Layout(gtx)
							})
						})
					}),
					layout.Flexed(1, layout.Spacer{}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnNextStep, func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.Accent, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(12), "Next: Verification & Activation →")
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

// Step 2: Verification & Activation
func (a *App) layoutWizardStep2_Verify(gtx layout.Context, pal brand.Palette) layout.Dimensions {
	if a.btnFinishWizard.Clicked(gtx) {
		a.store.AddConfiguredAgent(a.selectedAI.ID)
		a.isAddingAgent = false
		a.wizardStep = 0
	}

	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(20),
		Bottom: unit.Dp(20),
		Left:   unit.Dp(22),
		Right:  unit.Dp(22),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.theme, unit.Sp(16), "Activation & Mandatory Cooperation Mandate")
				lbl.Color = pal.TextPrimary
				lbl.Font.Weight = 700
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				info := fmt.Sprintf("Congratulations! %s is configured to communicate with the SailC Hub.\n\nMandate: When active, %s will monitor the central log and is strictly obligated to supply solutions (hub_answer) to unblock peer AI models on this computer.", a.selectedAI.Name, a.selectedAI.Name)
				lbl := material.Label(a.theme, unit.Sp(13), info)
				lbl.Color = pal.TextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DrawBadge(gtx, pal.Canvas, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.theme, unit.Sp(12), "✓ Protocol: Model Context Protocol (MCP) • Port 8765 / stdio • Status: Active")
					lbl.Color = pal.StatusOnline
					lbl.Font.Weight = 600
					return lbl.Layout(gtx)
				})
			}),
			layout.Flexed(1, layout.Spacer{}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnCancelWizard, func(gtx layout.Context) layout.Dimensions {
							return DrawBadge(gtx, pal.SurfaceElevated, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(12), "← Back")
								lbl.Color = pal.TextSecondary
								return lbl.Layout(gtx)
							})
						})
					}),
					layout.Flexed(1, layout.Spacer{}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.btnFinishWizard, func(gtx layout.Context) layout.Dimensions {
							macroBtn := op.Record(gtx.Ops)
							dimsBtn := layout.Inset{
								Top:    unit.Dp(10),
								Bottom: unit.Dp(10),
								Left:   unit.Dp(18),
								Right:  unit.Dp(18),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(13), "Finish & Add to Active Hub ✓")
								lbl.Color = pal.Canvas
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
							})
							callBtn := macroBtn.Stop()

							paint.FillShape(gtx.Ops, pal.StatusOnline, clip.UniformRRect(image.Rect(0, 0, dimsBtn.Size.X, dimsBtn.Size.Y), gtx.Dp(unit.Dp(brand.RadiusNormal))).Op(gtx.Ops))
							callBtn.Add(gtx.Ops)
							return dimsBtn
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
