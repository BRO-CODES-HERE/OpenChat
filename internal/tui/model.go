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
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	feedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1)



	userColors = []string{
		"197", // Vibrant pink-red
		"203", // Bright red-orange
		"208", // Bright orange
		"214", // Gold/yellow
		"220", // Warm yellow
		"118", // Lime green
		"76",  // Bright green
		"46",  // Mint green
		"43",  // Teal
		"39",  // Sky blue
		"33",  // Royal blue
		"99",  // Vibrant purple/indigo
		"135", // Lavender/violet
		"213", // Hot pink
	}
)

func getUsernameStyle(username string) lipgloss.Style {
	if username == "system" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Italic(true)
	}
	var hash int
	for _, char := range username {
		hash += int(char)
	}
	color := userColors[hash%len(userColors)]
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
}

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
	sub       <-chan chat.Message
}

type msgReceived struct{ msg chat.Message }


// New creates a chat TUI model.
func New(cfg Config) Model {
	m := Model{
		cfg:      cfg,
		messages: cfg.Hub.Messages(),
		sub:      cfg.Hub.Subscribe(),
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
	return func() tea.Msg {
		msg := <-m.sub
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
			runes := []rune(m.input)
			if len(runes) > 0 {
				m.input = string(runes[:len(runes)-1])
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

	tsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	msgContentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("253"))

	for _, msg := range m.messages[start:] {
		ts := msg.Timestamp.Format("15:04:05")
		tsStr := tsStyle.Render("[" + ts + "]")

		if msg.Sender == "system" {
			sysContent := lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true).Render("★ " + msg.Content)
			lines = append(lines, fmt.Sprintf("%s  %s", tsStr, sysContent))
			continue
		}

		senderStyle := getUsernameStyle(msg.Sender)
		if msg.Sender == m.cfg.LocalUser {
			senderStyle = senderStyle.Underline(true)
		}
		senderStr := senderStyle.Render(msg.Sender)
		separator := sepStyle.Render("│")
		contentStr := msgContentStyle.Render(msg.Content)

		lines = append(lines, fmt.Sprintf("%s  %s %s %s", tsStr, senderStr, separator, contentStr))
	}
	if len(lines) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No messages yet. Type below and press Enter."))
	}

	// Render title with full width and a nice emoji icon
	titleStr := titleStyle.Width(m.width).Render(" 💬  " + m.cfg.Title)

	feed := feedStyle.
		Width(m.width - 4).
		Height(feedHeight).
		Render(strings.Join(lines, "\n"))

	// Format input box with placeholder text if empty
	var inputVal string
	if m.input == "" {
		inputVal = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Type a message... (Esc to quit)")
	} else {
		inputVal = m.input + "█"
	}
	input := inputStyle.
		Width(m.width - 4).
		Render("> " + inputVal)

	// Format status bar as powerline-style segmented pills
	var statusBlocks []string
	segments := strings.Split(m.cfg.Status, " | ")
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		var style lipgloss.Style
		if strings.Contains(seg, "Ghost Mode") {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("208")). // Amber/orange
				Foreground(lipgloss.Color("16")).
				Bold(true).
				Padding(0, 1)
		} else if strings.Contains(seg, "Local Encrypted") {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("76")).  // Vibrant green
				Foreground(lipgloss.Color("16")).
				Bold(true).
				Padding(0, 1)
		} else if strings.HasPrefix(seg, "key:") {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("33")).  // Blue
				Foreground(lipgloss.Color("255")).
				Bold(true).
				Padding(0, 1)
		} else if strings.HasPrefix(seg, "peers:") {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("99")).  // Purple
				Foreground(lipgloss.Color("255")).
				Bold(true).
				Padding(0, 1)
		} else {
			// address/listen/p2p info
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("237")). // Dark gray
				Foreground(lipgloss.Color("86")).  // Cyan
				Bold(true).
				Padding(0, 1)
		}
		statusBlocks = append(statusBlocks, style.Render(seg))
	}
	status := lipgloss.JoinHorizontal(lipgloss.Top, statusBlocks...)

	return fmt.Sprintf("%s\n%s\n%s\n%s", titleStr, feed, input, status)
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
