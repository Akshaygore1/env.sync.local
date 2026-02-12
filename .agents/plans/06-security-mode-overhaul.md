# Security Mode Overhaul Plan (Major Upgrade)

## 1) Problem Statement

Current security assumptions are mixed:

- SCP/SSH transport implies broad host-to-host trust.
- AGE at-rest encryption adds limited protection when peer hosts can already SSH/SCP into each other with high privileges.
- In cross-owner scenarios (no passwordless SSH, no shared account trust), current SCP-centered sync and insecure HTTP fallback do not fit.

This plan separates trust models into explicit operation modes and introduces a secure non-SSH peer server architecture.

---

## 2) Security Goals

1. **Make threat models explicit** (no ambiguous “secure by default” claims across incompatible trust assumptions).
2. **Support same-owner convenience mode** without unnecessary crypto overhead.
3. **Support cross-owner collaboration mode** with:
   - encrypted transport,
   - authenticated peers,
   - encrypted at-rest data,
   - explicit authorization and re-encryption workflows.
4. **Avoid requiring passwordless SSH between peers** in collaborative mode.

---

## 3) Target Operation Modes

### Mode A: `dev-plaintext-http` (debug only)
- Storage: plaintext
- Transport: plaintext HTTP
- Discovery: mDNS
- Use case: local debugging only
- UX: loud warnings; never default

### Mode B: `trusted-owner-ssh` (same owner devices)
- Storage: plaintext (default for this mode)
- Transport: SCP/SSH
- Discovery: mDNS + SSH reachability filter
- Use case: all devices belong to one user and are mutually trusted
- AGE: optional/off by default in this mode (no false promise of isolation from SSH-trusted peers)
- Onboarding rule: **zero-touch for existing hosts**. Adding HostB must not require opening terminals or running commands on HostA/other existing hosts.
- Practical join rule: as long as HostB can set up SSH key access to at least one existing peer, HostB can join and sync.

### Mode C: `secure-peer` (cross-owner, no SSH trust)
- Storage: encrypted (AGE multi-recipient)
- Transport: encrypted/authenticated daemon API (HTTPS + mTLS)
- Discovery: mDNS (minimal metadata)
- Use case: different owners sharing selected secrets without shell access
- SSH/SCP: not required for peer communication
- Onboarding rule: invitation/approval is required, but only once per new host (not once per host pair).

---

## 4) Transport Security Decision (researched options)

### Option 1: Basic HTTP + encrypted payload only
- Pros: can keep HTTP plumbing simple
- Cons: harder auth, replay protection, endpoint identity, downgrade resistance

### Option 2: HTTPS with mTLS + peer authorization (**recommended**)
- Pros:
  - mature crypto in Go stdlib (`crypto/tls`, `net/http`)
  - transport confidentiality + integrity + mutual endpoint authentication
  - easier to reason about than custom message crypto
- Cons: certificate lifecycle and trust bootstrapping required

### Option 3: Custom Noise/hand-rolled secure channel
- Pros: flexible, compact protocol possibilities
- Cons: much higher implementation and audit risk; not ideal for first secure-server rollout

**Decision:** Mode C uses **HTTPS + mTLS + certificate/public-key pinning** and keeps AGE for at-rest secret encryption.

---

## 5) Mode C Architecture (Secure Peer Server)

### 5.1 Identity and Trust Material
- Each host gets:
  - a long-term **transport identity keypair** (for mTLS cert identity),
  - an **AGE keypair** (for secret recipient encryption).
- Maintain a peer registry:
  - peer ID/hostname,
  - transport identity fingerprint,
  - AGE public key,
  - authorization state (`pending`, `approved`, `revoked`),
  - capabilities (`read`, `request_reencrypt`, etc).

### 5.1.1 TLS Trust Anchor Policy (important)
- **No global env-sync root certificate will be shipped in code/binaries.**
- Rationale: a shared product-wide root creates unnecessary global trust and high blast radius if compromised.
- Trust is **deployment-local** in Mode C:
  - default: local trust bootstrap via invite + explicit approval + pinned peer identity/cert fingerprint,
  - optional: admin-managed local CA per trust group (not vendor-global).
- Provide explicit commands to inspect trusted cert fingerprints and active trust anchors.

### 5.1.2 How a new secure network bootstraps (HostA -> HostB)
**When HostA is first created (network genesis):**
1. Operator runs `env-sync mode set secure-peer`.
2. HostA generates:
   - transport identity keypair + self-signed TLS cert,
   - AGE keypair,
   - empty peer registry/trust store.
3. HostA starts secure daemon (`serve -d`) and advertises via mDNS.
4. At this point HostA trusts only itself; there is no shared/global root.

**When HostB joins later:**
1. On HostA, operator creates a one-time short-lived invite (`env-sync peer invite --expires <duration>`), producing:
   - enrollment token,
   - HostA transport fingerprint (or pinned cert reference),
   - optional pairing code for human verification.
