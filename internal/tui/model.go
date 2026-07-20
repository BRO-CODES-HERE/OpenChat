package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/storage"
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

func (m Model) getUsernameStyle(username string) lipgloss.Style {
	if username == "system" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Italic(true)
	}

	colorIndex := -1
	if m.cfg.Store != nil {
		if idx, err := m.cfg.Store.GetUserColor(username); err == nil {
			colorIndex = idx
		}
	}

	if colorIndex < 0 || colorIndex >= len(userColors) {
		var hash int
		for _, char := range username {
			hash += int(char)
		}
		colorIndex = hash % len(userColors)

		if m.cfg.Store != nil {
			_ = m.cfg.Store.SaveUserColor(username, colorIndex)
		}
	}

	color := userColors[colorIndex]
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
}

// RoomHost defines the interface required by the TUI to query room status and route DMs.
type RoomHost interface {
	ClientIDs() []string
	GetClientConn(id string) io.ReadWriteCloser
}

// Config configures the chat TUI.
type Config struct {
	Title        string
	LocalUser    string
	Status       string
	Hub          *chat.Hub
	Store        *storage.Store
	RoomHost     RoomHost
	OnQuit       func()
	VerifyScreen string
}

// Model is the Bubble Tea chat model.
type Model struct {
	cfg            Config
	messages       []chat.Message
	input          string
	width          int
	height         int
	quitting       bool
	verifyMsg      string
	awaiting       bool
	sub            <-chan chat.Message
	peerCount      int
	showTimestamps bool
	scrollOffset   int
	unreadCount    int
	lastSent       string
}

type msgReceived struct{ msg chat.Message }


// New creates a chat TUI model.
func New(cfg Config) Model {
	m := Model{
		cfg:            cfg,
		messages:       cfg.Hub.Messages(),
		sub:            cfg.Hub.Subscribe(),
		showTimestamps: true,
		scrollOffset:   0,
		unreadCount:    0,
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
		if msg.msg.Sender == "system:count" {
			var count int
			if _, err := fmt.Sscanf(msg.msg.Content, "%d", &count); err == nil {
				m.peerCount = count
			}
			return m, m.listen()
		}
		m.messages = append(m.messages, msg.msg)

		var cmds []tea.Cmd
		cmds = append(cmds, m.listen())

		if m.scrollOffset > 0 {
			m.unreadCount++
		}

		contentLower := strings.ToLower(msg.msg.Content)
		mentionLower := strings.ToLower("@" + m.cfg.LocalUser)
		if msg.msg.Sender != m.cfg.LocalUser && strings.Contains(contentLower, mentionLower) {
			cmds = append(cmds, func() tea.Msg {
				fmt.Print("\a")
				return nil
			})
		}
		return m, tea.Batch(cmds...)

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
			m.input = ""
			m.scrollOffset = 0
			m.unreadCount = 0

			if line != "" {
				m.lastSent = line
				if strings.HasPrefix(line, "/") {
					cmdMsg, isLocal := m.handleSlashCommand(line)
					if isLocal {
						if cmdMsg.Content != "" {
							m.messages = append(m.messages, cmdMsg)
						}
						return m, nil
					}
				}
				m.cfg.Hub.Send(m.cfg.LocalUser, line)
			}
		case "ctrl+n", "alt+enter":
			m.input += "\n"
			return m, nil
		case "up":
			if m.input == "" && m.lastSent != "" {
				m.input = m.lastSent
			}
			return m, nil
		case "backspace":
			runes := []rune(m.input)
			if len(runes) > 0 {
				m.input = string(runes[:len(runes)-1])
			}
		case "ctrl+t":
			m.showTimestamps = !m.showTimestamps
			return m, nil
		case "pageup":
			feedHeight := m.height - 6
			if feedHeight < 4 {
				feedHeight = 4
			}
			maxScroll := len(m.messages) - feedHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.scrollOffset += feedHeight / 2
			if m.scrollOffset > maxScroll {
				m.scrollOffset = maxScroll
			}
			return m, nil
		case "pagedown":
			feedHeight := m.height - 6
			if feedHeight < 4 {
				feedHeight = 4
			}
			m.scrollOffset -= feedHeight / 2
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
				m.unreadCount = 0
			}
			return m, nil
		default:
			if len(msg.Runes) > 0 {
				m.input += string(msg.Runes)
			}
		}
	}

	return m, nil
}

