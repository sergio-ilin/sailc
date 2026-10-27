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
	portFlag := flag.Int("port", 8765, "Порт для встроенного MCP сервера")
	stdioFlag := flag.Bool("mcp-stdio", false, "Запустить в режиме stdio MCP сервера (без графического окна)")
	flag.Parse()

	// 1. Инициализация центрального хранилища
	store := domain.NewStore()

	// 2. Инициализация MCP сервера
	mcpServer := mcp.NewServer(store, *portFlag)

	// Если запрошен режим stdio (для прямого запуска как MCP плагин)
	if *stdioFlag {
		mcpServer.RunStdio()
		return
	}

	// Запуск HTTP транспорта MCP
	if err := mcpServer.StartHTTP(); err != nil {
		log.Printf("Предупреждение: Не удалось запустить HTTP MCP сервер на порту %d: %v", *portFlag, err)
	} else {
		fmt.Printf("[SailC] MCP сервер запущен на http://127.0.0.1:%d/mcp\n", *portFlag)
	}

	// 3. Запуск фонового сканера локальных процессов и портов ИИ
	scanner := monitor.NewScanner(store, 3*time.Second)
	scanner.Start()
	defer scanner.Stop()

	// 4. Инициализация детектора системной темы Windows
	themeDetector := theme.NewDetector()

	// 5. Запуск нативного быстрого графического интерфейса Gio
	appUI := ui.NewApp(store, themeDetector, *portFlag)

	go func() {
		if err := appUI.Run(); err != nil {
			log.Fatalf("Ошибка в оконной подсистеме: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}
