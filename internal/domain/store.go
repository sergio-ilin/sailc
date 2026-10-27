package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Store управляет центральным состоянием агентов и лога.
type Store struct {
	mu           sync.RWMutex
	filePath     string
	agents       map[string]*Agent
	threads      []*Thread
	subscribers  []chan struct{}
}

type persistedData struct {
	Agents  []*Agent  `json:"agents"`
	Threads []*Thread `json:"threads"`
}

// NewStore создаёт и загружает хранилище.
func NewStore() *Store {
	storageDir := ""
	if appData := os.Getenv("APPDATA"); appData != "" {
		storageDir = filepath.Join(appData, "SailC")
	} else {
		storageDir = "./data"
	}
	_ = os.MkdirAll(storageDir, 0755)
	dataFile := filepath.Join(storageDir, "hub_data.json")

	s := &Store{
		filePath: dataFile,
		agents:   make(map[string]*Agent),
		threads:  make([]*Thread, 0),
	}

	s.load()

	// Если данных нет, добавим базовые известные ИИ и вводный лог
	if len(s.threads) == 0 {
		s.seedInitialData()
	}

	return s
}

func (s *Store) seedInitialData() {
	now := time.Now()

	s.threads = append(s.threads, &Thread{
		ID:          "thr-101",
		AuthorAgent: "Antigravity",
		Title:       "Инициализация центрального контекстного хаба",
		Content:     "Развернут нативный хаб SailC на Go. Всем локальным ИИ доступен поиск по контексту (hub_search) и синхронизация задач.",
		Status:      ThreadResolved,
		CreatedAt:   now.Add(-15 * time.Minute),
		UpdatedAt:   now.Add(-10 * time.Minute),
		Responses: []Response{
			{
				ID:          "resp-1",
				ThreadID:    "thr-101",
				AuthorAgent: "Claude Desktop",
				Content:     "Контекст принят. Подключился к единому логу. Готов отвечать на запросы по коду и архитектуре.",
				CreatedAt:   now.Add(-10 * time.Minute),
			},
		},
		Tags: []string{"система", "инициализация", "контекст"},
	})
}

func (s *Store) notify() {
	for _, ch := range s.subscribers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Subscribe возвращает канал для получения уведомлений об изменениях.
func (s *Store) Subscribe() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan struct{}, 1)
	s.subscribers = append(s.subscribers, ch)
	return ch
}

// UpsertAgent обновляет или добавляет данные об агенте.
func (s *Store) UpsertAgent(agent *Agent) {
	s.mu.Lock()
	existing, ok := s.agents[agent.ID]
	if ok {
		existing.Name = agent.Name
		existing.Type = agent.Type
		existing.Status = agent.Status
		if agent.CurrentTask != "" {
			existing.CurrentTask = agent.CurrentTask
		}
		if agent.TaskCount >= 0 {
			existing.TaskCount = agent.TaskCount
		}
		if agent.PID > 0 {
			existing.PID = agent.PID
		}
		if agent.Port > 0 {
			existing.Port = agent.Port
		}
		existing.LastSeen = time.Now()
	} else {
		agent.LastSeen = time.Now()
		s.agents[agent.ID] = agent
	}
	s.mu.Unlock()

	s.notify()
	s.save()
}

// ListAgents возвращает список всех агентов.
func (s *Store) ListAgents() []*Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*Agent, 0, len(s.agents))
	for _, a := range s.agents {
		copyAgent := *a
		list = append(list, &copyAgent)
	}
	return list
}

// GetAgent возвращает агента по ID.
func (s *Store) GetAgent(id string) (*Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agents[id]
	if !ok {
		return nil, false
	}
	cp := *a
	return &cp, true
}

