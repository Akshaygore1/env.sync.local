#!/bin/bash
# Common functions for env-sync
# This file is sourced by other scripts

set -euo pipefail

# Configuration
ENV_SYNC_VERSION="1.0.0"
ENV_SYNC_PORT="5739"
ENV_SYNC_SERVICE="_envsync._tcp"
SECRETS_FILE="${HOME}/.secrets.env"
CONFIG_DIR="${HOME}/.config/env-sync"
BACKUP_DIR="${CONFIG_DIR}/backups"
LOG_DIR="${CONFIG_DIR}/logs"
MAX_BACKUPS=5

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    mkdir -p "$LOG_DIR"
    echo "[$timestamp] [$level] $message" >> "$LOG_DIR/env-sync.log"
    
    if [[ "${ENV_SYNC_QUIET:-}" != "true" ]]; then
        case "$level" in
            ERROR) echo -e "${RED}ERROR:${NC} $message" >&2 ;;
            WARN)  echo -e "${YELLOW}WARN:${NC} $message" >&2 ;;
            INFO)  echo -e "${BLUE}INFO:${NC} $message" ;;
            SUCCESS) echo -e "${GREEN}SUCCESS:${NC} $message" ;;
        esac
    fi
}

# Get hostname
get_hostname() {
    hostname -f 2>/dev/null || hostname
}

# Generate timestamp
get_timestamp() {
    date -u '+%Y-%m-%dT%H:%M:%SZ'
}

# Generate checksum
generate_checksum() {
    local file="$1"
    if [[ -f "$file" ]]; then
        sha256sum "$file" | cut -d' ' -f1
    else
        echo ""
    fi
}

# Extract metadata from secrets file
extract_metadata() {
    local file="$1"
    local key="$2"
    
    if [[ ! -f "$file" ]]; then
        echo ""
        return
    fi
    
    grep "^# $key:" "$file" | head -1 | sed "s/^# $key: //"
}

# Get version from secrets file
get_file_version() {
    local file="$1"
    extract_metadata "$file" "VERSION"
}

# Get timestamp from secrets file
get_file_timestamp() {
    local file="$1"
    extract_metadata "$file" "TIMESTAMP"
}

# Get hostname from secrets file
get_file_host() {
    local file="$1"
    extract_metadata "$file" "HOST"
}

# Get checksum from secrets file
get_file_checksum() {
    local file="$1"
    extract_metadata "$file" "CHECKSUM"
}

# Compare two versions (semantic versioning)
# Returns: 0 if equal, 1 if v1 > v2, 2 if v1 < v2
compare_versions() {
    local v1="$1"
    local v2="$2"
    
    if [[ "$v1" == "$v2" ]]; then
        return 0
    fi
    
    # Use sort -V for version comparison
    local higher=$(printf '%s\n%s\n' "$v1" "$v2" | sort -V | tail -n1)
    
    if [[ "$v1" == "$higher" ]]; then
        return 1  # v1 > v2
    else
        return 2  # v1 < v2
    fi
}

# Compare two timestamps (ISO 8601 format)
# Returns: 0 if equal, 1 if t1 > t2, 2 if t1 < t2
compare_timestamps() {
    local t1="$1"
    local t2="$2"
    
    if [[ "$t1" == "$t2" ]]; then
        return 0
    fi
    
    # Convert to seconds since epoch and compare
    local s1=$(date -d "$t1" '+%s' 2>/dev/null || echo "0")
    local s2=$(date -d "$t2" '+%s' 2>/dev/null || echo "0")
    
    if [[ "$s1" -gt "$s2" ]]; then
        return 1  # t1 > t2
    else
        return 2  # t1 < t2
    fi
}

# Determine if file1 is newer than file2
# Returns 0 if file1 is newer, 1 otherwise
is_newer() {
    local file1="$1"
    local file2="$2"
    
    local v1=$(get_file_version "$file1")
    local v2=$(get_file_version "$file2")
    local t1=$(get_file_timestamp "$file1")
    local t2=$(get_file_timestamp "$file2")
    local h1=$(get_file_host "$file1")
    local h2=$(get_file_host "$file2")
    
    # Compare timestamps first
    if [[ -n "$t1" && -n "$t2" ]]; then
        compare_timestamps "$t1" "$t2"
        local ts_result=$?
        if [[ $ts_result -eq 1 ]]; then
            return 0  # file1 is newer
        elif [[ $ts_result -eq 2 ]]; then
            return 1  # file2 is newer
        fi
    fi
    
    # Timestamps equal, compare versions
    if [[ -n "$v1" && -n "$v2" ]]; then
        compare_versions "$v1" "$v2"
        local v_result=$?
        if [[ $v_result -eq 1 ]]; then
            return 0  # file1 is newer
        elif [[ $v_result -eq 2 ]]; then
            return 1  # file2 is newer
        fi
    fi
    
    # Both equal, use hostname as tiebreaker (deterministic)
    if [[ "$h1" < "$h2" ]]; then
        return 0
    else
        return 1
    fi
}

# Create backup of secrets file
create_backup() {
    local file="$1"
    
    if [[ ! -f "$file" ]]; then
        return
    fi
    
    mkdir -p "$BACKUP_DIR"
    
    # Rotate backups
    for i in $(seq $((MAX_BACKUPS - 1)) -1 1); do
        local j=$((i + 1))
        if [[ -f "$BACKUP_DIR/secrets.backup.$i" ]]; then
            mv "$BACKUP_DIR/secrets.backup.$i" "$BACKUP_DIR/secrets.backup.$j"
        fi
    done
    
    # Create new backup
    cp "$file" "$BACKUP_DIR/secrets.backup.1"
    chmod 600 "$BACKUP_DIR/secrets.backup.1"
    
    log "INFO" "Created backup: secrets.backup.1"
}

