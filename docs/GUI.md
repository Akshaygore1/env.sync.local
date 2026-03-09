# GUI Application

env-sync includes an optional graphical desktop application built with [Wails v2](https://wails.io/) and Vue 3. The GUI provides the same capabilities as the CLI in a visual interface.

## Overview

The GUI (`env-sync-gui`) is a standalone desktop application that operates on the **same** configuration files, secrets, keys, and peer registry as the CLI. Both can be installed and used side-by-side — changes in one are immediately reflected in the other.

- **Not a replacement** — the CLI remains the primary interface
- **Same state** — both read/write `~/.config/env-sync/`
- **Independent binary** — install one or both
- **Native desktop app** — uses system WebView (not Electron/Chromium)

## Installation

### From Source

```bash
# Ubuntu/Debian GUI prerequisites
sudo apt-get update
sudo apt-get install -y pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev

# Build and install GUI only
make build-gui
sudo ./install.sh --gui-only

# Build and install both CLI + GUI
make build-all
sudo ./install.sh --all
```

### Using install.sh

```bash
# Install GUI only
./install.sh --gui-only

# Install both CLI and GUI
./install.sh --all

# Default (CLI only)
./install.sh
```

- **macOS**: installs `env-sync.app` to `/Applications` or `~/Applications` with `--user`
- **macOS releases**: ship separate Apple Silicon and Intel DMG files, each containing `env-sync.app`
- **Linux**: installs the GUI payload under `/opt/env-sync` or `~/.local/lib/env-sync`, plus a desktop launcher and icon
- **Windows**: use the release installer executable from GitHub Releases

### Prerequisites

- **Go 1.24+** and **Node.js 18+** (for building)
- **Linux**: `pkg-config`, `libgtk-3-dev`, and `libwebkit2gtk-4.1-dev`
- **macOS**: Xcode Command Line Tools (WebView is built-in)
- **Windows**: WebView2 runtime (included in Windows 11, available for Windows 10)

On Ubuntu/Debian, build the Linux GUI on a machine or runner that matches the target architecture. `libwebkit2gtk-4.1-dev` conflicts with `libwebkit2gtk-4.1-dev:arm64`, so do not install both into the same root filesystem unless you are maintaining a dedicated cross-compilation sysroot yourself.

## Development

```bash
# Start in development mode (hot-reload frontend)
make dev-gui

# Run GUI tests
make test-gui
```

## Views

### Dashboard
Overview of system status: secrets file info, current mode, server status, discovered peers, and recent backups.

### Secrets
Full CRUD management for secret key-value pairs. Supports click-to-reveal values, add/edit/delete, and export in `.env` or JSON format.

### Sync
Trigger manual sync with all peers or a specific peer. Shows sync results and discovered peer list.

### Peers
Manage peer registry (secure-peer mode). View approved/pending/revoked peers, create invitations, approve or revoke access.

### Keys
Manage AGE encryption keys. View local and peer public keys, import/export keys, and handle pending access requests.

### Settings
Configure operation mode, manage cron jobs, start/stop background server, and view configuration paths.

### Logs
View recent application log entries with level filtering.

## Architecture

```
src/gui/
├── main.go                    # Wails entry point, embeds frontend
├── app.go                     # App lifecycle, version, config paths
├── status_service.go          # System status aggregation
├── secrets_service.go         # Secrets CRUD operations
├── sync_service.go            # Sync orchestration
├── discovery_service.go       # mDNS peer discovery
├── keys_service.go            # Key management
├── mode_service.go            # Mode management
├── peer_service.go            # Peer registry management
├── service_service.go         # Background server control
├── cron_service.go            # Cron job management
├── backup_service.go          # Backup management
├── log_service.go             # Log viewing
├── services_test.go           # Unit tests
├── wails.json                 # Wails project configuration
└── frontend/                  # Vue 3 + TypeScript frontend
    ├── src/
    │   ├── main.ts            # App entry point
    │   ├── App.vue            # Root component (wizard gate)
    │   ├── router/            # Vue Router (7 routes)
    │   ├── stores/            # Pinia state management
    │   ├── types/             # TypeScript interfaces
    │   ├── views/             # Page components
    │   ├── components/        # Reusable components
    │   ├── composables/       # Vue composables (useToast)
    │   └── assets/styles/     # CSS design system
    ├── package.json
    ├── vite.config.ts
    └── tsconfig.json
```

### How It Works

The GUI uses Wails' Go-to-JavaScript bridge. Each Go service struct is bound to the Wails runtime, making its exported methods callable from JavaScript as `window.go.main.ServiceName.MethodName()`. Wails auto-generates TypeScript bindings at build time.

All service methods call the same `internal/` packages that the CLI uses — there is no separate API layer or database. The GUI is simply a different frontend for the same backend logic.

### Design Decisions

1. **Separate `src/gui/` package** — GUI code lives in its own directory, cleanly isolated from CLI code
2. **No build tags** — directory separation means `go build ./cmd/env-sync` never touches GUI code
3. **Embedded frontend** — `//go:embed all:frontend/dist` bundles the built Vue app into the Go binary
4. **System WebView** — Wails uses the OS-native WebView, keeping binary size small (~10-20MB)
5. **First-run wizard** — guides new users through mode selection and initialization
6. **Dark/light theme** — respects system preference, toggleable in sidebar
