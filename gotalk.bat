@echo off
title GoTalk Launcher
cls
echo ==========================================
echo           GoTalk Launcher
echo ==========================================
echo.
echo [1] Run Server
echo [2] Run Client
echo.
set /p choice="Select option (1/2): "

if "%choice%"=="1" goto run_server
if "%choice%"=="2" goto run_client
goto end

:run_server
echo.
echo Starting Server...
go run cmd/server/main.go
goto end

:run_client
echo.
set /p user="Enter Username: "
set /p room="Enter Room (default: general): "
if "%room%"=="" set room=general
echo.
echo Connecting to room '%room%' as '%user%'...
go run cmd/client/main.go -user="%user%" -room="%room%"
goto end

:end
pause
