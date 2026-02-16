# GoTalk

![Go Version](https://img.shields.io/badge/Go-1.20%2B-blue?style=flat-square)
![Platform](https://img.shields.io/badge/Platform-Windows%7CLinux%7CmacOS-lightgrey?style=flat-square)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

> **GoTalk** is a lightweight, real-time chat server and terminal client built with Go for instant group communication.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Quick Start (Linux/macOS)](#quick-start-linuxmacos)
  - [Server](#server)
  - [Web Client](#web-client)
  - [Client](#client)
- [Configuration](#configuration)
- [Development](#development)
- [Project Structure](#project-structure)
- [Troubleshooting](#troubleshooting)
- [Technologies](#technologies)
- [License](#license)

---

## Overview

GoTalk demonstrates the power of Go's concurrency model and WebSockets to create a seamless chat experience. It includes a robust server capable of handling multiple concurrent connections and a stylish Terminal User Interface (TUI) client powered by Bubble Tea.

## Features

- **Real-time Messaging**: Instant communication using WebSockets.
- **Multi-Room Support**: Create or join any room (e.g., `general`, `dev`, `random`).
- **User Identity**: Simple username-based identity system.
- **System Notifications**: Broadcasts when users join or leave the room.
- **Rich TUI Client**: A beautiful terminal interface with viewport scrolling and input handling.
- **Web + TUI Clients**: Chat from either browser UI or terminal UI.
- **Robustness**: Graceful server shutdown, connection handling, and safer message framing.

## Installation

### Prerequisites

- **Go 1.20** or higher installed on your machine.

### Clone the Repository

```bash
git clone https://github.com/udaykiriti/Gotalk.git
cd Gotalk
go mod download
```

## Usage

You can run the project using the provided helper script or manual commands.

### Quick Start (Windows)

Simply run the `gotalk.bat` script:

```powershell
.\gotalk.bat
```

Select **Option 1** to start the server, then open a new terminal and select **Option 2** to start a client.

---

### Quick Start (Linux/macOS)

Run the launcher script:

```bash
./run.sh
```

Select **Option 1** to start the server, then open a second terminal and run **Option 2** for one or more clients.

### Server

Start the chat server on the default port `8080`.

```bash
go run cmd/server/main.go
```

> [!NOTE]
> The server runs on `localhost:8080` by default.

### Web Client

Start the server, then open this URL in your browser:

```text
http://localhost:8080/
```

Enter username and room, click **Connect**, then start chatting.

### Client

Connect to the server using the CLI client. You can open multiple terminal windows to simulate different users.

```bash
go run cmd/client/main.go -user="Alice" -room="general"
```

> [!TIP]
> **Pro Tip**: Use the `-room` flag to create private channels!

## Configuration

### Server Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | The HTTP network address to bind to. |

### Client Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-host` | `localhost:8080` | The address of the chat server. |
| `-user` | `Guest` | Your username for the session. |
| `-room` | `general` | The chat room to join. |

## Development

Run tests:

```bash
go test ./...
```

Run static checks:

```bash
go vet ./...
```

Run end-to-end websocket smoke test (starts server, dials 2 clients, verifies delivery):

```bash
go run scripts/smoke_e2e.go
```

## Project Structure

```text
Gotalk/
├── cmd/
│   ├── server/       # Server entry point
│   └── client/       # Client entry point (TUI)
├── internal/
│   ├── ws/           # WebSocket Hub & Client logic
│   ├── models/       # Shared data structures (Message types)
│   └── handlers/     # HTTP handlers
├── web/              # Static frontend resources
├── go.mod            # Go module definition
└── LICENSE           # Project License
```

## Troubleshooting

### Port Already in Use
**Error**: `bind: address already in use`
- **Solution**: The server is likely already running. Check your other terminals or kill the process using port 8080.

### Connection Refused
**Error**: `dial tcp ... connect: connection refused`
- **Solution**: Ensure the server is running *before* starting the client.

### Browser Page Not Loading
**Error**: `/` returns not found or empty page
- **Solution**: Make sure `web/index.html` exists and run the server from the project root, or deploy with a `web/` folder next to the server binary.

## Technologies

- [Go](https://go.dev/) - The programming language used.
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket implementation for Go.
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - A powerful little TUI framework.
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for nice terminal layouts.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
