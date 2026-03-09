<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useSecretsStore } from '@/stores/secrets'
import { useToast } from '@/composables/useToast'

const secrets = useSecretsStore()
const toast = useToast()

const showAdd = ref(false)
const newKey = ref('')
const newValue = ref('')
const revealedKeys = ref<Set<string>>(new Set())
const confirmDelete = ref<string | null>(null)

function toggleReveal(key: string) {
  if (revealedKeys.value.has(key)) {
    revealedKeys.value.delete(key)
  } else {
    revealedKeys.value.add(key)
  }
}

async function copyValue(key: string) {
  try {
    const entry = await secrets.getSecret(key)
    await navigator.clipboard.writeText(entry.value)
    toast.success(`Copied ${key} to clipboard`)
  } catch {
    toast.error('Failed to copy')
  }
}

async function addSecret() {
  if (!newKey.value || !newValue.value) return
  try {
    await secrets.addSecret(newKey.value, newValue.value)
    toast.success(`Added ${newKey.value}`)
    newKey.value = ''
    newValue.value = ''
    showAdd.value = false
  } catch (e) {
    toast.error('Failed to add secret: ' + e)
  }
}

async function deleteSecret(key: string) {
  try {
    await secrets.removeSecret(key)
    toast.success(`Removed ${key}`)
    confirmDelete.value = null
  } catch (e) {
    toast.error('Failed to remove: ' + e)
  }
}

async function exportAs(format: 'env' | 'json') {
  try {
    const content = format === 'env' ? await secrets.exportEnv() : await secrets.exportJSON()
    await navigator.clipboard.writeText(content)
    toast.success(`Exported as ${format.toUpperCase()} and copied to clipboard`)
  } catch (e) {
    toast.error('Export failed: ' + e)
  }
}

onMounted(() => {
  secrets.fetchAll()
})
</script>

<template>
  <div class="secrets-view">
    <div class="section-header">
      <div>
        <h1 class="section-title">Secrets</h1>
        <p class="section-subtitle">{{ secrets.count }} keys stored</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-primary" @click="showAdd = true">+ Add</button>
        <button class="btn btn-secondary" @click="exportAs('env')">⬇ .env</button>
        <button class="btn btn-secondary" @click="exportAs('json')">⬇ JSON</button>
      </div>
    </div>

    <!-- Filter -->
    <div class="filter-bar">
      <input
        class="input"
        type="text"
        placeholder="🔍 Filter keys..."
        v-model="secrets.filter"
      />
    </div>

    <!-- Secrets table -->
    <div class="card" style="padding: 0; margin-top: 12px">
      <table class="table" v-if="secrets.filtered.length > 0">
        <thead>
          <tr>
            <th>Key</th>
            <th>Value</th>
            <th>Updated</th>
            <th style="width: 120px">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="entry in secrets.filtered" :key="entry.key">
            <td class="mono" style="font-weight: 600">{{ entry.key }}</td>
            <td class="mono">
              <span v-if="!revealedKeys.has(entry.key)" class="masked">••••••••</span>
              <span v-else class="revealed">{{ entry.value }}</span>
              <button class="btn btn-ghost btn-sm" @click="toggleReveal(entry.key)" style="margin-left: 4px">
                {{ revealedKeys.has(entry.key) ? '🙈' : '👁' }}
              </button>
            </td>
            <td class="text-muted" style="font-size: 12px">
              {{ entry.updatedAt ? new Date(entry.updatedAt).toLocaleString() : '-' }}
            </td>
            <td>
              <div class="action-btns">
                <button class="btn btn-ghost btn-sm" @click="copyValue(entry.key)" title="Copy">📋</button>
                <button class="btn btn-ghost btn-sm" @click="confirmDelete = entry.key" title="Delete">🗑</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        <p class="text-muted">No secrets found</p>
      </div>
    </div>

    <!-- Add Secret Modal -->
    <div class="modal-overlay" v-if="showAdd" @click.self="showAdd = false">
      <div class="modal">
        <h2 class="modal-title">Add Secret</h2>
        <div class="form-group">
          <label class="form-label">Key</label>
          <input class="input input-mono" v-model="newKey" placeholder="MY_SECRET_KEY" @keyup.enter="addSecret" />
        </div>
        <div class="form-group">
          <label class="form-label">Value</label>
          <input class="input input-mono" v-model="newValue" placeholder="secret-value" type="password" @keyup.enter="addSecret" />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showAdd = false">Cancel</button>
          <button class="btn btn-primary" @click="addSecret" :disabled="!newKey || !newValue">Add</button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div class="modal-overlay" v-if="confirmDelete" @click.self="confirmDelete = null">
      <div class="modal">
        <h2 class="modal-title">Delete Secret</h2>
        <p>Are you sure you want to delete <strong class="mono">{{ confirmDelete }}</strong>?</p>
        <p class="text-muted" style="margin-top: 8px">This action cannot be undone.</p>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="confirmDelete = null">Cancel</button>
          <button class="btn btn-danger" @click="deleteSecret(confirmDelete!)">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.secrets-view {
  max-width: 1000px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.filter-bar {
  margin-top: 16px;
}

.masked {
  color: var(--text-muted);
  letter-spacing: 2px;
}

.revealed {
  word-break: break-all;
}

.action-btns {
  display: flex;
  gap: 2px;
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
