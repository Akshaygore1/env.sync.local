import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { StatusInfo, FileStatus, ServerStatus, BackupEntry, ModeInfo } from '@/types'

export const useStatusStore = defineStore('status', () => {
  const status = ref<StatusInfo | null>(null)
  const loading = ref(false)
  const lastRefresh = ref<string>('')
  const lastFileModTime = ref<string>('')

  const fileStatus = computed<FileStatus | null>(() => status.value?.secretsFile ?? null)
  const serverStatus = computed<ServerStatus | null>(() => status.value?.server ?? null)
  const backups = computed<BackupEntry[]>(() => status.value?.backups ?? [])
  const modeInfo = computed<ModeInfo | null>(() => status.value?.mode ?? null)

  async function fetchStatus() {
    loading.value = true
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.StatusService.GetStatus()
      status.value = result
      lastRefresh.value = new Date().toISOString()
    } catch (e) {
      console.error('Failed to fetch status:', e)
    } finally {
      loading.value = false
    }
  }

  async function checkFileChanged(): Promise<boolean> {
    try {
      // @ts-expect-error Wails bindings
      const modTime = await window.go.main.StatusService.GetFileModTime()
      if (modTime !== lastFileModTime.value) {
        lastFileModTime.value = modTime
        return true
      }
    } catch {
      // file might not exist yet
    }
    return false
  }

  return {
    status,
    loading,
    lastRefresh,
    fileStatus,
    serverStatus,
    backups,
    modeInfo,
    fetchStatus,
    checkFileChanged,
  }
})
