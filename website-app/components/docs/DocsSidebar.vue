<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()

const navGroups = [
  {
    title: 'Getting Started',
    links: [
      { name: 'Installation', path: '/installation' },
      { name: 'Quickstart', path: '/installation/quickstart' },
      { name: 'Trusted Peers', path: '/installation/trusted-peers' },
      { name: 'Secure Peers', path: '/installation/secure-peers' },
    ]
  },
  {
    title: 'Core Concepts',
    links: [
      { name: 'Modes', path: '/modes' },
      { name: 'Security', path: '/security' },
    ]
  },
  {
    title: 'Reference',
    links: [
      { name: 'Usage Guide', path: '/usage' },
    ]
  },
  {
    title: 'Comparisons',
    links: [
      { name: 'Comparisons Overview', path: '/comparisons' },
      { name: 'vs dotenvx', path: '/comparisons/dotenvx-vs-envsync' },
      { name: 'vs Doppler', path: '/comparisons/doppler-vs-envsync' },
      { name: 'vs Infisical', path: '/comparisons/infisical-vs-envsync' },
      { name: 'vs SOPS', path: '/comparisons/sops-vs-envsync' },
      { name: 'vs HashiCorp Vault', path: '/comparisons/vault-vs-envsync' },
    ]
  }
]
</script>

<template>
  <aside class="docs-sidebar">
    <div v-for="group in navGroups" :key="group.title" class="sidebar-group">
      <h3 class="sidebar-heading">{{ group.title }}</h3>
      <ul class="sidebar-list">
        <li v-for="link in group.links" :key="link.path">
          <NuxtLink
            :to="link.path"
            class="sidebar-link"
            :class="{ 'active': route.path === link.path }"
          >
            {{ link.name }}
          </NuxtLink>
        </li>
      </ul>
    </div>
  </aside>
</template>

<style scoped>
.docs-sidebar {
  width: 100%;
  margin-bottom: 2rem;
}
@media (min-width: 768px) {
  .docs-sidebar {
    width: 260px;
    flex-shrink: 0;
    position: sticky;
    top: 90px;
    height: calc(100vh - 90px);
    overflow-y: auto;
    border-right: 1px solid var(--border);
    padding-right: 1.5rem;
    margin-bottom: 0;
  }
}
.sidebar-group { margin-bottom: 2rem; }
.sidebar-heading {
  font-size: 0.75rem;
  font-family: var(--font-sans);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-muted);
  margin-bottom: 0.75rem;
  font-weight: 700;
}
.sidebar-list { list-style: none; padding: 0; margin: 0; }
.sidebar-list li { margin-bottom: 0.25rem; margin-top: 0; }
.sidebar-link {
  display: block;
  padding: 0.5rem 0.75rem;
  font-size: 0.9rem;
  color: var(--text-muted);
  border-radius: 6px;
  border-left: 2px solid transparent;
  transition: all 0.2s;
  text-decoration: none;
}
.sidebar-link:hover {
  background: var(--bg-panel-hover);
  color: var(--text-main);
  text-decoration: none;
}
.sidebar-link.active {
  background: var(--accent-light);
  color: var(--accent);
  border-left-color: var(--accent);
  font-weight: 600;
}
</style>
