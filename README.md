# env-sync

Distributed secrets synchronization tool for local networks. Sync your `.env` style secrets across multiple machines without a central server.

## Features

- **Distributed**: No master server, all machines are equal
- **Automatic Discovery**: Uses mDNS/Bonjour to find peers automatically
- **Easy Expansion**: Add new machines without changing existing ones
- **Version Control**: Built-in versioning and conflict resolution
- **Backup System**: Automatic backups before overwriting
- **Cross-Platform**: Works on Linux, macOS, and Windows (WSL2)

## Quick Start

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

# 3. Start the server
env-sync serve -d

# 4. Set up periodic sync (optional)
env-sync cron --install
```

That's it! The machines will automatically discover each other and sync.

## Usage

### Commands

```bash
env-sync                    # Sync secrets from network (default)
env-sync serve -d          # Start HTTP server as daemon
env-sync discover          # Find peers on network
env-sync status            # Show current status
env-sync init              # Create new secrets file
env-sync restore [n]       # Restore from backup (n=1-5)
env-sync cron --install    # Setup 30-min sync cron job
env-sync --help            # Show full help
```

### Sync Options

```bash
env-sync sync -a          # Sync from all discovered peers
env-sync sync -f          # Force sync even if local is newer
env-sync sync hostname    # Sync from specific host
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
- Each machine runs an HTTP server on port 5739
- Machines advertise themselves via mDNS (`_envsync._tcp`)
- Peers are automatically discovered on the local network

### 2. Sync Process
- When syncing, your machine queries all discovered peers
- Compares version numbers and timestamps
- Downloads the newest version
- Creates backup before overwriting
- Updates local metadata

### 3. Conflict Resolution
1. Compare timestamps (newer wins)
2. If equal, compare version numbers (higher wins)
3. If both equal, use hostname as tiebreaker

### 4. Triggers
- **Shell startup**: Quick background sync when opening terminal
- **Cron job**: Every 30 minutes (if installed)
- **Manual**: Run `env-sync` anytime

## Adding New Machines

To add a 4th machine (e.g., `surface.local`):

```bash
# On surface.local only:

# 1. Install env-sync
curl -fsSL https://raw.githubusercontent.com/yourusername/env-sync/main/install.sh | bash

# 2. Initialize
env-sync init

# 3. Start server
env-sync serve -d

# 4. Sync to get existing secrets
env-sync
```

**No changes needed on existing machines!** They will automatically discover `surface.local` on their next sync.

## Requirements

### Linux (Ubuntu/Debian)
```bash
sudo apt-get install avahi-daemon avahi-utils curl netcat-openbsd
```

### Linux (Fedora/RHEL)
```bash
sudo dnf install avahi avahi-tools curl nmap-ncat
```

### macOS
Built-in support (uses `dns-sd`). No additional dependencies.

### Windows
Use WSL2 with the Linux instructions above.

## Configuration

### Environment Variables

```bash
ENV_SYNC_QUIET=true      # Suppress output
ENV_SYNC_PORT=5739       # Change server port
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
~/.config/env-sync/server.pid     # Server process ID
```

## Network Details

- **Port**: 5739 (ENV-SYNC in T9)
- **Protocol**: HTTP (local network only)
- **Discovery**: mDNS/Bonjour (`_envsync._tcp`)
- **Hostname**: Uses `.local` domain (e.g., `beelink.local`)

## Troubleshooting

### No peers found
```bash
# Check if avahi is running (Linux)
sudo systemctl status avahi-daemon

# Manual discovery
avahi-browse -a  # List all mDNS services

# Test connectivity
env-sync discover -v  # Verbose discovery
curl http://beelink.local:5739/health  # Test specific host
```

### Server won't start
```bash
# Check if port is in use
lsof -i :5739

# Try different port
env-sync serve -p 5740
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

## Security Notes

### Current Implementation
- No encryption (secrets stored in plaintext)
- HTTP transport (local network only)
- File permissions: 600 (owner only)

### Future: Encryption Support
The tool is designed to easily add encryption later:
- Will encrypt values while keeping metadata plaintext
- Support for age/sops encryption
- Key management via age keys

### Recommendations
- Only use on trusted local networks
- Use firewall rules to block port 5739 from external networks
- Consider VPN if syncing across untrusted networks

## Development

### Project Structure
```
env-sync/
├── bin/
│   ├── env-sync              # Main CLI
│   ├── env-sync-discover     # Peer discovery
│   ├── env-sync-client       # Sync client
│   └── env-sync-serve        # HTTP server
├── lib/
│   └── common.sh             # Shared functions
├── install.sh                # Installation script
├── README.md                 # This file
└── PLAN.md                   # Implementation plan
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

- [ ] Native Windows support (without WSL)
- [ ] Encryption support (age/sops)
- [ ] Web UI for management
- [ ] Selective sync (whitelist/blacklist secrets)
- [ ] Conflict resolution UI
- [ ] Docker container support
