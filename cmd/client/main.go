package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/user"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"

	"gotalk/internal/models"
)

type state int

const (
	stateScanning state = iota
	statePicking
	stateChat
)

var (
	host     = flag.String("host", "", "The chat server host (auto-discovery if empty)")
	roomName = flag.String("room", "", "The room to join (picker if empty)")
	username = flag.String("user", "", "Your username (OS user if empty)")
)

type errMsg error

// connectionMsg is sent when we successfully connect
type connectionMsg *websocket.Conn

// incomingMsg is sent when a message arrives from the server
type incomingMsg models.Message

// serverFoundMsg is sent when mDNS finds a server
type serverFoundMsg string

// roomsFetchedMsg is sent when we get the room list
type roomsFetchedMsg map[string]int

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := lipgloss.NewStyle().PaddingLeft(2).Render
	if index == m.Index() {
		fn = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Render
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	state      state
	list       list.Model
	viewport   viewport.Model
	textarea   textarea.Model
	err        error
	conn       *websocket.Conn
	messages   []string
	width      int
	height     int
	connected  bool
	status     string
	discovered bool
}

func initialModel() model {
	// Textarea setup
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	// Viewport setup
	vp := viewport.New(30, 5)
	vp.SetContent(lipgloss.NewStyle().Foreground(lipgloss.Color("247")).Render("Connecting..."))

	// List setup (for picking rooms)
	l := list.New([]list.Item{}, itemDelegate{}, 0, 0)
	l.Title = "Active Rooms"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)

	// Determine initial state
	initialState := stateChat
	if *host == "" {
		initialState = stateScanning
	} else if *roomName == "" {
		initialState = statePicking
	}

	return model{
		state:     initialState,
		textarea:  ta,
		messages:  []string{lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("   Connecting to room...")},
		viewport:  vp,
		list:      l,
		err:       nil,
		status:    "Connecting...",
		connected: false,
	}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{textarea.Blink}

	switch m.state {
	case stateScanning:
		cmds = append(cmds, discoverServer)
	case statePicking:
		cmds = append(cmds, fetchRooms(*host))
	case stateChat:
		cmds = append(cmds, connectToWebsocket)
	}

	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		liCmd tea.Cmd
	)

	// Pre-process keys for navigation
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		switch kmsg.Type {
		case tea.KeyCtrlC:
			if m.conn != nil {
				_ = m.conn.Close()
			}
			return m, tea.Quit
		case tea.KeyEsc:
			if m.state == statePicking {
				m.state = stateScanning
				return m, discoverServer
			}
			if m.state == stateChat {
				if m.conn != nil {
					_ = m.conn.Close()
					m.conn = nil
				}
				m.connected = false
				m.state = statePicking
				return m, fetchRooms(*host)
			}
		}
	}

	// State-specific Updates
	switch m.state {
	case stateScanning:
		switch msg := msg.(type) {
		case serverFoundMsg:
			*host = string(msg)
			m.state = statePicking
			m.status = fmt.Sprintf("Found server: %s", *host)
			return m, fetchRooms(*host)
		case errMsg:
			m.err = msg
			return m, nil
		}

	case statePicking:
		// Handle list updates first, but capture if Enter was pressed
		if kmsg, ok := msg.(tea.KeyMsg); ok && kmsg.Type == tea.KeyEnter {
			selected := m.list.SelectedItem()
			if selected == nil {
				return m, nil // Safety: avoid panic on empty list
			}

			roomItem := selected.(item)
			if strings.HasPrefix(string(roomItem), "+") {
				if *roomName == "" {
					*roomName = "general"
				}
			} else {
				*roomName = strings.Fields(string(roomItem))[0]
			}
			m.state = stateChat
			m.status = "Connecting..."
			return m, connectToWebsocket
		}

		m.list, liCmd = m.list.Update(msg)

		if msg, ok := msg.(roomsFetchedMsg); ok {
			var items []list.Item
			for room, count := range msg {
				items = append(items, item(fmt.Sprintf("%s (%d online)", room, count)))
			}
			items = append(items, item("+ Create or Join Other..."))
			m.list.SetItems(items)
			m.status = "Select a room (Esc to rescan)"
		}

	case stateChat:
		switch msg := msg.(type) {
		case connectionMsg:
			m.conn = msg
			m.connected = true
			m.err = nil
			m.status = "Joined " + *roomName
			m.addSystemLine(fmt.Sprintf("Welcome to %s!", *roomName))
			return m, waitForIncomingMessage(m.conn)

		case incomingMsg:
			if msg.Type == models.TypeNotification {
				m.addSystemLine(msg.Content)
			} else {
				senderColor := lipgloss.Color("2")
				if msg.User == *username {
					senderColor = lipgloss.Color("6")
				}
				line := fmt.Sprintf("%s %s: %s",
					lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(time.Now().Format("15:04")),
					lipgloss.NewStyle().Foreground(senderColor).Bold(true).Render(msg.User),
					msg.Content,
				)
				m.messages = append(m.messages, line)
			}
			m.renderMessages()
			return m, waitForIncomingMessage(m.conn)

		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				text := strings.TrimSpace(m.textarea.Value())
				if text != "" {
					if text == "/clear" {
						m.messages = nil
						m.renderMessages()
					} else if m.conn != nil && m.connected {
						_ = m.conn.WriteMessage(websocket.TextMessage, []byte(text))
					}
					m.textarea.Reset()
				}
				return m, nil
			}
		}
		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)
	}

	// Universal Updates
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)

		m.textarea.SetWidth(msg.Width - 2)
		// Precise height calculation
		headerH := 2
		footerH := 2
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - m.textarea.Height() - headerH - footerH
		m.renderMessages()
	case errMsg:
		m.err = msg
	}

	return m, tea.Batch(tiCmd, vpCmd, liCmd)
}

