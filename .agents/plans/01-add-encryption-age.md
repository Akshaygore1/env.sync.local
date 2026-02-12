# env-sync V1.0 - AGE Encryption Implementation Plan

## Overview

Adding at-rest encryption using AGE to the distributed secrets sync system, while maintaining zero-configuration machine addition.

**Core Challenge**: How to add a 4th machine (D) without modifying encryption on machines A, B, C?

**Solution**: Hybrid encryption model where each machine has its own AGE key, and secrets are encrypted to multiple recipients. New machines discover existing public keys and request re-encryption.

## Security Model

### Threats Addressed
- ✅ Secrets not stored in plaintext on disk
- ✅ Secrets encrypted during sync (even with SCP/SSH)
- ✅ Compromised single machine doesn't expose all secrets
- ✅ Offline attacks on secrets file are prevented

### Threats NOT Addressed (out of scope)
- Active MITM attacks (mitigated by SSH in SCP mode)
- Memory dumps while secrets are in use
- Malicious machines in the sync group

## Key Architecture

### Individual Machine Keys
Each machine generates its own AGE key pair on first run:
- **Private Key**: `~/.config/env-sync/keys/age_key` (chmod 600)
- **Public Key**: Embedded in mDNS TXT records for auto-discovery

### Encryption Strategy

**Key Insight**: Instead of one shared key, each machine encrypts to ALL known recipients.

```
Machine A encrypts to: [A_pubkey, B_pubkey, C_pubkey]
Machine B encrypts to: [A_pubkey, B_pubkey, C_pubkey]
Machine C encrypts to: [A_pubkey, B_pubkey, C_pubkey]
```

When Machine D joins:
1. D generates its key pair
2. D broadcasts its public key via mDNS
3. A, B, C discover D's pubkey on next sync
4. Next sync from A/B/C re-encrypts to include D
5. D can now decrypt all secrets

### File Format (Encrypted)

```bash
# === ENV_SYNC_METADATA ===
# VERSION: 1.2.3
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# MODIFIED: 2025-02-07T15:30:45Z
# ENCRYPTED: true
# RECIPIENTS: age1xyz...,age1abc...,age1def...
# === END_METADATA ===

# Encrypted secrets block (BASE64 encoded AGE encrypted data)
-----BEGIN AGE ENCRYPTED FILE-----
YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSB...+500 chars...
-----END AGE ENCRYPTED FILE-----

# === ENV_SYNC_FOOTER ===
# VERSION: 1.2.3
# TIMESTAMP: 2025-02-07T15:30:45Z
# HOST: beelink.local
# === END_FOOTER ===
```

**Design Decision**: Metadata remains plaintext (version, timestamp, host, recipients) while only secret VALUES are encrypted.

## Public Key Distribution

### Method 1: mDNS TXT Records (Primary)

Each machine advertises its public key in mDNS service:

```
Service: _envsync._tcp
Port: 5739
TXT Records:
  - version=1.2.3
  - timestamp=2025-02-07T15:30:45Z
  - hostname=beelink.local
  - age_pubkey=age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
```

**Advantages**:
- Zero config - automatic discovery
- No central server needed
- Works with existing mDNS infrastructure

**Disadvantages**:
- TXT record size limits (~255 chars per record, 4KB total)
- AGE pubkeys are ~100 chars, well within limits

### Method 2: Direct Key Exchange (Fallback)

If mDNS unavailable or pubkey too large:

```bash
# Machine A requests pubkey from Machine B
env-sync key request B.local
# B responds with pubkey (encrypted over SSH)

# Or manual
env-sync key import --from hostname.local
```

### Public Key Cache

Each machine maintains a cache of known pubkeys:

```
~/.config/env-sync/keys/
├── age_key                  # This machine's private key
├── age_key.pub              # This machine's public key
├── known_hosts/
│   ├── beelink.local.pub    # beelink's pubkey
│   ├── mbp16.local.pub      # mbp16's pubkey
│   └── razer.local.pub      # razer's pubkey
└── cache/
    └── pubkey_cache.json    # Last seen timestamps
```

