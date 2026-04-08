# System Architecture

GoTalk uses a Hub-and-Spoke architecture to manage WebSocket connections and message broadcasting.

## Core Components

### 1. Hub (`internal/ws/hub.go`)
The Hub is the central manager for all active client connections. It handles:
- **Registration**: Adding new clients to specific rooms.
- **Unregistration**: Removing disconnected clients and cleaning up resources.
- **Broadcasting**: Routing messages to all clients in a specific room.
- **State Management**: Maintaining the map of active rooms and clients.
- **Persistence Integration**: Coordinating with the Storage layer to save and retrieve messages.

### 2. Storage (`internal/storage/storage.go`)
The Storage layer provides persistent message history using SQLite.
- **Schema Management**: Initializes the database and creates the `messages` table and indices.
- **Message Persistence**: Saves every broadcasted message to the database.
- **History Retrieval**: Fetches the last 50 messages for a room to populate new client feeds.

### 3. Client (`internal/ws/client.go`)
Each WebSocket connection is represented by a Client struct. It acts as a middleman between the WebSocket connection and the Hub.
- **ReadPump**: A goroutine that reads messages from the WebSocket and sends them to the Hub.
- **WritePump**: A goroutine that receives messages from the Hub and writes them to the WebSocket.
- **Validation**: Enforces size and character limits on incoming room and user names.

### 4. Message Flow
1. **User sends message**: Client JS sends JSON payload via WebSocket.
2. **ReadPump receives**: The `ReadPump` goroutine reads the message.
3. **Hub processes & Persists**: The Hub receives the message, saves it to the **SQLite database**, and then sends it to the `Broadcast` channel.
4. **Hub routes**: The Hub identifies the target room and iterates over all registered clients in that room.
5. **History Sync**: When a new user registers, the Hub fetches the last 50 messages from SQLite and sends them to the new client's `Send` channel immediately.
6. **WritePump sends**: The `WritePump` goroutine writes the messages (broadcast or history) to the WebSocket.
7. **User receives**: Client JS receives the message and updates the UI.

> [!NOTE]
> Messages are room-specific. A message sent in "general" will never be seen by users in "random".

## Concurrency Model
GoTalk leverages Go's concurrency primitives:
- **Goroutines**: Each client connection spawns two goroutines (ReadPump, WritePump) to handle I/O independently.
- **Channels**: Communication between the Hub and Clients is done via buffered channels (`Register`, `Unregister`, `Broadcast`, `Send`), ensuring thread safety without explicit locks.

> [!WARNING]
> The `Hub` logic runs in a single goroutine loop. Blocking operations inside the Hub's main loop will stall all message routing.

## Network Discovery & Join Logic
The server implements automatic IP detection and QR code generation:
1. **Local IP Detection**: On startup, the server identifies the primary non-loopback network interface.
2. **QR Generation**: A scannable QR code is printed to the terminal pointing to the server's network address.
3. **In-App Invitations**: The web client can generate room-specific QR codes for easier sharing among local peers.
