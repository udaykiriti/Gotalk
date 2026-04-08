@echo off
echo ==========================================
2: echo      GoTalk Cross-Platform Builder
3: echo ==========================================
4: echo.
5: 
6: if not exist "bin" mkdir bin
7: 
8: echo [1/2] Building for Windows (amd64)...
9: set GOOS=windows
10: set GOARCH=amd64
11: go build -o bin/gotalk_windows.exe ./cmd/server
12: 
13: echo [2/2] Building for Linux (amd64)...
14: set GOOS=linux
15: set GOARCH=amd64
16: go build -o bin/gotalk_linux ./cmd/server
17: 
18: echo.
19: echo Done! Binaries are in the 'bin' folder.
20: pause
