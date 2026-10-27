package monitor

import (
	"encoding/json"
	"fmt"
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

// Scanner performs periodic auditing of installed and active AI systems.
type Scanner struct {
	store    *domain.Store
	interval time.Duration
	stopCh   chan struct{}
}

// NewScanner creates a new scanner instance.
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

// Start begins the background monitoring loop.
func (s *Scanner) Start() {
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

// Stop terminates the scanner.
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
	SortOrder int
}

var knownAITargets = []targetAI{
	{
		ID:        "antigravity",
		Name:      "Antigravity",
		Type:      "IDE Assistant / DeepMind",
		ProcMatch: []string{"antigravity"},
		Port:      0,
		SortOrder: 1,
	},
	{
		ID:        "claude",
		Name:      "Claude Desktop",
		Type:      "Desktop AI / Anthropic",
		ProcMatch: []string{"claude"},
		Port:      0,
		SortOrder: 2,
	},
	{
		ID:        "cursor",
		Name:      "Cursor AI Editor",
		Type:      "IDE Assistant",
		ProcMatch: []string{"cursor"},
		Port:      0,
		SortOrder: 3,
	},
	{
		ID:        "ollama",
		Name:      "Ollama Engine",
		Type:      "Local LLM Server",
		ProcMatch: []string{"ollama", "ollama_llama_server"},
		Port:      11434,
		TestURL:   "http://127.0.0.1:11434/api/tags",
		SortOrder: 4,
	},
	{
		ID:        "lmstudio",
		Name:      "LM Studio",
		Type:      "Local Inference Server",
		ProcMatch: []string{"lmstudio", "lms"},
		Port:      1234,
		TestURL:   "http://127.0.0.1:1234/v1/models",
		SortOrder: 5,
	},
	{
		ID:        "vscode_copilot",
		Name:      "VS Code Copilot",
		Type:      "IDE Assistant",
		ProcMatch: []string{"code"},
		Port:      0,
		SortOrder: 6,
	},
}

func (s *Scanner) scanOnce() {
	runningProcs := getRunningProcesses()

	for _, target := range knownAITargets {
		var matchedPID int
		var matchedExeName string
		var isRunning bool

		// 1. Process matching
		for _, nameMatch := range target.ProcMatch {
			if matches, ok := runningProcs[strings.ToLower(nameMatch)]; ok && len(matches) > 0 {
				isRunning = true
				matchedPID = matches[0].pid
				matchedExeName = matches[0].exeName
				break
			}
		}

		// 2. Port check
		var portOpen bool
		var extraTaskInfo string
		if target.Port > 0 {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(target.Port)), 150*time.Millisecond)
			if err == nil {
				conn.Close()
				portOpen = true
				isRunning = true
			}

			if target.TestURL != "" && portOpen {
				extraTaskInfo = fetchActiveModels(target.TestURL)
			}
		}

		existing, hasExisting := s.store.GetAgent(target.ID)
		status := domain.StatusOffline
		taskCount := 0
		taskDesc := ""
		sourceText := ""

		if isRunning {
			status = domain.StatusOnline
			if matchedPID > 0 {
				sourceText = fmt.Sprintf("Process: %s (PID %d)", matchedExeName, matchedPID)
			} else if portOpen {
				sourceText = fmt.Sprintf("Port: 127.0.0.1:%d (REST API)", target.Port)
			}

			if hasExisting && existing.Status == domain.StatusWorking {
				status = domain.StatusWorking
				taskCount = existing.TaskCount
				taskDesc = existing.CurrentTask
			} else if extraTaskInfo != "" {
				status = domain.StatusWorking
				taskCount = 1
				taskDesc = extraTaskInfo
			} else {
				taskDesc = "Awaiting tasks"
			}
		} else {
			status = domain.StatusOffline
			taskDesc = "Process not detected"
			if target.Port > 0 {
				sourceText = fmt.Sprintf("Watched Port: :%d (Closed)", target.Port)
			} else {
				sourceText = fmt.Sprintf("Watched Process: %s.exe", target.ProcMatch[0])
			}
		}

		agent := &domain.Agent{
			ID:          target.ID,
			Name:        target.Name,
			Type:        target.Type,
			Source:      sourceText,
			Status:      status,
			CurrentTask: taskDesc,
			TaskCount:   taskCount,
			PID:         matchedPID,
			Port:        target.Port,
			SortOrder:   target.SortOrder,
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
			return "Loaded models: " + strings.Join(names, ", ")
		}
	}
	return "API server active"
}

type procInfo struct {
	pid     int
	exeName string
}

func getRunningProcesses() map[string][]procInfo {
	procs := make(map[string][]procInfo)

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
		rawExe := syscall.UTF16ToString(entry.ExeFile[:])
		nameLower := strings.ToLower(strings.TrimSuffix(rawExe, ".exe"))
		procs[nameLower] = append(procs[nameLower], procInfo{
			pid:     int(entry.ProcessID),
			exeName: rawExe,
		})

		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
	}

	return procs
}
