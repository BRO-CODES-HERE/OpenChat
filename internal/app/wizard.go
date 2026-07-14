package app

import (
	"errors"
	"os"
	"os/user"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/BRO-CODES-HERE/OpenChat/internal/storage"
)

var (
	ErrWizardAborted = errors.New("wizard aborted")

	// Wizard specific styles
	wizardTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("62")).
				Padding(0, 1).
				MarginBottom(1)

	wizardBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Width(65)

	wizardInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86")).
				Padding(0, 1).
				Width(40)

	wizardActiveOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	wizardNormalOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("246"))

	wizardHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)
)

type wizardStep int

const (
	stepUsername wizardStep = iota
	stepMode
	stepAddress
	stepP2P
	stepStorage
	stepPassphrase
	stepDone
)

type wizardModel struct {
	step        wizardStep
	username    string
	isServer    bool
	address     string
	useP2P      bool
	storageMode storage.Mode
	passphrase  string
	
	// Input buffers
	textVal     string
	cursor      string
	activeOpt   int // index of selected option in selection lists
	aborted     bool
}

func initialWizardModel() wizardModel {
	defaultUser := "me"
	if u, err := user.Current(); err == nil {
		defaultUser = u.Username
	} else if envUser := os.Getenv("USER"); envUser != "" {
		defaultUser = envUser
	} else if envWinUser := os.Getenv("USERNAME"); envWinUser != "" {
		defaultUser = envWinUser
	}

	return wizardModel{
		step:        stepUsername,
		username:    defaultUser,
		isServer:    true,
		address:     "127.0.0.1:2222",
		useP2P:      false,
		storageMode: storage.ModeLocal,
		passphrase:  "chatssh",
		textVal:     defaultUser,
		cursor:      "█",
		activeOpt:   0,
	}
}

