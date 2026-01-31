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
	viewport    viewport.Model
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
	conn        *websocket.Conn
	messages    []string
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to GoTalk!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
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
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - 2 // Subtract header/footer space
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	case connectionMsg:
		m.conn = msg
		return m, waitForIncomingMessage(m.conn)

	case incomingMsg:
		// Format the incoming message
		timestamp := time.Now().Format("15:04")
		var styledMsg string

		if msg.Type == models.TypeNotification {
			styledMsg = fmt.Sprintf("%s %s",
				lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(timestamp),
				lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true).Render(msg.Content),
			)
		} else {
			senderColor := lipgloss.Color("2") // Green
			if msg.User == *username {
				senderColor = lipgloss.Color("6") // Cyan
			}
			styledMsg = fmt.Sprintf("%s %s: %s",
				lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(timestamp),
				lipgloss.NewStyle().Foreground(senderColor).Bold(true).Render(msg.User),
				msg.Content,
			)
		}

		m.messages = append(m.messages, styledMsg)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, waitForIncomingMessage(m.conn)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.conn != nil && m.textarea.Value() != "" {
				// Send message
				err := m.conn.WriteMessage(websocket.TextMessage, []byte(m.textarea.Value()))
				if err != nil {
					m.err = err
					return m, nil
				}
				m.textarea.Reset()
			}
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
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

func main() {
	flag.Parse()
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
