# GoTalk - Real-time Chat Application

GoTalk is a robust, real-time group chat application built with Go (Golang) and pure JavaScript. It features a WebSocket-based server, a responsive web client, and a terminal user interface (TUI).

## Key Features

- **Real-time Messaging**: Instant message delivery using WebSockets.
- **Multi-room Support**: Create or join dynamic chat rooms.
- **Live User List**: See who is currently online in the room.
- **System Notifications**: Join/leave alerts and connection status updates.
- **Robust Reconnection**: Automatic reconnection with exponential backoff.
- **Secure by Default**: Strict CORS policies and input validation.

## Quick Start

### Prerequisites
- Go 1.20 or higher
- Make (optional)

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/udaykiriti/Gotalk.git
   cd Gotalk
   ```
2. Build the project:
   ```bash
   go build -o gotalk ./cmd/server
   ```

### Running the Server
Start the server on port 8080:
```bash
./gotalk --addr :8080
```
Visit `http://localhost:8080` in your browser to start chatting.

## Project Structure
- `cmd/server`: Main server entry point.
- `cmd/client`: TUI client entry point.
- `internal/ws`: WebSocket logic (Hub, Client, Message).
- `internal/handlers`: HTTP handlers.
- `web`: Frontend assets (HTML, CSS, JS).

For more detailed documentation, see the `docs/` directory.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [API Protocol](docs/API.md)
- [Frontend Guide](docs/FRONTEND.md)
- [Security](docs/SECURITY.md)
- [Development](docs/DEVELOPMENT.md)

## License
MIT
