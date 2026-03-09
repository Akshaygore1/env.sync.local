<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { usePeersStore } from '@/stores/peers'
import { useToast } from '@/composables/useToast'

const peers = usePeersStore()
const toast = useToast()
const isSecureMode = ref(false)
const showInvite = ref(false)
const inviteToken = ref('')
const confirmRevoke = ref<string | null>(null)

onMounted(async () => {
  isSecureMode.value = await peers.isSecurePeerMode()
  await peers.fetchPeers()
  if (isSecureMode.value) {
    await peers.fetchPending()
  }
})

async function createInvite() {
  try {
    const invite = await peers.createInvite(24)
    inviteToken.value = invite.token
    showInvite.value = true
  } catch (e) {
    toast.error('Failed to create invite: ' + e)
  }
}

async function copyInvite() {
  await navigator.clipboard.writeText(inviteToken.value)
  toast.success('Invite token copied to clipboard')
}

async function approve(peerID: string) {
  try {
    await peers.approvePeer(peerID)
    toast.success(`Approved ${peerID}`)
  } catch (e) {
    toast.error('Approval failed: ' + e)
  }
}

async function revoke(peerID: string) {
  try {
    await peers.revokePeer(peerID)
    toast.success(`Revoked ${peerID}`)
    confirmRevoke.value = null
  } catch (e) {
    toast.error('Revoke failed: ' + e)
  }
}

async function discoverSSH() {
  await peers.discoverSSH(5)
  toast.info(`Found ${peers.discovered.length} SSH-reachable peers`)
}
</script>

<template>
  <div class="peers-view">
    <div class="section-header">
      <div>
        <h1 class="section-title">Peers</h1>
        <p class="section-subtitle">Manage connected machines</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-primary" @click="createInvite" v-if="isSecureMode">
          ✉️ Create Invite
        </button>
        <button class="btn btn-secondary" @click="discoverSSH">
          📡 Discover SSH
        </button>
        <button class="btn btn-secondary" @click="peers.fetchPeers()">
          ↻ Refresh
        </button>
      </div>
    </div>

    <!-- Pending Approvals (secure-peer mode) -->
    <div class="card" style="margin-top: 16px" v-if="isSecureMode && peers.pending.length > 0">
      <div class="card-header">
        <span class="card-title">⏳ Pending Approvals</span>
        <span class="badge badge-warning">{{ peers.pending.length }}</span>
      </div>
      <div v-for="p in peers.pending" :key="p.hostname" class="peer-row">
        <div class="peer-info">
          <span class="peer-name mono">{{ p.hostname }}</span>
          <span class="text-muted" style="font-size: 12px">Requested: {{ new Date(p.addedAt).toLocaleString() }}</span>
        </div>
        <div class="peer-actions">
          <button class="btn btn-primary btn-sm" @click="approve(p.hostname)">✓ Approve</button>
          <button class="btn btn-danger btn-sm" @click="confirmRevoke = p.hostname">✗ Reject</button>
        </div>
      </div>
    </div>

    <!-- Registered Peers -->
    <div class="card" style="margin-top: 16px">
      <div class="card-header">
        <span class="card-title">Registered Peers</span>
        <span class="badge badge-info">{{ peers.registered.length }}</span>
      </div>
      <div v-if="peers.registered.length > 0">
        <div v-for="p in peers.registered" :key="p.hostname" class="peer-row">
          <div class="peer-info">
            <span class="peer-name mono">{{ p.hostname }}</span>
            <div class="peer-details">
              <span class="text-muted" style="font-size: 12px">
                Status: <span class="badge" :class="p.state === 'approved' ? 'badge-success' : 'badge-error'">{{ p.state }}</span>
              </span>
              <span class="text-muted" style="font-size: 12px" v-if="p.agePubKey">
                Key: <span class="mono">{{ p.agePubKey.substring(0, 16) }}...</span>
              </span>
            </div>
          </div>
          <div class="peer-actions" v-if="isSecureMode">
            <button class="btn btn-ghost btn-sm" @click="confirmRevoke = p.hostname" v-if="p.state === 'approved'">
              Revoke
            </button>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <p class="text-muted">No registered peers</p>
      </div>
    </div>

    <!-- Discovered Peers -->
    <div class="card" style="margin-top: 16px" v-if="peers.discovered.length > 0">
      <div class="card-header">
        <span class="card-title">Discovered on Network</span>
      </div>
      <div v-for="d in peers.discovered" :key="d.hostname" class="peer-row">
        <div class="peer-info">
          <span class="peer-name mono">{{ d.hostname }}</span>
          <span class="text-muted" style="font-size: 12px">{{ d.address }}:{{ d.port }}</span>
        </div>
        <span class="badge" :class="d.reachable ? 'badge-success' : 'badge-error'">
          {{ d.reachable ? 'Reachable' : 'Unreachable' }}
        </span>
      </div>
    </div>

    <!-- Invite Modal -->
    <div class="modal-overlay" v-if="showInvite" @click.self="showInvite = false">
      <div class="modal">
        <h2 class="modal-title">Invitation Token</h2>
        <p class="text-muted" style="margin-bottom: 12px">Share this token with the new peer:</p>
        <div class="token-display">
          <code class="mono">{{ inviteToken }}</code>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showInvite = false">Close</button>
          <button class="btn btn-primary" @click="copyInvite">📋 Copy</button>
        </div>
      </div>
    </div>

    <!-- Revoke Confirmation -->
    <div class="modal-overlay" v-if="confirmRevoke" @click.self="confirmRevoke = null">
      <div class="modal">
        <h2 class="modal-title">Revoke Peer Access</h2>
        <p>Revoke access for <strong class="mono">{{ confirmRevoke }}</strong>?</p>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="confirmRevoke = null">Cancel</button>
          <button class="btn btn-danger" @click="revoke(confirmRevoke!)">Revoke</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.peers-view {
  max-width: 1000px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.peer-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.peer-row:last-child {
  border-bottom: none;
}

.peer-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.peer-name {
  font-weight: 600;
}

.peer-details {
  display: flex;
  gap: 12px;
}

.peer-actions {
  display: flex;
  gap: 8px;
}

.empty-state {
  padding: 40px;
  text-align: center;
}

.token-display {
  background: var(--bg-primary);
  padding: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  word-break: break-all;
  margin-bottom: 16px;
}
</style>
