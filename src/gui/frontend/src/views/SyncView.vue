<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usePeersStore } from '@/stores/peers'
import { useToast } from '@/composables/useToast'

const peers = usePeersStore()
const toast = useToast()

const syncing = ref(false)
const syncResults = ref<{ peer: string; success: boolean; message: string }[]>([])
const showForceSync = ref(false)
const forcePullPeer = ref('')

async function syncAll() {
  syncing.value = true
  syncResults.value = []
  try {
    // @ts-expect-error Wails bindings
    const result = await window.go.main.SyncService.SyncAll()
    syncResults.value = [{ peer: 'all', success: result.success, message: result.message }]
    if (result.success) {
      toast.success('Sync completed')
    } else {
      toast.error('Sync failed: ' + result.message)
    }
  } catch (e) {
    toast.error('Sync error: ' + e)
  } finally {
    syncing.value = false
  }
}

async function syncFrom(peer: string) {
  syncing.value = true
  try {
    // @ts-expect-error Wails bindings
    const result = await window.go.main.SyncService.SyncFrom(peer)
    if (result.success) {
      toast.success(`Synced from ${peer}`)
    } else {
      toast.error(`Sync from ${peer} failed: ${result.message}`)
    }
  } catch (e) {
    toast.error('Sync error: ' + e)
  } finally {
    syncing.value = false
  }
}

async function forcePull(peer: string) {
  syncing.value = true
  try {
    // @ts-expect-error Wails bindings
    const result = await window.go.main.SyncService.ForcePull(peer)
    if (result.success) {
      toast.success(`Force pulled from ${peer}`)
    } else {
      toast.error(`Force pull failed: ${result.message}`)
    }
  } catch (e) {
    toast.error('Force pull error: ' + e)
  } finally {
    syncing.value = false
    showForceSync.value = false
  }
}

async function discoverPeers() {
  await peers.discover(5)
  toast.info(`Found ${peers.discovered.length} peers`)
}

onMounted(async () => {
  await peers.discover(5)
})
</script>

<template>
  <div class="sync-view">
    <div class="section-header">
      <div>
        <h1 class="section-title">Sync</h1>
        <p class="section-subtitle">Sync secrets across peers</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-primary" @click="syncAll" :disabled="syncing">
          <span v-if="syncing" class="spinner"></span>
          <span v-else>🔄</span>
          Sync All
        </button>
        <button class="btn btn-secondary" @click="discoverPeers" :disabled="peers.discovering">
          📡 Discover
        </button>
      </div>
    </div>

    <!-- Discovered Peers -->
    <div class="card" style="margin-top: 16px">
      <div class="card-header">
        <span class="card-title">Discovered Peers</span>
        <span class="badge badge-info" v-if="peers.discovering">Scanning...</span>
      </div>

      <div v-if="peers.discovered.length > 0">
        <div
          v-for="peer in peers.discovered"
          :key="peer.hostname"
          class="peer-row"
        >
          <div class="peer-info">
            <span class="peer-name mono">{{ peer.hostname }}</span>
            <span class="peer-addr text-muted">{{ peer.address }}:{{ peer.port }}</span>
          </div>
          <div class="peer-actions">
            <span class="badge" :class="peer.reachable ? 'badge-success' : 'badge-error'">
              {{ peer.reachable ? '● Reachable' : '● Unreachable' }}
            </span>
            <button
              class="btn btn-secondary btn-sm"
              @click="syncFrom(peer.hostname)"
              :disabled="syncing || !peer.reachable"
            >
              Sync
            </button>
            <button
              class="btn btn-ghost btn-sm"
              @click="forcePullPeer = peer.hostname; showForceSync = true"
              :disabled="syncing || !peer.reachable"
            >
              Force Pull
            </button>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <p class="text-muted">{{ peers.discovering ? 'Scanning network...' : 'No peers found' }}</p>
      </div>
    </div>

    <!-- Sync Results -->
    <div class="card" style="margin-top: 16px" v-if="syncResults.length > 0">
      <div class="card-header">
        <span class="card-title">Last Sync Result</span>
      </div>
      <div v-for="r in syncResults" :key="r.peer" class="result-row">
        <span class="badge" :class="r.success ? 'badge-success' : 'badge-error'">
          {{ r.success ? '✓' : '✗' }}
        </span>
        <span>{{ r.message }}</span>
      </div>
    </div>

    <!-- Force Pull Confirmation -->
    <div class="modal-overlay" v-if="showForceSync" @click.self="showForceSync = false">
      <div class="modal">
        <h2 class="modal-title">⚠️ Force Pull</h2>
        <p>This will <strong>overwrite</strong> your local secrets with those from <strong class="mono">{{ forcePullPeer }}</strong>.</p>
        <p class="text-muted" style="margin-top: 8px">A backup will be created first.</p>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showForceSync = false">Cancel</button>
          <button class="btn btn-danger" @click="forcePull(forcePullPeer)">Force Pull</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sync-view {
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

.peer-addr {
  font-size: 12px;
}

.peer-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.result-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
}

.empty-state {
  padding: 40px;
  text-align: center;
}
</style>
