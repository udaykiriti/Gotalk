# WebSocket API Protocol

The GoTalk client and server communicate using a JSON-based protocol over WebSockets.

## Connection Endpoint

`ws://<host>/ws`

> [!NOTE]
> The default host is `localhost:8080`.

### Query Parameters
- `room` (string, optional): The name of the room to join. Defaults to "general".
- `user` (string, optional): The username. Defaults to "Anonymous".
- `history` (implicit): Upon successful connection, the server automatically pushes the last 50 messages of the room's history to the client.

**Validation Rules:**
- Username and Room name must match regex `^[a-zA-Z0-9_-]+$`.
- Max length: Room (50 chars), User (30 chars).

> [!IMPORTANT]
> Invalid inputs will result in immediate connection closure with code `1008` (Policy Violation).

## Message Structure

All messages exchanged are JSON objects with the following structure:

```json
{
  "type": "string",
  "room": "string",
  "user": "string",
  "content": "string",
  "users": ["string"] // Optional, only for user_list type
}
```

### Message Types

#### 1. Chat Message (`message`)
Sent by users, broadcasted to the room.
- `type`: "message"
- `user`: Sender's username
- `content`: The text message

#### 2. System Notification (`notification`)
Sent by the server to announce events (e.g., user joined/left).
- `type`: "notification"
- `user`: "System"
- `content`: Event description

#### 3. User List Update (`user_list`)
Sent by the server when the list of online users changes.
- `type`: "user_list"
- `users`: Array of usernames currently in the room
- `content`: "user_list_update" (ignored by frontend)

> [!TIP]
> Clients should replace their local user list entirely when receiving this message, rather than diffing.

## Close Codes
- `1000`: Normal closure (User disconnected).
- `1001`: Going away (Browser navigation).
- `1008`: Policy violation (Invalid input or validation error).
