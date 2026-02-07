#!/bin/bash
# Entrypoint script for env-sync test containers
# Note: We don't use 'set -e' because we want to handle errors gracefully

USER_NAME="envsync"
USER_HOME="/home/envsync"

echo "[entrypoint] Starting env-sync test container: $HOSTNAME"

# Ensure SSH directory permissions (as root first, then switch)
chown -R ${USER_NAME}:${USER_NAME} ${USER_HOME}/.ssh 2>/dev/null || true
chmod 700 ${USER_HOME}/.ssh 2>/dev/null || true
chmod 600 ${USER_HOME}/.ssh/* 2>/dev/null || true
chmod 644 ${USER_HOME}/.ssh/*.pub 2>/dev/null || true
chmod 644 ${USER_HOME}/.ssh/authorized_keys 2>/dev/null || true

# Generate host SSH keys if not present
if [ ! -f /etc/ssh/ssh_host_ed25519_key ]; then
    echo "[entrypoint] Generating SSH host keys..."
    ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key -N "" 2>/dev/null || true
fi
if [ ! -f /etc/ssh/ssh_host_rsa_key ]; then
    ssh-keygen -t rsa -b 4096 -f /etc/ssh/ssh_host_rsa_key -N "" 2>/dev/null || true
fi

# Start SSH daemon
echo "[entrypoint] Starting SSH daemon..."
/usr/sbin/sshd

# Start D-Bus (required for avahi on some systems)
if [ -d /var/run/dbus ]; then
    rm -rf /var/run/dbus/* 2>/dev/null || true
fi
mkdir -p /var/run/dbus 2>/dev/null || true
dbus-daemon --system --fork 2>/dev/null || echo "[entrypoint] D-Bus note: non-fatal"

# Start Avahi for mDNS
echo "[entrypoint] Starting Avahi daemon..."
avahi-daemon --daemonize --no-drop-root 2>/dev/null || avahi-daemon --daemonize 2>/dev/null || echo "[entrypoint] Avahi warning: could not start"

# Wait for services to be ready
echo "[entrypoint] Waiting for services to start..."
sleep 3

# Test SSH is working
echo "[entrypoint] Testing SSH daemon..."
if pgrep -x sshd > /dev/null 2>&1; then
    echo "[entrypoint] SSH daemon is running"
else
    echo "[entrypoint] WARNING: SSH daemon may not be running"
fi

# Test mDNS is working
echo "[entrypoint] Testing Avahi daemon..."
if pgrep -x avahi-daemon > /dev/null 2>&1; then
    echo "[entrypoint] Avahi daemon is running"
else
    echo "[entrypoint] WARNING: Avahi daemon may not be running"
fi

# Populate known_hosts with other containers (as envsync user)
echo "[entrypoint] Populating SSH known_hosts..."
for host in alpha.local beta.local gamma.local; do
    if [ "$host" != "$HOSTNAME" ]; then
        su - ${USER_NAME} -c "ssh-keyscan -H $host >> ${USER_HOME}/.ssh/known_hosts 2>/dev/null" || echo "[entrypoint] Note: Could not scan $host yet"
    fi
done
chown ${USER_NAME}:${USER_NAME} ${USER_HOME}/.ssh/known_hosts 2>/dev/null || true
chmod 600 ${USER_HOME}/.ssh/known_hosts 2>/dev/null || true

echo "[entrypoint] Container is ready: $HOSTNAME"

# If no command provided or command is 'bash', keep container running as envsync user
if [ $# -eq 0 ] || [ "$1" = "bash" ] || [ "$1" = "/bin/bash" ]; then
    echo "[entrypoint] Keeping container running as ${USER_NAME}..."
    su - ${USER_NAME} -c "tail -f /dev/null"
else
    echo "[entrypoint] Executing command as ${USER_NAME}: $@"
    su - ${USER_NAME} -c "$@"
fi