// handleSlashCommand parses and processes in-chat slash commands.
// Returns a status message (if any) and a boolean indicating if the command was handled locally.
func (m *Model) handleSlashCommand(line string) (chat.Message, bool) {
	parts := strings.SplitN(line, " ", 3)
	cmd := parts[0]

	switch cmd {
	case "/clear":
		m.messages = nil
		return chat.Message{
			Sender:    "system",
			Timestamp: time.Now(),
			Content:   "Console cleared.",
		}, true

	case "/quit":
		m.quitting = true
		if m.cfg.OnQuit != nil {
			m.cfg.OnQuit()
		}
		return chat.Message{}, true

	case "/help":
		helpText := "Available commands:\n" +
			"  /help               - Show this help message\n" +
			"  /clear              - Clear the message screen\n" +
			"  /quit               - Exit OpenChat\n" +
			"  /users              - List all active users in the room\n" +
			"  /dm <user> <msg>    - Send a private message to a user"
		return chat.Message{
			Sender:    "system",
			Timestamp: time.Now(),
			Content:   helpText,
		}, true

	case "/users":
		if m.cfg.RoomHost != nil {
			var active []string
			active = append(active, m.cfg.LocalUser+" (host)")
			for _, cid := range m.cfg.RoomHost.ClientIDs() {
				active = append(active, cid)
			}
			return chat.Message{
				Sender:    "system",
				Timestamp: time.Now(),
				Content:   fmt.Sprintf("Active users: %s", strings.Join(active, ", ")),
			}, true
		}
		return chat.Message{}, false

	case "/dm":
		if len(parts) < 3 {
			return chat.Message{
				Sender:    "system",
				Timestamp: time.Now(),
				Content:   "Usage: /dm <username> <message>",
			}, true
		}
		target := parts[1]
		dmMsg := parts[2]

		if m.cfg.RoomHost != nil {
			if target == m.cfg.LocalUser {
				return chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   "You cannot DM yourself.",
				}, true
			}
			targetConn := m.cfg.RoomHost.GetClientConn(target)
			if targetConn != nil {
				replyToTarget := chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("[DM from %s]: %s", m.cfg.LocalUser, dmMsg),
				}
				lineTarget := FormatMessage(replyToTarget) + "\n"
				_, _ = io.WriteString(targetConn, lineTarget)

				return chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("[DM to %s]: %s", target, dmMsg),
				}, true
			} else {
				return chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("User '%s' not found.", target),
				}, true
			}
		}
		return chat.Message{}, false
	}

	return chat.Message{
		Sender:    "system",
		Timestamp: time.Now(),
		Content:   fmt.Sprintf("Unknown command: %s. Type /help for a list of commands.", cmd),
	}, true
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
	start := len(m.messages) - feedHeight - m.scrollOffset
	if start < 0 {
		start = 0
	}
	end := len(m.messages) - m.scrollOffset
	if end < 0 {
		end = 0
	}
	if end > len(m.messages) {
		end = len(m.messages)
	}

	tsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	msgContentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("253"))
	mentionHighlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")). // Yellow
		Foreground(lipgloss.Color("16")).  // Black text
		Bold(true)

	for _, msg := range m.messages[start:end] {
		var tsStr string
		if m.showTimestamps {
			ts := msg.Timestamp.Format("15:04:05")
			tsStr = tsStyle.Render("["+ts+"]") + "  "
		}

		var lineStr string
		if msg.Sender == "system" {
			sysContent := lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true).Render("★ " + msg.Content)
			lineStr = fmt.Sprintf("%s%s", tsStr, sysContent)
		} else {
			senderStyle := m.getUsernameStyle(msg.Sender)
			if msg.Sender == m.cfg.LocalUser {
				senderStyle = senderStyle.Underline(true)
			}
			senderStr := senderStyle.Render(msg.Sender)
			separator := sepStyle.Render("│")
			contentStr := msgContentStyle.Render(msg.Content)

			lineStr = fmt.Sprintf("%s%s %s %s", tsStr, senderStr, separator, contentStr)
		}

		contentLower := strings.ToLower(msg.Content)
		mentionLower := strings.ToLower("@" + m.cfg.LocalUser)
		if msg.Sender != m.cfg.LocalUser && strings.Contains(contentLower, mentionLower) {
			lineStr = mentionHighlightStyle.Render(lineStr)
		}

		lines = append(lines, lineStr)
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
		if seg == "" || strings.HasPrefix(seg, "peers:") {
			continue
		}

		var style lipgloss.Style
		if strings.Contains(seg, "Ghost") {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("208")). // Amber/orange
				Foreground(lipgloss.Color("16")).
				Bold(true).
				Padding(0, 1)
		} else if strings.Contains(seg, "Encrypted") {
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

	if m.peerCount > 0 {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("99")).  // Purple
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1)
		statusBlocks = append(statusBlocks, style.Render(fmt.Sprintf("peers:%d", m.peerCount)))
	}

	if m.scrollOffset > 0 {
		scrollStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("99")). // Purple
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1)
		statusBlocks = append(statusBlocks, scrollStyle.Render(fmt.Sprintf("history:-%d", m.scrollOffset)))
	}

	if m.unreadCount > 0 {
		unreadStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("196")). // Red
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1)
		statusBlocks = append(statusBlocks, unreadStyle.Render(fmt.Sprintf("↓ %d unread", m.unreadCount)))
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
