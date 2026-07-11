package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	feedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	msgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selfStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)
)

// Config configures the chat TUI.
type Config struct {
	Title            string
	LocalUser        string
	Status           string
	Hub              *chat.Hub
	OnQuit           func()
	VerifyScreen     string
}

// Model is the Bubble Tea chat model.
type Model struct {
	cfg       Config
	messages  []chat.Message
	input     string
	width     int
	height    int
	quitting  bool
	verifyMsg string
	awaiting  bool
}

type msgReceived struct{ msg chat.Message }
type windowSizeMsg struct{ width, height int }

// New creates a chat TUI model.
func New(cfg Config) Model {
	m := Model{
		cfg:      cfg,
		messages: cfg.Hub.Messages(),
	}
	if cfg.VerifyScreen != "" {
		m.verifyMsg = cfg.VerifyScreen
		m.awaiting = true
	}
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.listen(),
		tea.EnterAltScreen,
	)
}

func (m Model) listen() tea.Cmd {
	hub := m.cfg.Hub
	return func() tea.Msg {
		ch := hub.Subscribe()
		msg := <-ch
		return msgReceived{msg}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case msgReceived:
		m.messages = append(m.messages, msg.msg)
		return m, m.listen()

	case tea.KeyMsg:
		if m.awaiting {
			switch msg.String() {
			case "y", "Y":
				m.awaiting = false
				m.verifyMsg = ""
			case "n", "N", "ctrl+c":
				m.quitting = true
				if m.cfg.OnQuit != nil {
					m.cfg.OnQuit()
				}
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			if m.cfg.OnQuit != nil {
				m.cfg.OnQuit()
			}
			return m, tea.Quit
		case "enter":
			line := strings.TrimSpace(m.input)
			if line != "" {
				m.cfg.Hub.Send(m.cfg.LocalUser, line)
				m.input = ""
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(msg.Runes) > 0 {
				m.input += string(msg.Runes)
			}
		}
	}

	return m, nil
}

// SetVerification shows a host-key verification prompt.
func (m *Model) SetVerification(text string) {
	m.verifyMsg = text
	m.awaiting = true
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.awaiting && m.verifyMsg != "" {
		return feedStyle.Width(m.width - 4).Render(m.verifyMsg)
	}

	feedHeight := m.height - 6
	if feedHeight < 4 {
		feedHeight = 4
	}

	var lines []string
	start := 0
	if len(m.messages) > feedHeight {
		start = len(m.messages) - feedHeight
	}
	for _, msg := range m.messages[start:] {
		ts := msg.Timestamp.Format("15:04:05")
		style := msgStyle
		if msg.Sender == m.cfg.LocalUser {
			style = selfStyle
		}
		lines = append(lines, style.Render(fmt.Sprintf("[%s] %s: %s", ts, msg.Sender, msg.Content)))
	}
	if len(lines) == 0 {
		lines = append(lines, statusStyle.Render("No messages yet. Type below and press Enter."))
	}

	title := titleStyle.Render(m.cfg.Title)
	feed := feedStyle.
		Width(m.width - 4).
		Height(feedHeight).
		Render(strings.Join(lines, "\n"))
	input := inputStyle.
		Width(m.width - 4).
		Render("> " + m.input + "█")
	status := statusStyle.Render(m.cfg.Status)

	return fmt.Sprintf("%s\n%s\n%s\n%s", title, feed, input, status)
}

// Run starts the Bubble Tea program.
func Run(cfg Config) error {
	p := tea.NewProgram(New(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// FormatMessage formats a message for wire transport.
func FormatMessage(msg chat.Message) string {
	return fmt.Sprintf("%s|%s|%s", msg.Sender, msg.Timestamp.Format(time.RFC3339Nano), msg.Content)
}

// ParseMessage parses a wire-format message.
func ParseMessage(line string) (chat.Message, error) {
	parts := strings.SplitN(line, "|", 3)
	if len(parts) != 3 {
		return chat.Message{}, fmt.Errorf("invalid message format")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[1])
	if err != nil {
		ts = time.Now()
	}
	return chat.Message{
		Sender:    parts[0],
		Timestamp: ts,
		Content:   parts[2],
	}, nil
}
