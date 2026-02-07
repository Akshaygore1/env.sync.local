# env-sync V3 - Bash to Go Migration Plan

## Overview

Migrate the core env-sync logic from Bash to Go while preserving CLI behavior and keeping the repository layout intact. During migration, the Go binary will be emitted to `./target/env-sync` to avoid conflicting with the existing Bash entrypoint at `./bin/env-sync`. End-to-end tests remain in BATS and continue to invoke commands from Bash.

## Goals

- Preserve command-line compatibility (flags, help text, exit codes, output)
- Keep existing Bash invocation patterns during transition
- Move all core logic into Go with clear, testable packages
- Maintain current file formats, metadata, encryption semantics, and storage paths
- Allow progressive migration without breaking existing workflows

## Non-Goals

- No new repo or major CLI redesign
- No change to the secrets file format or metadata headers/footers
- No replacement of BATS tests with Go-only tests (Go tests are additive)

## Constraints and Requirements

- Go code lives under `./src`
- Top-level `./Makefile` compiles and installs the Go binary
- Final binary name is `env-sync`
- During migration, build output goes to `./target/env-sync`
- CLI must remain binary compatible with current Bash commands and subcommands
- Progressive migration within this repo only

## Current CLI Surface (to preserve)

- Primary command: `env-sync` with subcommands
- Subcommands: `sync`, `serve`, `discover`, `status`, `init`, `restore`, `cron`, `key`, `load`, `add`, `remove`, `list`, `show`, `help`
- Existing companion scripts: `env-sync-client`, `env-sync-discover`, `env-sync-serve`, `env-sync-key`, `env-sync-load`
- Key behaviors: metadata validation, backups, merge by per-line timestamps, SCP/SSH sync (default), HTTP sync (insecure), AGE encryption, mDNS discovery

## Target Go Code Structure

```
src/
  cmd/
    env-sync/
      main.go
  internal/
    cli/                # argument parsing, help, usage
    config/             # paths, env vars, defaults
    logging/            # log levels, file logging, quiet mode
    secrets/            # file format, metadata, parsing, set/get, merge
    metadata/           # header/footer parsing, checksum, version/timestamp
    backup/             # rotate and restore backups
    crypto/age/         # encrypt/decrypt values and files
    keys/               # pubkey cache, request workflow, known_hosts
    discovery/          # mDNS discovery (dns-sd/avahi wrappers)
    transport/ssh/      # SCP/SSH fetch and connectivity checks
    transport/http/     # HTTP fetch
    server/             # HTTP server for /health and /secrets.env
    sync/               # sync orchestration and merge strategy
    cron/               # crontab install/remove/show
    compat/             # formatting of legacy output and warnings
```

Notes:
- Initial Go implementation may shell out to `ssh`, `scp`, `age`, `dns-sd`, `avahi-browse` to preserve behavior, then optionally replace with Go libraries later.
- CLI help strings should match existing Bash output as closely as possible.

## Migration Strategy

Use a progressive replacement model that keeps the Bash entrypoints stable while routing implemented commands to Go. The Go binary will be built into `./target/env-sync`, and Bash scripts can delegate to it per-command when parity is achieved.

### Progressive Delegation Model

1. Keep `bin/env-sync` as the stable shell entrypoint.
2. Add a delegation layer inside `bin/env-sync` to detect `./target/env-sync` and forward subcommands that are already migrated.
3. Maintain fallback to existing Bash logic until each subcommand reaches parity.
4. When all commands are migrated, flip `bin/env-sync` into a thin wrapper that `exec`s the Go binary, or replace it in install steps.

## Phased Plan

### Phase 0 - Inventory and Compatibility Contract

- Document exact CLI behavior per command:
  - Usage strings and help output
  - Exit codes for success and failure
  - Output format (including quiet mode)
  - Error messages and warnings
- Capture secrets file format and metadata rules as a compatibility spec
- Build a BATS compatibility suite that treats current Bash output as golden
- Identify dependencies (age, ssh, scp, curl, dns-sd, avahi-browse, nc, jq, qrencode)

Exit criteria:
- Golden tests cover all commands and common error cases
- Compatibility spec is written and reviewed

### Phase 1 - Go Scaffolding and Build System

