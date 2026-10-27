package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Store manages thread-safe state for agents and knowledge threads.
type Store struct {
	mu            sync.RWMutex
	filePath      string
	agents        map[string]*Agent
	configuredIDs map[string]bool
	threads       []*Thread
	subscribers   []chan struct{}
}

type persistedData struct {
	Agents        []*Agent  `json:"agents"`
	ConfiguredIDs []string  `json:"configured_ids"`
	Threads       []*Thread `json:"threads"`
}

// NewStore initializes and loads the store.
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
		filePath:      dataFile,
		agents:        make(map[string]*Agent),
		configuredIDs: make(map[string]bool),
		threads:       make([]*Thread, 0),
	}

	s.load()

	// Seed initial example data if empty
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
		Title:       "Initialization of Decentralized AI Context Hub",
		Content:     "Native SailC hub initialized on Go. All local AI agents have access to proactive context search (hub_search) and inter-agent coordination.",
		Status:      ThreadResolved,
		CreatedAt:   now.Add(-15 * time.Minute),
		UpdatedAt:   now.Add(-10 * time.Minute),
		Responses: []Response{
			{
				ID:          "resp-1",
				ThreadID:    "thr-101",
				AuthorAgent: "Claude Desktop",
				Content:     "Context received. Connected to shared hub. Ready to collaborate on code, architecture, and task distribution.",
				CreatedAt:   now.Add(-10 * time.Minute),
			},
		},
		Tags: []string{"ProjectSailC", "SystemInit", "ContextHub"},
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

// Subscribe returns a channel notifying on state modifications.
func (s *Store) Subscribe() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan struct{}, 1)
	s.subscribers = append(s.subscribers, ch)
	return ch
}

// AddConfiguredAgent marks an agent as actively connected by the user.
func (s *Store) AddConfiguredAgent(id string) {
	s.mu.Lock()
	s.configuredIDs[id] = true
	if a, ok := s.agents[id]; ok {
		a.IsConfigured = true
	}
	s.mu.Unlock()

	s.notify()
	s.save()
}

// RemoveConfiguredAgent disconnects an agent from the active view.
func (s *Store) RemoveConfiguredAgent(id string) {
	s.mu.Lock()
	delete(s.configuredIDs, id)
	if a, ok := s.agents[id]; ok {
		a.IsConfigured = false
	}
	s.mu.Unlock()

	s.notify()
	s.save()
}

// IsAgentConfigured checks if an agent has been added.
func (s *Store) IsAgentConfigured(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configuredIDs[id]
}

// UpsertAgent updates or inserts an agent state.
func (s *Store) UpsertAgent(agent *Agent) {
	s.mu.Lock()
	existing, ok := s.agents[agent.ID]
	changed := false

	isConfigured := s.configuredIDs[agent.ID]
	agent.IsConfigured = isConfigured

	if ok {
		if existing.Status != agent.Status ||
			existing.CurrentTask != agent.CurrentTask ||
			existing.TaskCount != agent.TaskCount ||
			existing.PID != agent.PID ||
			existing.Port != agent.Port ||
			existing.Source != agent.Source ||
			existing.IsConfigured != agent.IsConfigured {
			changed = true
		}
		existing.Name = agent.Name
		existing.Type = agent.Type
		existing.Source = agent.Source
		existing.Status = agent.Status
		existing.SortOrder = agent.SortOrder
		existing.IsConfigured = isConfigured
		existing.CurrentTask = agent.CurrentTask
		existing.TaskCount = agent.TaskCount
		existing.PID = agent.PID
		existing.Port = agent.Port
		existing.LastSeen = time.Now()
	} else {
		changed = true
		agent.LastSeen = time.Now()
		s.agents[agent.ID] = agent
	}
	s.mu.Unlock()

	if changed {
		s.notify()
		s.save()
	}
}

// ListConfiguredAgents returns only agents that have been explicitly added/configured.
func (s *Store) ListConfiguredAgents() []*Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*Agent, 0)
	for id := range s.configuredIDs {
		if a, ok := s.agents[id]; ok {
			copyAgent := *a
			list = append(list, &copyAgent)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].SortOrder != list[j].SortOrder {
			return list[i].SortOrder < list[j].SortOrder
		}
		return list[i].Name < list[j].Name
	})

	return list
}

// ListAgents returns all agents (for MCP and internal queries).
func (s *Store) ListAgents() []*Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*Agent, 0, len(s.agents))
	for _, a := range s.agents {
		copyAgent := *a
		list = append(list, &copyAgent)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].SortOrder != list[j].SortOrder {
			return list[i].SortOrder < list[j].SortOrder
		}
		return list[i].Name < list[j].Name
	})

	return list
}

