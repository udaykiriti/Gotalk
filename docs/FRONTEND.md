# Frontend Architecture

The web client (`web/index.html`) is a single-page application built with vanilla JavaScript, HTML, and CSS. It implements a robust state machine for managing WebSocket connections.

## Connection States

The client tracks the connection state to provide visual feedback and manage reconnection logic.

1. **DISCONNECTED**: Initial state or after a fatal error.
2. **CONNECTING**: WebSocket handshake in progress.
3. **CONNECTED**: Connection established, ready to chat.
4. **RECONNECTING**: Connection lost, waiting to retry.
5. **MANUALLY_CLOSED**: User explicitly clicked "Disconnect". Auto-reconnection is disabled.

> [!NOTE]
> The connection state is visualized by the colored dot next to the "Connect" button.

## Reconnection Strategy

To ensure reliability, the client implements an exponential backoff strategy for reconnection attempts:
- **Base Delay**: 1 second
- **Multiplier**: 2x (1s, 2s, 4s, 8s...)
- **Max Delay**: 30 seconds
- **Jitter**: ±10% randomization to prevent thundering herd
- **Max Attempts**: 10

> [!WARNING]
> If max attempts (10) are reached, the client stops trying. The user must click "Connect" manually to retry.

## UI Components

- **Status Bar**: Shows current connection state (color-coded dots) and attempt counters.
- **Chat Log**: Scrollable area displaying messages and system notifications.
- **User List**: Sidebar showing real-time list of online users.
- **Input Area**: Message input field, disabled when disconnected.

## Security Features

- **HTML Escaping**: User content is inserted as text nodes, not innerHTML, preventing XSS.
- **Input Validation**: Username and room inputs are validated before connection.
