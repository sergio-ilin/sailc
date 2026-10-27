package domain

import (
	"time"
)

// AgentStatus represents the active operating state of an AI agent.
type AgentStatus string

const (
	StatusWorking AgentStatus = "working" // Currently executing tasks
	StatusOnline  AgentStatus = "online"  // Active and ready for requests
	StatusOffline AgentStatus = "offline" // Process / service not detected
)

// Label returns the human-readable English label for the status.
func (s AgentStatus) Label() string {
	switch s {
	case StatusWorking:
		return "Working"
	case StatusOnline:
		return "Ready"
	case StatusOffline:
		return "Offline"
	default:
		return "Offline"
	}
}

// AvailableAI describes an AI model/assistant available to be added.
type AvailableAI struct {
	ID          string
	Name        string
	Provider    string
	Description string
	AutoSupport bool // Supports 1-click automatic config injection
}

// GetAvailableAIs returns all supported AI assistants that can be connected.
func GetAvailableAIs() []AvailableAI {
	return []AvailableAI{
		{
			ID:          "claude",
			Name:        "Claude Desktop",
			Provider:    "Anthropic",
			Description: "Full desktop application with native MCP support via claude_desktop_config.json",
			AutoSupport: true,
		},
		{
			ID:          "cursor",
			Name:        "Cursor AI Editor",
			Provider:    "Anysphere",
			Description: "AI-first code editor with project-level or global MCP configuration",
			AutoSupport: true,
		},
		{
			ID:          "antigravity",
			Name:        "Antigravity",
			Provider:    "Google DeepMind",
			Description: "Advanced agentic coding platform with native tool manifests",
			AutoSupport: true,
		},
		{
			ID:          "vscode_copilot",
			Name:        "VS Code / Cline",
			Provider:    "Microsoft / Open Source",
			Description: "VS Code with GitHub Copilot or Cline / Roo Code MCP extension",
			AutoSupport: false,
		},
		{
			ID:          "ollama",
			Name:        "Ollama Engine",
			Provider:    "Ollama Inc.",
			Description: "Local open-source LLM inference engine listening on port 11434",
			AutoSupport: false,
		},
		{
			ID:          "lmstudio",
			Name:        "LM Studio",
			Provider:    "Element Labs",
			Description: "Local LLM desktop studio with OpenAI-compatible server on port 1234",
			AutoSupport: false,
		},
	}
}

// Agent represents an installed or connected AI on the machine.
type Agent struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`         // e.g. "IDE Assistant", "Desktop AI", "Local LLM Server"
	Source       string      `json:"source"`       // Origin of detection: "Process: Antigravity.exe", "Port: 127.0.0.1:11434", etc.
	Status       AgentStatus `json:"status"`
	CurrentTask  string      `json:"current_task"` // Summary of current active task
	TaskCount    int         `json:"task_count"`   // Number of active tasks
	PID          int         `json:"pid"`
	Port         int         `json:"port"`
	SortOrder    int         `json:"sort_order"`   // Stable deterministic position
	IsConfigured bool        `json:"is_configured"`// Whether explicitly connected by user
	LastSeen     time.Time   `json:"last_seen"`
	IsManaged    bool        `json:"is_managed"`   // Discovered via process scan or registered via MCP
}
