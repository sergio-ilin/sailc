package domain

import (
	"strings"
	"time"
	"unicode"
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
	Tags        []string     `json:"tags"`         // Standard PascalCase tags (e.g. "ProjectSailC", "GoBackend")
}

// NormalizeTag formats any tag into the strict SailC PascalCase tag standard without spaces:
// e.g., "project sailc" -> "ProjectSailC", "go-backend" -> "GoBackend", "#ast_parser" -> "AstParser".
func NormalizeTag(raw string) string {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "#")
	if cleaned == "" {
		return ""
	}

	// Split by space, hyphen, underscore, dot
	words := strings.FieldsFunc(cleaned, func(r rune) bool {
		return unicode.IsSpace(r) || r == '-' || r == '_' || r == '.' || r == '/'
	})

	var result strings.Builder
	for _, w := range words {
		runes := []rune(w)
		if len(runes) == 0 {
			continue
		}
		// Capitalize first letter of each word
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}

	res := result.String()
	if res == "" {
		return cleaned
	}
	return res
}

// NormalizeTags applies PascalCase formatting and deduplicates a slice of tags.
func NormalizeTags(tags []string) []string {
	seen := make(map[string]bool)
	var normalized []string

	for _, t := range tags {
		norm := NormalizeTag(t)
		if norm != "" && !seen[norm] {
			seen[norm] = true
			normalized = append(normalized, norm)
		}
	}
	return normalized
}
