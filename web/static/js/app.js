// Mapping old classes to new structure is implicit because I kept IDs same.
// However, I changed 'status-indicator' to 'status-dot' in CSS. 
// I need to update the script to target 'status-dot' or alias it.
// Let's update the script variable selector too.
// ============================================================================
// WebSocket Connection Manager with Robust Reconnection Strategy
// ============================================================================

// Connection state enum
const ConnectionState = {
    DISCONNECTED: 'DISCONNECTED',
    CONNECTING: 'CONNECTING',
    CONNECTED: 'CONNECTED',
    RECONNECTING: 'RECONNECTING',
    MANUALLY_CLOSED: 'MANUALLY_CLOSED'
};

// Configuration
const config = {
    baseReconnectDelay: 1000,      // 1 second
    maxReconnectDelay: 30000,      // 30 seconds
    reconnectDelayMultiplier: 2,   // Exponential backoff
    jitterFactor: 0.1,             // ±10% jitter
    maxReconnectAttempts: 10,      // Max retries
    logEnabled: true               // Console logging
};

// State variables
let conn = null;
let connectionState = ConnectionState.DISCONNECTED;
let reconnectAttempt = 0;
let reconnectDelay = config.baseReconnectDelay;
let reconnectTimeout = null;
let currentRoom = null;
let currentUser = null;

// DOM elements
const log = document.getElementById("log");
const msg = document.getElementById("msg");
const roomInput = document.getElementById("room");
const userInput = document.getElementById("username");
const statusText = document.getElementById("statusText");
const statusIndicator = document.querySelector(".status-dot");
const sendBtn = document.getElementById("sendBtn");
const msgInput = document.getElementById("msg");
const connectBtn = document.getElementById("connectBtn");
const disconnectBtn = document.getElementById("disconnectBtn");
const form = document.getElementById("form");
const userList = document.getElementById("user-list");
const usersContent = document.getElementById("users-content");

// ============================================================================
// Utility Functions
// ============================================================================

function log_debug(message, data) {
    if (config.logEnabled) {
        console.log(`[GoTalk] ${message}`, data || '');
    }
}

function log_error(message, error) {
    if (config.logEnabled) {
        console.error(`[GoTalk] ${message}`, error || '');
    }
}

function updateStatus(state, message) {
    connectionState = state;
    statusText.innerText = message;

    // Update status indicator
    statusIndicator.className = 'status-dot';
    switch (state) {
        case ConnectionState.CONNECTED:
            statusIndicator.classList.add('connected');
            break;
        case ConnectionState.CONNECTING:
            statusIndicator.classList.add('connecting');
            break;
        case ConnectionState.RECONNECTING:
            statusIndicator.classList.add('reconnecting');
            break;
        case ConnectionState.DISCONNECTED:
        case ConnectionState.MANUALLY_CLOSED:
            statusIndicator.classList.add('disconnected');
            break;
    }

    log_debug(`Status updated: ${state} - ${message}`);
}

function setInputsEnabled(enabled) {
    msgInput.disabled = !enabled;
    sendBtn.disabled = !enabled;
}

function setInputsLocked(locked) {
    roomInput.disabled = locked;
    userInput.disabled = locked;
}

function appendLog(element) {
    const doScroll = (log.scrollHeight - log.clientHeight - log.scrollTop) < 50;
    log.appendChild(element);
    if (doScroll) {
        log.scrollTop = log.scrollHeight;
    }
}

function appendNotification(message, type = 'info') {
    const item = document.createElement("div");
    item.className = `notification ${type}`;
    item.innerText = message;
    appendLog(item);
}

function calculateBackoffDelay(attempt) {
    // Base delay with exponential backoff
    let delay = config.baseReconnectDelay * Math.pow(config.reconnectDelayMultiplier, attempt - 1);

    // Cap at max delay
    delay = Math.min(delay, config.maxReconnectDelay);

    // Add jitter (±10%)
    const jitter = delay * config.jitterFactor * (Math.random() * 2 - 1);
    delay = Math.max(1000, delay + jitter); // Ensure minimum 1 second

    return Math.round(delay);
}

