# Plan 07: Add GUI Layer with Wails.io (Vue 3 + TypeScript)

## Problem Statement

env-sync currently operates exclusively via CLI. While this is powerful for automation and shell
integration, a desktop GUI would make the tool more accessible for visual management of secrets,
peer status monitoring, and mode configuration. The GUI should cover **all** CLI functionalities
while providing a richer visual experience.

## Proposed Approach

Use **Wails v2** (latest stable: v2.9.x) to build a native desktop application with:
- **Backend**: Go — reuse existing `internal/` packages directly (no API layer needed)
- **Frontend**: Vue 3 + TypeScript + Vite
- **Binding**: Wails auto-generates TypeScript bindings from Go struct methods
- **Distribution**: Single binary embedding frontend assets via `go:embed`

### CLI + GUI Coexistence Model

The GUI is **not a replacement** for the CLI — it is an alternative management interface. Both
operate on the exact same configuration, secrets files, keys, peer registry, and backups.
Users can freely switch between them, or use both simultaneously.

**Shared State (on disk — single source of truth):**
- `~/.config/env-sync/secrets` — Secrets file (read/written by both CLI and GUI)
- `~/.config/env-sync/keys/` — AGE keys and cached peer public keys
- `~/.config/env-sync/peer/` — Peer registry and membership log
- `~/.config/env-sync/requests/` — Pending access requests
- `~/.config/env-sync/backups/` — Backup files
- `~/.config/env-sync/config` — Mode and configuration

**What this means in practice:**
- `env-sync add MY_KEY="value"` on CLI → immediately visible when GUI refreshes its Secrets view
- Approving a peer in GUI → CLI's `env-sync peer list` reflects the change instantly
- `env-sync sync` on CLI → GUI's Dashboard shows updated status on next poll
- Cron job installed via GUI → same crontab entry that `env-sync cron --show` displays
- Backups created by GUI restore → same backups available to `env-sync restore`
- No database, no daemon state, no IPC — everything is file-based, so interop is automatic

**Installation model:**
- Both binaries can be installed side-by-side: `env-sync` (CLI) + `env-sync-gui` (GUI)
- `make install` installs CLI, `make install-gui` installs GUI — independent targets
- Neither requires the other — CLI-only, GUI-only, or both are all valid setups
- The background server/daemon (`env-sync serve --daemon`) is shared: if the CLI starts it,
  the GUI sees it as running; if the GUI stops it, the CLI reflects that
- Same `install.sh` can offer both: `./install.sh --cli`, `./install.sh --gui`, `./install.sh --all`

**Concurrency safety:**
- File-level operations (read/write secrets, backups) use the same `internal/` packages
  with the same file-locking and atomic-write patterns already used by the CLI
- The GUI does not hold long-lived locks on any files — each operation is atomic
- If CLI and GUI write simultaneously, the last-write-wins semantics match existing
  multi-peer sync behavior (backup is always created before overwrite)

### Why Wails v2 (not v3)?
- v3 is still in Alpha (as of early 2026) with potential breaking changes
- v2.9.x is stable, well-documented, and production-ready
- v2 has excellent Vue + TypeScript template support
- Migration to v3 can be done later when it stabilizes

### Architecture Overview

```
src/
├── cmd/env-sync/main.go          # Existing CLI entry point (unchanged)
├── cmd/env-sync-gui/             # NEW: GUI entry point
│   ├── main.go                   #   Wails app initialization
│   ├── app.go                    #   Main App struct (bound to frontend)
│   ├── sync_service.go           #   Sync operations service
│   ├── secrets_service.go        #   Secrets CRUD service
│   ├── discovery_service.go      #   Peer discovery service
│   ├── keys_service.go           #   Key management service
│   ├── mode_service.go           #   Mode management service
│   ├── peer_service.go           #   Peer registry service
│   ├── status_service.go         #   Status & health service
│   ├── service_service.go        #   Background service mgmt
│   └── cron_service.go           #   Cron management service
├── frontend/                     # NEW: Vue 3 + TypeScript frontend
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── router/
│   │   │   └── index.ts
│   │   ├── stores/               # Pinia state management
│   │   │   ├── secrets.ts
│   │   │   ├── peers.ts
│   │   │   ├── status.ts
│   │   │   └── settings.ts
│   │   ├── views/
│   │   │   ├── DashboardView.vue
│   │   │   ├── SecretsView.vue
│   │   │   ├── PeersView.vue
│   │   │   ├── KeysView.vue
│   │   │   ├── SyncView.vue
│   │   │   ├── SettingsView.vue
│   │   │   └── LogsView.vue
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   ├── secrets/
│   │   │   ├── peers/
│   │   │   ├── keys/
│   │   │   └── common/
│   │   ├── composables/          # Vue composables for Wails bindings
│   │   │   ├── useSync.ts
│   │   │   ├── useSecrets.ts
│   │   │   ├── usePeers.ts
│   │   │   └── useDiscovery.ts
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── assets/
│   │       └── styles/
│   ├── wailsjs/                  # Auto-generated by Wails
│   │   ├── go/                   #   Go bindings (TypeScript)
│   │   └── runtime/              #   Wails runtime API
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   └── env.d.ts
├── internal/                     # Existing packages (reused as-is)
│   ├── cli/
│   ├── sync/
│   ├── discovery/
│   ├── secrets/
│   ├── crypto/age/
│   ├── keys/
│   ├── mode/
│   ├── peer/
│   ├── config/
│   ├── backup/
│   ├── cron/
│   ├── server/
│   ├── service/
│   ├── transport/
│   ├── metadata/
│   ├── identity/
│   └── logging/
├── go.mod                        # Updated with wails dependency
└── go.sum
```

