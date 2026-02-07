ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SRC_DIR := $(ROOT_DIR)/src
TARGET_DIR := $(ROOT_DIR)/target
BIN := $(TARGET_DIR)/env-sync
GO ?= go
PREFIX ?= /usr/local

.PHONY: build test install clean

build:
	@mkdir -p $(TARGET_DIR)
	$(GO) -C $(SRC_DIR) build -o $(BIN) ./cmd/env-sync

test:
	$(GO) -C $(SRC_DIR) test ./...

install: build
	install -d $(PREFIX)/bin
	install -m 755 $(BIN) $(PREFIX)/bin/env-sync

clean:
	rm -rf $(TARGET_DIR)