- Add `src/go.mod` and basic module layout
- Create top-level `Makefile` targets:
  - `build` (outputs `./target/env-sync`)
  - `test` (Go tests plus BATS if desired)
  - `install` (copy binary to install location, optional)
  - `clean` (remove `./target`)
- Create `cmd/env-sync/main.go` with minimal CLI skeleton and help output
- Add basic config and logging packages

Exit criteria:
- `make build` produces `./target/env-sync`
- `./target/env-sync --help` matches existing help output (or is stubbed with clear TODO while not delegated)

### Phase 2 - Core File and Metadata Logic

- Implement metadata parsing and checksum logic
- Implement secrets file reader/writer with header/footer preservation
- Implement backup rotation and restore
- Implement version and timestamp comparison logic
- Add unit tests for parsing, checksum, metadata updates, merge logic

Exit criteria:
- Go packages mirror Bash behavior for metadata and file operations
- Unit tests cover parsing and checksum validation

### Phase 3 - Read-Only Commands (Low Risk)

- Port `status`, `list`, `show`, `load` to Go
- Preserve output formatting and quiet mode behavior
- Update Bash entrypoint to delegate these commands to Go binary

Exit criteria:
- Golden tests pass for read-only commands
- Bash entrypoint delegates these commands without behavior changes

### Phase 4 - Initialization and Editing Commands

- Port `init` (plain and encrypted) including overwrite prompts
- Port `add` and `remove` with timestamp updates and encryption rules
- Port `restore`
- Update delegation map in Bash entrypoint

Exit criteria:
- BATS tests pass for init/add/remove/restore
- Encryption paths match Bash behavior

### Phase 5 - Discovery and Server

- Port `discover` using external tools initially (dns-sd/avahi)
- Port `serve` with HTTP responses, headers, and daemon mode
- Maintain `/health` and `/secrets.env` behavior and headers

Exit criteria:
- Discovery results match Bash behavior for quiet/verbose
- Server responses match current headers and status codes

### Phase 6 - Sync Orchestration and Transport

- Port sync logic:
  - newest-peer selection
  - merge by per-line timestamps
  - SCP/SSH default transport
  - HTTP fallback with security warning
- Preserve SSH connectivity tests and quiet mode behavior

Exit criteria:
- BATS tests for sync pass for SCP and HTTP
- Same warnings and messages as Bash implementation

### Phase 7 - Key Management and Access Workflow

- Port `key` subcommands:
  - show/export/import/list
  - request-access/grant-access/approve-requests
  - remove/revoke
- Keep cache format compatible (`~/.config/env-sync/keys`)
- Preserve jq/qrencode optional behavior

Exit criteria:
- All key subcommands match current behavior
- Access workflows remain compatible with existing requests files

### Phase 8 - Cron Integration and Final Cutover

- Port cron operations (`install`, `remove`, `show`)
- Update `install.sh` and docs to use Go binary once parity is complete
- Replace or simplify Bash wrappers to exec Go binary

Exit criteria:
- Full parity across all commands
- Bash scripts become thin wrappers or are deprecated safely

## Compatibility Checklist (Per Command)

- `sync`: flags, warnings, newest-peer selection, merge behavior
- `serve`: port selection, daemon mode, headers, checksum, response body
- `discover`: timeout handling, quiet/verbose output, SSH filtering
- `status`: server health probe and peer listing format
- `init`: encrypted/unencrypted, overwrite prompt, metadata fields
- `restore`: backup rotation and error handling
- `key`: cache file paths, request and grant flow, optional tools
- `load`: env/json formats, decrypt-only, key filter
- `add/remove/list/show`: formatting, encryption semantics
- `cron`: crontab entries and update logic

## Risk Mitigation

- Keep a clear fallback to Bash for any command not fully ported
- Preserve file formats and on-disk layout to avoid data migration
- Use golden tests for each command before delegation
- Prefer shelling out to existing system utilities initially to match behavior

## Deliverables

- `./src` Go module with structured packages
- `./Makefile` for build/test/install
- `./target/env-sync` Go binary during migration
- Updated BATS test suite covering CLI compatibility
- Updated documentation explaining the hybrid Bash/Go transition

## Success Criteria

- All existing CLI commands behave identically to Bash implementation
- BATS tests pass using the Go binary via Bash entrypoint
- Go code owns core logic and Bash wrappers are minimal
- No change required for users invoking `env-sync` from the shell
