SHELL := /bin/bash

.PHONY: help fmt test vet smoke run-server run-client

help:
	@echo "Available targets:"
	@echo "  make fmt         - Format Go code"
	@echo "  make test        - Run unit/integration tests"
	@echo "  make vet         - Run static analysis"
	@echo "  make smoke       - Run websocket end-to-end smoke test"
	@echo "  make run-server  - Start chat server"
	@echo "  make run-client  - Start TUI chat client (USER/ROOM optional)"

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

smoke:
	go run scripts/smoke_e2e.go

run-server:
	go run ./cmd/server

run-client:
	go run ./cmd/client -user="$(if $(USER),$(USER),Guest)" -room="$(if $(ROOM),$(ROOM),general)"
