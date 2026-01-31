# GoTalk

![Go Version](https://img.shields.io/badge/Go-1.20%2B-blue)
![Platform](https://img.shields.io/badge/Platform-Windows%7CLinux%7CmacOS-lightgrey)

GoTalk is a lightweight, real-time chat server and terminal client built with Go for instant group communication.

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Architecture](#architecture)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Features

- **Fast Messaging**: Chat happens instantly.
- **Rooms**: Join different channels (e.g., general, random, tech).
- **User Names**: Pick any name you like.
- **Notifications**: See when people join or leave.
- **Terminal App**: A cool-looking chat app for your command line.
- **Reliable**: Handles errors and quitting safely.

## Project Structure

Gotalk/
├── cmd/
│   ├── server/       # Starts the server
│   └── client/       # Starts the chat app
├── internal/
│   ├── ws/           # Code for real-time chat
│   ├── models/       # Data shapes used everywhere
│   └── handlers/     # Handles web requests
├── web/              # Website files
├── go.mod            # Go project file
├── README.md         # This file
└── gotalk.bat        # Script to run everything easily
```


## Getting Started

### Prerequisites

- Go 1.20 or higher.

### Running the Server

1. Navigate to the project root.
2. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

> [!WARNING]
> Ensure port 8080 is available before starting the server. If it is in use, the server will fail to start.

> [!NOTE]
> The server runs on localhost:8080 by default. You can change the address using flags if needed.
3. The server will start on `localhost:8080`.

### Quick Start

We have a unified launcher script to run both the server and the client.

1. **Run the Server**:
   Double-click `gotalk.bat` (or run `.\gotalk.bat`) and select **Option 1**.

2. **Run the Client**:
   Open a new terminal, run `.\gotalk.bat`, select **Option 2**, and follow the prompts.

> [!TIP]
> You can open multiple terminal windows and run the client in each to simulate a chat between different users.

### Using the Chat

#### Web Interface
1. Ensure the server is running.
2. Open `http://localhost:8080`.
3. Enter a **Username** and **Room Name**.
4. Click "Connect".

You can mix and match web and CLI users in the same room!

## Architecture

- **Hub**: Controls chat rooms and sends messages to everyone.
- **Client**: Manages user connections and sending/receiving messages.
- **CLI**: A command-line chat app.

## Troubleshooting

### Port Already in Use
### Port Already in Use
If you see an error saying the address is "already in use", it means port 8080 is busy.
**Solution**: Stop the other program using that port or change the port settings.

### Firewall Issues
If remote users cannot connect, ensure your firewall allows traffic on port 8080.

## Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests.
