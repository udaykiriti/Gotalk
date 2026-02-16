# GoTalk

[![Go Version](https://img.shields.io/badge/Go-1.20%2B-blue?style=flat-square)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=flat-square)](#launcher-scripts)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Repo Size](https://img.shields.io/github/repo-size/udaykiriti/Gotalk?style=flat-square)](https://github.com/udaykiriti/Gotalk)
[![Last Commit](https://img.shields.io/github/last-commit/udaykiriti/Gotalk?style=flat-square)](https://github.com/udaykiriti/Gotalk/commits/main)
[![Issues](https://img.shields.io/github/issues/udaykiriti/Gotalk?style=flat-square)](https://github.com/udaykiriti/Gotalk/issues)
[![Stars](https://img.shields.io/github/stars/udaykiriti/Gotalk?style=flat-square)](https://github.com/udaykiriti/Gotalk/stargazers)

GoTalk is a lightweight real-time group chat application written in Go. It includes:
- a WebSocket chat server,
- a browser-based web client,
- a terminal UI (TUI) client.

> [!NOTE]
> `web/index.html` is embedded into the Go binary, so serving `/` does not depend on runtime working directory.

## Features

- Real-time messaging over WebSockets
- Multi-room chat support (`general`, `dev`, `random`, etc.)
- Username-based identity per session
- Join/leave system notifications
- Browser and TUI clients
- Graceful server shutdown
- End-to-end smoke test for runtime verification

## Quick Start

### 1. Clone and install dependencies

```bash
git clone https://github.com/udaykiriti/Gotalk.git
cd Gotalk
go mod download
```

### 2. Run the server

```bash
go run ./cmd/server
```

Server default address: `:8080`

> [!TIP]
> Run `make run-server` and `make run-client USER=Alice ROOM=general` for quicker local iteration.

### 3. Connect clients

Web client:
- Open `http://localhost:8080/`

TUI client:

```bash
go run ./cmd/client -user="Alice" -room="general"
```

You can run multiple clients in separate terminals.

> [!IMPORTANT]
> Start the server first, then connect clients. If the server is down, client connections will fail immediately.

## Launcher Scripts

Windows:

```powershell
.\gotalk.bat
```

Linux/macOS:

```bash
./run.sh
```

## Configuration

### Server flags

| Flag | Default | Description |
|---|---|---|
| `-addr` | `:8080` | HTTP address for the server |

### Client flags

| Flag | Default | Description |
|---|---|---|
| `-host` | `localhost:8080` | Chat server host |
| `-user` | `Guest` | Username for this session |
| `-room` | `general` | Room to join |

## Development

### Common commands

```bash
make fmt
make test
make vet
make smoke
```

> [!TIP]
> Run `make smoke` before pushing changes. It catches runtime WebSocket regressions that static checks cannot.

Equivalent direct commands:

```bash
go fmt ./...
go test ./...
go vet ./...
go run scripts/smoke_e2e.go
```

## Project Structure

```text
Gotalk/
├── cmd/
│   ├── server/                 # Server entrypoint
│   └── client/                 # TUI client entrypoint
├── internal/
│   ├── handlers/               # HTTP handlers
│   ├── models/                 # Shared data models
│   └── ws/                     # WebSocket hub/client logic
├── web/
│   ├── index.html              # Web chat UI source
│   └── assets.go               # Embedded web assets
├── scripts/
│   ├── smoke_e2e.go            # Runtime websocket smoke test
│   └── README.md               # Scripts documentation
├── Makefile                    # Developer task shortcuts
├── run.sh                      # Linux/macOS launcher
├── gotalk.bat                  # Windows launcher
├── go.mod
├── go.sum
└── LICENSE
```

## Troubleshooting

### Port already in use

Error: `bind: address already in use`

- Stop the process currently using port `8080`, or run server on another port:

```bash
go run ./cmd/server -addr :9090
```

> [!WARNING]
> If another process owns the port, clients may connect to the wrong service and you will see confusing behavior.

### Connection refused

Error: `dial tcp ... connect: connection refused`

- Start the server before launching clients.
- Verify host/port passed to `-host` in the TUI client.

### TUI client shows an error and stops receiving messages

- Restart the client session.
- Ensure server and client versions are from the same branch/commit.

> [!CAUTION]
> Mixing binaries from old and new commits can cause protocol or UX mismatches. Rebuild both sides after pulling updates.

## Tech Stack

- [Go](https://go.dev/)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Bubbles](https://github.com/charmbracelet/bubbles)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## Contributing

1. Fork the repo
2. Create a feature branch
3. Make changes with tests where applicable
4. Open a pull request with a clear summary

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
