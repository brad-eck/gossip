package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	// "github.com/brad-eck/gossip/internal/config"
	"github.com/brad-eck/gossip/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.InitialModel(5), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// cfg, err := config.load()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
	// 	os.Exit(1)
	// }

	// if len(cfg.Hosts) == 0 && len(cfg.Groups) == 0 {
	// 	fmt.Fprintf(os.Stderr, "No hosts defined. Use --hosts or config file.\n")
	// 	os.Exit(1)
	// }

	// p := tea.NewProgram(tui.InitialModel(cfg), tea.WithAltScreen())
	// if err := p.Start(); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error running gossip: %v\n", err)
	// 	os.Exit(1)
	// }
}