# env-sync

Distributed secrets synchronization tool for local networks. Sync your `.env` style secrets across multiple machines using **SCP/SSH by default** for security.

## ⚠️ Security Model

**Default Mode: SCP/SSH (Secure)**
- Uses SCP over SSH for encrypted peer-to-peer synchronization
- Requires SSH keys to be set up between machines
- Secrets are never transmitted in plaintext

**Fallback Mode: HTTP (Insecure)**
- Available with `--insecure-http` flag
- ⚠️ **WARNING**: Transmits secrets in plaintext
- ⚠️ Accessible to any device on your local network
- Only use if SSH keys cannot be set up

## Features

- **Secure by Default**: Uses SCP/SSH for encrypted sync
- **Distributed**: No master server, all machines are equal
- **Automatic Discovery**: Uses mDNS/Bonjour to find peers
- **Easy Expansion**: Add new machines without changing existing ones
- **Version Control**: Built-in versioning and conflict resolution
- **Backup System**: Automatic backups before overwriting
- **Cross-Platform**: Works on Linux, macOS, and Windows (WSL2)

## Quick Start

### Prerequisites

Ensure SSH keys are set up between your machines:

```bash
# On each machine, copy your SSH key to other machines
ssh-copy-id beelink.local
ssh-copy-id mbp16.local
ssh-copy-id razer.local
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
# 1. Initialize secrets file
env-sync init

# 2. Edit secrets
nano ~/.secrets.env

# 3. Set up periodic sync (optional)
env-sync cron --install
```

That's it! The machines will automatically discover each other and sync via SCP.

## Usage

### Commands

```bash
env-sync                    # Sync secrets via SCP (secure, default)
env-sync serve -d          # Start HTTP server (for HTTP mode only)
env-sync discover          # Find peers on network
env-sync status            # Show current status
env-sync init              # Create new secrets file
env-sync restore [n]       # Restore from backup (n=1-5)
env-sync cron --install    # Setup 30-min sync cron job
env-sync --help            # Show full help
```

### Insecure HTTP Mode (Not Recommended)

Only use if SSH keys cannot be set up:

```bash
# 1. Start HTTP server on all machines
env-sync serve -d

# 2. Sync using HTTP (displays security warning)
env-sync --insecure-http
```

**⚠️ Security Warning when using HTTP:**
```
╔════════════════════════════════════════════════════════════════════════════╗
║  ⚠️  SECURITY WARNING: USING INSECURE HTTP MODE                            ║
║                                                                            ║
║  You are using the insecure HTTP sync mode. This exposes your secrets:     ║
║  • Transmitted in PLAINTEXT over the network                               ║
║  • Accessible to ANY device on your local network                          ║
║  • No authentication or encryption                                         ║
╚════════════════════════════════════════════════════════════════════════════╝
```

### Setup SSH Keys Between Machines

```bash
# Generate SSH key if you don't have one
ssh-keygen -t ed25519 -C "your-email@example.com"

# Copy to other machines
ssh-copy-id hostname.local

# Test connection
ssh hostname.local "echo Connected!"
```

### Secrets File Format

Your `~/.secrets.env` file includes metadata headers:

```bash
# === ENV_SYNC_METADATA ===
# VERSION: 1.0.1
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# MODIFIED: 2025-02-07T15:30:45Z
# CHECKSUM: sha256:abc123...
# === END_METADATA ===

# Add your secrets here
OPENAI_API_KEY="sk-..."
AWS_ACCESS_KEY_ID="AKIA..."
DATABASE_URL="postgres://..."

# === ENV_SYNC_FOOTER ===
# VERSION: 1.0.1
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# === END_FOOTER ===
```

## How It Works

### 1. Discovery
- Machines advertise themselves via mDNS (`_envsync._tcp`)
- Peers are automatically discovered on the local network
- In SCP mode: Filters for hosts with SSH access available

### 2. Sync Process (SCP Mode - Default)
- When syncing, your machine queries all discovered peers
- Uses SCP over SSH to fetch secrets file from each peer
- Compares version numbers and timestamps
- Downloads the newest version
- Creates backup before overwriting

### 3. Sync Process (HTTP Mode - Insecure)
- Each machine runs an HTTP server on port 5739
- Fetches via unencrypted HTTP
- ⚠️ **WARNING**: Secrets transmitted in plaintext!

### 4. Conflict Resolution
1. Compare timestamps (newer wins)
2. If equal, compare version numbers (higher wins)
3. If both equal, use hostname as tiebreaker

### 5. Triggers
- **Shell startup**: Quick background sync when opening terminal
- **Cron job**: Every 30 minutes (if installed)
- **Manual**: Run `env-sync` anytime

## Adding New Machines