---

## CLI → GUI Feature Coverage Matrix

Every CLI command maps to a GUI feature. Nothing is left behind.

| # | CLI Command | GUI Location | GUI Implementation |
|---|-------------|-------------|-------------------|
| 1 | `sync [host]` | Dashboard + Sync View | One-click sync button, host selector dropdown |
| 2 | `sync --all` | Dashboard | "Sync All" button with progress indicator |
| 3 | `sync --force` | Sync View | Force sync toggle/checkbox |
| 4 | `sync --force-pull <host>` | Sync View | Force pull button per peer |
| 5 | `serve [--port] [--daemon]` | Settings View | Server toggle switch, port config input |
| 6 | `discover` | Peers View | Auto-refresh peer list, manual "Scan" button |
| 7 | `discover --ssh` | Peers View | Filter toggle: "SSH reachable only" |
| 8 | `discover --collect-keys` | Peers View | "Collect Keys" action button |
| 9 | `status` | Dashboard (home) | Real-time status cards (file, server, peers, backups) |
| 10 | `init [--encrypted]` | First-run wizard / Settings | Guided initialization flow |
| 11 | `init --encrypt-existing` | Settings View | "Encrypt existing secrets" button |
| 12 | `restore [n]` | Backups section in Settings | Backup list with "Restore" button per entry |
| 13 | `cron --install` | Settings View | "Enable periodic sync" toggle |
| 14 | `cron --remove` | Settings View | Same toggle (off state) |
| 15 | `cron --show` | Settings View | Current cron interval display |
| 16 | `cron --interval N` | Settings View | Interval input (minutes) |
| 17 | `add KEY="value"` | Secrets View | "Add Secret" form/modal |
| 18 | `remove KEY` | Secrets View | Delete button per secret row |
| 19 | `show KEY` | Secrets View | Click-to-reveal value with copy button |
| 20 | `list` | Secrets View | Table of all keys (always visible) |
| 21 | `load --format env` | Secrets View | "Export as .env" button |
| 22 | `load --format json` | Secrets View | "Export as JSON" button |
| 23 | `load --key KEY` | Secrets View | Copy single key value |
| 24 | `path` | Settings View | Display config paths |
| 25 | `path --backup` | Settings View | Display backup directory path |
| 26 | `key show` | Keys View | Display local public key |
| 27 | `key show --private` | Keys View | Reveal private key (with warning dialog) |
| 28 | `key export [--qr]` | Keys View | Export button + QR code display |
| 29 | `key import <pubkey> <host>` | Keys View | "Import Key" form |
| 30 | `key import --from <host>` | Keys View | "Import from peer" dropdown |
| 31 | `key list` | Keys View | Table of all cached keys |
| 32 | `key request-access` | Keys View | "Request Access" workflow |
| 33 | `key grant-access` | Keys View | "Grant Access" form |
| 34 | `key approve-requests` | Keys View / Notification | Approval queue with accept/deny |
| 35 | `key remove <host>` | Keys View | Delete button per key row |
| 36 | `key revoke <host>` | Keys View | "Revoke" button with confirmation |
| 37 | `mode get` | Settings View | Current mode display |
| 38 | `mode set <mode>` | Settings View | Mode selector (radio/dropdown) with confirmation |
| 39 | `peer invite` | Peers View | "Create Invite" button → token display + copy |
| 40 | `peer request <host> <token>` | Peers View | "Join Network" form |
| 41 | `peer approve <id>` | Peers View | Approve button in pending peers list |
| 42 | `peer revoke <id>` | Peers View | Revoke button with confirmation dialog |
| 43 | `peer list` | Peers View | Peer registry table (always visible) |
| 44 | `peer trust show` | Peers View | Trust details panel |
| 45 | `service stop` | Settings View | Stop server button |
| 46 | `service restart` | Settings View | Restart server button |
| 47 | `service uninstall` | Settings View | Uninstall service button |
| 48 | `version` | Settings View / About | Version display in sidebar footer |
| 49 | `help` | N/A | GUI is self-explanatory; add tooltips everywhere |
| 50 | `--verbose` | Logs View | Real-time log viewer panel |

---

## Implementation Steps

### Phase 1: Project Scaffolding & Wails Setup

#### Step 1.1: Install Wails CLI and Verify Prerequisites
- Install Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Run `wails doctor` to verify all dependencies (Go, npm, platform tools)
- On macOS: Ensure Xcode command-line tools are installed
- On Linux: Ensure `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` are available