2. Operator transfers invite details to HostB via an out-of-band channel (chat, terminal copy, QR, etc).
3. On HostB, operator runs `env-sync mode set secure-peer` (HostB generates its own transport cert + AGE key).
4. HostB discovers HostA via mDNS and sends `peer request` with token + HostB identity + HostB AGE public key.
5. HostA validates token, records HostB as `pending` (or auto-approves if policy allows), then operator approves HostB.
6. After approval, both hosts persist each other's transport identity as pinned trust material and store AGE public keys.
7. HostB can now establish mTLS to HostA and request re-encryption; HostA re-encrypts to include HostB recipient; HostB syncs and decrypts locally.

**When HostC joins an existing mesh (e.g., HostA + HostB already trusted):**
1. HostC can be invited by **either** HostA or HostB (any already-trusted peer can sponsor join).
2. Sponsor peer approves HostC and emits a **signed membership event** (HostC transport identity + HostC AGE pubkey + event id + expiry).
3. Signed membership events are replicated mesh-wide as append-only metadata; any online trusted peer can relay them.
4. Existing peers that receive a valid event auto-register HostC as trusted/approved and re-encrypt to include HostC recipient.
5. Result: HostC onboarding effort happens once; HostC does not need separate manual invite/approval with every existing peer.

**Offline HostB scenario (explicit):**
1. If HostB is offline when HostC is approved by HostA, HostB misses that event initially.
2. When HostB returns, it requests missing signed membership events from **any reachable peer** (HostA or HostC).
3. HostB verifies event signature against an already-trusted sponsor identity (e.g., HostA), checks event freshness/expiry and monotonic event id, then pins HostC trust material.
4. After pinning, HostB can establish normal mTLS with HostC and sync secrets; no second manual invite is required.

### 5.1.3 Trust propagation mechanism for offline peers
- Use an append-only `membership-events` log with signed records (gossip/eventual consistency model).
- Membership records are **self-verifying**; transport source does not need to be pre-trusted if signature validation succeeds.
- For practical bootstrap, expose a read-only membership-events endpoint that can serve signed events even before full peer trust is established; apply strict rate limits and minimal metadata disclosure.
- Defend against replay/rollback with:
  - monotonically increasing event IDs (per mesh),
  - event timestamps + expiry windows,
  - last-applied event cursor persisted locally.

**Optional deployment profile:** a local per-group CA can replace pure pinning, but CA root distribution still happens per deployment (not from env-sync vendor/binary).

### 5.2 Discovery
- Continue `_envsync._tcp` mDNS.
- Advertise minimal metadata only (avoid leaking sensitive details in TXT records).
- In secure mode, discovery only tells “reachable peer candidate”, not trust.

### 5.3 Pairing / Initial Registration
- Pairing flow (first contact):
  1. Host A creates short-lived enrollment token (`env-sync peer invite`).
  2. Host B calls Host A pairing endpoint with token + Host B identity + AGE pubkey.
  3. Host A marks request pending or auto-approves by policy.
  4. On approval, both persist each other in peer registry and enable mTLS trust.
- Optional later hardening: SAS/QR verification for MITM-resistant first contact.

### 5.4 Secure API Surface (v2)
- `GET /v2/health` (authn optional, limited data)
- `GET /v2/secrets` (mTLS + authorization required)
- `GET /v2/membership/events` (returns signed membership events for trust propagation/catch-up)
- `POST /v2/peer/request-access` (pair/request flow)
- `POST /v2/peer/approve` (local/admin action)
- `POST /v2/secrets/request-reencrypt` (peer asks owner to add recipient and re-encrypt)

### 5.5 Re-encryption Workflow
- Host B requests access/re-encryption from Host A via authenticated API.
- Host A validates policy and approval state.
- Host A updates recipient set (includes Host B AGE pubkey), re-encrypts secrets, bumps metadata/version.
- Host B next sync fetches encrypted file and decrypts locally with its own AGE private key.

---

## 6) Data/Config Model Changes

Add explicit mode config:
- `sync_mode`: `dev-plaintext-http | trusted-owner-ssh | secure-peer`
- `storage_encryption`: `plaintext | age`
- `transport`: `http | scp | https-mtls`

Add peer registry files (new):
- transport trust store (pinned peer certs/fingerprints)
- peer authorization policy
- pending access requests
- append-only membership events log + local last-applied event cursor

Keep backward compatibility by mapping current behavior to `trusted-owner-ssh` unless user opts into secure-peer.

---

## 7) CLI and UX Plan

### New/updated commands
- `env-sync mode get|set <mode>`
- `env-sync peer invite`
- `env-sync peer request --to <host> --token <token>`
- `env-sync peer approve <peer>`
- `env-sync peer revoke <peer>`
- `env-sync peer list`
- `env-sync peer trust show <peer>`
- `env-sync mode set <mode> [--yes] [--prune-old-material]`

