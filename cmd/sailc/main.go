package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"github.com/sergio-ilin/sailc/internal/daemon"
	"github.com/sergio-ilin/sailc/internal/domain"
	"github.com/sergio-ilin/sailc/internal/mcp"
	"github.com/sergio-ilin/sailc/internal/theme"
	"github.com/sergio-ilin/sailc/internal/ui"
)

func main() {
	portFlag := flag.Int("port", 8765, "Port for the embedded MCP server")
	stdioFlag := flag.Bool("mcp-stdio", false, "Run in stdio MCP server mode (headless without GUI)")
	installStartup := flag.Bool("install-startup", false, "Register background daemon in Windows startup")
	uninstallStartup := flag.Bool("uninstall-startup", false, "Remove background daemon from Windows startup")
	flag.Parse()

	if *installStartup {
		daemon.AttachParentConsole()
		if err := daemon.InstallStartup(*portFlag); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[SailC] Background daemon registered in Windows startup.")
		return
	}

	if *uninstallStartup {
		daemon.AttachParentConsole()
		if err := daemon.UninstallStartup(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[SailC] Background daemon removed from Windows startup.")
		return
	}

	// 1. Initialize central domain store
	store := domain.NewStore()

	// Mode A: STDIO mode (direct pipe to IDEs)
	if *stdioFlag {
		mcpServer := mcp.NewServer(store, *portFlag)
		mcpServer.RunStdio()
		return
	}

	// Mode B: Interactive GUI desktop window (strictly for observation)
	// Automatically ensure background daemon (sailcd.exe) is running so AI models always have 24/7 access
	_ = daemon.EnsureRunning(*portFlag)

	// Watch for store changes made by background daemon or AI tools
	store.StartWatcher(1 * time.Second)

	// Windows registry system theme detector
	themeDetector := theme.NewDetector()

	// Launch native Gio interface
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