#### Step 1.2: Initialize Wails Project Structure
- Run `wails init -n env-sync-gui -t vue-ts` in a temporary location
- Copy the generated structure into `src/` adapting to project layout:
  - `frontend/` → `src/frontend/`
  - `main.go`, `app.go` → `src/cmd/env-sync-gui/`
  - `wails.json` → `src/wails.json`
- Update `src/go.mod` to add `github.com/wailsapp/wails/v2` dependency
- Update `wails.json` paths to match project layout
- Verify `wails dev` runs the skeleton app from `src/`

#### Step 1.3: Configure Frontend Tooling
- Upgrade Vue to latest Vue 3.x in `package.json`
- Ensure all frontend code is TypeScript (no `.js` files in `src/`)
- Add dependencies:
  - `vue-router@4` - Client-side routing
  - `pinia` - State management
  - `@vueuse/core` - Vue composables utilities (optional, but useful)
- Configure `tsconfig.json` with strict mode enabled
- Configure `vite.config.ts` for path aliases (`@/` → `src/`)
- Add ESLint + Prettier for TypeScript + Vue

#### Step 1.4: Update Build System
- Add `Makefile` targets:
  - `build-gui`: Runs `wails build` from `src/` directory
  - `dev-gui`: Runs `wails dev` for development with hot-reload
  - `install-gui`: Install GUI binary alongside CLI binary
- The GUI binary should be named `env-sync-gui` (separate from CLI `env-sync`)
- Both binaries coexist in `/usr/local/bin/`: `env-sync` + `env-sync-gui`
- Update `install.sh` to support `--cli`, `--gui`, and `--all` flags
- Update `.gitignore` for `frontend/node_modules/`, `frontend/dist/`, `frontend/wailsjs/`

---

### Phase 2: Go Backend Services (Wails Bindings)

These services are Go structs whose public methods are auto-bound to the frontend.
They wrap existing `internal/` packages — **no business logic duplication**.

#### Step 2.1: App Service (`app.go`)
Core application struct bound to Wails.

```go
type App struct {
    ctx context.Context
}

// Lifecycle hooks
func (a *App) startup(ctx context.Context)    // Store context, init logging
func (a *App) shutdown(ctx context.Context)    // Cleanup resources

// General
func (a *App) GetVersion() string              // Returns version string
func (a *App) GetConfigPaths() ConfigPaths     // Returns all config paths
func (a *App) IsInitialized() bool             // Check if secrets file exists
```

#### Step 2.2: Status Service (`status_service.go`)
Maps to: `env-sync status`

```go
type StatusService struct{}

type StatusInfo struct {
    SecretsFile   FileStatus
    Server        ServerStatus
    Peers         []PeerStatus
    Backups       []BackupInfo
    Mode          ModeInfo
}

func (s *StatusService) GetStatus() (StatusInfo, error)
func (s *StatusService) GetFileStatus() (FileStatus, error)
func (s *StatusService) GetServerStatus() (ServerStatus, error)
func (s *StatusService) IsServerRunning() bool
func (s *StatusService) GetFileModTime() (string, error)  // For staleness detection vs CLI changes
```

#### Step 2.3: Secrets Service (`secrets_service.go`)
Maps to: `add`, `remove`, `show`, `list`, `load`, `init`

```go
type SecretsService struct{}

type SecretEntry struct {
    Key       string
    Value     string    // Decrypted value
    UpdatedAt string
}

func (s *SecretsService) List() ([]SecretEntry, error)
func (s *SecretsService) Get(key string) (SecretEntry, error)
func (s *SecretsService) Add(key, value string) error
func (s *SecretsService) Remove(key string) error
func (s *SecretsService) ExportEnv() (string, error)
func (s *SecretsService) ExportJSON() (string, error)
func (s *SecretsService) Initialize(encrypted bool) error
func (s *SecretsService) EncryptExisting() error
func (s *SecretsService) IsEncrypted() bool
```

#### Step 2.4: Sync Service (`sync_service.go`)
Maps to: `sync`

```go
type SyncService struct{}

type SyncResult struct {
    Success  bool
    Message  string
    Source   string
    Changes  int
}

func (s *SyncService) SyncAll() ([]SyncResult, error)
func (s *SyncService) SyncFrom(hostname string) (SyncResult, error)
func (s *SyncService) ForcePull(hostname string) (SyncResult, error)
func (s *SyncService) ForceSync(hostname string) (SyncResult, error)
```

#### Step 2.5: Discovery Service (`discovery_service.go`)
Maps to: `discover`

```go
type DiscoveryService struct{}

type DiscoveredPeer struct {
    Hostname    string
    SSHAccess   bool
    HasPubKey   bool
    Version     string
}

func (d *DiscoveryService) Discover(timeout int) ([]DiscoveredPeer, error)
func (d *DiscoveryService) DiscoverSSH(timeout int) ([]DiscoveredPeer, error)
func (d *DiscoveryService) CollectKeys(timeout int) (int, error)
```

#### Step 2.6: Keys Service (`keys_service.go`)
Maps to: `key` subcommands

