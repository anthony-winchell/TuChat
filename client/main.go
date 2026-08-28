package main

import (
	"flag"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tuchat/client/tui"
)

var addr = flag.String("addr", "", "server address (default: $TUCHAT_ADDR or localhost:8080)")

func main() {
	flag.Parse()

	serverAddr := *addr
	if serverAddr == "" {
		serverAddr = os.Getenv("TUCHAT_ADDR")
	}
	if serverAddr == "" {
		serverAddr = "localhost:8080"
	}

	p := tea.NewProgram(
		tui.New(serverAddr),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()

	if err != nil {
		os.Exit(1)
	}
}