// GetAgent fetches an agent by ID.
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

// CreateThread creates a new question/task in the central log.
func (s *Store) CreateThread(author, title, content string, tags []string) *Thread {
	s.mu.Lock()
	now := time.Now()
	normTags := NormalizeTags(tags)

	thr := &Thread{
		ID:          fmt.Sprintf("thr-%d", now.UnixNano()/1e6),
		AuthorAgent: author,
		Title:       title,
		Content:     content,
		Status:      ThreadOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
		Responses:   make([]Response, 0),
		Tags:        normTags,
	}
	s.threads = append([]*Thread{thr}, s.threads...)

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

// AddResponse appends a reply from another AI agent to an existing thread.
func (s *Store) AddResponse(threadID, author, content string, additionalTags []string) (*Response, []string, error) {
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
		return nil, nil, fmt.Errorf("discussion thread %s not found", threadID)
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

	if len(additionalTags) > 0 {
		merged := append(target.Tags, additionalTags...)
		target.Tags = NormalizeTags(merged)
	}

	finalTags := target.Tags

	if a, ok := s.agents[author]; ok {
		a.LastSeen = now
	}

	s.mu.Unlock()

	s.notify()
	s.save()
	return &resp, finalTags, nil
}

// UpdateThreadTags updates or appends PascalCase tags to an existing thread.
func (s *Store) UpdateThreadTags(threadID string, tags []string, appendMode bool) ([]string, error) {
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
		return nil, fmt.Errorf("discussion thread %s not found", threadID)
	}

	var updated []string
	if appendMode {
		updated = NormalizeTags(append(target.Tags, tags...))
	} else {
		updated = NormalizeTags(tags)
	}
	target.Tags = updated
	target.UpdatedAt = time.Now()

	s.mu.Unlock()

	s.notify()
	s.save()
	return updated, nil
}

// ListThreads returns all threads.
func (s *Store) ListThreads() []*Thread {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*Thread, len(s.threads))
	copy(res, s.threads)
	return res
}

// SearchThreads searches titles, content, responses, and PascalCase tags.
func (s *Store) SearchThreads(query string) []*Thread {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	q = strings.TrimPrefix(q, "#")

	if q == "" {
		res := make([]*Thread, len(s.threads))
		copy(res, s.threads)
		return res
	}

	var results []*Thread
	for _, t := range s.threads {
		match := false

		for _, tag := range t.Tags {
			if strings.Contains(strings.ToLower(tag), q) {
				match = true
				break
			}
		}

		if !match {
			if strings.Contains(strings.ToLower(t.Title), q) ||
				strings.Contains(strings.ToLower(t.Content), q) ||
				strings.Contains(strings.ToLower(t.AuthorAgent), q) {
				match = true
			}
		}

		if !match {
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

// GetThreadByID retrieves a thread by ID.
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

// TotalStats returns total configured agents, active working agents, and task count.
func (s *Store) TotalStats() (totalAgents int, workingAgents int, activeTasks int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalAgents = len(s.configuredIDs)
	for id := range s.configuredIDs {
		if a, ok := s.agents[id]; ok {
			if a.Status == StatusWorking {
				workingAgents++
				activeTasks += a.TaskCount
			} else if a.Status == StatusOnline {
				activeTasks += a.TaskCount
			}
		}
	}
	return
}

func (s *Store) save() {
	s.mu.RLock()
	data := persistedData{
		Agents:        make([]*Agent, 0, len(s.agents)),
		ConfiguredIDs: make([]string, 0, len(s.configuredIDs)),
		Threads:       s.threads,
	}
	for _, a := range s.agents {
		data.Agents = append(data.Agents, a)
	}
	for id := range s.configuredIDs {
		data.ConfiguredIDs = append(data.ConfiguredIDs, id)
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
	for _, id := range data.ConfiguredIDs {
		s.configuredIDs[id] = true
	}
	s.threads = data.Threads
}

// StartWatcher starts a background poller that checks if hub_data.json was modified
// on disk (e.g. by the background MCP daemon or other AI agents) and reloads it.
func (s *Store) StartWatcher(interval time.Duration) {
	go func() {
		var lastMod time.Time
		if fi, err := os.Stat(s.filePath); err == nil {
			lastMod = fi.ModTime()
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			fi, err := os.Stat(s.filePath)
			if err != nil {
				continue
			}
			if fi.ModTime().After(lastMod) {
				lastMod = fi.ModTime()
				s.mu.Lock()
				s.load()
				s.mu.Unlock()
				s.notify()
			}
		}
	}()
}
