# AGENTS.md - Project Guide for LLM Coding Agents

## Project Overview

**env-sync** is a distributed secrets synchronization tool for local networks. It allows multiple machines to sync `.env` style secrets without a central server, using peer-to-peer architecture with mDNS discovery.

**Current Version**: v2.0 - Rewritten in Go with built-in AGE encryption

## Architecture

### Core Philosophy
- **Distributed**: No master server, all machines are equal
- **Zero Configuration**: New machines auto-discover without touching existing ones
- **Local Network Only**: Uses mDNS/Bonjour for discovery, SCP/SSH for file transfer
- **Eventually Consistent**: Syncs on shell startup, cron (30min), or manual trigger
- **Secure by Default**: AGE encryption at rest, SSH for transport

### Sync Strategy
1. **Discovery**: Find peers via mDNS (`_envsync._tcp` service on port 5739)
2. **Fetch**: Download encrypted secrets from peer via SCP/SSH
3. **Decrypt**: Decrypt using local AGE private key
4. **Comparison**: Compare version/timestamp to find newest file
5. **Merge**: Combine changes (if needed)
6. **Re-encrypt**: Encrypt to all known recipient keys
7. **Backup**: Always backup before overwriting (keep last 5)
8. **Update**: Replace local file and update metadata

## File Structure

```
env.sync.local/
├── PLAN.md                    # Detailed implementation plan & roadmap
├── README.md                  # User documentation & installation guide
├── AGENTS.md                  # This file - internal dev documentation
├── install.sh                 # Installation script (builds Go binary by default)
├── Makefile                   # Build automation
├── src/                       # Go source code (v2.0)
│   ├── cmd/env-sync/          # Main entry point
│   │   └── main.go
│   ├── internal/              # Internal packages
│   │   ├── cli/               # CLI interface and command routing
│   │   ├── sync/              # Sync logic
│   │   ├── discovery/         # mDNS peer discovery
│   │   ├── crypto/age/        # AGE encryption/decryption
│   │   ├── transport/         # SSH/HTTP transport
│   │   │   ├── ssh/           # SCP transport (default)
│   │   │   └── http/          # HTTP transport (fallback)
│   │   ├── server/            # HTTP server
│   │   ├── metadata/          # File metadata handling
│   │   ├── secrets/           # Secrets file management
│   │   ├── backup/            # Backup management
│   │   ├── keys/              # Key management
│   │   ├── config/            # Configuration
│   │   ├── logging/           # Logging utilities
│   │   └── cron/              # Cron job management
│   ├── go.mod                 # Go module definition
│   └── go.sum                 # Go module checksums
├── target/                    # Build output
│   └── env-sync               # Compiled Go binary
├── legacy/                    # Legacy bash version (v1.x)
│   ├── bin/                   # Bash scripts
│   │   ├── env-sync           # Main CLI entry point
│   │   ├── env-sync-discover  # mDNS peer discovery tool
│   │   ├── env-sync-client    # HTTP client for fetching secrets
│   │   ├── env-sync-serve     # HTTP server for serving secrets
│   │   ├── env-sync-key       # Key management CLI
│   │   └── env-sync-load      # Shell integration helper
│   └── lib/                   # Shared libraries
│       └── common.sh          # Common functions & utilities
└── tests/                     # Integration tests (BATS)
    ├── bats/                  # BATS test files
    ├── docker/                # Docker test environment
    └── utils/                 # Test utilities
```

## Go Implementation (v2.0)

### Main Components

#### cmd/env-sync/main.go
**Purpose**: Entry point for the Go binary
**Key Features**:
- Parses command-line arguments
- Routes to appropriate CLI handlers
- Initializes configuration and logging

#### internal/cli/cli.go
**Purpose**: Command-line interface implementation
**Commands**:
- `sync`: Run sync process (default)
- `init`: Initialize secrets file
- `add`: Add/update a secret
- `remove`: Remove a secret
- `list`: List secret keys
- `show`: Show secret value
- `load`: Load secrets for shell
- `key`: Key management subcommands
- `serve`: Start HTTP server
- `discover`: Find peers on network
- `status`: Show current status
- `restore`: Restore from backup
- `cron`: Manage cron job

