package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sergio-ilin/sailc/internal/daemon"
	"github.com/sergio-ilin/sailc/internal/domain"
	"github.com/sergio-ilin/sailc/internal/mcp"
	"github.com/sergio-ilin/sailc/internal/monitor"
)

func main() {
	portFlag := flag.Int("port", 8765, "Port for the embedded MCP server")
	installStartup := flag.Bool("install-startup", false, "Register daemon in Windows startup for 24/7 background availability")
	uninstallStartup := flag.Bool("uninstall-startup", false, "Remove daemon from Windows startup")
	flag.Parse()

	if *installStartup {
		daemon.AttachParentConsole()
		if err := daemon.InstallStartup(*portFlag); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[SailC] Daemon registered in Windows startup.")
		return
	}

	if *uninstallStartup {
		daemon.AttachParentConsole()
		if err := daemon.UninstallStartup(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[SailC] Daemon removed from Windows startup.")
		return
	}

	store := domain.NewStore()
	store.StartWatcher(1 * time.Second)

	mcpServer := mcp.NewServer(store, *portFlag)
	if err := mcpServer.StartHTTP(); err != nil {
		log.Fatalf("Fatal: Failed to start HTTP MCP server on port %d: %v", *portFlag, err)
	}

	scanner := monitor.NewScanner(store, 3*time.Second)
	scanner.Start()
	defer scanner.Stop()

	// Block forever in background
	select {}
}
