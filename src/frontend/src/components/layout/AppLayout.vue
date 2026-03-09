<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import ToastContainer from '@/components/common/ToastContainer.vue'

const router = useRouter()
const route = useRoute()
const settings = useSettingsStore()
const collapsed = ref(false)
const isDark = ref(true)

const navItems = [
  { path: '/', name: 'Dashboard', icon: '🏠' },
  { path: '/secrets', name: 'Secrets', icon: '🔑' },
  { path: '/sync', name: 'Sync', icon: '🔄' },
  { path: '/peers', name: 'Peers', icon: '👥' },
  { path: '/keys', name: 'Keys', icon: '🗝️' },
  { path: '/settings', name: 'Settings', icon: '⚙️' },
  { path: '/logs', name: 'Logs', icon: '📋' },
]

function isActive(path: string): boolean {
  return route.path === path
}

function navigate(path: string) {
  router.push(path)
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.setAttribute('data-theme', isDark.value ? 'dark' : 'light')
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(async () => {
  // Load saved theme
  const saved = localStorage.getItem('theme')
  if (saved === 'light') {
    isDark.value = false
    document.documentElement.setAttribute('data-theme', 'light')
  }
  await settings.fetchVersion()
  await settings.fetchMode()
})
</script>

<template>
  <div class="app-layout">
    <aside class="sidebar" :class="{ collapsed }">
      <div class="sidebar-header">
        <div class="logo" v-if="!collapsed">
          <span class="logo-icon">🔐</span>
          <span class="logo-text">env-sync</span>
        </div>
        <button class="btn-ghost toggle-btn" @click="collapsed = !collapsed">
          {{ collapsed ? '→' : '←' }}
        </button>
      </div>

      <nav class="sidebar-nav">
        <button
          v-for="item in navItems"
          :key="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
          @click="navigate(item.path)"
        >
          <span class="nav-icon">{{ item.icon }}</span>
          <span class="nav-label" v-if="!collapsed">{{ item.name }}</span>
        </button>
      </nav>

      <div class="sidebar-footer" v-if="!collapsed">
        <button class="btn btn-ghost btn-sm" @click="toggleTheme" aria-label="Toggle theme">
          {{ isDark ? '☀️ Light' : '🌙 Dark' }}
        </button>
        <div class="mode-badge" v-if="settings.currentMode">
          <span class="badge badge-info">{{ settings.currentMode.current }}</span>
        </div>
        <div class="version text-muted">v{{ settings.version }}</div>
      </div>
    </aside>

    <main class="main-content">
      <router-view />
    </main>

    <ToastContainer />
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.sidebar {
  width: var(--sidebar-width);
  min-width: var(--sidebar-width);
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  transition: width 0.2s ease, min-width 0.2s ease;
}

.sidebar.collapsed {
  width: 60px;
  min-width: 60px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-icon {
  font-size: 20px;
}

.logo-text {
  font-size: 16px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.toggle-btn {
  padding: 4px 8px;
  font-size: 12px;
}

.sidebar-nav {
  flex: 1;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: left;
  width: 100%;
}

.nav-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.nav-item.active {
  background: var(--accent-color);
  color: white;
}

.nav-icon {
  font-size: 16px;
  width: 24px;
  text-align: center;
}

.nav-label {
  font-size: 14px;
  font-weight: 500;
}

.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.version {
  font-size: 12px;
  font-family: var(--font-mono);
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
</style>
