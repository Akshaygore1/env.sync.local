import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { SecretEntry } from '@/types'

export const useSecretsStore = defineStore('secrets', () => {
  const entries = ref<SecretEntry[]>([])
  const loading = ref(false)
  const filter = ref('')

  const filtered = computed(() => {
    if (!filter.value) return entries.value
    const q = filter.value.toLowerCase()
    return entries.value.filter(e => e.key.toLowerCase().includes(q))
  })

  const count = computed(() => entries.value.length)

  async function fetchAll() {
    loading.value = true
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.SecretsService.List()
      entries.value = result || []
    } catch (e) {
      console.error('Failed to fetch secrets:', e)
    } finally {
      loading.value = false
    }
  }

  async function addSecret(key: string, value: string) {
    // @ts-expect-error Wails bindings
    await window.go.main.SecretsService.Add(key, value)
    await fetchAll()
  }

  async function removeSecret(key: string) {
    // @ts-expect-error Wails bindings
    await window.go.main.SecretsService.Remove(key)
    await fetchAll()
  }

  async function getSecret(key: string): Promise<SecretEntry> {
    // @ts-expect-error Wails bindings
    return await window.go.main.SecretsService.Get(key)
  }

  async function exportEnv(): Promise<string> {
    // @ts-expect-error Wails bindings
    return await window.go.main.SecretsService.ExportEnv()
  }

  async function exportJSON(): Promise<string> {
    // @ts-expect-error Wails bindings
    return await window.go.main.SecretsService.ExportJSON()
  }

  return {
    entries,
    loading,
    filter,
    filtered,
    count,
    fetchAll,
    addSecret,
    removeSecret,
    getSecret,
    exportEnv,
    exportJSON,
  }
})
