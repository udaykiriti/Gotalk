#!/bin/bash
echo "=========================================="
echo "     GoTalk Cross-Platform Builder"
echo "=========================================="
echo ""

mkdir -p bin

echo "[1/2] Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o bin/gotalk_windows.exe ./cmd/server

echo "[2/2] Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o bin/gotalk_linux ./cmd/server

echo ""
echo "Done! Binaries are in the 'bin' folder."
chmod +x bin/gotalk_linux
