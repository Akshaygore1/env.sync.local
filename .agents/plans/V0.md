# Distributed Secrets Sync Tool - Implementation Plan

## Overview
A distributed, peer-to-peer secrets synchronization system for local networks using mDNS discovery.

## Key Design Principles

### 1. Truly Distributed (No Master)
- No single point of failure or coordination
- Every machine is equal - can both serve and fetch secrets
- New machines automatically integrate without configuration on existing machines

### 2. Automatic Discovery
- Uses mDNS/Bonjour for automatic peer discovery
- No manual configuration of peer addresses required
- Works with `.local` hostnames (beelink.local, mbp16.local, razer.local)

### 3. Zero Configuration for New Machines
To add a 4th machine (e.g., `surface.local`):
1. Install the tool on the new machine
2. That's it - existing machines will automatically discover it

## Architecture

### File Format: `.secrets.env`
```bash
# === ENV_SYNC_METADATA ===
# VERSION: 1.2.3
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# MODIFIED: 2025-02-07T15:30:45Z
# CHECKSUM: sha256:abc123...
# === END_METADATA ===

OPENAI_API_KEY="sk-xxx"
ANTHROPIC_API_KEY="sk-xxx"
AWS_ACCESS_KEY_ID="AKIA..."

# === ENV_SYNC_FOOTER ===
# VERSION: 1.2.3
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# === END_FOOTER ===
```

### Components

#### 1. `env-sync-discover` - Peer Discovery
- Uses `avahi-browse` (Linux) / `dns-sd` (macOS) / custom solution (Windows)
- Discovers all `_envsync._tcp` services on local network
- Returns list of available peers with their hostnames and ports
- Includes self-detection and filtering

#### 2. `env-sync-serve` - HTTP Server
- Lightweight HTTP server for serving secrets file
- Runs on port 5739 (ENV-SYNC in T9)
- Serves `.secrets.env` with proper headers
- Includes version info in response headers
- Implements simple authentication token (optional)

#### 3. `env-sync-client` - Sync Client
- Compares local version with all discovered peers
- Fetches from peer with newest version
- Implements conflict resolution (latest timestamp wins)
- Validates checksums after download
- Creates backups before overwriting

#### 4. `env-sync` - Main CLI
- Orchestrates discover → compare → sync workflow
- Triggers: bash startup, cron (30min), manual
- Commands:
  - `env-sync` - Run sync
  - `env-sync status` - Show local version and peers
  - `env-sync init` - Initialize secrets file
  - `env-sync serve` - Start HTTP server

#### 5. Shell Integration
- Auto-run on bash/zsh startup (non-blocking, background)
- Source `.secrets.env` in `.profile` / `.bashrc`
- Cron job for periodic sync

## Versioning & Conflict Resolution

### Version Format
- Uses semantic versioning: MAJOR.MINOR.PATCH
- MAJOR: Breaking changes to file format
- MINOR: New secrets added
- PATCH: Existing secrets modified

### Conflict Resolution Strategy
1. Compare timestamps (newer wins)
2. If timestamps equal, compare version numbers (higher wins)
3. If both equal, use lexicographic hostname order (deterministic)
4. Always create backup before overwriting

### Backup Strategy
- Keep last 5 versions: `.secrets.env.backup.1` through `.secrets.env.backup.5`
- Rotate backups on each sync
- Allow manual restore: `env-sync restore [backup_number]`

## Cross-Platform Support

### Linux (Ubuntu/Debian)
- Uses `avahi-daemon` and `avahi-utils`
- Service discovery via `avahi-browse`
- Service publication via `avahi-publish`

### macOS
- Built-in Bonjour/mDNS support
- Uses `dns-sd` command
- No additional dependencies

