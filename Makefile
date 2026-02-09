ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SRC_DIR := $(ROOT_DIR)/src
TARGET_DIR := $(ROOT_DIR)/target
BIN := $(TARGET_DIR)/env-sync
GO ?= go
PREFIX ?= /usr/local

.PHONY: build test install clean

build:
	@mkdir -p $(TARGET_DIR)
	cd $(SRC_DIR) && $(GO) build -o $(BIN) ./cmd/env-sync

test:
	cd $(SRC_DIR) && $(GO) test ./...

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

clean:
	rm -rf $(TARGET_DIR)