```go
type KeysService struct{}

type KeyInfo struct {
    Hostname  string
    PublicKey string
    IsLocal   bool
}

func (k *KeysService) GetLocalKey() (KeyInfo, error)
func (k *KeysService) GetPrivateKey() (string, error)              // ⚠ Warning: sensitive
func (k *KeysService) ExportPublicKey() (string, error)
func (k *KeysService) ExportQRCode() (string, error)               // Returns base64 PNG
func (k *KeysService) ImportKey(pubkey, hostname string) error
func (k *KeysService) ImportFromPeer(hostname string) error
func (k *KeysService) ListKeys() ([]KeyInfo, error)
func (k *KeysService) RemoveKey(hostname string) error
func (k *KeysService) RevokeKey(hostname string) error
func (k *KeysService) RequestAccess(hostname string) error
func (k *KeysService) RequestAccessAll() (int, error)
func (k *KeysService) TriggerReencryption(hostname string) error
func (k *KeysService) TriggerReencryptionAll() (int, error)
func (k *KeysService) GrantAccess(hostname, pubkey string) error
func (k *KeysService) GetPendingRequests() ([]AccessRequest, error)
func (k *KeysService) ApproveRequest(hostname string) error
func (k *KeysService) DenyRequest(hostname string) error
```

#### Step 2.7: Mode Service (`mode_service.go`)
Maps to: `mode` subcommands

```go
type ModeService struct{}

type ModeInfo struct {
    Current     string   // "trusted-owner-ssh", "secure-peer", "dev-plaintext-http"
    Description string
    Features    []string // List of mode capabilities
}

func (m *ModeService) GetMode() (ModeInfo, error)
func (m *ModeService) SetMode(mode string, pruneOldMaterial bool) error
func (m *ModeService) GetAvailableModes() []ModeInfo
```

#### Step 2.8: Peer Service (`peer_service.go`)
Maps to: `peer` subcommands (secure-peer mode only)

```go
type PeerService struct{}

type PeerInfo struct {
    ID          string
    Hostname    string
    State       string   // "approved", "revoked", "pending"
    TLSFingerprint string
    AGEPubKey   string
}

type InviteInfo struct {
    Token       string
    CreatedBy   string
    Fingerprint string
    ExpiresAt   string
    Command     string   // Full command for new peer
}

type TrustInfo struct {
    LocalHostname    string
    TLSFingerprint   string
    CertValidUntil   string
    AGEPublicKey     string
    TrustedPeers     []PeerInfo
}

func (p *PeerService) ListPeers() ([]PeerInfo, error)
func (p *PeerService) ListPending() ([]PeerInfo, error)
func (p *PeerService) CreateInvite(expiry string) (InviteInfo, error)
func (p *PeerService) RequestAccess(host, token string) error
func (p *PeerService) ApprovePeer(peerID string) error
func (p *PeerService) RevokePeer(peerID string) error
func (p *PeerService) GetTrustInfo() (TrustInfo, error)
func (p *PeerService) IsSecurePeerMode() bool
```

#### Step 2.9: Service Management Service (`service_service.go`)
Maps to: `serve`, `service` subcommands

```go
type ServiceMgmtService struct{}

func (s *ServiceMgmtService) StartServer(port int, daemon bool) error
func (s *ServiceMgmtService) StopServer() error
func (s *ServiceMgmtService) RestartServer() error
func (s *ServiceMgmtService) UninstallService() error
func (s *ServiceMgmtService) GetServerPort() int
func (s *ServiceMgmtService) IsServerRunning() bool
```

#### Step 2.10: Cron Service (`cron_service.go`)
Maps to: `cron` subcommands

```go
type CronService struct{}

type CronInfo struct {
    Installed bool
    Interval  int    // minutes
    NextRun   string // estimated
}

func (c *CronService) GetCronStatus() (CronInfo, error)
func (c *CronService) InstallCron(intervalMinutes int) error
func (c *CronService) RemoveCron() error
```

#### Step 2.11: Backup Service (`backup_service.go`)
Maps to: `restore`

```go
type BackupService struct{}

type BackupInfo struct {
    Number    int
    Timestamp string
    Size      int64
    Path      string
}

func (b *BackupService) ListBackups() ([]BackupInfo, error)
func (b *BackupService) Restore(n int) error
func (b *BackupService) GetBackupDir() string
```

#### Step 2.12: Log Service (for real-time GUI log viewer)
New functionality — maps to `--verbose` flag behavior.

```go
type LogService struct{}

type LogEntry struct {
    Timestamp string
    Level     string
    Message   string
}

func (l *LogService) GetRecentLogs(count int) ([]LogEntry, error)
func (l *LogService) GetLogFile() string
```

#### Step 2.13: Wire Up main.go
Register all services in `src/cmd/env-sync-gui/main.go`:

