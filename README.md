# env-sync

Distributed secrets synchronization tool for local networks with **AGE encryption**. Sync your `.env` style secrets securely across multiple machines using SCP/SSH with at-rest encryption.

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

Install AGE encryption:

```bash
# macOS
brew install age

# Linux (Ubuntu/Debian)
sudo apt-get install age

# Linux (Fedora)
sudo dnf install age
```

### Installation

```bash
# Clone or download the repository
git clone https://github.com/yourusername/env-sync.git
cd env-sync

# Install to /usr/local/bin (requires sudo)
sudo ./install.sh

# Or install to ~/.local/bin (user-only)
./install.sh --user
```

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

# Key management
env-sync key show                           # Show your public key
env-sync key list                           # List known peer keys
env-sync key import age1xyz... hostname     # Import peer's key
env-sync key request-access --trigger-all   # Request access on new machine

# Load secrets for shell
env-sync load                               # Output: export KEY=value

# Other commands
env-sync serve -d          # Start HTTP server (for HTTP mode only)
env-sync discover          # Find peers on network
env-sync status            # Show current status
env-sync init              # Create new secrets file
env-sync init --encrypted  # Create with encryption
env-sync restore [n]       # Restore from backup (n=1-5)
env-sync cron --install    # Setup 30-min sync cron job
env-sync --help            # Show full help
```

### Adding a New Machine (Machine D joining A, B, C)

**On the new machine (D) only:**

```bash
# 1. Install env-sync and AGE
curl -fsSL .../install.sh | bash
brew install age  # or apt install age

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

-----BEGIN AGE ENCRYPTED FILE-----
YWdlLWVuY3J5cHRpb24ub3JnL3Yx...
-----END AGE ENCRYPTED FILE-----

# === ENV_SYNC_FOOTER ===
# VERSION: 1.2.3
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# === END_FOOTER ===
```

**Metadata stays plaintext** (for discovery/versioning), only **secret values encrypted**.

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

# Test encryption
echo "test" | age -r $(env-sync key show) | age -d -i ~/.config/env-sync/keys/age_key
```

## Security Considerations

### Current Implementation

**SCP Mode (Default - Recommended)**
- ✅ Encrypted transmission via SSH
- ✅ Requires SSH key authentication
- ✅ File permissions: 600
- ✅ AGE encryption at rest

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
├── bin/
│   ├── env-sync              # Main CLI
│   ├── env-sync-discover     # Peer discovery with key collection
│   ├── env-sync-client       # Sync client with encryption
│   ├── env-sync-serve        # HTTP server
│   ├── env-sync-key          # Key management CLI
│   └── env-sync-load         # Shell integration
├── lib/
│   └── common.sh             # Shared functions + AGE encryption
├── install.sh                # Installation script
├── README.md                 # This file
└── AGENTS.md                 # Developer documentation
```

## License

MIT License

## Roadmap

- [x] AGE encryption (v1.0)
- [x] Multi-recipient encryption
- [x] Remote trigger for machine addition
- [x] Transparent decryption
- [ ] Hardware key support (YubiKey)
- [ ] Web UI for management
- [ ] Key rotation
- [ ] Selective sync (whitelist/blacklist)