// ============================================================================
// Connection Lifecycle Management
// ============================================================================

function setupWebSocket(room, user) {
    currentRoom = room;
    currentUser = user;

    if (window["WebSocket"]) {
        try {
            const protocol = document.location.protocol === "https:" ? "wss" : "ws";
            const url = `${protocol}://${document.location.host}/ws?room=${encodeURIComponent(room)}&user=${encodeURIComponent(user)}`;

            log_debug(`Creating WebSocket connection to: ${url}`);

            const socket = new WebSocket(url);
            conn = socket;

            socket.onopen = onConnectionOpen;
            socket.onclose = onConnectionClose;
            socket.onerror = onConnectionError;
            socket.onmessage = onConnectionMessage;

            updateStatus(ConnectionState.CONNECTING, "Connecting...");
        } catch (error) {
            log_error("Error creating WebSocket", error);
            appendNotification(`Failed to create connection: ${error.message}`, 'warning');
            updateStatus(ConnectionState.DISCONNECTED, "Connection failed");
        }
    } else {
        const message = "Your browser does not support WebSockets.";
        const item = document.createElement("div");
        item.innerHTML = `<b>${message}</b>`;
        appendLog(item);
        log_error(message);
        updateStatus(ConnectionState.DISCONNECTED, "WebSocket not supported");
    }
}

function onConnectionOpen(evt) {
    if (conn !== evt.target) {
        log_debug("Ignoring onopen event for stale connection");
        return;
    }

    log_debug("WebSocket connected");

    // Reset reconnection counters
    reconnectAttempt = 0;
    reconnectDelay = config.baseReconnectDelay;

    // Update UI
    updateStatus(ConnectionState.CONNECTED, `Connected to ${currentRoom}`);
    setInputsEnabled(true);
    setInputsLocked(true);
    connectBtn.classList.add('hidden');
    disconnectBtn.classList.remove('hidden');
    userList.style.display = "flex"; // Show user list

    // Clear chat log
    log.innerHTML = "";

    appendNotification(`Joined room "${currentRoom}" as "${currentUser}"`, 'info');
}

function onConnectionError(evt) {
    if (conn !== evt.target) {
        log_debug("Ignoring onerror event for stale connection");
        return;
    }

    log_error("WebSocket error occurred", evt);
    appendNotification("Connection error occurred. Attempting to reconnect...", 'warning');

    // Trigger reconnection
    if (connectionState !== ConnectionState.MANUALLY_CLOSED) {
        scheduleReconnection();
    }
}

function onConnectionClose(evt) {
    if (conn !== evt.target) {
        log_debug("Ignoring onclose event for stale connection");
        return;
    }

    log_debug(`WebSocket closed - Code: ${evt.code}, Clean: ${evt.wasClean}, Reason: ${evt.reason}`);

    conn = null;
    setInputsEnabled(false);

    // Determine if this was a clean close or unexpected disconnect
    const isCleanClose = evt.wasClean && evt.code === 1000;

    if (connectionState === ConnectionState.MANUALLY_CLOSED) {
        // User manually closed the connection
        updateStatus(ConnectionState.MANUALLY_CLOSED, "Disconnected - Click Connect to rejoin");
        appendNotification("You disconnected from the chat.", 'info');
        setInputsLocked(false);
        connectBtn.classList.remove('hidden');
        disconnectBtn.classList.add('hidden');
        userList.style.display = "none";
        usersContent.innerHTML = "";
    } else if (isCleanClose) {
        // Server closed connection cleanly (shouldn't auto-reconnect)
        updateStatus(ConnectionState.DISCONNECTED, "Disconnected");
        appendNotification("Connection was closed by the server.", 'warning');
        setInputsLocked(false);
        connectBtn.classList.remove('hidden');
        disconnectBtn.classList.add('hidden');
        userList.style.display = "none";
        usersContent.innerHTML = "";
    } else {
        // Unexpected disconnect - attempt reconnection
        appendNotification("Connection lost unexpectedly. Attempting to reconnect...", 'warning');
        scheduleReconnection();
    }
}

