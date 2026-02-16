# Scripts

This folder contains developer automation scripts that are not part of the runtime binaries.

## Available scripts

- `smoke_e2e.go`: boots a local server, connects two websocket clients, sends a message, and verifies delivery.

## Usage

```bash
go run scripts/smoke_e2e.go
```
