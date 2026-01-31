#!/bin/bash

# Clear the screen
clear

echo "           GoTalk Launcher"
echo ""
echo "[1] Run Server"
echo "[2] Run Client"
echo ""
read -p "Select option (1/2): " choice

if [ "$choice" == "1" ]; then
    echo ""
    echo "Starting Server..."
    go run cmd/server/main.go

elif [ "$choice" == "2" ]; then
    echo ""
    read -p "Enter Username: " user
    read -p "Enter Room (default: general): " room
    
    # Set default room if empty
    if [ -z "$room" ]; then
        room="general"
    fi
    
    echo ""
    echo "Connecting to room '$room' as '$user'..."
    go run cmd/client/main.go -user="$user" -room="$room"

else
    echo "Invalid option selected."
    exit 1
fi
