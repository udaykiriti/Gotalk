package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"

	"gotalk/internal/models"
)

var (
	host     = flag.String("host", "localhost:8080", "The chat server host")
	roomName = flag.String("room", "general", "The room to join")
	username = flag.String("user", "Guest", "Your username")
)

type errMsg error

// connectionMsg is sent when we successfully connect
type connectionMsg *websocket.Conn

// incomingMsg is sent when a message arrives from the server
type incomingMsg models.Message

type model struct {
	viewport  viewport.Model
	textarea  textarea.Model
	err       error
	conn      *websocket.Conn
	messages  []string
	width     int
	height    int
	connected bool
	status    string
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Type a message... (/clear clears local chat)"
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(
		lipgloss.NewStyle().Foreground(lipgloss.Color("247")).Render("Connecting to server..."),
	)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:  ta,
		messages:  []string{},
		viewport:  vp,
		err:       nil,
		status:    "Connecting...",
		connected: false,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, connectToWebsocket)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 3
		footerHeight := 3

		inputWidth := msg.Width - 2
		if inputWidth < 20 {
			inputWidth = 20
		}
		m.textarea.SetWidth(inputWidth)

		viewportHeight := msg.Height - m.textarea.Height() - headerHeight - footerHeight
		if viewportHeight < 3 {
			viewportHeight = 3
		}
		m.viewport.Width = msg.Width
		m.viewport.Height = viewportHeight
		m.renderMessages()

	case connectionMsg:
		m.conn = msg
		m.connected = true
		m.err = nil
		m.status = fmt.Sprintf("Connected to %s as %s", *roomName, *username)
		m.addSystemLine("Connected. Start chatting.")
		return m, waitForIncomingMessage(m.conn)

	case incomingMsg:
		if msg.Type == models.TypeNotification {
			m.addSystemLine(msg.Content)
		} else if msg.Type == models.TypeUserList {
			// Ignore user list updates in TUI for now
			return m, waitForIncomingMessage(m.conn)
		} else {
			timestamp := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(time.Now().Format("15:04"))
			senderColor := lipgloss.Color("2")
			if msg.User == *username {
				senderColor = lipgloss.Color("6")
			}
			line := fmt.Sprintf("%s %s: %s",
				timestamp,
				lipgloss.NewStyle().Foreground(senderColor).Bold(true).Render(msg.User),
				msg.Content,
			)
			m.messages = append(m.messages, line)
		}
		m.renderMessages()
		return m, waitForIncomingMessage(m.conn)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.conn != nil {
				_ = m.conn.Close()
			}
			return m, tea.Quit
		case tea.KeyEnter:
			text := strings.TrimSpace(m.textarea.Value())
			if text == "" {
				return m, tea.Batch(tiCmd, vpCmd)
			}

			if text == "/clear" {
				m.messages = nil
				m.renderMessages()
				m.textarea.Reset()
				m.status = "Chat cleared"
				return m, tea.Batch(tiCmd, vpCmd)
			}

			if m.conn != nil && m.connected {
				err := m.conn.WriteMessage(websocket.TextMessage, []byte(text))
				if err != nil {
					m.err = err
					m.status = "Disconnected"
					m.connected = false
					_ = m.conn.Close()
					m.conn = nil
					return m, nil
				}
				m.textarea.Reset()
			} else {
				m.status = "Not connected. Press r to reconnect."
			}
		case tea.KeyRunes:
			if len(msg.Runes) == 1 && (msg.Runes[0] == 'r' || msg.Runes[0] == 'R') {
				if !m.connected {
					m.status = "Reconnecting..."
					m.err = nil
					return m, connectToWebsocket
				}
			}
		}

	case errMsg:
		m.err = msg
		m.connected = false
		m.status = "Disconnected. Press r to reconnect."
		if m.conn != nil {
			_ = m.conn.Close()
			m.conn = nil
		}
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("25")).
		Padding(0, 1).
		Render(fmt.Sprintf("GoTalk  room:%s  user:%s", *roomName, *username))

	connState := "offline"
	connColor := lipgloss.Color("1")
	if m.connected {
		connState = "online"
		connColor = lipgloss.Color("2")
	}
	statusLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Padding(0, 1).
		Render(fmt.Sprintf("Status: %s  |  %s", m.status, lipgloss.NewStyle().Foreground(connColor).Bold(true).Render(connState)))

	helpLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(0, 1).
		Render("Enter send • /clear clear chat • r reconnect • esc/ctrl+c quit")

	charCount := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(fmt.Sprintf("%d/%d", len([]rune(m.textarea.Value())), m.textarea.CharLimit))

	inputBlock := m.textarea.View() + "\n" + charCount

	if m.err != nil {
		errText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Padding(0, 1).
			Render("Error: " + m.err.Error())
		return lipgloss.JoinVertical(lipgloss.Left, title, statusLine, m.viewport.View(), errText, inputBlock, helpLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, statusLine, m.viewport.View(), inputBlock, helpLine)
}

// Commands

func connectToWebsocket() tea.Msg {
	u := url.URL{
		Scheme: "ws",
		Host:   *host,
		Path:   "/ws",
		RawQuery: fmt.Sprintf("room=%s&user=%s",
			url.QueryEscape(*roomName),
			url.QueryEscape(*username),
		),
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return errMsg(err)
	}
	return connectionMsg(conn)
}

func waitForIncomingMessage(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return errMsg(err)
		}

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			return errMsg(err)
		}
		return incomingMsg(msg)
	}
}

func (m *model) addSystemLine(content string) {
	timestamp := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(time.Now().Format("15:04"))
	line := fmt.Sprintf("%s %s",
		timestamp,
		lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true).Render(content),
	)
	m.messages = append(m.messages, line)
}

func (m *model) renderMessages() {
	m.viewport.SetContent(strings.Join(m.messages, "\n"))
	m.viewport.GotoBottom()
}

func main() {
	flag.Parse()
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
