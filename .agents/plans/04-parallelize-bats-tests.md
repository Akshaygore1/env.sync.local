# Plan 04 - Parallelization Opportunities in Bats Tests

## Goal
Identify Bats tests in `tests/bats/` with internal steps that can run concurrently using GNU parallel to reduce wall-clock time without changing test semantics.

## Scope
Files reviewed:
- `tests/bats/01_setup.bats`
- `tests/bats/10_basic_sync.bats`
- `tests/bats/20_encrypted_sync.bats`
- `tests/bats/30_propagation.bats`
- `tests/bats/40_add_machine.bats`
- `tests/bats/99_teardown.bats`
- `tests/bats/test_helper.bash`

## General Guidance for Parallelization
- Prefer parallelizing read-only or independent network checks (e.g., `getent`, `ssh`, `env-sync show`).
- Avoid parallel writes within the same container when the command modifies the same file or keyring (e.g., multiple `env-sync add` or `env-sync key import` in the same container at the same time).
- Use GNU parallel with failure propagation to keep test behavior consistent, for example:
  - `parallel --halt now,fail=1 -k ::: "cmd1" "cmd2" "cmd3"`
- Use grouping to parallelize by container when each group still runs sequentially inside a container.

## File-by-File Plan

### `tests/bats/01_setup.bats`
Parallelization candidates inside tests:
- **"Containers can reach each other via mDNS"**
  - Independent checks:
    - alpha -> beta `getent hosts beta.local`
    - alpha -> gamma `getent hosts gamma.local`
    - beta -> alpha `getent hosts alpha.local`
  - These can run concurrently via `parallel`, each as a `container_exec` command.

- **"SSH connectivity between containers works"**
  - Independent checks:
    - alpha -> beta `ssh ... beta.local echo OK`
    - alpha -> gamma `ssh ... gamma.local echo OK`
  - Safe to run in parallel since they are read-only checks.

Potential helper improvement (optional, not required for this plan):
- `wait_for_containers` in `tests/bats/test_helper.bash` loops over containers sequentially. Consider replacing the per-iteration checks with parallelized `wait_for_container` calls (one per container) and then aggregating results. This is a helper change, not a test-level change.

### `tests/bats/10_basic_sync.bats`
Parallelization candidates inside tests:
- **"Add second secret to beta and verify propagation"**
  - After `add_secret` completes:
    - `trigger_sync` on alpha and gamma can run in parallel.
    - After syncs complete, `get_secret` on alpha and gamma can run in parallel.
  - Avoid parallelizing `add_secret` with syncs; the add must finish first.

### `tests/bats/20_encrypted_sync.bats`
Parallelization candidates inside tests:
- **"Verify encrypted file exists on alpha"**
  - Two read-only checks (file exists and contains "ENCRYPTED: true"). Can run concurrently.

- **"Get beta's and gamma's AGE public keys"**
  - `get_pubkey` for beta and gamma can run in parallel since they are independent and read-only.

- **"Exchange public keys between all containers"**
  - Grouped parallelization by target container:
    - Group 1: imports into alpha (beta + gamma) sequential inside alpha.
    - Group 2: imports into beta (alpha + gamma) sequential inside beta.
    - Group 3: imports into gamma (alpha + beta) sequential inside gamma.
  - Run the three groups in parallel to reduce total time, while avoiding concurrent imports into the same container.

- **"Trigger sync on all containers to exchange encrypted secrets"**
  - `trigger_sync` for alpha, beta, gamma can run in parallel.

- **"Verify all containers have encrypted files"**
  - The loop over containers can be replaced with `parallel` to check each container concurrently.

### `tests/bats/30_propagation.bats`
Parallelization candidates inside tests:
- **"Verify original encrypted secret still exists on all containers"**
  - Three `get_secret` calls can run in parallel.

- **"Sync all containers and verify multiple secrets propagated"**
  - `trigger_sync` on alpha and beta can run in parallel.
  - After syncs complete, `get_secret` checks on alpha and beta can run in parallel.

Not recommended for parallelization:
- **"Add multiple secrets to gamma"**
  - Multiple `env-sync add` calls in the same container likely modify the same file and should remain sequential to avoid races.

### `tests/bats/40_add_machine.bats`
Parallelization candidates inside tests:
- **"Share delta's pubkey with alpha"**
  - After `DELTA_PUBKEY` is captured, imports into beta and gamma can run in parallel.
  - The import into alpha can run concurrently with beta/gamma only if the test still checks alpha's status explicitly; otherwise, keep alpha separate to preserve its `run` assertions.

- **"Sync other containers to get re-encrypted file"**
  - `trigger_sync` on beta and gamma can run in parallel.

Not recommended for parallelization:
- **"Collect public keys on delta from existing machines"**
  - Multiple `env-sync key import` commands into the same delta keyring should remain sequential to avoid conflicts.
- **"Add secret on delta and verify it syncs back to others"**
  - Each sync should complete before verifying; keep sequential per container.

### `tests/bats/99_teardown.bats`
Parallelization candidates inside tests:
- **"Clean up all test containers and volumes"**
  - `docker stop`/`docker rm` for delta could run in parallel, but current ordering is safe and quick; gains are minimal.

- **"Clean up Docker volumes"**
  - Volume removals for alpha/beta/gamma can run in parallel.
  - `wait_for_network_removed` should remain sequential after removals complete.

## Output and Safety Notes
- Use `parallel --halt now,fail=1` to stop on the first failure and preserve test correctness.
- Ensure Bats captures exit codes; for parallel invocations, capture failures by checking the `parallel` exit code.
- Keep concurrency limited to avoid overloading Docker; consider `parallel -j 2` or `-j 3` where appropriate.

## Next Step (When Implementing)
- Add a small helper in `tests/bats/test_helper.bash` to wrap GNU parallel calls with consistent error handling, then refactor the listed tests to call it.
