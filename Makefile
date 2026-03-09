ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SRC_DIR := $(ROOT_DIR)/src
TARGET_DIR := $(ROOT_DIR)/target
BIN := $(TARGET_DIR)/env-sync
BIN_GUI := $(TARGET_DIR)/env-sync-gui
GO ?= go
PREFIX ?= /usr/local

.PHONY: build build-gui build-all test test-gui install install-gui install-all clean dev-gui

build:
	@mkdir -p $(TARGET_DIR)
	cd $(SRC_DIR) && $(GO) build -o $(BIN) ./cmd/env-sync

build-gui:
	@mkdir -p $(TARGET_DIR)
	cd $(SRC_DIR)/gui/frontend && npm install && npm run build
	cd $(SRC_DIR) && $(GO) build -o $(BIN_GUI) ./gui

build-all: build build-gui

dev-gui:
	cd $(SRC_DIR)/gui && PATH="$$HOME/go/bin:$$PATH" wails dev

test:
	cd $(SRC_DIR) && $(GO) test ./cmd/... ./internal/...

test-gui:
	cd $(SRC_DIR) && $(GO) test ./gui/...

install: build
	@# Stop service if running
	@if command -v env-sync >/dev/null 2>&1; then \
		echo "Stopping env-sync service if running..."; \
		env-sync service stop >/dev/null 2>&1 || true; \
	fi
	install -d $(PREFIX)/bin
	install -m 755 $(BIN) $(PREFIX)/bin/env-sync
	@# Restart service if it was stopped
	@if command -v systemctl >/dev/null 2>&1; then \
		if systemctl --user is-active --quiet env-sync 2>/dev/null || systemctl --user is-enabled --quiet env-sync 2>/dev/null; then \
			echo "Restarting env-sync service..."; \
			$(PREFIX)/bin/env-sync service restart >/dev/null 2>&1 || true; \
		fi \
	elif command -v launchctl >/dev/null 2>&1; then \
		if launchctl print gui/$$(id -u)/env-sync >/dev/null 2>&1; then \
			echo "Restarting env-sync service..."; \
			$(PREFIX)/bin/env-sync service restart >/dev/null 2>&1 || true; \
		fi \
	fi

install-gui: build-gui
	install -d $(PREFIX)/bin
	install -m 755 $(BIN_GUI) $(PREFIX)/bin/env-sync-gui

install-all: install install-gui

clean:
	rm -rf $(TARGET_DIR)
	rm -rf $(SRC_DIR)/gui/frontend/node_modules
	rm -rf $(SRC_DIR)/gui/frontend/dist
