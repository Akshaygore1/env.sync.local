# Security Modes

env-sync v3.0+ operates in three distinct security modes, each designed for different trust scenarios. This document explains the security guarantees, trade-offs, and appropriate use cases for each mode.

## Table of Contents

- [Overview](#overview)
- [Mode A: dev-plaintext-http](#mode-a-dev-plaintext-http)
- [Mode B: trusted-owner-ssh](#mode-b-trusted-owner-ssh)
- [Mode C: secure-peer](#mode-c-secure-peer)
- [Comparison Matrix](#comparison-matrix)
- [Choosing a Mode](#choosing-a-mode)
- [Mode Switching](#mode-switching)
- [Security Threat Models](#security-threat-models)

## Overview

| Mode | Storage | Transport | Use Case |
|------|---------|-----------|----------|
| **dev-plaintext-http** | Plaintext | Plaintext HTTP | Local debugging only |
| **trusted-owner-ssh** | Plaintext (default) | SCP/SSH | Same owner, mutually trusted devices |
| **secure-peer** | AGE Encrypted | HTTPS+mTLS | Cross-owner, no shared SSH trust |

Each mode makes different security trade-offs. Choose based on your actual trust boundaries, not assumptions about what's "most secure."

## Mode A: dev-plaintext-http

**Purpose**: Local development and debugging only

### Security Characteristics

- **Storage**: Secrets stored in plaintext on disk
- **Transport**: Unencrypted HTTP (port 5739)
- **Authentication**: None
- **Encryption**: None

### When to Use

- Debugging env-sync itself
- Local testing in isolated environments
- Learning how the system works

### When NOT to Use

- Production environments
- Shared networks
- Any scenario involving real secrets

### Warnings

This mode displays prominent warnings on every command:

```
⚠️  WARNING: Running in dev-plaintext-http mode
   Secrets are stored and transmitted in plaintext!
   This mode is for debugging only.
```

### Commands

```bash
# Switch to dev mode (requires explicit confirmation)
env-sync mode set dev-plaintext-http

# Start HTTP server (no encryption)
env-sync serve -d
```

## Mode B: trusted-owner-ssh

**Purpose**: Synchronize secrets across devices owned by the same person/organization

### Security Characteristics

- **Storage**: Plaintext by default (optional AGE encryption)
- **Transport**: SCP over SSH
- **Authentication**: SSH key-based
- **Trust Model**: Implicit - any peer with SSH access is trusted

### Rationale

In this mode, all devices are mutually trusted through SSH. If Machine A can SSH into Machine B, it already has significant access to that system. Adding AGE encryption provides limited additional protection because:

1. Compromising one machine likely exposes SSH keys to others
2. The attacker can use those keys to access other machines directly
3. Once on another machine, they can read the secrets regardless of encryption

Therefore, **plaintext storage is the honest default** - it doesn't claim security that SSH trust doesn't actually provide.

### When to Use

- All devices belong to you
- SSH keys are already set up between devices
- You trust all devices equally
- You want zero-config peer addition

### Key Features

#### Zero-Touch Peer Addition

Adding a new machine requires **zero action on existing machines**:

```bash
# On new machine only:
env-sync init                    # Creates plaintext secrets file
env-sync discover                # Find peers
env-sync                         # Sync (SSH key must be set up first)
```

That's it. No commands needed on existing peers.

#### Practical Join Rule

As long as the new machine can set up SSH key access to at least one existing peer, it can join and sync.

### Optional Encryption

You can still enable AGE encryption if you want:

```bash
# Initialize with encryption
env-sync init --encrypted

# Or add encryption later
env-sync mode set trusted-owner-ssh --encrypted
```

**Note**: This provides defense-in-depth for specific scenarios (e.g., laptop theft), but doesn't protect against network-level attacks from compromised peers that have SSH access.

### Commands

```bash
# Set mode (default)
env-sync mode set trusted-owner-ssh

# Initialize with optional encryption
env-sync init --encrypted

# Discover peers (filters by SSH reachability)
env-sync discover

# Request access on new machine
env-sync key request-access --trigger hostname.local
```

### Security Considerations

**Strengths:**
- ✅ SSH provides encrypted transport
- ✅ SSH key authentication required
- ✅ File permissions enforced (600)
- ✅ Well-understood security model

**Limitations:**
- ⚠️ TOFU (Trust On First Use) for SSH host keys
- ⚠️ Compromised peer can access all SSH-trusted machines
- ⚠️ No fine-grained authorization controls

## Mode C: secure-peer

**Purpose**: Share secrets across different owners without requiring SSH trust

### Security Characteristics

- **Storage**: AGE encrypted (mandatory)
- **Transport**: HTTPS with mutual TLS (mTLS)
- **Authentication**: Certificate-based mutual authentication
- **Trust Model**: Explicit - peers must be invited and approved

### Rationale

When different owners want to share selected secrets, SSH trust is inappropriate:

1. Owner A shouldn't have shell access to Owner B's machine
2. Owner B shouldn't trust Owner A's SSH key blindly
3. At-rest encryption becomes meaningful when transport peers aren't fully trusted

This mode uses **mTLS** (mutual TLS) where both client and server present certificates, enabling authentication without granting shell access.

### When to Use

- Collaboration between different people/teams
- Shared secrets without shared system access
- Environments where SSH trust is inappropriate
- When you need explicit authorization controls

### Architecture

#### Identity and Trust Material

Each host has:

1. **Transport Identity**: Keypair + self-signed certificate for mTLS
2. **AGE Keypair**: For at-rest secret encryption
3. **Peer Registry**: Database of known peers and their authorization status

```
~/.config/env-sync/
├── keys/
│   ├── transport_key          # mTLS private key
│   ├── transport_cert.pem     # Self-signed certificate
│   ├── age_key                # AGE private key
│   └── age_key.pub            # AGE public key
└── peers/
    ├── registry.json          # Peer registry
    └── trust_store/           # Pinned peer certificates
```

#### Trust Bootstrap (No Global Root CA)

**No global env-sync root certificate is shipped.** Trust is deployment-local:

1. First host (genesis) creates its own identity
2. New hosts are invited via enrollment tokens
3. Existing peers approve new hosts explicitly
4. Approved peers exchange and pin each other's certificates

This avoids global trust and blast radius issues.

### Key Features

#### Invitation-Based Onboarding

```bash
# On existing trusted peer
env-sync peer invite --expires 1h
# Outputs: token, hostname, fingerprint

# On new machine
env-sync mode set secure-peer
env-sync peer request-access --to hostname.local --token <token>

# Back on existing peer
env-sync peer approve new-host.local
```

#### Membership Propagation

Once approved by one trusted peer, a new host is automatically learned by all other trusted peers (including those that were offline):

1. Signed membership events are created when a peer is approved
2. Events are replicated mesh-wide
3. Offline peers catch up when they come online
4. No need for separate invites with every peer

#### Explicit Authorization

```bash
# View pending requests
env-sync peer list --pending

# Approve or revoke
env-sync peer approve hostname.local
env-sync peer revoke hostname.local

# View authorization status
env-sync peer list
```

### Commands

```bash
# Switch to secure-peer mode
env-sync mode set secure-peer

# Create invitation
env-sync peer invite --expires 30m --description "John's laptop"

# Request access
env-sync peer request-access --to hostname.local --token <token>

# Manage peers
env-sync peer list
env-sync peer approve hostname.local
env-sync peer revoke hostname.local
env-sync peer trust show hostname.local

# Sync (uses mTLS)
env-sync
```

### Security Considerations

**Strengths:**
- ✅ Encrypted at-rest (AGE)
- ✅ Encrypted transport (TLS 1.3)
- ✅ Mutual authentication (mTLS)
- ✅ Explicit authorization required
- ✅ No SSH access needed
- ✅ Certificate pinning (no global CA)
- ✅ Membership events with replay protection

**Limitations:**
- ⚠️ First contact vulnerable to MITM (TOFU) without SAS/QR verification
- ⚠️ More complex setup than trusted-owner mode
- ⚠️ Requires running daemon (HTTPS server)

## Comparison Matrix

| Feature | dev-plaintext-http | trusted-owner-ssh | secure-peer |
|---------|-------------------|-------------------|-------------|
| **Storage** | Plaintext | Plaintext (opt: encrypted) | Encrypted (AGE) |
| **Transport** | HTTP | SCP/SSH | HTTPS+mTLS |
| **Auth** | None | SSH keys | Certificates |
| **Onboarding** | Automatic | Zero-touch | Invitation + approval |
| **Encryption** | None | Optional | Mandatory |
| **Use Case** | Debugging only | Same owner | Cross-owner |
| **Complexity** | Minimal | Low | Medium |
| **Daemon Required** | Optional | Optional | Yes |

## Choosing a Mode

### Use `trusted-owner-ssh` if:

- All devices belong to you
- You have SSH keys set up
- You want simple setup
- You don't need fine-grained access control

**This is the default mode for new installs.**

### Use `secure-peer` if:

- Collaborating with others
- Different people own different machines
- You need explicit authorization
- SSH trust is inappropriate

### Use `dev-plaintext-http` if:

- Debugging the application
- Learning how it works
- Never for production use

## Mode Switching

### Non-Destructive by Default

Mode switches preserve existing data:

```bash
# Switch modes (keys and secrets preserved)
env-sync mode set secure-peer
```

Your existing keys, certificates, and secrets remain available.

### Cleanup Option

To remove old mode's material:

```bash
# Switch and clean up
env-sync mode set trusted-owner-ssh --prune-old-material --yes
```

**Warning**: This deletes mode-specific data (certificates, peer registry, etc.)

### Downgrade Warnings

Switching from secure-peer to less secure modes requires explicit confirmation:

```bash
$ env-sync mode set dev-plaintext-http
⚠️  SECURITY WARNING: You are switching from secure-peer to dev-plaintext-http.
   This will disable encryption and mTLS.
   Secrets will be stored and transmitted in plaintext!

   Use --yes to confirm you understand the risks.
```

## Security Threat Models

### Mode A Threat Model

**Assumption**: Isolated local development environment

**Protects against**: None

**Vulnerable to**:
- Network sniffing
- Unauthorized access
- File system access
- Everything

### Mode B Threat Model

**Assumption**: All peers are equally trusted; SSH compromise is game-over anyway

**Protects against**:
- Passive network eavesdropping (SSH encryption)
- Unauthorized hosts without SSH keys

**Vulnerable to**:
- Compromised peer with SSH access to others
- SSH host key MITM on first connection (TOFU)
- Insider threats from SSH-trusted machines

**Mitigations**:
- Enable AGE encryption for defense-in-depth
- Use `ENV_SYNC_STRICT_SSH=true` and pre-populate known_hosts
- Regular key rotation

### Mode C Threat Model

**Assumption**: Peers don't inherently trust each other; explicit authorization required

**Protects against**:
- Passive network eavesdropping (TLS)
- Active MITM (mTLS mutual auth)
- Unauthorized peer access (explicit approval)
- Compromised peer lateral movement (no SSH trust)
- At-rest file exposure (AGE encryption)

**Vulnerable to**:
- First-contact MITM without verification (TOFU)
- Approved but malicious peer
- Compromised approved peer's private keys

**Mitigations**:
- Verify transport fingerprints out-of-band
- Regular peer audits and revocations
- Principle of least privilege
- Monitor membership events for anomalies

## Recommendations

1. **Start with `trusted-owner-ssh`** for personal use - it's simple and sufficient
2. **Use `secure-peer`** for collaboration - it's designed for that scenario
3. **Never use `dev-plaintext-http`** for real secrets
4. **Enable encryption** in trusted-owner mode if you want defense-in-depth
5. **Verify fingerprints** in secure-peer mode when security is critical
6. **Regular backups** regardless of mode
7. **Monitor logs** for unauthorized access attempts