function onConnectionMessage(evt) {
    if (conn !== evt.target) {
        log_debug("Ignoring onmessage event for stale connection");
        return;
    }

    try {
        const data = JSON.parse(evt.data);
        let item = document.createElement("div");

        if (data.type === "notification") {
            item.className = "notification";
            item.innerText = data.content;
            appendLog(item);
        } else if (data.type === "user_list") {
            // Update user list
            usersContent.innerHTML = "";
            if (data.users) {
                data.users.forEach(u => {
                    const uItem = document.createElement("div");
                    uItem.className = "user-item";
                    
                    const avatar = document.createElement("div");
                    avatar.className = "user-avatar";
                    
                    const nameSpan = document.createElement("span");
                    nameSpan.innerText = u;

                    if (u === currentUser) {
                        nameSpan.style.fontWeight = "600";
                        nameSpan.innerText += " (You)";
                    }
                    
                    uItem.appendChild(avatar);
                    uItem.appendChild(nameSpan);
                    usersContent.appendChild(uItem);
                });
            }
        } else if (data.type === "message") {
            const isMine = data.user === currentUser;

            const row = document.createElement("div");
            row.className = isMine ? "msg-row msg-mine" : "msg-row msg-other";

            const meta = document.createElement("div");
            meta.className = "meta";
            
            const userSpan = document.createElement("span");
            userSpan.innerText = isMine ? "You" : data.user;
            
            const timeSpan = document.createElement("span");
            timeSpan.className = "timestamp";
            timeSpan.innerText = new Date().toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});

            meta.appendChild(userSpan);
            meta.appendChild(timeSpan);

            const bubble = document.createElement("div");
            bubble.className = "message";
            bubble.innerText = data.content;

            row.appendChild(meta);
            row.appendChild(bubble);
            item = row;
            appendLog(item);
        }
    } catch (error) {
        log_error("Error parsing message", error);
    }
}

// ============================================================================
// Reconnection Logic
// ============================================================================

function scheduleReconnection() {
    if (connectionState === ConnectionState.MANUALLY_CLOSED) {
        log_debug("Not scheduling reconnection - user manually closed connection");
        return;
    }

    // Clear any existing reconnection timeout
    if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
    }

    reconnectAttempt++;

    if (reconnectAttempt > config.maxReconnectAttempts) {
        log_error(`Max reconnection attempts (${config.maxReconnectAttempts}) reached`);
        updateStatus(ConnectionState.DISCONNECTED, "Connection failed - Max retries reached");
        appendNotification(
            `Failed to reconnect after ${config.maxReconnectAttempts} attempts. Click Connect to try again.`,
            'warning'
        );
        setInputsLocked(false);
        connectBtn.classList.remove('hidden');
        disconnectBtn.classList.add('hidden');
        return;
    }

    // Calculate backoff delay
    const delay = calculateBackoffDelay(reconnectAttempt);
    reconnectDelay = delay;

    const delaySeconds = (delay / 1000).toFixed(1);
    const message = `Reconnecting... (Attempt ${reconnectAttempt}/${config.maxReconnectAttempts}, next in ${delaySeconds}s)`;

    updateStatus(ConnectionState.RECONNECTING, message);
    appendNotification(`${message}`, 'info');

    log_debug(`Scheduling reconnection attempt ${reconnectAttempt} in ${delay}ms`);

    // Schedule reconnection
    reconnectTimeout = setTimeout(() => {
        if (connectionState === ConnectionState.MANUALLY_CLOSED) {
            log_debug("Not reconnecting - connection was manually closed");
            return;
        }

        log_debug(`Attempting reconnection #${reconnectAttempt}`);
        setupWebSocket(currentRoom, currentUser);
    }, delay);
}

