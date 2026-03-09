import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ModeInfo, CronInfo, ConfigPaths } from '@/types'

export const useSettingsStore = defineStore('settings', () => {
  const currentMode = ref<ModeInfo | null>(null)
  const availableModes = ref<ModeInfo[]>([])
  const cronInfo = ref<CronInfo | null>(null)
  const configPaths = ref<ConfigPaths | null>(null)
  const version = ref('')
  const initialized = ref(false)

  async function fetchMode() {
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.ModeService.GetMode()
      currentMode.value = result
    } catch (e) {
      console.error('Failed to fetch mode:', e)
    }
  }

  async function fetchAvailableModes() {
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.ModeService.GetAvailableModes()
      availableModes.value = result || []
    } catch (e) {
      console.error('Failed to fetch available modes:', e)
    }
  }

  async function setMode(mode: string, pruneOldMaterial = false) {
    // @ts-expect-error Wails bindings
    await window.go.main.ModeService.SetMode(mode, pruneOldMaterial)
    await fetchMode()
  }

  async function fetchCron() {
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.CronService.GetCronStatus()
      cronInfo.value = result
    } catch (e) {
      console.error('Failed to fetch cron status:', e)
    }
  }

  async function installCron(interval: number) {
    // @ts-expect-error Wails bindings
    await window.go.main.CronService.InstallCron(interval)
    await fetchCron()
  }

  async function removeCron() {
    // @ts-expect-error Wails bindings
    await window.go.main.CronService.RemoveCron()
    await fetchCron()
  }

  async function fetchConfigPaths() {
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.App.GetConfigPaths()
      configPaths.value = result
    } catch (e) {
      console.error('Failed to fetch config paths:', e)
    }
  }

  async function fetchVersion() {
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.App.GetVersion()
      version.value = result
    } catch (e) {
      console.error('Failed to fetch version:', e)
    }
  }

  async function checkInitialized() {
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.App.IsInitialized()
      initialized.value = result
    } catch (e) {
      console.error('Failed to check initialization:', e)
    }
  }

  async function initialize(encrypted: boolean) {
    // @ts-expect-error Wails bindings
    await window.go.main.SecretsService.Initialize(encrypted)
    initialized.value = true
  }

  return {
    currentMode,
    availableModes,
    cronInfo,
    configPaths,
    version,
    initialized,
    fetchMode,
    fetchAvailableModes,
    setMode,
    fetchCron,
    installCron,
    removeCron,
    fetchConfigPaths,
    fetchVersion,
    checkInitialized,
    initialize,
  }
})
