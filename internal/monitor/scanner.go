package monitor

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/sergio-ilin/sailc/internal/domain"
	"golang.org/x/sys/windows"
)

// Scanner выполняет периодический аудит установленных и активных ИИ на компьютере.
type Scanner struct {
	store    *domain.Store
	interval time.Duration
	stopCh   chan struct{}
}

// NewScanner создаёт новый экземпляр сканера.
func NewScanner(store *domain.Store, interval time.Duration) *Scanner {
	if interval < 2*time.Second {
		interval = 3 * time.Second
	}
	return &Scanner{
		store:    store,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start запускает фоновый процесс мониторинга.
func (s *Scanner) Start() {
	// Первичное мгновенное сканирование
	s.scanOnce()

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.scanOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop останавливает мониторинг.
func (s *Scanner) Stop() {
	close(s.stopCh)
}

type targetAI struct {
	ID        string
	Name      string
	Type      string
	ProcMatch []string
	Port      int
	TestURL   string
}

var knownAITargets = []targetAI{
	{
		ID:        "antigravity",
		Name:      "Antigravity",
		Type:      "IDE Ассистент / DeepMind",
		ProcMatch: []string{"antigravity"},
		Port:      0,
	},
	{
		ID:        "claude",
		Name:      "Claude Desktop",
		Type:      "Десктопный ИИ / Anthropic",
		ProcMatch: []string{"claude"},
		Port:      0,
	},
	{
		ID:        "cursor",
		Name:      "Cursor AI Editor",
		Type:      "IDE Ассистент",
		ProcMatch: []string{"cursor"},
		Port:      0,
	},
	{
		ID:        "ollama",
		Name:      "Ollama Engine",
		Type:      "Локальный LLM Сервер",
		ProcMatch: []string{"ollama", "ollama_llama_server"},
		Port:      11434,
		TestURL:   "http://127.0.0.1:11434/api/tags",
	},
	{
		ID:        "lmstudio",
		Name:      "LM Studio",
		Type:      "Локальный инференс LLM",
		ProcMatch: []string{"lmstudio", "lms"},
		Port:      1234,
		TestURL:   "http://127.0.0.1:1234/v1/models",
	},
	{
		ID:        "vscode_copilot",
		Name:      "VS Code Copilot",
		Type:      "IDE Ассистент",
		ProcMatch: []string{"code"},
		Port:      0,
	},
}

func (s *Scanner) scanOnce() {
	runningProcs := getRunningProcesses()

	for _, target := range knownAITargets {
		var matchedPID int
		var isRunning bool

		// 1. Поиск по процессам
		for _, nameMatch := range target.ProcMatch {
			if pids, ok := runningProcs[strings.ToLower(nameMatch)]; ok && len(pids) > 0 {
				isRunning = true
				matchedPID = pids[0]
				break
			}
		}

		// 2. Проверка сетевого порта, если указан
		var portOpen bool
		var extraTaskInfo string
		if target.Port > 0 {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(target.Port)), 150*time.Millisecond)
			if err == nil {
				conn.Close()
				portOpen = true
				isRunning = true
			}

			// Если порт открыт и есть URL, попытаемся узнать загруженные модели
			if target.TestURL != "" && portOpen {
				extraTaskInfo = fetchActiveModels(target.TestURL)
			}
		}

		// Определяем текущее состояние в Store
		existing, hasExisting := s.store.GetAgent(target.ID)
		status := domain.StatusOffline
		taskCount := 0
		taskDesc := ""

		if isRunning {
			status = domain.StatusOnline
			if hasExisting && existing.Status == domain.StatusWorking {
				status = domain.StatusWorking
				taskCount = existing.TaskCount
				taskDesc = existing.CurrentTask
			} else if extraTaskInfo != "" {
				status = domain.StatusWorking
				taskCount = 1
				taskDesc = extraTaskInfo
			} else {
				taskDesc = "Ожидание запросов"
			}
		} else {
			taskDesc = "Служба не запущена"
		}

		agent := &domain.Agent{
			ID:          target.ID,
			Name:        target.Name,
			Type:        target.Type,
			Status:      status,
			CurrentTask: taskDesc,
			TaskCount:   taskCount,
			PID:         matchedPID,
			Port:        target.Port,
			LastSeen:    time.Now(),
			IsManaged:   false,
		}

		s.store.UpsertAgent(agent)
	}
}

func fetchActiveModels(url string) string {
	client := http.Client{Timeout: 300 * time.Millisecond}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var data struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		var names []string
		for _, m := range data.Models {
			names = append(names, m.Name)
		}
		for _, d := range data.Data {
			names = append(names, d.ID)
		}
		if len(names) > 0 {
			return "Загружены модели: " + strings.Join(names, ", ")
		}
	}
	return "Активен API сервер"
}

// getRunningProcesses использует Windows Toolhelp32 для быстрого снятия снимка процессов
func getRunningProcesses() map[string][]int {
	procs := make(map[string][]int)

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return procs
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return procs
	}

	for {
		name := syscall.UTF16ToString(entry.ExeFile[:])
		nameLower := strings.ToLower(strings.TrimSuffix(name, ".exe"))
		procs[nameLower] = append(procs[nameLower], int(entry.ProcessID))

		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
	}

	return procs
}
