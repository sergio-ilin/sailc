package domain

import (
	"time"
)

// AgentStatus определяет рабочее состояние ИИ-агента.
type AgentStatus string

const (
	StatusWorking AgentStatus = "working" // Выполняет задачи
	StatusOnline  AgentStatus = "online"  // Запущен и готов к работе
	StatusOffline AgentStatus = "offline" // Не запущен в системе
)

// RussianLabel возвращает русскоязычное обозначение статуса.
func (s AgentStatus) RussianLabel() string {
	switch s {
	case StatusWorking:
		return "В работе"
	case StatusOnline:
		return "Готов"
	case StatusOffline:
		return "Не запущен"
	default:
		return "Неизвестно"
	}
}

// Agent представляет установленный или подключённый ИИ на компьютере.
type Agent struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`        // Например: "IDE Ассистент", "Локальная LLM", "MCP Агент"
	Status      AgentStatus `json:"status"`
	CurrentTask string      `json:"current_task"` // Краткое описание текущей выполняемой задачи
	TaskCount   int         `json:"task_count"`   // Количество активных задач
	PID         int         `json:"pid"`
	Port        int         `json:"port"`
	LastSeen    time.Time   `json:"last_seen"`
	IsManaged   bool        `json:"is_managed"` // Обнаружен автоматически процессом или зарегистрирован по MCP
}