```go
func main() {
    app := NewApp()
    syncSvc := &SyncService{}
    secretsSvc := &SecretsService{}
    discoverySvc := &DiscoveryService{}
    keysSvc := &KeysService{}
    modeSvc := &ModeService{}
    peerSvc := &PeerService{}
    statusSvc := &StatusService{}
    serviceSvc := &ServiceMgmtService{}
    cronSvc := &CronService{}
    backupSvc := &BackupService{}
    logSvc := &LogService{}

    err := wails.Run(&options.App{
        Title:  "env-sync",
        Width:  1200,
        Height: 800,
        MinWidth: 900,
        MinHeight: 600,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        OnStartup:  app.startup,
        OnShutdown: app.shutdown,
        Bind: []interface{}{
            app, syncSvc, secretsSvc, discoverySvc, keysSvc,
            modeSvc, peerSvc, statusSvc, serviceSvc, cronSvc,
            backupSvc, logSvc,
        },
    })
}
```

---

### Phase 3: Frontend Foundation (Vue 3 + TypeScript)

#### Step 3.1: App Shell & Layout
- Create `AppLayout.vue` with sidebar navigation + main content area
- Sidebar links:
  - 🏠 Dashboard
  - 🔑 Secrets
  - 🔄 Sync
  - 👥 Peers
  - 🗝️ Keys
  - ⚙️ Settings
  - 📋 Logs
- Footer: version number, current mode badge
- Responsive layout (min-width: 900px)

#### Step 3.2: Vue Router Setup
```typescript
// router/index.ts
const routes = [
  { path: '/',           name: 'dashboard', component: DashboardView },
  { path: '/secrets',    name: 'secrets',   component: SecretsView },
  { path: '/sync',       name: 'sync',      component: SyncView },
  { path: '/peers',      name: 'peers',     component: PeersView },
  { path: '/keys',       name: 'keys',      component: KeysView },
  { path: '/settings',   name: 'settings',  component: SettingsView },
  { path: '/logs',       name: 'logs',      component: LogsView },
]
```

#### Step 3.3: Pinia Stores
- `stores/secrets.ts` — Secret entries, CRUD operations
- `stores/peers.ts` — Peer list, discovery results, invitations
- `stores/status.ts` — File status, server status, sync status
- `stores/settings.ts` — Mode, cron config, server config, paths

Each store wraps Wails-generated bindings with reactive state:
```typescript
// stores/secrets.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { List, Add, Remove, Get } from '../../wailsjs/go/main/SecretsService'

export const useSecretsStore = defineStore('secrets', () => {
  const entries = ref<SecretEntry[]>([])
  const loading = ref(false)

  async function fetchAll() {
    loading.value = true
    entries.value = await List()
    loading.value = false
  }

  async function addSecret(key: string, value: string) {
    await Add(key, value)
    await fetchAll()
  }

  // ... remove, get, export methods
  return { entries, loading, fetchAll, addSecret }
})
```

#### Step 3.4: Composables for Wails Bindings
Create thin TypeScript composables that wrap Wails binding calls:

```typescript
// composables/useSync.ts
import { SyncAll, SyncFrom, ForcePull } from '../../wailsjs/go/main/SyncService'

export function useSync() {
  const syncAll = async () => await SyncAll()
  const syncFrom = async (host: string) => await SyncFrom(host)
  const forcePull = async (host: string) => await ForcePull(host)
  return { syncAll, syncFrom, forcePull }
}
```

#### Step 3.5: TypeScript Types
Define frontend types mirroring Go structs in `types/index.ts`.
Wails auto-generates these in `wailsjs/go/models.ts`, but we can re-export
with any frontend-specific additions.

---

### Phase 4: View Implementation

#### Step 4.1: Dashboard View (`DashboardView.vue`)
**Maps to: `status`, quick `sync`**

Layout:
```
┌──────────────────────────────────────────────┐
│  env-sync Dashboard                          │
├──────────┬──────────┬──────────┬────────────┤
│ Secrets  │ Server   │ Peers    │ Last Sync  │
│ 12 keys  │ ● Online │ 3 found │ 5 min ago  │
│ 🔒 enc   │ :5739    │ 2 SSH   │            │
├──────────┴──────────┴──────────┴────────────┤
│  [🔄 Sync Now]  [🔄 Sync All]               │
├──────────────────────────────────────────────┤
│  Recent Activity / Sync Log                  │
│  - Synced from beelink.local (2 keys updated)│
│  - Backup created #3                         │
│  - Peer nuc.local discovered                 │
└──────────────────────────────────────────────┘
```

Components:
- `StatusCard.vue` — Reusable card for each status metric
- `QuickActions.vue` — Sync Now, Sync All buttons
- `ActivityFeed.vue` — Recent sync/event log

Coexistence behavior:
- Auto-polls `GetFileModTime()` every 5 seconds to detect CLI-side changes
- When file modification detected, triggers store refresh automatically
- Window-focus event also triggers a full status refresh
- Manual "Refresh" button in header for explicit reload

#### Step 4.2: Secrets View (`SecretsView.vue`)
**Maps to: `list`, `show`, `add`, `remove`, `load`**