#### internal/sync/sync.go
**Purpose**: Core sync logic
**Functions**:
- `Sync()`: Main sync orchestration
- `FetchFromPeers()`: Download from discovered peers
- `CompareVersions()`: Determine newest file
- `MergeSecrets()`: Combine changes from multiple sources

#### internal/discovery/discovery.go
**Purpose**: mDNS peer discovery
**Platform Support**:
- **Linux**: Uses `avahi-browse` command
- **macOS**: Uses `dns-sd` command
- **Fallback**: Network scanning

**Functions**:
- `DiscoverPeers()`: Find all peers on network
- `AdvertiseService()`: Announce this machine via mDNS

#### internal/crypto/age/age.go
**Purpose**: AGE encryption/decryption
**Key Features**:
- Built-in AGE library (filippo.io/age)
- Multi-recipient encryption
- Key generation and management
- Individual value encryption (keys stay plaintext)

**Functions**:
- `GenerateKey()`: Create new AGE key pair
- `Encrypt(value, recipients)`: Encrypt a value to multiple recipients
- `Decrypt(encrypted, privateKey)`: Decrypt a value
- `LoadPrivateKey()`: Load private key from disk
- `LoadPublicKey()`: Load public key

#### internal/transport/ssh/ssh.go
**Purpose**: SCP/SSH transport (default, secure)
**Functions**:
- `FetchFile(host, remotePath)`: Download file via SCP
- `TestConnection(host)`: Verify SSH connectivity
- `TriggerReEncrypt(host)`: Remotely trigger re-encryption

#### internal/transport/http/http.go
**Purpose**: HTTP transport (fallback, insecure)
**Functions**:
- `FetchFile(url)`: Download file via HTTP
- `GetHealth(host)`: Check server health

#### internal/server/server.go
**Purpose**: HTTP server for serving secrets
**Endpoints**:
- `GET /health`: JSON status response
- `GET /secrets.env`: Encrypted secrets file with metadata headers

**HTTP Headers**:
- `X-EnvSync-Version`: Semantic version
- `X-EnvSync-Timestamp`: ISO 8601 timestamp
- `X-EnvSync-Host`: Hostname
- `X-EnvSync-Encrypted`: "true" or "false"
- `X-EnvSync-Recipients`: Comma-separated AGE public keys

#### internal/secrets/secrets.go
**Purpose**: Secrets file management
**Functions**:
- `Read()`: Read and parse secrets file
- `Write()`: Write secrets file with metadata
- `AddSecret(key, value)`: Add or update a secret
- `RemoveSecret(key)`: Remove a secret
- `ListKeys()`: Get list of secret keys
- `GetValue(key)`: Get decrypted value for a key

#### internal/metadata/metadata.go
**Purpose**: File metadata handling
**Functions**:
- `Parse(file)`: Extract metadata from file
- `Update(file, version)`: Update metadata
- `Compare(meta1, meta2)`: Compare versions/timestamps
- `Generate()`: Create new metadata

#### internal/backup/backup.go
**Purpose**: Backup management
**Functions**:
- `Create()`: Create backup before modification
- `List()`: List available backups
- `Restore(n)`: Restore from backup number n
- `Rotate()`: Keep only MAX_BACKUPS (5) backups

#### internal/keys/keys.go
**Purpose**: AGE key management
**Functions**:
- `GenerateKeyPair()`: Create new AGE key pair
- `LoadPrivateKey()`: Load private key
- `LoadPublicKey()`: Load public key
- `ImportPeerKey(pubkey, hostname)`: Save peer's public key
- `ListPeerKeys()`: Get all known peer keys
- `GetRecipients()`: Get list of all recipients for encryption

#### internal/config/config.go
**Purpose**: Configuration management
**Constants**:
```go
ENV_SYNC_VERSION = "2.0.0"
ENV_SYNC_PORT = "5739"
SECRETS_FILE = "~/.config/env-sync/.secrets.env"
CONFIG_DIR = "~/.config/env-sync"
BACKUP_DIR = CONFIG_DIR + "/backups"
KEYS_DIR = CONFIG_DIR + "/keys"
LOG_DIR = CONFIG_DIR + "/logs"
MAX_BACKUPS = 5
```

## Secrets File Format (v2.0)

**Location**: `~/.config/env-sync/.secrets.env`

