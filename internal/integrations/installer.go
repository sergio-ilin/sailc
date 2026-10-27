package integrations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// GetSailcExePath returns the absolute path to the running or companion sailc.exe.
func GetSailcExePath() string {
	exePath, err := os.Executable()
	if err == nil {
		return exePath
	}
	// Fallback to current working directory
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "sailc.exe")
}

// ConfigureClaudeDesktop injects SailC MCP server into claude_desktop_config.json
func ConfigureClaudeDesktop() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not found")
	}

	claudeDir := filepath.Join(appData, "Claude")
	_ = os.MkdirAll(claudeDir, 0755)
	configPath := filepath.Join(claudeDir, "claude_desktop_config.json")

	var configMap map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &configMap)
	}
	if configMap == nil {
		configMap = make(map[string]interface{})
	}

	var mcpServers map[string]interface{}
	if existing, ok := configMap["mcpServers"].(map[string]interface{}); ok {
		mcpServers = existing
	} else {
		mcpServers = make(map[string]interface{})
	}

	exePath := GetSailcExePath()
	mcpServers["sailc"] = map[string]interface{}{
		"command": exePath,
		"args":    []string{"-mcp-stdio"},
	}
	configMap["mcpServers"] = mcpServers

	out, err := json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully injected SailC MCP into Claude Desktop!\nFile: %s\nRestart Claude Desktop to activate.", configPath), nil
}

// ConfigureAntigravity sets up MCP tools in ~/.gemini/antigravity/mcp/sailc
func ConfigureAntigravity() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	targetDir := filepath.Join(home, ".gemini", "antigravity", "mcp", "sailc")
	_ = os.MkdirAll(targetDir, 0755)

	tools := []string{"hub_search", "hub_ask", "hub_answer", "hub_update_tags", "agent_report_status", "agent_list"}
	for _, t := range tools {
		schema := map[string]interface{}{
			"name":        t,
			"description": fmt.Sprintf("SailC AI Hub Tool: %s", t),
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		}
		raw, _ := json.MarshalIndent(schema, "", "  ")
		_ = os.WriteFile(filepath.Join(targetDir, t+".json"), raw, 0644)
	}

	instructions := "# SailC Central AI Hub Tools\nConnected to local hub.\n- Use hub_search before asking.\n- Use hub_ask with PascalCase tags.\n- Obligated to answer open threads via hub_answer."
	_ = os.WriteFile(filepath.Join(targetDir, "instructions.md"), []byte(instructions), 0644)

	return fmt.Sprintf("Successfully configured Antigravity MCP provider!\nFolder: %s", targetDir), nil
}

// ConfigureCursor writes .cursor/mcp.json in the current working directory or user profile
func ConfigureCursor() (string, error) {
	cwd, _ := os.Getwd()
	cursorDir := filepath.Join(cwd, ".cursor")
	_ = os.MkdirAll(cursorDir, 0755)
	configPath := filepath.Join(cursorDir, "mcp.json")

	exePath := GetSailcExePath()
	config := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"sailc": map[string]interface{}{
				"command": exePath,
				"args":    []string{"-mcp-stdio"},
			},
		},
	}

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully generated Cursor MCP config!\nFile: %s\nReload Cursor window to apply.", configPath), nil
}

// ProbeOllama checks if Ollama service is reachable
func ProbeOllama() (string, error) {
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		return "", fmt.Errorf("Ollama is not running on port 11434. Start Ollama and retry.")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "Ollama engine successfully detected on port 11434!", nil
	}
	return "Ollama port reached but returned non-200 code.", nil
}

// ProbeLMStudio checks if LM Studio local inference server is reachable
func ProbeLMStudio() (string, error) {
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get("http://127.0.0.1:1234/v1/models")
	if err != nil {
		return "", fmt.Errorf("LM Studio server is not running on port 1234. Start local server in LM Studio and retry.")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "LM Studio local inference server verified on port 1234!", nil
	}
	return "LM Studio port reached.", nil
}
