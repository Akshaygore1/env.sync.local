# env-sync

[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![SSH](https://img.shields.io/badge/SSH-Secure%20Transfer-green?logo=openssh&logoColor=white)](https://www.openssh.com/)
[![AGE](https://img.shields.io/badge/AGE-Encryption-orange?logo=age&logoColor=white)](https://age-encryption.org/)

[![Linux](https://img.shields.io/badge/Linux-Supported-FCC624?logo=linux&logoColor=white)](https://kernel.org/)
[![macOS](https://img.shields.io/badge/macOS-Supported-000000?logo=apple&logoColor=white)](https://apple.com/macos/)
[![Windows](https://img.shields.io/badge/Windows%20(WSL2)-Supported-0078D6?logo=windows11&logoColor=white)](https://docs.microsoft.com/en-us/windows/wsl/)
[![BATS Tests](https://github.com/championswimmer/env.sync.local/actions/workflows/bats-tests.yml/badge.svg)](https://github.com/championswimmer/env.sync.local/actions/workflows/bats-tests.yml)

Distributed secrets synchronization tool for local networks with **AGE encryption**. Sync your `.env` style secrets securely across multiple machines using SCP/SSH with at-rest encryption.

![](./docs/cover.png)

## 🆕 What's New in v2.0

**Major Rewrite in Go!**

- **Single Binary**: No more bash scripts - everything is now a single, statically compiled Go binary
- **Built-in AGE Encryption**: AGE encryption library is built-in, no need to install separate `age` package
- **Improved Performance**: Faster sync operations and better resource usage
- **Better Cross-Platform**: More consistent behavior across Linux, macOS, and Windows (WSL2)
- **Easier Installation**: Just build and install one binary instead of multiple scripts
- **Backward Compatible**: v2.0 can sync with v1.x bash-based installations

**Legacy Support**: The bash-based v1.x version is still available in the `legacy/` directory and can be installed with `./install.sh --legacy`

## 🆕 What's New in v1.0 - Encryption Support

- **AGE Encryption**: Secrets are encrypted at rest using AGE encryption
- **Multi-Recipient Encryption**: Each machine has its own key, encrypted to all authorized recipients
- **Transparent Decryption**: Automatic decryption during sync, shell loading, and cron jobs
- **Zero-Config Machine Addition**: Add new machines without modifying existing ones
- **Remote Trigger**: New machines can trigger re-encryption remotely via SSH

## ⚠️ Security Model

**Two Layers of Security:**

1. **Transport Security (SCP/SSH)** - Default
   - Uses SCP over SSH for encrypted, authenticated file transfer
   - Requires SSH keys between machines
   - Prevents eavesdropping during sync

2. **At-Rest Encryption (AGE)** - Optional but Recommended
   - Secrets encrypted on disk using AGE
   - Each machine has its own key pair
   - Multi-recipient encryption (encrypted to all authorized machines)
   - If a machine is compromised, other machines' secrets remain safe

## Features

- ✅ **Secure by Default**: SCP/SSH transport + AGE encryption
- ✅ **Distributed**: No master server, all machines are equal
- ✅ **Automatic Discovery**: Uses mDNS/Bonjour to find peers
- ✅ **Easy Expansion**: Add new machines without touching existing ones
- ✅ **Zero-Config Addition**: New machines trigger re-encryption remotely
- ✅ **Transparent Decryption**: Works seamlessly in shell, cron, and manual sync
- ✅ **Version Control**: Built-in versioning and conflict resolution
- ✅ **Backup System**: Automatic backups before overwriting
- ✅ **Cross-Platform**: Works on Linux, macOS, and Windows (WSL2)

## Quick Start

### Prerequisites

Ensure SSH keys are set up between your machines:

```bash
# On each machine, copy your SSH key to other machines
ssh-copy-id beelink.local
ssh-copy-id mbp16.local
ssh-copy-id razer.local
```

AGE encryption is built into the Go binary, so no additional package install is required.
If you want to troubleshoot encryption manually, you can optionally install the `age` CLI.

### Installation

**Quick Install (Web-based)**

Download and install the latest release directly:

```bash
# Install to /usr/local/bin (requires sudo)
curl -fsSL https://raw.githubusercontent.com/championswimmer/env.sync.local/main/install.sh | sudo bash

# Or install to ~/.local/bin (user-only, no sudo)
curl -fsSL https://raw.githubusercontent.com/championswimmer/env.sync.local/main/install.sh | bash -s -- --user
```

**Install from Source**

```bash
# Clone or download the repository
git clone https://github.com/championswimmer/env.sync.local.git
cd env.sync.local

# Install v2.0 (Go binary) to /usr/local/bin (requires sudo)
sudo ./install.sh

# Or install to ~/.local/bin (user-only)
./install.sh --user

# For legacy bash version (v1.x)
sudo ./install.sh --legacy
```

**Note**: The installation script automatically handles running services. If env-sync is running as a background service (via `env-sync serve -d`), the installer will:
1. Stop the service before installation
2. Replace the binary
3. Restart the service automatically

This ensures seamless upgrades without manual intervention.

### Initial Setup

**On each machine:**

```bash
# 1. Initialize secrets file with encryption
env-sync init --encrypted

# 2. Edit secrets
nano ~/.secrets.env

# 3. Sync with peers (they'll learn your public key)
env-sync

# 4. Set up periodic sync (optional)
env-sync cron --install
```

That's it! The machines will automatically discover each other and sync encrypted secrets.

## Usage

### Commands

```bash
# Sync (auto-decrypts if encrypted, re-encrypts to all recipients)
env-sync
env-sync sync mbp16.local                   # Sync from specific host
env-sync sync --force-pull mbp16.local      # Force overwrite all local secrets from specific host

# Key management
env-sync key show                           # Show your public key
env-sync key list                           # List known peer keys
env-sync key import age1xyz... hostname     # Import peer's key
env-sync key request-access --trigger-all   # Request access on new machine

# Load secrets for shell
env-sync load                               # Output: export KEY=value

# Secret management
env-sync add KEY="value"                    # Add or update a secret
env-sync add OPENAI_API_KEY="sk-..."        # Example: add API key
env-sync remove KEY                         # Remove a secret
env-sync list                               # List all keys (values hidden)
env-sync show KEY                           # Show value of specific key

# Service management
env-sync service stop                       # Stop the background service
env-sync service restart                    # Restart the background service
env-sync service uninstall                  # Uninstall the service completely

# Other commands
env-sync serve -d          # Start HTTP server as a background service (HTTP mode only)
env-sync discover          # Find peers on network
env-sync status            # Show current status
env-sync init              # Create new secrets file
env-sync init --encrypted  # Create with encryption
env-sync restore [n]       # Restore from backup (n=1-5)
env-sync cron --install    # Setup 30-min sync cron job
env-sync --help            # Show full help
```

### Running as a background service

`env-sync serve -d` installs a user-level service that keeps the HTTP server running, advertises `_envsync._tcp` via mDNS/Bonjour, and performs a sync every 30 minutes. The service restarts automatically after login or reboot.

- Linux (systemd user): `systemctl --user status env-sync` (logs: `journalctl --user -u env-sync`)
- macOS (LaunchAgent): `launchctl print gui/$(id -u)/env-sync` (restart: `launchctl kickstart -k gui/$(id -u)/env-sync`)

**Managing the service:**
```bash
env-sync service stop        # Stop the service
env-sync service restart     # Restart the service
env-sync service uninstall   # Remove the service completely
```

The service commands use the native OS service manager (systemd on Linux, launchd on macOS), ensuring proper lifecycle management across platforms.

### Adding a New Machine (Machine D joining A, B, C)

**On the new machine (D) only:**

```bash
# 1. Install env-sync
curl -fsSL .../install.sh | bash

# 2. Initialize with encryption
env-sync init --encrypted

# 3. Discover peers and collect their pubkeys
env-sync discover --collect-keys

# 4. Request access (triggers re-encryption on existing machines)
env-sync key request-access --trigger beelink.local
# OR trigger all online machines:
# env-sync key request-access --trigger-all

# 5. Sync to get encrypted secrets
env-sync
```

**What happens:**
1. D generates its AGE key pair
2. D SSHes into beelink.local and adds its pubkey to recipients
3. beelink.local re-encrypts secrets to include D
4. D syncs and can now decrypt the secrets

**No changes needed on A, B, or C!**

### Managing Secrets

env-sync provides commands to add, remove, list, and view secrets without manually editing the file.

#### Adding Secrets

```bash
# Add a new secret key-value pair
env-sync add OPENAI_API_KEY="sk-abc123xyz"

# Values can include spaces (use quotes)
env-sync add DATABASE_URL="postgres://user:pass@localhost/db"

# Updates existing key if it already exists
env-sync add API_KEY="new-value"
```

**Features:**
- Works with both encrypted and plaintext files
- Automatically creates backup before modification
- Updates metadata (timestamp, checksum)
- Properly handles quotes in values

#### Removing Secrets

```bash
# Remove a secret by key name
env-sync remove OPENAI_API_KEY

# Safe to run - warns if key doesn't exist
env-sync remove NONEXISTENT_KEY
```

#### Listing Secrets

```bash
# List all secret keys (values are hidden for security)
env-sync list

# Output:
# Secrets keys:
#   • OPENAI_API_KEY
#   • DATABASE_URL
#   • AWS_ACCESS_KEY
#
# Total: 3 keys
```

#### Viewing Secrets

```bash
# Show the value of a specific key
env-sync show OPENAI_API_KEY

# Output: sk-abc123xyz
```

**Full Example Workflow:**

```bash
# Initialize with encryption
env-sync init --encrypted

# Add your secrets
env-sync add OPENAI_API_KEY="sk-..."
env-sync add DATABASE_URL="postgres://..."
env-sync add AWS_ACCESS_KEY="AKIA..."

# Verify what you have
env-sync list

# View a specific value when needed
env-sync show DATABASE_URL

# Remove if needed
env-sync remove OLD_API_KEY

# Changes are automatically backed up
ls ~/.config/env-sync/backups/
```

### Force Pull from a Specific Host

Sometimes you want to completely overwrite your local secrets with those from a specific machine, ignoring timestamps and local changes. This is useful when:
- You want to reset your local secrets to match a trusted source
- You've made incorrect local changes and want to revert
- You want to ensure exact consistency with a specific machine

```bash
# Force pull all secrets from a specific host (overwrites local)
env-sync sync --force-pull nodeA.local

# This will:
# 1. Create a backup of your local secrets
# 2. Download secrets from nodeA.local
# 3. Completely overwrite local file (no merging)
# 4. Ignore local timestamps and version comparisons
```

**Important Notes:**
- Requires a specific hostname (won't work without it)
- Creates a backup before overwriting (can restore with `env-sync restore`)
- All local changes will be lost (replaced with remote values)
- Use with caution - normal sync is safer as it merges changes

### Shell Integration

Add to `~/.bashrc` or `~/.zshrc`:

```bash
# Auto-load secrets (decrypts automatically if encrypted)
eval "$(env-sync load 2>/dev/null)"

# Auto-sync on shell startup (background)
if command -v env-sync &> /dev/null; then
    (env-sync --quiet &)
fi
```

### Encrypted File Format

```bash
# === ENV_SYNC_METADATA ===
# VERSION: 1.2.3
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# MODIFIED: 2025-02-07T15:30:45Z
# ENCRYPTED: true
# RECIPIENTS: age1xyz...,age1abc...,age1def...
# === END_METADATA ===

OPENAI_API_KEY="YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOS..." # ENVSYNC_UPDATED_AT=2025-02-07T15:30:45Z
DATABASE_URL="YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOS..." # ENVSYNC_UPDATED_AT=2025-02-07T15:30:45Z

# === ENV_SYNC_FOOTER ===
# VERSION: 1.2.3
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# === END_FOOTER ===
```

**Metadata stays plaintext** (for discovery/versioning).
**Keys are plaintext**, but **values are individually encrypted** using AGE.
Timestamps track when each key was last updated for granular merging.

## How It Works

### 1. Encryption Model

Each machine has its own AGE key pair:
- **Private Key**: `~/.config/env-sync/keys/age_key` (chmod 600)
- **Public Key**: Shared with peers, cached locally

**Multi-Recipient Encryption:**
```
Machine A encrypts to: [A_pubkey, B_pubkey, C_pubkey]
Machine B encrypts to: [A_pubkey, B_pubkey, C_pubkey]
Machine C encrypts to: [A_pubkey, B_pubkey, C_pubkey]
```

When Machine D joins:
1. D generates its key pair
2. D triggers re-encryption on A/B/C via SSH
3. A/B/C re-encrypt to [A, B, C, D]
4. D can now decrypt

### 2. Sync Process

1. **Discovery**: Find peers via mDNS
2. **Fetch**: Download encrypted secrets via SCP
3. **Decrypt**: Decrypt using local private key (if in recipient list)
4. **Compare**: Check versions
5. **Merge**: Use newest version
6. **Re-encrypt**: Encrypt to all known recipients
7. **Save**: Store encrypted file

### 3. Adding New Machines

**Remote Trigger (Preferred):**
```bash
# On new machine D:
env-sync key request-access --trigger beelink.local
```

This:
- SSHes into beelink.local
- Adds D's pubkey to beelink's cache
- Triggers sync (re-encrypts with D as recipient)
- D can immediately sync and decrypt

**No manual approval needed** - works because D must have SSH access anyway.

### 4. Transparent Decryption

```bash
# During sync (automatic)
env-sync

# During shell load
eval "$(env-sync load)"

# During cron
*/30 * * * * env-sync sync --quiet && eval "$(env-sync load --quiet)"
```

## File Locations

```
~/.config/env-sync/
├── config                          # Config file
├── .secrets.env                    # Encrypted secrets
├── .secrets.env.backup.1..5        # Encrypted backups
├── keys/
│   ├── age_key                     # Private key (chmod 600)
│   ├── age_key.pub                 # Public key
│   ├── known_hosts/                # Cached peer pubkeys
│   │   ├── beelink.local.pub
│   │   └── mbp16.local.pub
│   └── cache/
│       └── pubkey_cache.json       # Metadata about known keys
└── logs/                           # Sync logs
```

## Troubleshooting

### Cannot decrypt - not in recipient list
```bash
# Request access from an existing machine
env-sync key request-access --trigger beelink.local

# Or manually add your key on any existing machine:
# On existing machine:
env-sync key import $(new_machine_pubkey) new_machine.local
env-sync  # Re-encrypts with new recipient
```

### SSH connection fails
```bash
# Test SSH connectivity
ssh -v hostname.local

# Copy SSH key again
ssh-copy-id hostname.local
```

### Sync not working
```bash
# Check status
env-sync status

# View logs
tail -f ~/.config/env-sync/logs/env-sync.log

# Test encryption (optional, if age CLI is installed)
echo "test" | age -r $(env-sync key show) | age -d -i ~/.config/env-sync/keys/age_key
```

## Security Considerations

### Current Implementation

**SCP Mode (Default - Recommended)**
- ✅ Encrypted transmission via SSH
- ✅ Requires SSH key authentication
- ✅ File permissions: 600
- ✅ AGE encryption at rest
- ⚠️ SSH host keys are auto-accepted on first connect (StrictHostKeyChecking=accept-new)
  - This is TOFU behavior and can enable MITM attacks on first connection
  - Set `ENV_SYNC_STRICT_SSH=true` and pre-populate known_hosts for production

**HTTP Mode (Fallback - Insecure)**
- ❌ Secrets transmitted in plaintext
- ❌ Accessible to any device on network
- ⚠️ Use only for testing or completely trusted networks

### Key Management

1. **Private Key Protection**
   - Never transmit private key
   - File permissions: 600
   - Backup securely (offline or encrypted)

2. **Revoking Access**
   ```bash
   # Remove compromised machine
   env-sync key revoke compromised.local
   # Re-encrypts without their key
   ```

3. **Lost Key Recovery**
   - If you lose your private key, you cannot decrypt
   - Must get re-encrypted file from another machine
   - Or restore from backup with old key

## Development

### Project Structure
```
env-sync/
├── src/                       # Go source code
│   ├── cmd/env-sync/          # Main entry point
│   ├── internal/              # Internal packages
│   │   ├── cli/               # CLI interface
│   │   ├── sync/              # Sync logic
│   │   ├── discovery/         # mDNS discovery
│   │   ├── crypto/age/        # AGE encryption
│   │   ├── transport/         # SSH/HTTP transport
│   │   └── ...
│   └── go.mod                 # Go module definition
├── target/
│   └── env-sync               # Built Go binary
├── legacy/                    # Legacy bash v1.x (for reference)
│   ├── bin/                   # Bash scripts
│   └── lib/                   # Bash libraries
├── install.sh                 # Installation script
├── Makefile                   # Build automation
├── README.md                  # This file
└── AGENTS.md                  # Developer documentation
```

### Building from Source

```bash
# Build the Go binary
make build

# Run tests
make test

# Install locally
make install

# Or use install.sh
./install.sh --user
```

### Legacy Bash Version

To use the legacy bash-based version (v1.x):

```bash
# Install legacy version
./install.sh --legacy

# Or force use of bash scripts even if Go binary is present
ENV_SYNC_USE_BASH=true env-sync status
```

## License

MIT License

## Roadmap

- [x] AGE encryption (v1.0)
- [x] Multi-recipient encryption
- [x] Remote trigger for machine addition
- [x] Transparent decryption
- [x] CLI secret management (add/remove/list/show)
- [ ] Hardware key support (YubiKey)
- [ ] Web UI for management
- [ ] Key rotation
- [ ] Selective sync (whitelist/blacklist)
