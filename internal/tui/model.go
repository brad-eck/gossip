package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/brad-eck/gossip/internal/session"
	"github.com/charmbracelet/bubbles/viewport"
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

	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#44475A")).
			Padding(1, 2)
)

type Model struct {
	ready      bool
	width      int
	height     int
	broadcast  bool
	selected   int
	hosts      []string
	sessions   map[int]*session.Session
	viewports  map[int]viewport.Model
	outputBuf  map[int]*strings.Builder
	hostListVP viewport.Model
}

func InitialModel(hosts []string) Model {
	m := Model{
		hosts:     hosts,
		sessions:  make(map[int]*session.Session),
		viewports: make(map[int]viewport.Model),
		outputBuf: make(map[int]*strings.Builder),
	}

	for i, h := range hosts {
		sess := session.NewSession(h)
        m.sessions[i] = sess
        m.outputBuf[i] = &strings.Builder{}
        vp := viewport.New(80, 20)
        vp.SetContent("")
        m.viewports[i] = vp
	}
	m.hostListVP = viewport.New(0, 0)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.tickOutput(),
	)
}

func (m Model) tickOutput() tea.Cmd {
	return tea.Every(80*time.Millisecond, func(t time.Time) tea.Msg {
		return outputTickMsg{}
	})
}

type outputTickMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Resize viewports
		reservedTop := 4 + m.hostListVP.Height + 4
		reservedBottom := 5
		vpHeight := msg.Height - reservedTop - reservedBottom
		if vpHeight < 10 {
			vpHeight = 10 // minimum useful height
		}
		
		desiredHostHeight := len(m.hosts) + 2
		hostVPHeight := min(15, max(4, desiredHostHeight))
		m.hostListVP.Width = m.width - 10
		m.hostListVP.Height = hostVPHeight

		for i := range m.hosts {
			vp := m.viewports[i]
			vp.Width = msg.Width - 8 // padding
			vp.Height = vpHeight
			m.viewports[i] = vp
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			for _, s := range m.sessions {
				s.Close()
			}
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

		default:
			// Send keystroke to selected or all
			input := msg.String()
			if input == "enter" {
				input = "\r\n"
			} else if len(input) == 1 {
				input += "\r" // most terminals expect \r for enter
			}

			var targets []*session.Session
			if m.broadcast {
				for _, s := range m.sessions {
					targets = append(targets, s)
				}
			} else {
				targets = append(targets, m.sessions[m.selected])
			}

			for _, s := range targets {
				_ = s.Write([]byte(input))
			}
		}

	case outputTickMsg:
		updated := false
		for i, sess := range m.sessions {
		drainLoop:
			for { // drain all pending data in one tick
				select {
				case data := <-sess.Output:
					m.outputBuf[i].Write(data)
					updated = true
				default:
					break drainLoop
				}
			}
			if updated {
				vp := m.viewports[i]
				vp.SetContent(m.outputBuf[i].String())
				vp.GotoBottom()
				m.viewports[i] = vp
			}
		}
		// Always continue ticking for smooth feel
		cmds = append(cmds, m.tickOutput())
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Connecting..."
	}

	var sb strings.Builder

	// Title
	sb.WriteString(titleStyle.Render(" 🗣️ Gossip — TUI Cluster SSH ") + "\n")

	// Hosts list
	var hostContent strings.Builder
	for i, host := range m.hosts {
		cursor := "  "
		style := hostStyle
		if i == m.selected {
			cursor = "▶ "
			style = selectedStyle
		}
		hostContent.WriteString(cursor + style.Render(host) + "\n")
	}

	// Set the content
	m.hostListVP.SetContent(hostContent.String())

	// Scroll to make the selected item visible (auto-center-ish)
	targetLine := m.selected // each host = 1 line
	halfHeight := m.hostListVP.Height / 2
	if targetLine > halfHeight {
		m.hostListVP.SetYOffset(targetLine - halfHeight)
	} else {
		m.hostListVP.SetYOffset(0)
	}
	// Alternative: always scroll to bottom if you prefer tail behavior
	// m.hostListVP.GotoBottom()

	sb.WriteString("\n" + headerStyle.Render("Hosts") + "\n")
	sb.WriteString(viewportStyle.Render(m.hostListVP.View()) + "\n")

	// Status
	mode := "Single"
	modeColor := lipgloss.Color("#FF5555")
	if m.broadcast {
		mode = "Broadcast"
		modeColor = lipgloss.Color("#50FA7B")
	}
	status := fmt.Sprintf("Mode: %s • Selected: %d/%d",
		lipgloss.NewStyle().Foreground(modeColor).Render(mode),
		m.selected+1,
		len(m.hosts),
	)
	sb.WriteString(statusStyle.Render(status))

	// Help
	sb.WriteString(helpStyle.Render("j/k: navigate • b: toggle broadcast • q: quit"))

	return sb.String()
}