**Cache Format** (pubkey_cache.json):
```json
{
  "beelink.local": {
    "pubkey": "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p",
    "last_seen": "2025-02-07T15:30:45Z",
    "first_seen": "2025-02-01T10:00:00Z"
  },
  "mbp16.local": {
    "pubkey": "age1abc...",
    "last_seen": "2025-02-07T14:00:00Z",
    "first_seen": "2025-02-01T10:00:00Z"
  }
}
```

## Adding New Machines (Zero-Config)

### Scenario: Add Machine D (surface.local)

**On surface.local ONLY:**

```bash
# 1. Install env-sync
curl -fsSL .../install.sh | bash

# 2. Initialize (generates AGE key pair automatically)
env-sync init --encrypted

# 3. Discover peers (automatically collects pubkeys)
env-sync discover --collect-keys

# 4. Sync (will fail initially - see below)
env-sync
```

**What Happens:**

1. **Initialization**: D generates its key pair
   - Pubkey broadcast via mDNS
   - Other machines (A, B, C) discover D's pubkey → cache it

2. **First Sync Attempt**: D tries to sync
   - D downloads encrypted file from A/B/C
   - File encrypted to [A, B, C] recipients
   - **D CANNOT DECRYPT** (not in recipient list)

3. **Re-encryption Request**: D triggers re-encryption on existing machines
   
   **Option A: Remote Trigger (Preferred for Trusted Networks)**
   ```bash
   # On D: SSH into an existing machine and trigger sync
   # This adds D's pubkey to recipients and re-encrypts immediately
   env-sync key request-access --trigger beelink.local
   
   # Or trigger on all online machines
   env-sync key request-access --trigger-all
   ```
   
   **How it works:**
   - D SSHes into beelink.local (must have SSH access)
   - Runs `env-sync key add --pubkey D_pubkey` on beelink
   - Triggers `env-sync` on beelink to re-encrypt and sync
   - File is now encrypted to [A, B, C, D]
   - D can immediately sync and decrypt
   
   **Requirements:**
   - D must have SSH access to at least one existing machine
   - This is already required for SCP-based sync
   - No manual approval needed - D can trigger directly
   
   **Option B: Automatic Request (For Unattended Approval)**
   ```bash
   # D requests existing machines to re-encrypt
   env-sync key request-access --from A.local --from B.local --from C.local
   
   # Or simpler - request from all discovered peers
   env-sync key request-access --all
   ```
   
   On A/B/C, a notification appears:
   ```
   New machine surface.local (age1xyz...) requests access to secrets.
   Grant access? [Y/n]: 
   ```
   
   If Y: A/B/C re-encrypt secrets to include D and sync
   If n: Access denied, D cannot decrypt

   **Option C: Manual**
   ```bash
   # On D, show pubkey
   env-sync key show
   # Copy pubkey
   
   # On any existing machine (A, B, or C)
   env-sync key add --pubkey "age1xyz..."
   env-sync  # Re-encrypt and sync
   
   # Now D can sync and decrypt
   ```

4. **Immediate Sync**: D downloads newly encrypted file
   - Now encrypted to [A, B, C, D]
   - D can decrypt with its private key
   - Success! No waiting period needed.

**No changes needed on A, B, C:**
- **With Remote Trigger**: D SSHes in and triggers sync automatically
  - No manual action needed on A/B/C
  - D controls the process entirely
  - Works because D must have SSH access anyway for sync to work
  
- **Without Remote Trigger**: They auto-discover and cache D's pubkey
  - Manual approval or waiting for next sync cycle
  - One-time security measure

**Why Remote Trigger Works:**
- D already needs SSH access to sync via SCP (our default mode)
- If D can SSH into A, it can run commands on A
- D triggers: "add my pubkey and re-encrypt now"
- Immediate result - no waiting for 30min cron or manual sync
- A, B, C need ZERO changes - they just respond to SSH commands