### Windows
- Options:
  1. Use Bonjour for Windows (Apple's implementation)
  2. Use WSL2 with Linux approach
  3. Use Python-based mDNS library (zeroconf)
- For simplicity: Start with WSL2 support, add native later

## Security (Initial - No Encryption)

### Transport
- HTTP (not HTTPS) for local network only
- Optional: Simple token-based authentication header
- Bind to localhost + local network interfaces only

### File Permissions
- `.secrets.env`: 600 (owner read/write only)
- Backup files: 600
- Configuration: 600

### Future: Encryption Support
- Designed to easily add age/sops
- Will encrypt values while keeping metadata plaintext
- Key management via age keys stored separately

## Network Protocol

### SCP Mode (Default - Secure)
Uses SCP over SSH for encrypted, authenticated file transfer:
```bash
# From source machine to local temp file
scp -o ConnectTimeout=5 hostname.local:~/.secrets.env /tmp/secrets.tmp
```

**Requirements**:
- SSH keys must be set up between machines (`ssh-copy-id`)
- SCP command available on all machines

### HTTP Mode (Fallback - Insecure)
HTTP endpoints for plaintext transfer (not recommended):
```
GET /secrets.env
  Response Headers:
    X-EnvSync-Version: 1.2.3
    X-EnvSync-Timestamp: 2025-02-07T15:30:45Z
    X-EnvSync-Host: beelink.local
    X-EnvSync-Checksum: sha256:abc123...
  Response Body: Raw .secrets.env content

GET /health
  Response: {"status": "ok", "version": "1.2.3", "timestamp": "..."}
```

**Warning**: Only use HTTP mode on completely trusted networks. Displays large security warning when used.

### mDNS Service Registration
- Service Type: `_envsync._tcp`
- Port: 5739
- TXT Records:
  - `version=1.2.3`
  - `timestamp=2025-02-07T15:30:45Z`
  - `hostname=beelink.local`

## Directory Structure

```
~/.config/env-sync/
├── config          # Configuration file
├── .secrets.env    # The secrets file
├── .secrets.env.backup.1..5  # Backups
└── logs/           # Sync logs

/opt/env-sync/ (or /usr/local/env-sync/)
├── bin/
│   ├── env-sync           # Main CLI
│   ├── env-sync-discover  # Discovery tool
│   ├── env-sync-client    # Sync client
│   └── env-sync-serve     # HTTP server
├── lib/
│   └── common.sh          # Shared functions
└── systemd/               # Service files (Linux)
```

## Triggers & Automation

### 1. Bash/Zsh Startup
```bash
# In .bashrc/.zshrc
if command -v env-sync &> /dev/null; then
    (env-sync --background &)
fi
```

### 2. Cron Job (30 minutes)
```bash
# Run sync every 30 minutes
*/30 * * * * /usr/local/bin/env-sync --quiet
```

### 3. Manual Trigger
```bash
env-sync  # Immediate foreground sync
```

## Implementation Phases

### Phase 1: Core Sync (MVP)
- [ ] File format with metadata header/footer
- [ ] HTTP server (env-sync-serve)
- [ ] HTTP client (env-sync-client)
- [ ] Version comparison logic
- [ ] Backup system
- [ ] env-sync CLI with sync/status/init commands

### Phase 2: Discovery
- [ ] mDNS service registration
- [ ] mDNS peer discovery
- [ ] Integration with sync workflow

### Phase 3: Automation
- [ ] Shell integration (bash/zsh)
- [ ] Cron setup
- [ ] Systemd service files (Linux)

### Phase 4: Cross-Platform
- [ ] Linux support (avahi)
- [ ] macOS support (dns-sd)
- [ ] Windows support (WSL2 or native)

### Phase 5: Future Enhancements
- [ ] Encryption support (age/sops)
- [ ] Web UI for viewing/managing secrets
- [ ] Conflict resolution UI
- [ ] Sync history/log viewer

## Testing Plan

### Unit Tests
- Version comparison logic
- File parsing (metadata extraction)
- Checksum validation

### Integration Tests
- Full sync workflow between two VMs
- Conflict resolution scenarios
- Network failure handling

### Manual Testing
- Test with 3 machines (Linux, macOS, Windows/WSL)
- Add 4th machine and verify auto-discovery
- Simulate network partitions

## Adding New Machines

### Scenario: Adding 4th machine (surface.local)

**On surface.local:**
```bash
# 1. Install dependencies
sudo apt install avahi-daemon avahi-utils curl  # Linux
# Or just use built-in tools on macOS

# 2. Download and install env-sync
curl -fsSL https://github.com/user/env-sync/raw/main/install.sh | bash

# 3. Initialize (creates .secrets.env with metadata)
env-sync init

# 4. Add your secrets
nano ~/.secrets.env

# 5. Start the service
env-sync serve --daemon

# 6. Run first sync (optional - will auto-sync anyway)
env-sync
```

**Result:**
- surface.local automatically discovers beelink.local, mbp16.local, razer.local
- Other machines automatically discover surface.local
- surface.local gets the latest secrets from any available peer
- All existing machines will sync from surface.local when it has newer versions
- **No changes needed on existing machines!**

## Success Criteria

- [ ] Secrets sync automatically across all machines within 30 minutes
- [ ] Adding a new machine requires no changes to existing machines
- [ ] Works on Linux, macOS, and Windows (WSL2)
- [ ] No single point of failure
- [ ] Graceful handling of offline machines
- [ ] Clear conflict resolution
- [ ] Backup system prevents data loss
