package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Bold(true).
			Margin(1, 2)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Margin(0, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Margin(1, 2)
)

type Model struct {
	ready      bool
	width      int
	height     int
	broadcast  bool
	selected   int
	hostCount  int
}

func InitialModel(hostCount int) Model {
	return Model{
		hostCount: hostCount,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "b":
			m.broadcast = !m.broadcast
		case "j", "down":
			if m.selected < m.hostCount-1 {
				m.selected++
			}
		case "k", "up":
			if m.selected > 0 {
				m.selected--
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing gossip..."
	}

	var s string

	// Title
	s += titleStyle.Render("🗣️  Gossip — TUI Cluster SSH") + "\n\n"

	// Host list placeholder
	s += lipgloss.NewStyle().Bold(true).Render("Connected Hosts:\n")
	for i := 0; i < m.hostCount; i++ {
		prefix := "  "
		if i == m.selected {
			prefix = "▶ "
		}
		s += fmt.Sprintf("%sHost %d (placeholder)\n", prefix, i+1)
	}

	// Status bar
	mode := "Single"
	if m.broadcast {
		mode = "Broadcast 🌐"
	}
	s += "\n" + statusStyle.Render(fmt.Sprintf("Mode: %s | Selected: %d/%d", mode, m.selected+1, m.hostCount))

	// Help
	s += "\n\n" + helpStyle.Render("b: toggle broadcast • j/k: navigate • q: quit")

	return s
}