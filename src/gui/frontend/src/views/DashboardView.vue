<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useStatusStore } from '@/stores/status'
import { useSecretsStore } from '@/stores/secrets'
import { useToast } from '@/composables/useToast'

const status = useStatusStore()
const secrets = useSecretsStore()
const toast = useToast()
const syncing = ref(false)

async function syncAll() {
  syncing.value = true
  try {
    // @ts-expect-error Wails bindings
    const result = await window.go.main.SyncService.SyncAll()
    if (result.success) {
      toast.success(result.message)
    } else {
      toast.error(result.message)
    }
    await refresh()
  } catch (e) {
    toast.error('Sync failed: ' + e)
  } finally {
    syncing.value = false
  }
}

async function refresh() {
  await status.fetchStatus()
  await secrets.fetchAll()
}

onMounted(async () => {
  await refresh()

  // Poll for changes every 5 seconds
  setInterval(async () => {
    const changed = await status.checkFileChanged()
    if (changed) {
      await refresh()
    }
  }, 5000)
})
</script>

<template>
  <div class="dashboard">
    <div class="section-header">
      <div>
        <h1 class="section-title">Dashboard</h1>
        <p class="section-subtitle">env-sync status overview</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-primary" @click="syncAll" :disabled="syncing">
          <span v-if="syncing" class="spinner"></span>
          <span v-else>🔄</span>
          Sync All
        </button>
        <button class="btn btn-secondary" @click="refresh">
          ↻ Refresh
        </button>
      </div>
    </div>

    <div class="status-grid" v-if="status.status">
      <!-- Secrets Card -->
      <div class="card status-card">
        <div class="status-icon">🔑</div>
        <div class="status-info">
          <div class="status-label">Secrets</div>
          <div class="status-value">{{ secrets.count }} keys</div>
          <div class="status-detail">
            <span class="badge" :class="status.fileStatus?.encrypted ? 'badge-success' : 'badge-warning'">
              {{ status.fileStatus?.encrypted ? '🔒 Encrypted' : '🔓 Plaintext' }}
            </span>
          </div>
        </div>
      </div>

      <!-- Server Card -->
      <div class="card status-card">
        <div class="status-icon">🖥️</div>
        <div class="status-info">
          <div class="status-label">Server</div>
          <div class="status-value">
            <span class="badge" :class="status.serverStatus?.running ? 'badge-success' : 'badge-error'">
              {{ status.serverStatus?.running ? '● Online' : '● Offline' }}
            </span>
          </div>
          <div class="status-detail text-muted" v-if="status.serverStatus?.running">
            Port {{ status.serverStatus.port }}
          </div>
        </div>
      </div>

      <!-- Mode Card -->
      <div class="card status-card">
        <div class="status-icon">🛡️</div>
        <div class="status-info">
          <div class="status-label">Mode</div>
          <div class="status-value">
            <span class="badge badge-info">{{ status.modeInfo?.current }}</span>
          </div>
          <div class="status-detail text-muted">{{ status.modeInfo?.description }}</div>
        </div>
      </div>

      <!-- Backups Card -->
      <div class="card status-card">
        <div class="status-icon">💾</div>
        <div class="status-info">
          <div class="status-label">Backups</div>
          <div class="status-value">{{ status.backups.length }} available</div>
          <div class="status-detail text-muted" v-if="status.backups.length > 0">
            Latest: {{ new Date(status.backups[status.backups.length - 1]?.timestamp).toLocaleString() }}
          </div>
        </div>
      </div>
    </div>

    <div class="card" v-if="!status.fileStatus?.exists" style="margin-top: 24px">
      <div class="card-header">
        <span class="card-title">⚠️ Not Initialized</span>
      </div>
      <p class="text-secondary">
        No secrets file found. Go to <strong>Settings</strong> to initialize env-sync.
      </p>
    </div>

    <div class="card" style="margin-top: 16px" v-if="status.fileStatus?.exists">
      <div class="card-header">
        <span class="card-title">File Info</span>
      </div>
      <div class="file-info">
        <div class="info-row">
          <span class="text-muted">Path</span>
          <span class="mono">{{ status.fileStatus?.path }}</span>
        </div>
        <div class="info-row">
          <span class="text-muted">Version</span>
          <span class="mono">{{ status.fileStatus?.version }}</span>
        </div>
        <div class="info-row">
          <span class="text-muted">Last Modified</span>
          <span>{{ status.fileStatus?.modTime ? new Date(status.fileStatus.modTime).toLocaleString() : '-' }}</span>
        </div>
        <div class="info-row">
          <span class="text-muted">Host</span>
          <span class="mono">{{ status.fileStatus?.host }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: 1000px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.status-card {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.status-icon {
  font-size: 28px;
  line-height: 1;
}

.status-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.status-value {
  font-size: 16px;
  font-weight: 600;
}

.status-detail {
  font-size: 13px;
}

.file-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}

.info-row span:first-child {
  font-size: 13px;
  min-width: 120px;
}
</style>
