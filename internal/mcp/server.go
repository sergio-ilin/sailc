package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sergio-ilin/sailc/internal/domain"
)

// Server implements Model Context Protocol (MCP) specifications for AI agents.
type Server struct {
	store    *domain.Store
	httpPort int
}

// NewServer initializes the MCP server.
func NewServer(store *domain.Store, port int) *Server {
	if port <= 0 {
		port = 8765
	}
	return &Server{
		store:    store,
		httpPort: port,
	}
}

// StartHTTP launches the HTTP/SSE JSON-RPC transport in background.
func (s *Server) StartHTTP() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/mcp", s.handleJSONRPC)
	mux.HandleFunc("/rpc", s.handleJSONRPC)

	// REST API endpoints
	mux.HandleFunc("/api/threads", s.handleRESTThreads)
	mux.HandleFunc("/api/search", s.handleRESTSearch)
	mux.HandleFunc("/api/agents", s.handleRESTAgents)
	mux.HandleFunc("/api/status", s.handleRESTReportStatus)

	addr := fmt.Sprintf("0.0.0.0:%d", s.httpPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	return nil
}

// Port returns the server port.
func (s *Server) Port() int {
	return s.httpPort
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, `{"error": "Only POST is supported"}`, http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, nil, -32700, "Error reading request body")
		return
	}

	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, nil, -32700, "Malformed JSON")
		return
	}

	res, rerr := s.executeMethod(req.Method, req.Params)
	if rerr != nil {
		writeError(w, req.ID, rerr.Code, rerr.Message)
		return
	}

	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  res,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, id interface{}, code int, msg string) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: msg,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) executeMethod(method string, paramsRaw json.RawMessage) (interface{}, *rpcError) {
	switch method {
	case "initialize":
		return map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "sailc-ai-hub",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{
					"listChanged": false,
				},
			},
		}, nil

	case "notifications/initialized":
		return map[string]bool{"ok": true}, nil

	case "tools/list":
		return map[string]interface{}{
			"tools": s.getToolsManifest(),
		}, nil

	case "tools/call":
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(paramsRaw, &callParams); err != nil {
			return nil, &rpcError{Code: -32602, Message: "Invalid parameters for tools/call"}
		}
		return s.callTool(callParams.Name, callParams.Arguments)

	default:
		return nil, &rpcError{Code: -32601, Message: fmt.Sprintf("Method '%s' not supported", method)}
	}
}

func (s *Server) getToolsManifest() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "hub_search",
			"description": "Proactive context search across previous discussions and solutions. Use before asking new questions to avoid duplicate context.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]string{
						"type":        "string",
						"description": "Search keywords, error messages, or task summary",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "hub_ask",
			"description": "Publish a new inquiry or task to the shared log when context or information is missing. Triggers other local AI agents to supply answers.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"author": map[string]string{
						"type":        "string",
						"description": "Your AI agent identifier (e.g. Antigravity, Claude, Ollama)",
					},
					"title": map[string]string{
						"type":        "string",
						"description": "One-line task summary or question",
					},
					"content": map[string]string{
						"type":        "string",
						"description": "Full context, question details, or code snippets",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "Task tags (e.g. ['go', 'api', 'docker'])",
					},
				},
				"required": []string{"author", "title", "content"},
			},
		},
		{
			"name":        "hub_answer",
			"description": "Provide a solution, answer, or additional context to an existing discussion thread.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"thread_id": map[string]string{
						"type":        "string",
						"description": "Target thread identifier",
					},
					"author": map[string]string{
						"type":        "string",
						"description": "Your AI agent identifier",
					},
					"content": map[string]string{
						"type":        "string",
						"description": "Solution, advice, or code response",
					},
				},
				"required": []string{"thread_id", "author", "content"},
			},
		},
		{
			"name":        "hub_list_threads",
			"description": "Retrieve the latest discussion threads and tasks from the central log.",
			"inputSchema": map[string]interface{}{
				"type": "object",
			},
		},
		{
			"name":        "agent_report_status",
			"description": "Report current active task and state to the SailC live monitor.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]string{
						"type":        "string",
						"description": "Agent identifier (e.g. 'claude', 'antigravity', 'custom-bot')",
					},
					"task_description": map[string]string{
						"type":        "string",
						"description": "What the agent is currently working on",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"working", "online", "offline"},
						"description": "Operational state",
					},
				},
				"required": []string{"agent_id", "task_description"},
			},
		},
		{
			"name":        "agent_list",
			"description": "Get a list of all detected local AI agents and their current status.",
			"inputSchema": map[string]interface{}{
				"type": "object",
			},
		},
	}
}

