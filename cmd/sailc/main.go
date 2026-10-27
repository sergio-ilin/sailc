package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"github.com/sergio-ilin/sailc/internal/domain"
	"github.com/sergio-ilin/sailc/internal/mcp"
	"github.com/sergio-ilin/sailc/internal/monitor"
	"github.com/sergio-ilin/sailc/internal/theme"
	"github.com/sergio-ilin/sailc/internal/ui"
)

func main() {
	portFlag := flag.Int("port", 8765, "Port for the embedded MCP server")
	stdioFlag := flag.Bool("mcp-stdio", false, "Run in stdio MCP server mode (headless without GUI window)")
	flag.Parse()

	// 1. Initialize central thread-safe domain store
	store := domain.NewStore()

	// 2. Initialize Model Context Protocol server
	mcpServer := mcp.NewServer(store, *portFlag)

	if *stdioFlag {
		mcpServer.RunStdio()
		return
	}

	// Start background HTTP transport for MCP
	if err := mcpServer.StartHTTP(); err != nil {
		log.Printf("Warning: Failed to start HTTP MCP server on port %d: %v", *portFlag, err)
	} else {
		fmt.Printf("[SailC] MCP server listening on http://127.0.0.1:%d/mcp\n", *portFlag)
	}

	// 3. Start background AI process & port scanner
	scanner := monitor.NewScanner(store, 3*time.Second)
	scanner.Start()
	defer scanner.Stop()

	// 4. Initialize Windows registry theme detector
	themeDetector := theme.NewDetector()

	// 5. Start fast native Gio desktop interface
	appUI := ui.NewApp(store, themeDetector, *portFlag)

	go func() {
		if err := appUI.Run(); err != nil {
			log.Fatalf("Window subsystem error: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}