To add a 4th machine (e.g., `surface.local`):

```bash
# On surface.local only:

# 1. Install env-sync
curl -fsSL https://raw.githubusercontent.com/yourusername/env-sync/main/install.sh | bash

# 2. Set up SSH keys
ssh-copy-id beelink.local
ssh-copy-id mbp16.local
ssh-copy-id razer.local

# 3. Initialize
env-sync init

# 4. Sync to get existing secrets
env-sync
```

**No changes needed on existing machines!** They will automatically discover `surface.local` on their next sync.

## Requirements

### Linux (Ubuntu/Debian)
```bash
sudo apt-get install avahi-daemon avahi-utils curl openssh-client
```

### Linux (Fedora/RHEL)
```bash
sudo dnf install avahi avahi-tools curl openssh-clients
```

### macOS
Built-in support. No additional dependencies.

### Windows
Use WSL2 with the Linux instructions above.

## Configuration

### Environment Variables

```bash
ENV_SYNC_QUIET=true      # Suppress output
ENV_SYNC_PORT=5739       # Change server port (HTTP mode only)
```

### Shell Integration

Add to `~/.bashrc` or `~/.zshrc`:

```bash
# Auto-sync on shell startup (background)
if command -v env-sync &> /dev/null; then
    (env-sync --quiet &)
fi

# Source secrets
[[ -f ~/.secrets.env ]] && source ~/.secrets.env
```

## File Locations

```
~/.secrets.env                    # Main secrets file
~/.config/env-sync/config         # Configuration
~/.config/env-sync/backups/       # Backup files
~/.config/env-sync/logs/          # Sync logs
```

## Troubleshooting

### No peers found
```bash
# Check SSH connectivity
ssh hostname.local "echo OK"

# Test discovery
env-sync discover -v  # Verbose discovery

# Check if avahi is running (Linux)
sudo systemctl status avahi-daemon
```

### SSH connection fails
```bash
# Test SSH connectivity
ssh -v hostname.local

# Copy SSH key again
ssh-copy-id hostname.local

# Check SSH key permissions
ls -la ~/.ssh/
chmod 700 ~/.ssh
chmod 600 ~/.ssh/id_*
chmod 644 ~/.ssh/id_*.pub
```

### Sync not working
```bash
# Check status
env-sync status

# Manual sync with verbose output
env-sync sync -f

# View logs
tail -f ~/.config/env-sync/logs/env-sync.log
```

### Permission denied
```bash
# Fix secrets file permissions
chmod 600 ~/.secrets.env
```

## Security Considerations

### Current Implementation

**SCP Mode (Default - Recommended)**
- ✅ Encrypted transmission via SSH
- ✅ Requires SSH key authentication
- ✅ File permissions: 600

**HTTP Mode (Fallback - Insecure)**
- ❌ Secrets transmitted in plaintext
- ❌ Accessible to any device on network
- ❌ No authentication required

### Recommendations

1. **Always use SCP mode (default)**
2. Set up SSH keys between all machines
3. Only use HTTP mode on completely trusted networks
4. Use firewall rules to block port 5739 externally if using HTTP mode
5. Regular backups

### Future: Encryption Support

Even in SCP mode, we plan to add encryption for defense in depth:
- Encrypt values while keeping metadata plaintext
- Support for age/sops encryption
- Key management via age keys

## Development

### Project Structure
```
env-sync/
├── bin/
│   ├── env-sync              # Main CLI
│   ├── env-sync-discover     # Peer discovery
│   ├── env-sync-client       # Sync client (SCP/HTTP)
│   └── env-sync-serve        # HTTP server
├── lib/
│   └── common.sh             # Shared functions
├── install.sh                # Installation script
├── README.md                 # This file
└── AGENTS.md                 # Developer documentation
```

### Testing

```bash
# Test individual components
./bin/env-sync-discover --verbose
./bin/env-sync-client --force
./bin/env-sync-serve --port 9999

# View logs
tail -f ~/.config/env-sync/logs/env-sync.log
```

## License

MIT License - See LICENSE file

## Contributing

Contributions welcome! Please read CONTRIBUTING.md for guidelines.

## Roadmap

- [x] SCP/SSH sync (secure by default)
- [ ] Native Windows support (without WSL)
- [ ] Encryption support (age/sops)
- [ ] Web UI for management
- [ ] Selective sync (whitelist/blacklist secrets)
- [ ] Conflict resolution UI
- [ ] Docker container support

## Resources

- SSH Key Setup: https://www.ssh.com/academy/ssh/copy-id
- mDNS: https://www.ietf.org/rfc/rfc6762.txt
- Avahi: https://www.avahi.org/
- Bonjour: https://developer.apple.com/bonjour/