Layout:
```
┌──────────────────────────────────────────────┐
│  Secrets (12 keys)       [+ Add] [⬇ Export]  │
├──────────────────────────────────────────────┤
│  🔍 Filter keys...                           │
├──────────────────────────────────────────────┤
│  KEY              │ VALUE           │ Actions │
│  OPENAI_API_KEY   │ ●●●●●● [👁]    │ 📋 🗑  │
│  DATABASE_URL     │ ●●●●●● [👁]    │ 📋 🗑  │
│  AWS_SECRET       │ ●●●●●● [👁]    │ 📋 🗑  │
│  ...              │                │         │
├──────────────────────────────────────────────┤
│  Export: [.env] [JSON] [Copy All]            │
└──────────────────────────────────────────────┘
```

Components:
- `SecretTable.vue` — Sortable, filterable table
- `SecretRow.vue` — Row with masked value, reveal toggle, copy, delete
- `AddSecretModal.vue` — Form modal for KEY=value input
- `ExportPanel.vue` — Export format buttons

Features:
- Click-to-reveal values (eye icon toggle)
- Copy to clipboard buttons
- Search/filter bar
- Inline edit (double-click value to edit)
- Confirmation dialog on delete

#### Step 4.3: Sync View (`SyncView.vue`)
**Maps to: `sync` with all flags**

Layout:
```
┌──────────────────────────────────────────────┐
│  Sync                                         │
├──────────────────────────────────────────────┤
│  Discovered Peers:                           │
│  ┌────────────────────────────────────────┐  │
│  │ beelink.local  │ SSH ✓ │ v3.0 │ [Sync]│  │
│  │ nuc.local      │ SSH ✓ │ v3.0 │ [Sync]│  │
│  │ mbp16.local    │ SSH ✗ │ v3.0 │ [Sync]│  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Options:                                    │
│  ☐ Force sync (overwrite even if local newer)│
│  ☐ Force pull (replace local entirely)       │
│                                              │
│  [🔄 Sync All Peers]  [🔍 Re-discover]       │
├──────────────────────────────────────────────┤
│  Sync History:                               │
│  • 15:30 — Synced from beelink.local ✓      │
│  • 15:00 — Synced from nuc.local ✓          │
│  • 14:30 — Sync failed: SSH timeout ✗       │
└──────────────────────────────────────────────┘
```

Components:
- `PeerSyncCard.vue` — Per-peer sync controls
- `SyncOptions.vue` — Force/force-pull toggles
- `SyncHistory.vue` — Log of recent syncs

#### Step 4.4: Peers View (`PeersView.vue`)
**Maps to: `discover`, `peer` subcommands**

Layout:
```
┌──────────────────────────────────────────────┐
│  Peers                    [🔍 Scan] [✉ Invite]│
├──────────────────────────────────────────────┤
│  Mode: secure-peer                           │
│                                              │
│  Peer Registry:                              │
│  ┌────────────────────────────────────────┐  │
│  │ ✓ beelink.local │ Approved │ [Revoke] │  │
│  │ ✓ nuc.local     │ Approved │ [Revoke] │  │
│  │ ⏳ mbp16.local   │ Pending  │ [Approve]│  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Trust Details:                              │
│  Local: macbook.local                        │
│  TLS:   sha256:abc123...                     │
│  AGE:   age1xyz...                           │
├──────────────────────────────────────────────┤
│  Network Discovery (mDNS):                   │
│  3 peers found on local network              │
│  └─ [Collect Keys from All]                  │
└──────────────────────────────────────────────┘
```

Components:
- `PeerRegistry.vue` — Full peer table with actions
- `PeerCard.vue` — Individual peer with status, fingerprint
- `InviteModal.vue` — Create invite, show token + QR + copyable command
- `JoinNetworkModal.vue` — Enter host + token to request access
- `TrustPanel.vue` — Local trust identity details
- `DiscoveryPanel.vue` — Live mDNS discovery results

#### Step 4.5: Keys View (`KeysView.vue`)
**Maps to: `key` subcommands**

Layout:
```
┌──────────────────────────────────────────────┐
│  Key Management                               │
├──────────────────────────────────────────────┤
│  Local Key:                                  │
│  ┌────────────────────────────────────────┐  │
│  │ Hostname: macbook.local                │  │
│  │ Public:   age1abc123...  [📋 Copy]     │  │
│  │ [📤 Export] [📷 QR Code] [🔒 Show Private]│ │
│  └────────────────────────────────────────┘  │
│                                              │
│  Cached Peer Keys:                           │
│  ┌────────────────────────────────────────┐  │
│  │ beelink.local │ age1xyz... │ [🗑 Remove]│  │
│  │ nuc.local     │ age1def... │ [⛔ Revoke]│  │
│  └────────────────────────────────────────┘  │
│                                              │
│  [📥 Import Key]  [🤝 Request Access]         │
│                                              │
│  Pending Access Requests: (2)                │
│  ┌────────────────────────────────────────┐  │
│  │ mbp16.local │ age1... │ [✓ Approve] [✗]│  │
│  │ rpi.local   │ age1... │ [✓ Approve] [✗]│  │
│  └────────────────────────────────────────┘  │
└──────────────────────────────────────────────┘
```

