<script setup lang="ts">
import { onMounted, ref, onUnmounted } from 'vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const logs = ref('')
const loading = ref(false)
const autoRefresh = ref(false)
const lines = ref(200)
let refreshTimer: ReturnType<typeof setInterval> | null = null

async function fetchLogs() {
  loading.value = true
  try {
    // @ts-expect-error Wails bindings
    const result = await window.go.main.LogService.GetRecentLogs(lines.value)
    logs.value = result
  } catch (e) {
    console.error('Failed to fetch logs:', e)
  } finally {
    loading.value = false
  }
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    refreshTimer = setInterval(fetchLogs, 3000)
    toast.info('Auto-refresh enabled (3s)')
  } else {
    if (refreshTimer) clearInterval(refreshTimer)
    refreshTimer = null
    toast.info('Auto-refresh disabled')
  }
}

async function copyLogs() {
  await navigator.clipboard.writeText(logs.value)
  toast.success('Logs copied to clipboard')
}

onMounted(fetchLogs)

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <div class="logs-view">
    <div class="section-header">
      <div>
        <h1 class="section-title">Logs</h1>
        <p class="section-subtitle">Application logs</p>
      </div>
      <div class="header-actions">
        <div class="inline-form">
          <label class="text-muted" style="font-size: 13px">Lines:</label>
          <input class="input" style="width: 80px" type="number" v-model="lines" min="50" max="5000" @change="fetchLogs" />
        </div>
        <button class="btn" :class="autoRefresh ? 'btn-primary' : 'btn-secondary'" @click="toggleAutoRefresh">
          {{ autoRefresh ? '⏸ Auto' : '▶ Auto' }}
        </button>
        <button class="btn btn-secondary" @click="fetchLogs" :disabled="loading">
          ↻ Refresh
        </button>
        <button class="btn btn-secondary" @click="copyLogs">
          📋 Copy
        </button>
      </div>
    </div>

    <div class="log-container" style="margin-top: 16px">
      <pre class="log-output">{{ logs || 'No logs available' }}</pre>
    </div>
  </div>
</template>

<style scoped>
.logs-view {
  max-width: 100%;
  height: calc(100vh - 48px);
  display: flex;
  flex-direction: column;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.inline-form {
  display: flex;
  align-items: center;
  gap: 4px;
}

.log-container {
  flex: 1;
  overflow: auto;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}

.log-output {
  padding: 16px;
  margin: 0;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--text-secondary);
}
</style>
