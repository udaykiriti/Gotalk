# Security Model

GoTalk implements several security measures to protect the server and users.

## 1. Cross-Origin Resource Sharing (CORS)

The WebSocket server enforces strict origin checking. By default, it only accepts connections from the same origin (host). The permissive `CheckOrigin: true` setting has been removed to prevent Cross-Site WebSocket Hijacking (CSWSH).

> [!CAUTION]
> If you are running the frontend on a different port (e.g., during development), you will need to adjust the `CheckOrigin` policy in `internal/ws/client.go`.

## 2. Input Validation

All user inputs are validated on the server side to prevent abuse and injection attacks.
- **Regex**: Usernames and room names must match `^[a-zA-Z0-9_-]+$`.
- **Length Limits**:
  - Room Name: Max 50 characters
  - Username: Max 30 characters
  - Message: Max 512 bytes (WebSocket read limit)

## 3. Denial of Service (DoS) Protection

- **Read Limits**: The server sets a maximum message size (512 bytes) to prevent memory exhaustion.
- **Timeouts**: Read and Write deadlines are enforced to close idle or slow connections.
- **Safe Channel Closing**: The Hub uses synchronous map checks to prevent panics and ensure resource cleanup.

## 4. Cross-Site Scripting (XSS)

The frontend uses `innerText` assignment (or `textContent`) when rendering user messages. No user-generated content is ever rendered as raw HTML (`innerHTML`), neutralizing script injection attacks.

> [!IMPORTANT]
> Never use `innerHTML` for displaying user-provided content.
