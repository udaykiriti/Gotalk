SHELL := /bin/bash

.PHONY: help fmt test vet run-server run-client build-all

help:
	@echo "Available targets:"
	@echo "  make fmt         - Format Go code"
	@echo "  make build-all   - Build for Windows and Linux (put in bin/)"
	@echo "  make run-server  - Start chat server"
	@echo "  make run-client  - Start TUI chat client (USER/ROOM optional)"

fmt:
	go fmt ./...

build-all:
	@chmod +x scripts/build.sh
	./scripts/build.sh

run-server:
	go run ./cmd/server

run-client:
	go run ./cmd/client -user="$(if $(USER),$(USER),Guest)" -room="$(if $(ROOM),$(ROOM),general)"