func (m wizardModel) Init() tea.Cmd {
	return nil
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit

		case "enter":
			return m.nextStep()

		case "backspace":
			if m.isTextInputStep() {
				runes := []rune(m.textVal)
				if len(runes) > 0 {
					m.textVal = string(runes[:len(runes)-1])
				}
			}
			return m, nil

		case "up", "k":
			if !m.isTextInputStep() {
				m.activeOpt = 0
			}
			return m, nil

		case "down", "j":
			if !m.isTextInputStep() {
				m.activeOpt = 1
			}
			return m, nil

		case "tab":
			if !m.isTextInputStep() {
				if m.activeOpt == 0 {
					m.activeOpt = 1
				} else {
					m.activeOpt = 0
				}
			}
			return m, nil

		default:
			if m.isTextInputStep() && len(msg.Runes) > 0 {
				m.textVal += string(msg.Runes)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m wizardModel) isTextInputStep() bool {
	return m.step == stepUsername || m.step == stepAddress || m.step == stepPassphrase
}

func (m wizardModel) nextStep() (tea.Model, tea.Cmd) {
	// Save current step data
	switch m.step {
	case stepUsername:
		m.username = strings.TrimSpace(m.textVal)
		if m.username == "" {
			m.username = "me"
		}
		m.step = stepMode
		m.activeOpt = 0 // default to host

	case stepMode:
		m.isServer = (m.activeOpt == 0)
		if m.isServer {
			m.step = stepP2P
			m.activeOpt = 1 // default to direct TCP
		} else {
			m.step = stepAddress
			m.textVal = m.address
		}

	case stepAddress:
		m.address = strings.TrimSpace(m.textVal)
		if m.address == "" {
			m.address = "127.0.0.1:2222"
		}
		m.step = stepP2P
		m.activeOpt = 1 // default to direct TCP

	case stepP2P:
		m.useP2P = (m.activeOpt == 0)
		m.step = stepStorage
		m.activeOpt = 0 // default to local encrypted

	case stepStorage:
		if m.activeOpt == 0 {
			m.storageMode = storage.ModeLocal
			m.step = stepPassphrase
			m.textVal = m.passphrase
		} else {
			m.storageMode = storage.ModeGhost
			m.step = stepDone
			return m, tea.Quit
		}

	case stepPassphrase:
		m.passphrase = m.textVal
		m.step = stepDone
		return m, tea.Quit
	}

	return m, nil
}

func (m wizardModel) View() string {
	if m.aborted {
		return "\nSetup Wizard aborted. Goodbye!\n\n"
	}

	if m.step == stepDone {
		return "\nSetup complete! Starting Chat...\n\n"
	}

	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render(" OpenChat Setup Wizard ") + "\n\n")

	switch m.step {
	case stepUsername:
		content.WriteString("Step 1: Choose Display Name\n")
		content.WriteString("Enter the username other participants will see in the chat room:\n\n")
		content.WriteString(wizardInputStyle.Render(m.textVal + m.cursor) + "\n")
		content.WriteString(wizardHelpStyle.Render("Press Enter to continue • Esc to quit"))

	case stepMode:
		content.WriteString("Step 2: Choose Connection Mode\n")
		content.WriteString("Select whether you want to host a new room or connect to an existing room:\n\n")
		
		optHost := "  [ ] Host a new Chat Room (Server)"
		optJoin := "  [ ] Connect to an existing room (Client)"
		if m.activeOpt == 0 {
			optHost = wizardActiveOptionStyle.Render("  [▸] Host a new Chat Room (Server)")
			optJoin = wizardNormalOptionStyle.Render(optJoin)
		} else {
			optHost = wizardNormalOptionStyle.Render(optHost)
			optJoin = wizardActiveOptionStyle.Render("  [▸] Connect to an existing room (Client)")
		}
		
		content.WriteString(optHost + "\n" + optJoin + "\n")
		content.WriteString(wizardHelpStyle.Render("Use Up/Down or Tab to navigate • Enter to select • Esc to quit"))

	case stepAddress:
		content.WriteString("Step 3: Enter Host Address\n")
		content.WriteString("Enter the IP address and port of the room host (e.g. 192.168.1.100:2222):\n\n")
		content.WriteString(wizardInputStyle.Render(m.textVal + m.cursor) + "\n")
		content.WriteString(wizardHelpStyle.Render("Press Enter to continue • Esc to quit"))

	case stepP2P:
		content.WriteString("Step 4: Network Transport\n")
		content.WriteString("Choose the connection transport method:\n\n")

		optP2P := "  [ ] P2P Mode (AutoNAT hole-punching, NAT traversal)"
		optTCP := "  [ ] Direct TCP (Requires port forwarding if connecting worldwide)"
		if m.activeOpt == 0 {
			optP2P = wizardActiveOptionStyle.Render("  [▸] P2P Mode (AutoNAT hole-punching, NAT traversal)")
			optTCP = wizardNormalOptionStyle.Render(optTCP)
		} else {
			optP2P = wizardNormalOptionStyle.Render(optP2P)
			optTCP = wizardActiveOptionStyle.Render("  [▸] Direct TCP (Requires port forwarding if connecting worldwide)")
		}

		content.WriteString(optP2P + "\n" + optTCP + "\n")
		content.WriteString(wizardHelpStyle.Render("Use Up/Down or Tab to navigate • Enter to select • Esc to quit"))

	case stepStorage:
		content.WriteString("Step 5: Storage Selection\n")
		content.WriteString("Choose how your chat messages are persisted:\n\n")

		optLocal := "  [ ] Local Encrypted Storage (Saves encrypted logs to database)"
		optGhost := "  [ ] Ghost Mode (RAM cache only, scrubbed and erased on exit)"
		if m.activeOpt == 0 {
			optLocal = wizardActiveOptionStyle.Render("  [▸] Local Encrypted Storage (Saves encrypted logs to database)")
			optGhost = wizardNormalOptionStyle.Render(optGhost)
		} else {
			optLocal = wizardNormalOptionStyle.Render(optLocal)
			optGhost = wizardActiveOptionStyle.Render("  [▸] Ghost Mode (RAM cache only, scrubbed and erased on exit)")
		}

		content.WriteString(optLocal + "\n" + optGhost + "\n")
		content.WriteString(wizardHelpStyle.Render("Use Up/Down or Tab to navigate • Enter to select • Esc to quit"))

	case stepPassphrase:
		content.WriteString("Step 6: Set Database Passphrase\n")
		content.WriteString("Provide a password to encrypt your local database logs:\n\n")
		
		// Star the password for security
		var starred string
		for range m.textVal {
			starred += "*"
		}
		content.WriteString(wizardInputStyle.Render(starred + m.cursor) + "\n")
		content.WriteString(wizardHelpStyle.Render("Press Enter to start OpenChat • Esc to quit"))
	}

	return "\n" + wizardBoxStyle.Render(content.String()) + "\n"
}

// RunWizard launches the setup TUI and returns configured Options.
func RunWizard() (Options, error) {
	p := tea.NewProgram(initialWizardModel())
	m, err := p.Run()
	if err != nil {
		return Options{}, err
	}

	model := m.(wizardModel)
	if model.aborted {
		return Options{}, ErrWizardAborted
	}

	// Map Wizard fields back to App Options
	runMode := "server"
	if !model.isServer {
		runMode = "connect"
	}

	opts := Options{
		Mode:       runMode,
		Addr:       model.address,
		UseP2P:     model.useP2P,
		Storage:    model.storageMode,
		Passphrase: model.passphrase,
		LocalUser:  model.username,
		ListenPort: 4001,
		Bootnodes:  DefaultBootnodes(),
	}

	// Host a public room by default when starting server in wizard mode
	if model.isServer {
		opts.Room = true
		opts.RoomName = "public"
		// If using direct TCP server, default to port 2222 listener
		if !model.useP2P {
			opts.Addr = ":2222"
		}
	}

	return opts, nil
}