Components:
- `LocalKeyPanel.vue` — Show/export/QR local key
- `PeerKeyTable.vue` — Cached peer keys with actions
- `ImportKeyModal.vue` — Import form (manual or from peer)
- `AccessRequestQueue.vue` — Approve/deny incoming requests
- `RequestAccessModal.vue` — Request access to peer
- `QRCodeDisplay.vue` — QR code overlay/modal
- `PrivateKeyWarningModal.vue` — Warning before showing private key

#### Step 4.6: Settings View (`SettingsView.vue`)
**Maps to: `mode`, `cron`, `serve`, `init`, `restore`, `path`, `service`**

Layout:
```
┌──────────────────────────────────────────────┐
│  Settings                                     │
├──────────────────────────────────────────────┤
│  Security Mode                               │
│  ┌────────────────────────────────────────┐  │
│  │ ○ dev-plaintext-http (Debug only)      │  │
│  │ ● trusted-owner-ssh (Default)          │  │
│  │ ○ secure-peer (Cross-owner)            │  │
│  │ [Apply Mode Change]                    │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Server                                      │
│  ┌────────────────────────────────────────┐  │
│  │ Status: ● Running on port 5739         │  │
│  │ Port: [5739]  Daemon: [ON/OFF]         │  │
│  │ [Start] [Stop] [Restart] [Uninstall]   │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Periodic Sync (Cron)                        │
│  ┌────────────────────────────────────────┐  │
│  │ Enabled: [ON/OFF]                      │  │
│  │ Interval: [30] minutes                 │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Initialization                              │
│  ┌────────────────────────────────────────┐  │
│  │ [Initialize Secrets File]              │  │
│  │ [Encrypt Existing Secrets]             │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Backups                                     │
│  ┌────────────────────────────────────────┐  │
│  │ 5 backups available                    │  │
│  │ #1 - 2025-02-16 15:30 [Restore]       │  │
│  │ #2 - 2025-02-16 14:00 [Restore]       │  │
│  │ #3 - 2025-02-16 12:30 [Restore]       │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Paths                                       │
│  ┌────────────────────────────────────────┐  │
│  │ Secrets:  ~/.config/env-sync/secrets   │  │
│  │ Backups:  ~/.config/env-sync/backups/  │  │
│  │ Keys:     ~/.config/env-sync/keys/     │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  About: env-sync v3.0.0                      │
└──────────────────────────────────────────────┘
```

Components:
- `ModeSelector.vue` — Radio group for modes with descriptions
- `ServerPanel.vue` — Start/stop/restart/port config
- `CronPanel.vue` — Enable/disable toggle + interval slider
- `InitPanel.vue` — Initialize and encrypt buttons
- `BackupList.vue` — Backup restore list
- `PathDisplay.vue` — Config paths display

#### Step 4.7: Logs View (`LogsView.vue`)
**Maps to: `--verbose` behavior**

Layout:
```
┌──────────────────────────────────────────────┐
│  Logs                          [Clear] [⬇]   │
├──────────────────────────────────────────────┤
│  Filter: [All ▾] [🔍 Search...]              │
├──────────────────────────────────────────────┤
│  15:30:45 INFO  Sync completed from beelink  │
│  15:30:44 DEBUG Fetching from beelink:5739   │
│  15:30:43 INFO  Discovered 3 peers           │
│  15:30:40 DEBUG mDNS browse started          │
│  15:00:00 INFO  Cron sync triggered          │
│  ...                                         │
└──────────────────────────────────────────────┘
```

Components:
- `LogViewer.vue` — Scrollable, auto-tail log display
- `LogFilter.vue` — Level filter (ERROR, WARN, INFO, DEBUG)
- `LogEntry.vue` — Color-coded log line

---

### Phase 5: First-Run Experience & Polish

#### Step 5.1: First-Run Wizard
If `IsInitialized()` returns false, show an onboarding wizard:

1. **Welcome** — "Welcome to env-sync" with brief description
2. **Mode Selection** — Choose security mode with explanations
3. **Initialize** — Create secrets file (optionally encrypted)
4. **Key Setup** — Generate/show AGE key (if encrypted mode)
5. **Done** — Redirect to Dashboard

#### Step 5.2: Confirmation Dialogs
Add confirmation dialogs for destructive operations:
- Delete secret
- Revoke peer/key
- Force pull (overwrites local)
- Mode change (security implications)
- Restore backup (overwrites current)
- Show private key

#### Step 5.3: Toast Notifications
Non-blocking notifications for:
- Sync completed successfully
- Secret added/removed
- Peer approved/revoked
- Errors and warnings

#### Step 5.4: Loading States
All async operations show:
- Loading spinners on buttons
- Skeleton loaders for lists/tables
- Progress indicators for sync operations

#### Step 5.5: Error Handling
- Display Go errors in user-friendly format
- Suggest resolution steps (e.g., "Run init first", "Check SSH access")
- Never expose raw stack traces

---

### Phase 6: Styling & UX

#### Step 6.1: Design System
- Use a minimal CSS framework or utility classes (e.g., a custom design system, or UnoCSS)
- Color scheme: dark mode by default (matches terminal-native aesthetic), light mode toggle
- Monospace font for secrets, keys, fingerprints
- System font for UI elements