**Encrypted File Structure**:
```bash
# === ENV_SYNC_METADATA ===
# VERSION: 2.0.0
# TIMESTAMP: 2025-02-08T15:30:45Z
# HOST: beelink.local
# MODIFIED: 2025-02-08T15:30:45Z
# ENCRYPTED: true
# PUBLIC_KEYS: beelink.local:age1xyz...,mbp16.local:age1abc...,nuc.local:age1def...
# === END_METADATA ===

OPENAI_API_KEY="YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+..." # ENVSYNC_UPDATED_AT=2025-02-08T15:30:45Z
DATABASE_URL="YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+..." # ENVSYNC_UPDATED_AT=2025-02-08T14:20:10Z

# === ENV_SYNC_FOOTER ===
# VERSION: 2.0.0
# TIMESTAMP: 2025-02-08T15:30:45Z
# HOST: beelink.local
# === END_FOOTER ===
```

**Important Notes**:
- Metadata in header AND footer (for validation)
- Metadata is plaintext (for versioning/discovery)
- Keys are plaintext (for easy viewing/editing)
- Values are individually encrypted using AGE
- Each value has a timestamp for granular merging
- Multi-recipient encryption (encrypted to all known peers)
- File permissions: 600 (owner only)

## Dependencies

### Required (v2.0 Go)
- Go 1.24 or later (for building)
- `avahi-browse` (Linux - for mDNS discovery)
- `dns-sd` (macOS - built-in, for mDNS discovery)

### Go Modules
- `filippo.io/age` - AGE encryption library
- `golang.org/x/crypto` - Cryptographic primitives
- `github.com/kardianos/service` - Cross-platform service management

### Legacy (v1.x Bash) - Optional
- `bash` (v4.0+)
- `curl` (for HTTP requests)
- `nc` or `netcat` (for HTTP server)
- `age` and `age-keygen` (external binaries)
- `avahi-browse` (Linux) or `dns-sd` (macOS)

## Data Flow

### Discovery Flow
```
env-sync discover
└── discovery.DiscoverPeers()
    ├── Detect OS (Linux/macOS/Windows)
    ├── Platform-specific discovery
    │   ├── Linux: exec avahi-browse _envsync._tcp
    │   ├── macOS: exec dns-sd -B _envsync._tcp
    │   └── Fallback: Network scan
    ├── Parse results
    ├── Filter self
    └── Return sorted unique hostnames
```

### Sync Flow (v2.0)
```
env-sync (or cron/shell trigger)
└── sync.Sync()
    ├── discovery.DiscoverPeers()
    ├── For each peer:
    │   ├── transport/ssh.FetchFile() [default]
    │   │   └── exec scp peer:~/.config/env-sync/.secrets.env /tmp/
    │   ├── secrets.Read() and decrypt
    │   └── metadata.Compare()
    ├── Find newest version
    ├── If remote is newer:
    │   ├── backup.Create()
    │   ├── Merge secrets (per-key timestamps)
    │   ├── crypto/age.Encrypt() to all recipients
    │   ├── secrets.Write()
    │   └── metadata.Update()
    └── Return status
```

### Server Flow (v2.0)
```
env-sync serve -d
└── server.Start()
    ├── Check port availability (5739)
    ├── Create HTTP server
    ├── Register handlers:
    │   ├── GET /health → JSON status
    │   └── GET /secrets.env → Encrypted file + headers
    ├── Start listening
    └── Log requests
```

### Key Management Flow (v2.0)
```
env-sync key request-access --trigger beelink.local
└── keys.RequestAccess()
    ├── keys.LoadPublicKey() (local)
    ├── transport/ssh.TriggerReEncrypt(beelink.local)
    │   └── SSH to beelink.local:
    │       ├── env-sync key import <pubkey> <hostname>
    │       └── env-sync sync (re-encrypts with new recipient)
    └── env-sync sync (fetch newly encrypted file)
```

## Building and Testing

### Build Commands
```bash
# Build Go binary
make build

# Run Go unit tests
make test

# Install to /usr/local/bin
sudo make install

# Install using install.sh
./install.sh --user

# Build legacy bash version
./install.sh --legacy
```

### Testing
```bash
# Run all integration tests (Go + legacy)
./test-dockers.sh

# Run Go-only tests
ENV_SYNC_USE_BASH=false ./test-dockers.sh

# Run legacy bash-only tests
ENV_SYNC_SKIP_GO_BUILD=true ENV_SYNC_USE_BASH=true ./test-dockers.sh --skip-go-build

# Run Go unit tests
cd src && go test ./...
```

