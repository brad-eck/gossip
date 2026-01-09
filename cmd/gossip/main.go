package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/brad-eck/gossip/internal/tui"
	"github.com/brad-eck/gossip/internal/hosts"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: gossip <host1> <host2> ... [host-{1..N}]\n")
		fmt.Fprintf(os.Stderr, "Example: gossip user@web{1..4} db{1,2}\n")
		os.Exit(1)
	}

	rawHosts := os.Args[1:]
	expandedHosts := hosts.ExpandBrace(rawHosts)

	if len(expandedHosts) == 0 {
		fmt.Fprintf(os.Stderr, "No valid hosts provided\n")
		os.Exit(1)
	}

	fmt.Printf("Connecting to %d hosts: %v\n", len(expandedHosts), expandedHosts)

	p := tea.NewProgram(tui.InitialModel(expandedHosts), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running gossip: %v\n", err)
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