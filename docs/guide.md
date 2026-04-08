# GoTalk User Guide

Welcome to GoTalk! This guide will help you get started with the server and chat clients on different operating systems.

## Magic Features (Zero-Config)

GoTalk v2 is designed to be as simple as possible.

### 1. Auto-Discovery (mDNS)
You no longer need to type IP addresses! If you start the server, any GoTalk client on the same Wi-Fi will automatically find it. 
- Just run the client without the `-host` flag: `go run ./cmd/client`
- The client will scan and connect to your server instantly.

### 2. Interactive Room Picker
Forgotten the room name? No problem.
- Run the client without a `-room` flag.
- GoTalk will fetch all active rooms from the server and show you a beautiful **Interactive Menu** to pick from.

### 3. Smart Username
GoTalk automatically detects your computer's username so you can start chatting without typing a name every time.

## Getting Started on Windows

### 1. Simple Launcher
The easiest way to start on Windows is using the included batch file:
- Double-click `gotalk.bat`.
- Select **[1]** to start the Server or **[2]** to start the TUI Client.

### 2. Manual Commands
If you prefer using the terminal (PowerShell or Command Prompt):

**Start Server:**
```powershell
go run ./cmd/server
```
*Note: Look for the **QR code** that appears in the terminal—you can scan it with your phone to join!*

**Start TUI Client:**
```powershell
go run ./cmd/client -user="YourName" -room="general"
```

### 3. Firewall Setup (Optional)
If your phone cannot connect after scanning the QR code, run this in **Admin PowerShell**:
```powershell
New-NetFirewallRule -DisplayName "GoTalk Server" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
```

---

## Getting Started on Linux / macOS

### 1. Simple Launcher
- Open your terminal.
- Run the bash script:
  ```bash
  chmod +x run.sh
  ./run.sh
  ```
- Select **1** for Server or **2** for Client.

### 2. Using Makefile
If you have `make` installed:
- **Start Server**: `make run-server`
- **Start Client**: `make run-client USER=YourName ROOM=general`

---

## How to Use GoTalk

### 1. Web Interface
1. Start the server and visit `http://localhost:8080` in your browser.
2. Enter your **Username** and the **Room** you want to join.
3. Click **Connect**.
4. **Invite Others**: Click the "Invite" button in the header to show a QR code that peers can scan to join your room.

### 2. Mobile Access
1. Ensure your phone is on the **same Wi-Fi** as your computer.
2. Scan the **QR Code** displayed in your server's terminal.
3. The mobile-responsive UI will automatically load in your phone's browser.

## Practical Examples

Here are some common ways to use the terminal client:

### Scenario 1: Starting a Local Chat
1. Start the server (Terminal 1):
   `go run ./cmd/server`
2. Join as Alice (Terminal 2):
   `go run ./cmd/client -user="Alice" -room="vibe-check"`
3. Join as Bob (Terminal 3):
   `go run ./cmd/client -user="Bob" -room="vibe-check"`

### Scenario 2: Connecting to a Specific Server
If your friend is hosting the server on their IP (e.g., `192.168.1.50`):
```bash
go run ./cmd/client -host="192.168.1.50:8080" -user="YourName" -room="general"
```

### Scenario 3: Quick Room Join (Default Guest)
If you just want to jump in quickly without setting a name:
```bash
go run ./cmd/client -room="secret-room"
```

### 3. TUI (Terminal) Client Reference
The TUI client is a high-performance terminal interface built with BubbleTea.

**Command Line Flags:**
- `-host`: The chat server host (default: `localhost:8080`)
- `-user`: Your display name (default: `Guest`)
- `-room`: The chat room to join (default: `general`)

**Keyboard Shortcuts:**
- **Enter**: Send message.
- **Up/Down Arrows**: Scroll through the chat history.
- **PgUp/PgDn**: Fast scroll through long conversations.
- **R / r**: Reconnect to the server if disconnected.
- **Esc / Ctrl+C**: Quit the application safely.

**Special Commands:**
- Type **/clear** and press Enter to clear your local chat screen (this does not delete messages on the server).

## Troubleshooting & Tips

### Connections
- **Same Network**: Ensure your mobile phone and laptop are on the same Wi-Fi network for the QR code to work.
- **IP Changes**: If your computer's IP changes (e.g., you change Wi-Fi networks), you must restart the server to generate a new QR code.

### Layout Issues
- If the TUI looks distorted, try resizing your terminal window. The interface is designed to adapt to your terminal size dynamically.
- For the best experience on Windows, use **Windows Terminal** or **PowerShell**.
