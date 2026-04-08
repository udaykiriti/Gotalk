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
- **User List (Desktop)**: Sidebar showing real-time list of online users.
- **User List (Mobile)**: Slide-out drawer accessible via the "Users" icon in the header.
- **Invite Modal**: A glassmorphism overlay containing a scannable QR code for the current room URL.
- **Input Area**: Message input field, disabled when disconnected.

## Responsive Design

GoTalk uses a mobile-first responsive strategy:
- **Breakpoints**: 
    - `768px`: Layout shifts from desktop sidebar to mobile drawer. Full-screen container used.
    - `480px`: Header elements stack to maintain usability.
- **Drawer Logic**: On small screens, the user list is hidden by default. Toggling it opens a drawer that can be closed by clicking outside.
- **Bubble Adaptive Width**: Message bubbles expand to `85%` on mobile to maximize readability.

## QR Invitation System

The frontend uses the `QRious` library to generate room-specific invitations:
1. **Dynamic URL**: The invitation URL is constructed using the current `window.location.href` and the active room name.
2. **Instant Scan**: Generations occur entirely on the client-side, ensuring fast and private link sharing.

## Security Features

- **HTML Escaping**: User content is inserted as text nodes, not innerHTML, preventing XSS.
- **Input Validation**: Username and room inputs are validated before connection.
