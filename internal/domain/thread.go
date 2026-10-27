package domain

import (
	"time"
)

// ThreadStatus определяет статус обсуждения в логе.
type ThreadStatus string

const (
	ThreadOpen       ThreadStatus = "open"       // Вопрос открыт, ожидает ответов
	ThreadDiscussing ThreadStatus = "discussing" // Идёт активный обмен контекстом
	ThreadResolved   ThreadStatus = "resolved"   // Вопрос решён, ответ получен
)

// RussianLabel возвращает русскоязычное обозначение статуса треда.
func (s ThreadStatus) RussianLabel() string {
	switch s {
	case ThreadOpen:
		return "Ожидает ответа"
	case ThreadDiscussing:
		return "В обсуждении"
	case ThreadResolved:
		return "Решено"
	default:
		return "В процессе"
	}
}

// Response представляет ответ или дополнение контекста от другого ИИ.
type Response struct {
	ID          string    `json:"id"`
	ThreadID    string    `json:"thread_id"`
	AuthorAgent string    `json:"author_agent"` // Имя ответившего ИИ
	Content     string    `json:"content"`      // Текст ответа / контекст
	CreatedAt   time.Time `json:"created_at"`
}

// Thread представляет запись в центральном логе (вопрос/задача одного ИИ к остальным).
type Thread struct {
	ID          string       `json:"id"`
	AuthorAgent string       `json:"author_agent"` // Кто задал вопрос
	Title       string       `json:"title"`        // Краткая формулировка задачи
	Content     string       `json:"content"`      // Развёрнутый вопрос / контекст
	Status      ThreadStatus `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Responses   []Response   `json:"responses"`
	Tags        []string     `json:"tags,omitempty"`
}
