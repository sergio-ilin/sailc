package domain

import (
	"time"
)

// ThreadStatus defines the status of a discussion thread in the central log.
type ThreadStatus string

const (
	ThreadOpen       ThreadStatus = "open"       // Open, awaiting answers
	ThreadDiscussing ThreadStatus = "discussing" // Active context exchange in progress
	ThreadResolved   ThreadStatus = "resolved"   // Resolved, solution provided
)

// Label returns the human-readable English label for the thread status.
func (s ThreadStatus) Label() string {
	switch s {
	case ThreadOpen:
		return "Awaiting Answer"
	case ThreadDiscussing:
		return "Discussing"
	case ThreadResolved:
		return "Resolved"
	default:
		return "In Progress"
	}
}

// Response represents an answer or context addition from another AI agent.
type Response struct {
	ID          string    `json:"id"`
	ThreadID    string    `json:"thread_id"`
	AuthorAgent string    `json:"author_agent"` // Name of the responding AI
	Content     string    `json:"content"`      // Answer or code snippet
	CreatedAt   time.Time `json:"created_at"`
}

// Thread represents a central log entry (an inquiry from one AI to the rest).
type Thread struct {
	ID          string       `json:"id"`
	AuthorAgent string       `json:"author_agent"` // Who asked
	Title       string       `json:"title"`        // Short task summary
	Content     string       `json:"content"`      // Detailed question and context
	Status      ThreadStatus `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Responses   []Response   `json:"responses"`
	Tags        []string     `json:"tags,omitempty"`
}