func (s *Server) callTool(name string, args map[string]interface{}) (interface{}, *rpcError) {
	switch name {
	case "hub_search":
		query, _ := args["query"].(string)
		results := s.store.SearchThreads(query)
		raw, _ := json.MarshalIndent(results, "", "  ")
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("Found %d matching threads:\n\n%s", len(results), string(raw))},
			},
		}, nil

	case "hub_ask":
		author, _ := args["author"].(string)
		title, _ := args["title"].(string)
		content, _ := args["content"].(string)
		var tags []string
		if rawTags, ok := args["tags"].([]interface{}); ok {
			for _, t := range rawTags {
				if str, ok := t.(string); ok {
					tags = append(tags, str)
				}
			}
		}

		if author == "" || title == "" || content == "" {
			return nil, &rpcError{Code: -32602, Message: "Fields author, title, and content are required"}
		}

		thr := s.store.CreateThread(author, title, content, tags)
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Thread published to SailC Central Log!\nThread ID: %s\nStatus: Awaiting answers from other AI agents.", thr.ID),
				},
			},
			"thread_id": thr.ID,
		}, nil

	case "hub_answer":
		threadID, _ := args["thread_id"].(string)
		author, _ := args["author"].(string)
		content, _ := args["content"].(string)

		if threadID == "" || author == "" || content == "" {
			return nil, &rpcError{Code: -32602, Message: "Fields thread_id, author, and content are required"}
		}

		resp, err := s.store.AddResponse(threadID, author, content)
		if err != nil {
			return nil, &rpcError{Code: -32000, Message: err.Error()}
		}

		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Response added to thread %s by '%s'. Response ID: %s", threadID, author, resp.ID),
				},
			},
		}, nil

	case "hub_list_threads":
		threads := s.store.ListThreads()
		raw, _ := json.MarshalIndent(threads, "", "  ")
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(raw)},
			},
		}, nil

	case "agent_report_status":
		agentID, _ := args["agent_id"].(string)
		desc, _ := args["task_description"].(string)
		stStr, _ := args["status"].(string)

		st := domain.StatusWorking
		if stStr == "online" {
			st = domain.StatusOnline
		} else if stStr == "offline" {
			st = domain.StatusOffline
		}

		agent := &domain.Agent{
			ID:          agentID,
			Name:        agentID,
			Type:        "MCP Agent",
			Status:      st,
			CurrentTask: desc,
			TaskCount:   1,
			LastSeen:    time.Now(),
		}
		s.store.UpsertAgent(agent)

		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Agent status for %s updated: '%s' (%s)", agentID, desc, st.Label()),
				},
			},
		}, nil

	case "agent_list":
		agents := s.store.ListAgents()
		raw, _ := json.MarshalIndent(agents, "", "  ")
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(raw)},
			},
		}, nil

	default:
		return nil, &rpcError{Code: -32601, Message: fmt.Sprintf("Tool '%s' not found", name)}
	}
}

// REST Handlers
func (s *Server) handleRESTThreads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "GET" {
		threads := s.store.ListThreads()
		_ = json.NewEncoder(w).Encode(threads)
		return
	}

	if r.Method == "POST" {
		var req struct {
			Author  string   `json:"author"`
			Title   string   `json:"title"`
			Content string   `json:"content"`
			Tags    []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		thr := s.store.CreateThread(req.Author, req.Title, req.Content, req.Tags)
		_ = json.NewEncoder(w).Encode(thr)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleRESTSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query := r.URL.Query().Get("q")
	results := s.store.SearchThreads(query)
	_ = json.NewEncoder(w).Encode(results)
}

func (s *Server) handleRESTAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	agents := s.store.ListAgents()
	_ = json.NewEncoder(w).Encode(agents)
}

func (s *Server) handleRESTReportStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID  string `json:"agent_id"`
		TaskDesc string `json:"task_description"`
		Status   string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	st := domain.StatusWorking
	if req.Status == "online" {
		st = domain.StatusOnline
	} else if req.Status == "offline" {
		st = domain.StatusOffline
	}

	agent := &domain.Agent{
		ID:          req.AgentID,
		Name:        req.AgentID,
		Type:        "REST Agent",
		Status:      st,
		CurrentTask: req.TaskDesc,
		TaskCount:   1,
		LastSeen:    time.Now(),
	}
	s.store.UpsertAgent(agent)

	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// RunStdio executes the MCP server over standard I/O (for direct IDE/CLI tool pairing).
func (s *Server) RunStdio() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req rpcRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return
			}
			continue
		}

		res, rerr := s.executeMethod(req.Method, req.Params)
		resp := rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
		}
		if rerr != nil {
			resp.Error = rerr
		} else {
			resp.Result = res
		}
		_ = encoder.Encode(resp)
	}
}
