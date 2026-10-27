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

// Agent represents an installed or connected AI on the machine.
type Agent struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`         // e.g. "IDE Assistant", "Desktop AI", "Local LLM Server"
	Source      string      `json:"source"`       // Origin of detection: "Process: Antigravity.exe", "Port: 127.0.0.1:11434", etc.
	Status      AgentStatus `json:"status"`
	CurrentTask string      `json:"current_task"` // Summary of current active task
	TaskCount   int         `json:"task_count"`   // Number of active tasks
	PID         int         `json:"pid"`
	Port        int         `json:"port"`
	SortOrder   int         `json:"sort_order"`   // Stable deterministic position
	LastSeen    time.Time   `json:"last_seen"`
	IsManaged   bool        `json:"is_managed"`   // Discovered via process scan or registered via MCP
}