// CreateThread создаёт новый вопрос в центральном логе.
func (s *Store) CreateThread(author, title, content string, tags []string) *Thread {
	s.mu.Lock()
	now := time.Now()
	thr := &Thread{
		ID:          fmt.Sprintf("thr-%d", now.UnixNano()/1e6),
		AuthorAgent: author,
		Title:       title,
		Content:     content,
		Status:      ThreadOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
		Responses:   make([]Response, 0),
		Tags:        tags,
	}
	// Добавляем в начало списка (свежие сверху)
	s.threads = append([]*Thread{thr}, s.threads...)

	// Обновляем статус автора на "В работе"
	if a, ok := s.agents[author]; ok {
		a.Status = StatusWorking
		a.CurrentTask = title
		a.TaskCount++
	}

	s.mu.Unlock()

	s.notify()
	s.save()
	return thr
}

// AddResponse добавляет ответ от другого ИИ в существующий тред.
func (s *Store) AddResponse(threadID, author, content string) (*Response, error) {
	s.mu.Lock()
	var target *Thread
	for _, t := range s.threads {
		if t.ID == threadID {
			target = t
			break
		}
	}

	if target == nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("обсуждение с id %s не найдено", threadID)
	}

	now := time.Now()
	resp := Response{
		ID:          fmt.Sprintf("resp-%d", now.UnixNano()/1e6),
		ThreadID:    threadID,
		AuthorAgent: author,
		Content:     content,
		CreatedAt:   now,
	}

	target.Responses = append(target.Responses, resp)
	target.UpdatedAt = now
	target.Status = ThreadDiscussing

	// Обновим статус ответившего ИИ
	if a, ok := s.agents[author]; ok {
		a.LastSeen = now
	}

	s.mu.Unlock()

	s.notify()
	s.save()
	return &resp, nil
}

// ListThreads возвращает список всех тредов.
func (s *Store) ListThreads() []*Thread {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*Thread, len(s.threads))
	copy(res, s.threads)
	return res
}

// SearchThreads производит поиск по ключевым словам в заголовках, содержании и ответах.
func (s *Store) SearchThreads(query string) []*Thread {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		res := make([]*Thread, len(s.threads))
		copy(res, s.threads)
		return res
	}

	var results []*Thread
	for _, t := range s.threads {
		match := false
		if strings.Contains(strings.ToLower(t.Title), q) ||
			strings.Contains(strings.ToLower(t.Content), q) ||
			strings.Contains(strings.ToLower(t.AuthorAgent), q) {
			match = true
		} else {
			for _, r := range t.Responses {
				if strings.Contains(strings.ToLower(r.Content), q) ||
					strings.Contains(strings.ToLower(r.AuthorAgent), q) {
					match = true
					break
				}
			}
		}
		if match {
			results = append(results, t)
		}
	}
	return results
}

// GetThreadByID ищет тред по идентификатору.
func (s *Store) GetThreadByID(id string) (*Thread, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.threads {
		if t.ID == id {
			return t, true
		}
	}
	return nil, false
}

// TotalStats возвращает общую статистику (всего агентов, активных, задач).
func (s *Store) TotalStats() (totalAgents int, workingAgents int, activeTasks int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalAgents = len(s.agents)
	for _, a := range s.agents {
		if a.Status == StatusWorking {
			workingAgents++
			activeTasks += a.TaskCount
		} else if a.Status == StatusOnline {
			activeTasks += a.TaskCount
		}
	}
	return
}

func (s *Store) save() {
	s.mu.RLock()
	data := persistedData{
		Agents:  make([]*Agent, 0, len(s.agents)),
		Threads: s.threads,
	}
	for _, a := range s.agents {
		data.Agents = append(data.Agents, a)
	}
	s.mu.RUnlock()

	raw, err := json.MarshalIndent(data, "", "  ")
	if err == nil {
		_ = os.WriteFile(s.filePath, raw, 0644)
	}
}

func (s *Store) load() {
	raw, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var data persistedData
	if err := json.Unmarshal(raw, &data); err != nil {
		return
	}

	for _, a := range data.Agents {
		s.agents[a.ID] = a
	}
	s.threads = data.Threads
}
