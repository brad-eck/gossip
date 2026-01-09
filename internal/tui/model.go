package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#282A36")).
			Bold(true).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFB86C"))

	hostStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1)
)

type Model struct {
	ready      bool
	width      int
	height     int
	broadcast  bool
	selected   int
	hosts  	   []string
}

func InitialModel(hosts []string) Model {
	return Model{
		hosts: hosts,
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
			if m.selected < len(m.hosts)-1 {
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

	var s strings.Builder

	title := "🗣️ Gossip — TUI Cluster SSH"
	s.WriteString(titleStyle.Render(title) + "\n\n")

	s.WriteString(headerStyle.Render("Hosts") + "\n")

	for i, host := range m.hosts {
		cursor := "  "
		style := hostStyle
		if i == m.selected {
			cursor = "▶ "
			style = selectedStyle
		}

		line := cursor + style.Render(host)
		s.WriteString(line + "\n")
	}

	mode := "Single"
	modeColor := lipgloss.Color("#FF5555") // red = single
	if m.broadcast {
		mode = "Broadcast"
		modeColor = lipgloss.Color("#50FA7B") // green = broadcast
	}
	status := fmt.Sprintf("Mode: %s • Selected: %d/%d",
		lipgloss.NewStyle().Foreground(modeColor).Render(mode),
		m.selected+1,
		len(m.hosts),
	)
	s.WriteString(statusStyle.Render(status) + "\n")

	// Help
	helpText := "j/k: navigate • b: toggle broadcast • q: quit"
	s.WriteString(helpStyle.Render(helpText))

	return s.String()
}