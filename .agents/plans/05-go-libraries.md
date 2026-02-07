# Go library review (current Go port)

Scope: review the Go port and only suggest libraries where there is a clear, material advantage over the current stdlib/external-tool approach.

Code touchpoints reviewed:
- src/internal/cron/cron.go
- src/internal/transport/ssh/ssh.go
- src/internal/crypto/age/age.go
- src/internal/server/server.go
- src/internal/discovery/discovery.go (secondary)

## Recommendations

### AGE cryptography
Current: shelling out to `age` and `age-keygen` in `src/internal/crypto/age/age.go`.

Recommendation: consider adopting the official Go age library.
- Library: `filippo.io/age` (official, widely used)
- Why it is worth it:
  - Removes runtime dependency on external `age` and `age-keygen` binaries.
  - Cleaner error handling and easier unit tests (no exec).
  - Works well with static builds and environments without package managers.
  - Still compatible with existing key formats (X25519 identity strings).
- Why not mandatory:
  - Current approach is simple and already uses the canonical age CLI.
  - Migration requires careful mapping of key generation and storage.

If you keep the CLI approach, no library is required.

### SSH transport
Current: calls out to `ssh` and `scp` binaries in `src/internal/transport/ssh/ssh.go`.

Recommendation: keep current approach unless you need a pure-Go SSH stack.
- Potential libraries if you want a pure-Go transport:
  - `golang.org/x/crypto/ssh` (official Go SSH package)
  - `github.com/pkg/sftp` (file transfer over SSH)
  - or `github.com/bramvdbogaerde/go-scp` (SCP protocol helper)
- Why not recommended right now:
  - You would need to re-implement SSH config parsing, agent support, and known_hosts handling.
  - External `ssh`/`scp` already respect user config, ProxyCommand, and agent.
  - Current exec usage is not shell-based, so injection risk is low.
- When it becomes beneficial:
  - You want a self-contained binary without `ssh`/`scp` installed.
  - You need Windows-native support without external OpenSSH.
  - You want fine-grained telemetry or connection pooling.

### Cron scheduling
Current: edits the user's crontab in `src/internal/cron/cron.go`.

Recommendation: no library needed for the current model.
- Common library: `github.com/robfig/cron/v3` (in-process scheduler)
- Why it is not a fit today:
  - That library schedules jobs in-process and requires a long-running daemon.
  - Your current model uses OS cron to run the binary periodically, which is simpler.
- When it becomes beneficial:
  - You move to a persistent daemon/service and want internal scheduling.

### Server / daemonization
Current: stdlib `net/http` with a background re-exec in `src/internal/server/server.go`.

Recommendation: keep stdlib HTTP server; only consider a service manager if you want OS-level install/start/stop.
- Potential library if you want installable services:
  - `github.com/kardianos/service` (popular cross-platform service manager)
- Why not recommended right now:
  - Adds complexity and OS-specific concerns.
  - Current behavior (start in background) is adequate for a simple local server.
- Low-effort improvements without a library:
  - Add signal handling to clean up PID files and allow graceful shutdown.

## Considered but not recommended

### mDNS discovery
Current: uses `avahi-browse` and `dns-sd` in `src/internal/discovery/discovery.go`.

Potential libraries:
- `github.com/grandcat/zeroconf`
- `github.com/hashicorp/mdns`

Why not recommended right now:
- OS tools are reliable and already installed on macOS/Linux targets.
- Go mDNS libraries can be finicky across networks and add more surface area.

If you later want a single static binary with no external tools, this is a candidate for revisiting.