#### Step 6.2: Accessibility
- Keyboard navigation for all interactive elements
- ARIA labels on icon-only buttons
- Focus indicators
- Screen reader friendly status announcements

#### Step 6.3: Platform Adaptations
- macOS: Native title bar integration if possible
- Linux: Standard window decorations
- Window min size: 900×600

---

### Phase 7: Testing

#### Step 7.1: Go Backend Tests
- Unit tests for each service struct
- Mock `internal/` packages for isolated testing
- Test all error paths return proper error types

#### Step 7.2: Frontend Tests
- Vue component tests with Vitest + Vue Test Utils
- Store tests (Pinia stores with mocked Wails bindings)
- No E2E tests initially (Wails doesn't have great E2E support)

#### Step 7.3: Build Verification
- Verify `wails build` produces working binary on macOS and Linux
- Verify binary size is reasonable (< 30MB)
- Verify all frontend assets are embedded correctly

---

### Phase 8: Build & Distribution

#### Step 8.1: Makefile Integration
```makefile
# Existing CLI targets (unchanged)
build:
	cd $(SRC_DIR) && $(GO) build -o $(TARGET_DIR)/env-sync ./cmd/env-sync

# New GUI targets (coexist alongside CLI)
build-gui:
	cd $(SRC_DIR) && wails build -o $(TARGET_DIR)/env-sync-gui

dev-gui:
	cd $(SRC_DIR) && wails dev

# Install both or individually
install: build
	install -m 755 $(TARGET_DIR)/env-sync $(PREFIX)/bin/env-sync

install-gui: build-gui
	install -m 755 $(TARGET_DIR)/env-sync-gui $(PREFIX)/bin/env-sync-gui

install-all: install install-gui

# Build both
build-all: build build-gui
```

#### Step 8.2: Update install.sh
- Add `--gui` flag: install GUI binary only
- Add `--all` flag: install both CLI and GUI
- Default (`--cli` or no flag): install CLI only (backward-compatible)
- GUI install checks for WebView dependencies (libwebkit2gtk on Linux)

#### Step 8.3: App Metadata
- Application icon (all platforms)
- `wails.json` configuration:
  - App name: "env-sync"
  - Output filename: "env-sync-gui"
  - Frontend build commands
  - Author info

#### Step 8.3: Documentation
- Update README.md with GUI section explaining:
  - GUI is an alternative interface, not a replacement for CLI
  - Both can be installed and used side-by-side
  - Changes in one are immediately reflected in the other
- Add `docs/GUI.md` with GUI-specific docs
- Update `docs/INSTALLATION.md` with GUI install instructions
- Update `docs/USAGE.md` to mention GUI equivalents for each CLI command
- Screenshot gallery

---

## Key Design Decisions

1. **Dual-interface, shared state**: `env-sync` (CLI) and `env-sync-gui` (GUI) are two independent binaries that operate on the **exact same** config files, secrets, keys, and peer registry. You can use either, or both — the choice is purely about which management interface you prefer.
2. **No API server between CLI and GUI**: Both link the same `internal/` Go packages. The GUI uses Wails bindings (in-process Go calls), while the CLI uses direct function calls. No IPC, no socket, no REST.
3. **File-based interop**: All state lives on disk in `~/.config/env-sync/`. Changes made by the CLI are immediately visible to the GUI on refresh, and vice versa. No migration, no sync between the two.
4. **Independent installation**: `make install` for CLI, `make install-gui` for GUI. Neither depends on the other. Users can install just one or both.
5. **Shared daemon**: The background server (`env-sync serve --daemon`) is a single process. Whether started from CLI or GUI, the same server instance serves peers. Both interfaces can start/stop/restart it.
6. **Vue 3 + Composition API + TypeScript**: Modern, type-safe frontend
7. **Pinia stores**: Centralized state management with Wails binding integration
8. **Wails v2**: Stable, production-ready (upgrade to v3 when it stabilizes)

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Wails v2 EOL when v3 stabilizes | Architecture is clean enough to migrate; service layer is framework-agnostic |
| Large binary size | Wails uses system WebView (not Chromium) — typically 10-20MB |
| WebView differences across Linux distros | Require `libwebkit2gtk-4.0` in install docs |
| Long-running operations block UI | Run sync/discovery in goroutines, use Wails events for progress |
| Sensitive data in frontend memory | Values only fetched on demand (click-to-reveal), not pre-loaded |
| CLI and GUI writing secrets simultaneously | Atomic writes + backup-before-overwrite already in `internal/` packages; last-write-wins matches multi-peer sync semantics |
| GUI showing stale data after CLI change | Dashboard auto-polls status on focus/interval; manual refresh button on all views |

## Dependencies to Add

```
# Go (in go.mod)
github.com/wailsapp/wails/v2  (latest v2.x)

# Frontend (in package.json)
vue@^3.5           # Vue 3 latest
vue-router@^4      # Client-side routing
pinia@^3           # State management
typescript@^5.7    # TypeScript
vite@^6            # Build tool
@vitejs/plugin-vue # Vite Vue plugin
```