## Transparent Decryption

### During Sync
```bash
env-sync  # Automatic
```

Flow:
1. Discover peers and fetch latest secrets file
2. Detect `ENCRYPTED: true` in metadata
3. Check if local machine is in `RECIPIENTS` list
4. Decrypt using `age -d -i ~/.config/env-sync/keys/age_key`
5. Validate decrypted content
6. Compare versions, merge if needed
7. Re-encrypt for all known recipients
8. Save encrypted file locally

### During Shell Load (.profile)

```bash
# In .bashrc/.zshrc
if command -v env-sync &> /dev/null; then
    # Load secrets (decrypts automatically if encrypted)
    eval "$(env-sync load)"
fi
```

New `env-sync load` command:
- Reads `~/.secrets.env`
- If encrypted, decrypts to stdout
- Exports as environment variables
- Never writes plaintext to disk

**Example Output** (suitable for eval):
```bash
export OPENAI_API_KEY="sk-xxx"
export AWS_ACCESS_KEY_ID="AKIA..."
```

### Cron Job

```bash
# Crontab entry
*/30 * * * * env-sync sync --quiet && eval "$(env-sync load --quiet)"
```

Or:
```bash
# New auto-load flag
*/30 * * * * env-sync sync --quiet --auto-load
```

## Component Updates

### 1. `env-sync-init`

New options:
```bash
env-sync init                    # Plaintext (backward compat)
env-sync init --encrypted        # Generate AGE key, encrypted by default
env-sync init --encrypt-existing # Encrypt existing plaintext secrets
```

Actions:
- Check if AGE installed
- Generate key pair: `age-keygen -o ~/.config/env-sync/keys/age_key`
- Extract pubkey: `age-keygen -y ~/.config/env-sync/keys/age_key`
- If encrypting existing: encrypt current secrets to [local_pubkey]

### 2. `env-sync-discover`

New options:
```bash
env-sync discover                # Standard discovery
env-sync discover --pubkeys      # Show discovered pubkeys
env-sync discover --collect-keys # Update pubkey cache
```

Enhancements:
- Parse `age_pubkey` from mDNS TXT records
- Update `~/.config/env-sync/keys/cache/pubkey_cache.json`
- Show which machines have encryption enabled

### 3. `env-sync-client`

Encryption handling:
```bash
# Detect if file is encrypted
if grep -q "^# ENCRYPTED: true" "$file"; then
    # Decrypt to temp file
    age -d -i "$AGE_KEY_FILE" -o "$temp_plaintext" "$file"
    # Work with plaintext version
    # ... sync logic ...
    # Re-encrypt for all recipients before saving
fi
```

New encryption logic during sync:
1. Load all known recipient pubkeys from cache
2. Decrypt fetched file (if encrypted)
3. Merge/update secrets
4. Encrypt to all recipients: `age -r pubkey1 -r pubkey2 ...`
5. Save encrypted result

### 4. `env-sync-key` (NEW COMMAND)

Key management CLI:

```bash
# Show this machine's key
env-sync key show               # Show pubkey
env-sync key show --private     # ⚠️ Show private key (careful!)

# Import/Export
env-sync key export             # Export pubkey to stdout
env-sync key export --qr        # Export as QR code
env-sync key import <pubkey>    # Import someone's pubkey
env-sync key import --from hostname.local  # Import from peer

# List known keys
env-sync key list               # List cached pubkeys
env-sync key list --local       # List only local machine's key

# Request access (new machine joining)
# Method 1: Remote trigger (SSH into existing machine and trigger sync)
env-sync key request-access --trigger hostname.local    # Trigger specific host
env-sync key request-access --trigger-all               # Trigger all online hosts

# Method 2: Send request and wait for approval
env-sync key request-access --from hostname.local
env-sync key request-access --all  # Request from all discovered

# Grant access (existing machine approving)
env-sync key grant-access --to hostname.local --pubkey "age1..."
env-sync key approve-requests   # Interactive approval of pending requests

# Remote trigger implementation:
# 1. SSH into target machine
# 2. Add D's pubkey to target's recipient list
# 3. Trigger sync on target (which re-encrypts to include D)
# 4. D can now sync the encrypted file

# Remove/Revoke
env-sync key remove hostname.local  # Remove from recipients
env-sync key revoke hostname.local  # Remove and re-encrypt without them
```