func (m model) View() string {
	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Background(lipgloss.Color("1")).Foreground(lipgloss.Color("230")).Padding(0, 1).Render(" ERROR "),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(" "+m.err.Error()),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" Press Ctrl+C to quit or Esc to retry"),
		)
	}

	switch m.state {
	case stateScanning:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🔍 Scanning Network..."),
				"",
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Looking for GoTalk servers"),
			),
		)

	case statePicking:
		return lipgloss.JoinVertical(lipgloss.Left,
			m.list.View(),
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1).Render(m.status),
		)

	case stateChat:
		// Title bar
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("25")).Padding(0, 1).Render(fmt.Sprintf("GoTalk v2 | Room: %s | User: %s", *roomName, *username))

		// Status bar
		connState := "ONLINE"
		connColor := lipgloss.Color("2")
		if !m.connected {
			connState = "OFFLINE"
			connColor = lipgloss.Color("1")
		}

		statusLine := lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Padding(0, 1).Render(fmt.Sprintf("Status: %s  |  %s", m.status, lipgloss.NewStyle().Foreground(connColor).Bold(true).Render(connState)))

		// Textarea input
		input := m.textarea.View()

		// Bottom Help HUD
		helpLine := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Padding(0, 1).Render("Enter: Send • Esc: Leave Room • Ctrl+C: Quit")

		return lipgloss.JoinVertical(lipgloss.Left, title, statusLine, m.viewport.View(), input, helpLine)
	}
	return "Initializing..."
}

// Commands

func discoverServer() tea.Msg {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return errMsg(fmt.Errorf("failed to create mDNS resolver: %v", err))
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			// Found a GoTalk server
			if len(entry.AddrIPv4) > 0 {
				entries <- entry // We use the first one found for simplicity
				return
			}
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = resolver.Browse(ctx, "_gotalk._tcp", "local.", entries)
	if err != nil {
		return errMsg(fmt.Errorf("discovery failed: %v", err))
	}

	select {
	case entry := <-entries:
		return serverFoundMsg(fmt.Sprintf("%s:%d", entry.AddrIPv4[0], entry.Port))
	case <-ctx.Done():
		return errMsg(fmt.Errorf("no GoTalk servers found on network"))
	}
}

func fetchRooms(h string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("http://%s/api/rooms", h)
		resp, err := http.Get(url)
		if err != nil {
			return errMsg(err)
		}
		defer resp.Body.Close()

		var rooms map[string]int
		if err := json.NewDecoder(resp.Body).Decode(&rooms); err != nil {
			return errMsg(err)
		}
		return roomsFetchedMsg(rooms)
	}
}

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

	// Detect OS username if not provided
	if *username == "" {
		cur, err := user.Current()
		if err == nil {
			name := cur.Username
			// Clean up Windows domain (e.g., HOST\user -> user)
			if idx := strings.LastIndex(name, "\\"); idx != -1 {
				name = name[idx+1:]
			}
			*username = name
		} else {
			*username = "Guest"
		}
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
