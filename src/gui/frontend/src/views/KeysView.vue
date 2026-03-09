<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useToast } from '@/composables/useToast'
import type { KeyInfo } from '@/types'

const toast = useToast()
const localKey = ref<KeyInfo | null>(null)
const allKeys = ref<KeyInfo[]>([])
const loading = ref(false)
const showImport = ref(false)
const importPublicKey = ref('')
const importHostname = ref('')
const showFullKey = ref<Set<string>>(new Set())

async function fetchKeys() {
  loading.value = true
  try {
    // @ts-expect-error Wails bindings
    localKey.value = await window.go.main.KeysService.GetLocalKey()
    // @ts-expect-error Wails bindings
    allKeys.value = await window.go.main.KeysService.ListKeys()
  } catch (e) {
    console.error('Failed to fetch keys:', e)
  } finally {
    loading.value = false
  }
}

async function copyPublicKey() {
  if (localKey.value?.publicKey) {
    await navigator.clipboard.writeText(localKey.value.publicKey)
    toast.success('Public key copied to clipboard')
  }
}

async function importKey() {
  if (!importPublicKey.value || !importHostname.value) return
  try {
    // @ts-expect-error Wails bindings
    await window.go.main.KeysService.ImportKey(importPublicKey.value, importHostname.value)
    toast.success(`Imported key for ${importHostname.value}`)
    importPublicKey.value = ''
    importHostname.value = ''
    showImport.value = false
    await fetchKeys()
  } catch (e) {
    toast.error('Import failed: ' + e)
  }
}

function toggleKey(hostname: string) {
  if (showFullKey.value.has(hostname)) {
    showFullKey.value.delete(hostname)
  } else {
    showFullKey.value.add(hostname)
  }
}

onMounted(fetchKeys)
</script>

<template>
  <div class="keys-view">
    <div class="section-header">
      <div>
        <h1 class="section-title">Keys</h1>
        <p class="section-subtitle">AGE encryption keys</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-primary" @click="showImport = true">+ Import Key</button>
        <button class="btn btn-secondary" @click="fetchKeys">↻ Refresh</button>
      </div>
    </div>

    <!-- Local Key -->
    <div class="card" style="margin-top: 16px" v-if="localKey">
      <div class="card-header">
        <span class="card-title">🔑 Local Key</span>
        <button class="btn btn-secondary btn-sm" @click="copyPublicKey">📋 Copy Public Key</button>
      </div>
      <div class="key-info">
        <div class="info-row">
          <span class="text-muted">Hostname</span>
          <span class="mono">{{ localKey.hostname }}</span>
        </div>
        <div class="info-row">
          <span class="text-muted">Public Key</span>
          <span class="mono key-value">{{ localKey.publicKey }}</span>
        </div>
        <div class="info-row" v-if="localKey.fingerprint">
          <span class="text-muted">Fingerprint</span>
          <span class="mono">{{ localKey.fingerprint }}</span>
        </div>
      </div>
    </div>

    <!-- All Known Keys -->
    <div class="card" style="margin-top: 16px">
      <div class="card-header">
        <span class="card-title">Known Keys</span>
        <span class="badge badge-info">{{ allKeys.length }}</span>
      </div>
      <div v-if="allKeys.length > 0">
        <div v-for="key in allKeys" :key="key.hostname" class="key-row">
          <div class="key-row-info">
            <span class="mono" style="font-weight: 600">{{ key.hostname }}</span>
            <span class="mono key-preview" @click="toggleKey(key.hostname)">
              {{ showFullKey.has(key.hostname) ? key.publicKey : key.publicKey?.substring(0, 24) + '...' }}
            </span>
          </div>
          <span class="badge" :class="key.isLocal ? 'badge-success' : 'badge-info'">
            {{ key.isLocal ? 'Local' : 'Remote' }}
          </span>
        </div>
      </div>
      <div v-else class="empty-state">
        <p class="text-muted">No keys found</p>
      </div>
    </div>

    <!-- Import Key Modal -->
    <div class="modal-overlay" v-if="showImport" @click.self="showImport = false">
      <div class="modal">
        <h2 class="modal-title">Import Public Key</h2>
        <div class="form-group">
          <label class="form-label">Hostname</label>
          <input class="input input-mono" v-model="importHostname" placeholder="peer.local" />
        </div>
        <div class="form-group">
          <label class="form-label">Public Key</label>
          <input class="input input-mono" v-model="importPublicKey" placeholder="age1..." />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showImport = false">Cancel</button>
          <button class="btn btn-primary" @click="importKey" :disabled="!importPublicKey || !importHostname">
            Import
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.keys-view {
  max-width: 1000px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.key-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row span:first-child {
  font-size: 13px;
  min-width: 120px;
}

.key-value {
  font-size: 12px;
  word-break: break-all;
  max-width: 500px;
  text-align: right;
}

.key-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color);
}

.key-row:last-child {
  border-bottom: none;
}

.key-row-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.key-preview {
  font-size: 12px;
  color: var(--text-muted);
  cursor: pointer;
}

.empty-state {
  padding: 40px;
  text-align: center;
}

.form-group {
  margin-bottom: 12px;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--text-secondary);
}
</style>