### 5. `env-sync-serve` (HTTP Mode)

New endpoint:
```
GET /pubkey
  Response: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p

POST /request-access
  Body: {"hostname": "surface.local", "pubkey": "age1..."}
  Response: {"status": "pending", "message": "Request sent to admin"}
```

### 6. `env-sync-load` (NEW COMMAND)

Load secrets for shell integration:

```bash
env-sync load                    # Decrypt and output export statements
env-sync load --format json      # Output as JSON
env-sync load --format env       # Output as .env format (default)
env-sync load --key KEY_NAME     # Load only specific key
```

Example integration:
```bash
# In .bashrc
eval "$(env-sync load 2>/dev/null)"
```

## Backward Compatibility

### Unencrypted → Encrypted Migration

**Phase 1: Opt-in Encryption**
```bash
# Existing plaintext setup continues working
env-sync  # Works with plaintext files

# User opts into encryption
env-sync init --encrypt-existing
```

**Phase 2: Mixed Mode Support**
- Detect if peer has encryption enabled (mDNS TXT record flag)
- If syncing encrypted → unencrypted peer: decrypt before sending
- If syncing unencrypted → encrypted peer: show warning, allow with flag

**Phase 3: Full Encryption (Optional)**
- Config option: `ENFORCE_ENCRYPTION=true`
- Refuse to sync with unencrypted peers
- Require all machines to use encryption

### Version Compatibility

Encrypted files use metadata flag:
```bash
# VERSION: 2.0.0  # Major version bump for encryption
# ENCRYPTED: true
```

Older clients see:
- Unknown metadata field (ignored)
- Binary content after header (treated as corrupt)
- Error: "Please upgrade env-sync to support encrypted files"

## Implementation Phases

### Phase 1: Core Encryption (Week 1-2)
- [ ] Add AGE dependency check to install.sh
- [ ] Generate AGE keys on init
- [ ] Encrypt/decrypt functions in common.sh
- [ ] Update secrets file format
- [ ] `env-sync key` command
- [ ] `env-sync load` command

### Phase 2: Public Key Distribution (Week 2-3)
- [ ] mDNS TXT record with pubkey
- [ ] Pubkey cache system
- [ ] `env-sync discover --collect-keys`
- [ ] Key listing and management

### Phase 3: Sync Integration (Week 3-4)
- [ ] Update client to decrypt before comparing
- [ ] Re-encrypt to all recipients after merge
- [ ] Handle unencrypted peers gracefully
- [ ] Version compatibility checks

### Phase 4: New Machine Onboarding (Week 4)
- [ ] `env-sync key request-access` command
  - [ ] `--trigger hostname` to SSH in and trigger sync remotely
  - [ ] `--trigger-all` to trigger on all online hosts
  - [ ] `--from` and `--all` for manual approval workflow
- [ ] `env-sync key grant-access` command
- [ ] Interactive approval workflow
- [ ] Auto-discovery and caching
- [ ] Remote trigger implementation (SSH command execution)

### Phase 5: Shell Integration (Week 5)
- [ ] Transparent decryption in .profile
- [ ] Cron job auto-load
- [ ] Performance optimization (cache decrypted values?)
- [ ] Documentation and testing

## Dependencies

### Required
- `age` - File encryption tool
  - macOS: `brew install age`
  - Linux: `apt install age` or download binary
  - Also available as static binary

