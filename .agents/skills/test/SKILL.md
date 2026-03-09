---
name: Testing env-sync
description: Testing the mDNS discovery and sync logic of env-sync across multiple virtual machines.
---
# Testing env-sync

The project uses `bats-core` for integration testing within Docker containers. This ensures that the mDNS discovery and sync logic work across multiple virtual machines.

## Running Tests

To run the full suite of integration tests:

```bash
./test-dockers.sh
```

### Options

- `--filter <pattern>`: Run only tests matching the pattern (e.g., `--filter basic`).
- `--no-cleanup`: Keep the Docker containers running after tests. Useful for debugging.
- `--setup-only`: Prepare the environment (SSH keys, Docker images) without running tests.
- `--formatter <fmt>`: Choose the output format. Supported values include `pretty`, `tap`, `junit`.

## Agent Guidelines

When running tests as an AI agent, **always use the TAP formatter**:

```bash
./test-dockers.sh --formatter tap
```

### Why TAP?
1. **Linear Output**: Unlike the `pretty` formatter, TAP produces a simple, line-by-line stream. This avoids issues with terminal buffer refreshes or "jumping" cursors that can confuse log analysis.
2. **Readability**: It provides a clear, machine-readable sequence of passed and failed tests.
3. **Reliability**: TAP is the standard for CI environments and works best with captured stdout.
