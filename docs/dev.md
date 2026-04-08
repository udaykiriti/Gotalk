# Development Guide

## Build Instructions

To build the server binary:
```bash
go build -o gotalk ./cmd/server
```

> [!NOTE]
> The server runs on Linux, macOS, and Windows.

To build the TUI client (if applicable):
```bash
go build -o gotalk-client ./cmd/client
```

## Running Tests

Run all unit tests (including race detection):
```bash
go test -race ./...
```

> [!TIP]
> Always use `-race` when testing concurrent applications like this one to catch data races early.

Run specific validation scripts:
```bash
go test -v scripts/validation_test.go
```

## Database Management

GoTalk uses a local SQLite database named `gotalk.db` to persist chat history.
- **Location**: The database file is created in the project root directory.
- **Resetting History**: To clear all messages, delete the `gotalk.db` file and restart the server. The schema will be re-initialized automatically.
- **Inspection**: You can use standard SQLite tools (e.g., `sqlite3`, `DB Browser for SQLite`) to inspect the `messages` table.

## Directory Structure

- `cmd/`: Main applications (server, client).
- `internal/`: Private application code.
  - `handlers/`: HTTP handlers.
  - `models/`: Data structures.
  - `ws/`: WebSocket core logic.
- `web/`: Static assets for the web client.
- `docs/`: Project documentation.

## Adding Features

1. **Protocol Changes**: Update `internal/models/message.go` and `docs/API.md`.
2. **Backend Logic**: Update `internal/ws/hub.go` to handle new message types.
3. **Frontend**: Update `web/index.html` to render new features.

> [!IMPORTANT]
> When adding new features, update the frontend `web/index.html` carefully as it is embedded into the final binary.