### Optional (for enhanced security)
- `age-plugin-yubikey` - Hardware key support
- `age-plugin-fido` - FIDO2 hardware key support

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
├── logs/                           # Sync logs
└── requests/                       # Pending access requests
    └── surface.local.request
```

## Security Best Practices

### Key Management
1. **Private Key Protection**
   - Never transmit private key over network
   - Never log private key
   - File permissions: 600
   - Consider hardware key (YubiKey) for production

2. **Public Key Verification**
   - First-time pubkey acceptance (like SSH known_hosts)
   - Visual verification (fingerprint display)
   - Out-of-band verification for high-security environments

3. **Rotation Strategy**
   - Keys don't expire (AGE design)
   - To rotate: generate new key, add to recipients, remove old
   - `env-sync key rotate` command

### Recovery Scenarios

**Lost Private Key:**
- Cannot decrypt secrets (by design)
- Must get re-encrypted file from another machine
- Or restore from backup with old key
- Generate new key pair and request access again

**All Machines Lost:**
- Secrets are lost (by design)
- Maintain offline backup of one machine's key
- Or use paper backup of key

**Rogue Machine:**
- If malicious machine joins:
  - It can decrypt future secrets
  - Cannot decrypt old versions (backups)
- Revoke: `env-sync key revoke rogue.local`
- Re-encrypt without rogue's pubkey

## Testing Strategy

### Unit Tests
- [ ] Encryption/decryption round-trip
- [ ] Recipient list parsing
- [ ] Key generation
- [ ] Cache management

### Integration Tests
- [ ] Two machines: encrypted sync
- [ ] Three machines: add third, verify all can decrypt
- [ ] Remove machine: verify excluded from new encryption
- [ ] Lost key: restore and re-request access

### Manual Testing
- [ ] Add 4th machine to existing 3-machine setup
- [ ] Verify no changes needed on A, B, C
- [ ] Shell integration (.profile load)
- [ ] Cron job with auto-load
- [ ] Mixed encrypted/unencrypted sync

## Migration Guide

### Existing Users (v0.x → v1.0)

```bash
# 1. Upgrade env-sync
curl -fsSL .../install.sh | bash

# 2. Check age is installed
which age || echo "Install age first"

# 3. Encrypt existing secrets
env-sync init --encrypt-existing

# 4. Share pubkey with peers
env-sync key export
# Send pubkey to other machines

# 5. Import peer pubkeys
env-sync key import --from beelink.local
env-sync key import --from mbp16.local

# 6. Sync (re-encrypts to all recipients)
env-sync

# 7. Update .bashrc to use load command
echo 'eval "$(env-sync load 2>/dev/null)"' >> ~/.bashrc
```

### New Users

```bash
# Same as before, just with encryption
env-sync init  # Now generates keys automatically
env-sync       # Sync with encryption
```

## Success Criteria

- [ ] Secrets encrypted at rest (not plaintext on disk)
- [ ] Transparent decryption during sync/shell load
- [ ] Adding new machine requires zero changes to existing machines
- [ ] Only authorized machines can decrypt
- [ ] Compromised single machine doesn't expose secrets
- [ ] Backward compatible with unencrypted mode
- [ ] Performance acceptable (<1s overhead per sync)
- [ ] Works on Linux, macOS, Windows (WSL2)

## Open Questions

1. **Approval UX**: Should access requests be interactive or automatic?
   - Interactive: More secure, requires user presence
   - Automatic: More convenient, less secure
   - **Decision**: Interactive with `--auto-approve` flag for trusted networks

2. **Key Backup**: Should we support passphrase-protected key backup?
   - Yes: Add `env-sync key backup --passphrase "..."`
   - Store encrypted key backup in cloud/drive

3. **Multiple Keys**: Should machines support multiple AGE keys?
   - Probably not - adds complexity
   - Rotation handles key changes

4. **Hardware Keys**: Priority for YubiKey support?
   - Nice-to-have for v1.1
   - Not required for v1.0 MVP