// ============================================================================
// QR Code Handling
// ============================================================================

function showInviteQR() {
    const room = roomInput.value.trim() || currentRoom || "general";
    const url = new URL(window.location.href);
    url.searchParams.set('room', room);
    
    document.getElementById('qrRoomName').innerText = room;
    
    new QRious({
        element: document.getElementById('qrCanvas'),
        value: url.toString(),
        size: 200,
        padding: 12
    });
    
    document.getElementById('qrModal').style.display = 'flex';
}

function hideInviteQR() {
    document.getElementById('qrModal').style.display = 'none';
}

// ============================================================================
// Responsiveness / UI Logic
// ============================================================================

function toggleUserList() {
    const ul = document.getElementById('user-list');
    ul.classList.toggle('active');
    
    // Close list if user clicks outside on mobile
    if (ul.classList.contains('active')) {
        const closeHandler = (e) => {
            if (!ul.contains(e.target) && !document.getElementById('mobile-users-btn').contains(e.target)) {
                ul.classList.remove('active');
                document.removeEventListener('click', closeHandler);
            }
        };
        setTimeout(() => document.addEventListener('click', closeHandler), 10);
    }
}

// ============================================================================
// User Event Handlers
// ============================================================================

function handleConnectClick() {
    const room = roomInput.value.trim() || "general";
    const user = userInput.value.trim() || "Anonymous";

    if (!room || !user) {
        appendNotification("Please enter a username and room name.", 'warning');
        return;
    }

    if (connectionState === ConnectionState.CONNECTING || connectionState === ConnectionState.CONNECTED) {
        log_debug("Already connecting or connected");
        return;
    }

    // Clear any pending reconnection
    if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
    }

    // Reset reconnection counters
    reconnectAttempt = 0;
    reconnectDelay = config.baseReconnectDelay;

    setupWebSocket(room, user);
}

function handleDisconnectClick() {
    log_debug("User initiated disconnect");

    // Set flag to prevent auto-reconnection
    connectionState = ConnectionState.MANUALLY_CLOSED;

    // Clear any pending reconnection timeout
    if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
    }

    // Close WebSocket connection
    if (conn) {
        conn.close(1000, "User disconnect"); // 1000 = normal closure
    } else {
        // Fallback if conn is already null
        updateStatus(ConnectionState.MANUALLY_CLOSED, "Disconnected");
        setInputsEnabled(false);
        setInputsLocked(false);
        connectBtn.classList.remove('hidden');
        disconnectBtn.classList.add('hidden');
        appendNotification("You have disconnected.", 'info');
    }
}

// Handle form submission
form.onsubmit = function () {
    if (!conn || connectionState !== ConnectionState.CONNECTED) {
        return false;
    }

    const trimmedMsg = msg.value.trim();
    if (!trimmedMsg) {
        return false;
    }

    try {
        conn.send(trimmedMsg);
        msg.value = "";
    } catch (error) {
        log_error("Error sending message", error);
        appendNotification("Failed to send message. Not connected.", 'warning');
    }

    return false;
};

// ============================================================================
// Initialization
// ============================================================================

log_debug("GoTalk WebSocket client initialized");

// Optional: Setup online/offline event listeners
window.addEventListener('online', () => {
    log_debug("Browser came online");
    if (connectionState === ConnectionState.DISCONNECTED && reconnectAttempt > 0) {
        appendNotification("Connection restored. Reconnecting...", 'info');
        scheduleReconnection();
    }
});

window.addEventListener('offline', () => {
    log_debug("Browser went offline");
    appendNotification("Browser is now offline. Reconnection will be attempted when online.", 'warning');
});