### Existing command behavior updates
- `sync`: route by configured mode
- `serve -d`: in `secure-peer`, starts HTTPS+mTLS server; in dev mode can still allow plaintext HTTP with warnings
- `init`: selects encryption defaults by mode

### Safety UX
- Strong warnings when using Mode A.
- Prevent silent downgrade from secure mode to plaintext modes without explicit confirmation flag.
- **Mode switching is non-destructive by default**: do not auto-delete existing AGE keys, certs, peer registry, or secrets.
- If user requests cleanup (`--prune-old-material`), show impact summary and require explicit approval prompt.
- In non-interactive use, destructive mode transitions require both `--prune-old-material` and `--yes`.
- Switching away from Mode C must show a clear confidentiality downgrade warning before applying changes.

---

## 8) Implementation Phases

- [ ] **Phase 0: Threat model + docs alignment**
  - Update README security model section to mode-based semantics.
  - Clarify that SCP trust and AGE at-rest solve different problems.

- [ ] **Phase 1: Mode framework**
  - Add mode config, validation, and migration defaults.
  - Route sync/server logic by mode.
  - Add safe mode-switch workflow with non-destructive default and explicit destructive confirmation path.

- [ ] **Phase 2: Trusted-owner SSH mode cleanup**
  - Make plaintext default in Mode B.
  - Keep optional AGE but adjust messaging to avoid over-claiming security.

- [ ] **Phase 3: Secure transport foundation**
  - Add HTTPS server/client stack with mTLS support.
  - Add local identity generation and cert/fingerprint persistence.

- [ ] **Phase 4: Peer registry + pairing**
  - Implement invite token flow, pending approvals, peer trust persistence.
  - Implement signed append-only membership events.
  - Implement mesh membership propagation so a host approved by one trusted peer is auto-learned by other trusted peers (including peers that were offline).
  - Add CLI peer management commands.

- [ ] **Phase 5: Secure sync + re-encrypt API**
  - Implement secure endpoints for fetch/request/approve/re-encrypt.
  - Implement membership catch-up endpoint + cursor-based replay protection.
  - Remove SSH dependence for Mode C host↔host operations.
  - Ensure new approved peers are included mesh-wide without per-pair manual onboarding.

- [ ] **Phase 6: Authorization hardening**
  - Endpoint-level authorization checks.
  - Revocation and deny behavior.
  - Replay/rate-limit safeguards for sensitive endpoints.

- [ ] **Phase 7: Test matrix + migration**
  - Unit and integration tests for all three modes.
  - Cross-mode compatibility and downgrade/upgrade migration tests.

---

## 9) Security Validation Plan

- **Mode A tests**: warnings and explicit opt-in behavior.
- **Mode B tests**: SCP sync remains functional; plaintext defaults behave as expected.
- **Mode B onboarding tests**:
  - adding HostB requires no manual action on existing hosts,
  - HostB joins successfully once SSH key access to an existing host is configured.
- **Mode C tests**:
  - unauthenticated peers denied,
  - unapproved peers cannot fetch secrets,
  - approved peer can request re-encryption and then decrypt synced data,
  - revoked peer loses access on next policy update.
- **Mode C transitive onboarding tests**:
  - HostC invited via HostA is auto-accepted by HostB through membership propagation,
  - HostC can sync through either HostA or HostB after single onboarding flow.
- **Mode C offline-catchup tests**:
  - HostB offline during HostC approval later learns HostC from HostA membership events,
  - HostB offline during HostC approval can also learn HostC from HostC-relayed signed events when HostA is offline,
  - replayed/stale membership events are rejected by cursor + expiry checks.
- **Adversarial tests**:
  - MITM attempt on first contact (document TOFU risk until SAS/QR hardening lands),
  - replayed re-encryption request,
  - rogue mDNS advertisement without trusted transport identity.
- **Mode-switch safety tests**:
  - mode switch preserves existing key/cert material by default,
  - destructive cleanup cannot proceed without explicit approval (`--prune-old-material` + confirmation/`--yes`),
  - downgrade from Mode C emits required security warning.

---

## 10) Open Design Questions (to settle before implementation)

1. Pairing UX default: token-only (faster) vs token + SAS/QR (stronger MITM resistance).
2. Deployment-local trust default in Mode C: pinned peer certs only vs local-CA-per-trust-group.
3. Access granularity: all-secrets recipient model now vs future per-secret ACL support.
4. Secure mode default for new installs: immediate vs staged rollout after stabilization.

---

## 11) References Used for Planning

- RFC 8446 (TLS 1.3): https://www.rfc-editor.org/rfc/rfc8446
- RFC 5280 (X.509 PKI/certificate path validation): https://www.rfc-editor.org/rfc/rfc5280
- RFC 8882 (DNS-SD privacy requirements): https://www.rfc-editor.org/rfc/rfc8882
- OWASP REST Security Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html
- age project/spec overview: https://github.com/FiloSottile/age and https://age-encryption.org/v1