### Test Environment
- Uses Docker containers to simulate multiple machines
- BATS (Bash Automated Testing System) for integration tests
- Tests both Go and legacy bash implementations
- Validates interoperability between versions

## Version History & Roadmap

### v2.0 (Current)
- ✅ Complete rewrite in Go
- ✅ Built-in AGE encryption (no external age binary needed)
- ✅ Single static binary
- ✅ Backward compatible with v1.x bash version
- ✅ All v1.x features preserved
- ✅ Improved performance and reliability

### v1.0.0 (Legacy - in legacy/ directory)
- ✅ Bash-based implementation
- ✅ AGE encryption (external binary)
- ✅ SCP/SSH sync (secure by default)
- ✅ mDNS discovery (Linux/macOS)
- ✅ HTTP server/client (insecure fallback)
- ✅ Version comparison
- ✅ Backup system
- ✅ Cron automation
- ✅ Multi-recipient encryption
- ✅ Remote trigger for re-encryption
- ✅ CLI secret management

### Future Enhancements
- [ ] Native Windows support (no WSL)
- [ ] Web UI for management
- [ ] Key rotation
- [ ] Selective sync (whitelist/blacklist)
- [ ] Conflict resolution UI
- [ ] Docker container support
- [ ] REST API for programmatic access
- [ ] Hardware key support (YubiKey)

## Security Considerations

### Transport Security (SCP/SSH) - Default
- ✅ Encrypted transmission via SSH
- ✅ Requires SSH key authentication
- ✅ Accessible only to authorized machines
- ⚠️ SSH host keys auto-accepted on first connect (TOFU)

### At-Rest Security (AGE Encryption)
- ✅ Secrets encrypted on disk
- ✅ Multi-recipient encryption
- ✅ Each machine has its own key pair
- ✅ Values encrypted, keys/metadata plaintext
- ✅ Built-in AGE library (no external dependencies)

### HTTP Mode (Fallback - Insecure)
- ❌ Secrets transmitted in plaintext (even if encrypted at rest)
- ❌ Accessible to any device on network
- ❌ No authentication required
- ⚠️ Use only for testing or completely trusted networks

### Recommendations
- Always use SCP mode (default)
- Set up SSH keys between all machines
- Use encrypted secrets (--encrypted flag on init)
- Only use HTTP mode on completely trusted networks
- Use firewall to block port 5739 externally if using HTTP mode
- Regular backups
- Securely backup AGE private keys offline

## Code Style Guidelines

### Go Best Practices
- Follow standard Go conventions (gofmt, go vet)
- Use standard library where possible
- Error handling: explicit error returns
- Package organization: internal/ for private code
- Constants in SCREAMING_SNAKE_CASE
- Functions and variables in camelCase
- Exported names start with capital letter

### Error Handling
- Always check and handle errors explicitly
- Use custom error types for domain-specific errors
- Log errors with context
- Exit with non-zero on failure

### Logging
- Use internal/logging package
- Levels: ERROR, WARN, INFO, DEBUG
- Respect quiet mode flags
- Log to both console and file

## Contributing

When making changes to v2.0:
1. Update relevant documentation (README, AGENTS.md)
2. Test on both Linux and macOS if possible
3. Maintain backward compatibility with v1.x
4. Follow Go code style guidelines
5. Add unit tests for new functionality
6. Run integration tests before submitting
7. Update version number if needed

For legacy (v1.x) changes:
- Legacy code is in `legacy/` directory
- Only critical bug fixes should go to legacy
- New features should be implemented in v2.0

## Resources

- Go: https://go.dev/
- AGE: https://age-encryption.org/
- filippo.io/age: https://pkg.go.dev/filippo.io/age
- mDNS: https://www.ietf.org/rfc/rfc6762.txt
- DNS-SD: https://www.ietf.org/rfc/rfc6763.txt
- Semantic Versioning: https://semver.org/

## Questions?

For implementation questions:
1. Check PLAN.md for design decisions
2. Check README.md for user documentation
3. Review this file (AGENTS.md) for technical details
4. Look at Go source code in `src/` for current implementation
5. Check `legacy/` for bash v1.x reference implementation