# Restore from backup
restore_backup() {
    local backup_num="${1:-1}"
    local backup_file="$BACKUP_DIR/secrets.backup.$backup_num"
    
    if [[ ! -f "$backup_file" ]]; then
        log "ERROR" "Backup $backup_num not found"
        return 1
    fi
    
    create_backup "$SECRETS_FILE"
    cp "$backup_file" "$SECRETS_FILE"
    chmod 600 "$SECRETS_FILE"
    log "SUCCESS" "Restored from backup $backup_num"
}

# Initialize new secrets file
init_secrets_file() {
    local file="$1"
    local hostname=$(get_hostname)
    local timestamp=$(get_timestamp)
    local version="1.0.0"
    
    mkdir -p "$(dirname "$file")"
    
    cat > "$file" << EOF
# === ENV_SYNC_METADATA ===
# VERSION: $version
# TIMESTAMP: $timestamp
# HOST: $hostname
# MODIFIED: $timestamp
# CHECKSUM: 
# === END_METADATA ===

# Add your secrets below this line
# Example:
# OPENAI_API_KEY="sk-..."

# === ENV_SYNC_FOOTER ===
# VERSION: $version
# TIMESTAMP: $timestamp
# HOST: $hostname
# === END_FOOTER ===
EOF
    
    chmod 600 "$file"
    
    # Update checksum
    local checksum=$(generate_checksum "$file")
    sed -i.bak "s/^# CHECKSUM: /# CHECKSUM: $checksum/" "$file"
    rm -f "$file.bak"
    
    log "SUCCESS" "Initialized secrets file: $file"
}

# Update secrets file metadata
update_metadata() {
    local file="$1"
    local new_version="${2:-}"
    
    if [[ ! -f "$file" ]]; then
        log "ERROR" "File not found: $file"
        return 1
    fi
    
    local hostname=$(get_hostname)
    local timestamp=$(get_timestamp)
    local current_version=$(get_file_version "$file")
    local version="${new_version:-$current_version}"
    
    # Update header
    sed -i.bak \
        -e "s/^# TIMESTAMP: .*/# TIMESTAMP: $timestamp/" \
        -e "s/^# HOST: .*/# HOST: $hostname/" \
        -e "s/^# MODIFIED: .*/# MODIFIED: $timestamp/" \
        "$file"
    rm -f "$file.bak"
    
    # Update footer
    sed -i.bak \
        -e "s/^# VERSION: .*/# VERSION: $version/" \
        -e "s/^# TIMESTAMP: .*/# TIMESTAMP: $timestamp/" \
        -e "s/^# HOST: .*/# HOST: $hostname/" \
        "$file"
    rm -f "$file.bak"
    
    # Update checksum
    local checksum=$(generate_checksum "$file")
    sed -i.bak "s/^# CHECKSUM: .*/# CHECKSUM: $checksum/" "$file"
    rm -f "$file.bak"
}

# Validate secrets file
validate_secrets_file() {
    local file="$1"
    
    if [[ ! -f "$file" ]]; then
        log "ERROR" "Secrets file not found: $file"
        return 1
    fi
    
    # Check metadata exists
    if ! grep -q "^# === ENV_SYNC_METADATA ===" "$file"; then
        log "ERROR" "Invalid secrets file: missing metadata header"
        return 1
    fi
    
    if ! grep -q "^# === ENV_SYNC_FOOTER ===" "$file"; then
        log "ERROR" "Invalid secrets file: missing metadata footer"
        return 1
    fi
    
    # Validate checksum
    local stored_checksum=$(get_file_checksum "$file")
    if [[ -n "$stored_checksum" ]]; then
        local current_checksum=$(generate_checksum "$file")
        if [[ "$stored_checksum" != "$current_checksum" ]]; then
            log "WARN" "Checksum mismatch - file may be corrupted"
            return 1
        fi
    fi
    
    return 0
}

# Get secrets content (without metadata)
get_secrets_content() {
    local file="$1"
    
    if [[ ! -f "$file" ]]; then
        return
    fi
    
    # Extract content between metadata header and footer
    awk '/^# === END_METADATA ===/{found=1; next} /^# === ENV_SYNC_FOOTER ===/{found=0} found' "$file"
}

# Set secrets content (update file while preserving metadata)
set_secrets_content() {
    local file="$1"
    local content="$2"
    
    if [[ ! -f "$file" ]]; then
        init_secrets_file "$file"
    fi
    
    # Create temporary file
    local tmp_file=$(mktemp)
    
    # Write header
    awk '/^# === ENV_SYNC_METADATA ===/,/^# === END_METADATA ===/' "$file" > "$tmp_file"
    
    # Write content
    echo "" >> "$tmp_file"
    echo "$content" >> "$tmp_file"
    echo "" >> "$tmp_file"
    
    # Write footer
    awk '/^# === ENV_SYNC_FOOTER ===/,/^# === END_FOOTER ===/' "$file" >> "$tmp_file"
    
    # Move temp file to target
    mv "$tmp_file" "$file"
    chmod 600 "$file"
    
    # Update metadata
    update_metadata "$file"
}
