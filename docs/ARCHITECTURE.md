# System Architecture

GoTalk uses a Hub-and-Spoke architecture to manage WebSocket connections and message broadcasting.

## Core Components

### 1. Hub (`internal/ws/hub.go`)
The Hub is the central manager for all active client connections. It handles:
- **Registration**: Adding new clients to specific rooms.
- **Unregistration**: Removing disconnected clients and cleaning up resources.
- **Broadcasting**: Routing messages to all clients in a specific room.
- **State Management**: Maintaining the map of active rooms and clients.

### 2. Client (`internal/ws/client.go`)
Each WebSocket connection is represented by a Client struct. It acts as a middleman between the WebSocket connection and the Hub.
- **ReadPump**: A goroutine that reads messages from the WebSocket and sends them to the Hub.
- **WritePump**: A goroutine that receives messages from the Hub and writes them to the WebSocket.

### 3. Message Flow
1. **User sends message**: Client JS sends JSON payload via WebSocket.
2. **ReadPump receives**: The `ReadPump` goroutine reads the message.
3. **Hub processes**: The message is sent to the Hub's `Broadcast` channel.
4. **Hub routes**: The Hub identifies the target room and iterates over all registered clients in that room.
5. **WritePump sends**: The Hub sends the message to each client's `Send` channel. The `WritePump` goroutine writes it to the WebSocket.
6. **User receives**: Client JS receives the message and updates the UI.

> [!NOTE]
> Messages are room-specific. A message sent in "general" will never be seen by users in "random".

## Concurrency Model
GoTalk leverages Go's concurrency primitives:
- **Goroutines**: Each client connection spawns two goroutines (ReadPump, WritePump) to handle I/O independently.
- **Channels**: Communication between the Hub and Clients is done via buffered channels (`Register`, `Unregister`, `Broadcast`, `Send`), ensuring thread safety without explicit locks.

> [!WARNING]
> The `Hub` logic runs in a single goroutine loop. Blocking operations inside the Hub's main loop will stall all message routing.
