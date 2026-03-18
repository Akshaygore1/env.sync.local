<script setup lang="ts">
</script>

<template>
  <header>
    <div class="nav">
      <NuxtLink class="brand" to="/">env-sync</NuxtLink>
      
      <nav class="hidden md:block">
        <ul>
          <li class="dropdown">
            <NuxtLink to="/installation">Docs</NuxtLink>
            <ul class="dropdown-menu">
              <li><NuxtLink to="/installation/quickstart">Quickstart</NuxtLink></li>
              <li><NuxtLink to="/installation/trusted-peers">Trusted Peers mode</NuxtLink></li>
              <li><NuxtLink to="/installation/secure-peers">Secure Peers mode</NuxtLink></li>
            </ul>
          </li>
          <li><NuxtLink to="/usage">Usage</NuxtLink></li>
          <li><NuxtLink to="/modes">Security Modes</NuxtLink></li>
          <li><NuxtLink to="/comparisons">Compare</NuxtLink></li>
        </ul>
      </nav>

      <div class="flex items-center gap-4">
        <a class="github-link" href="https://github.com/championswimmer/env.sync.local">
          GitHub <span class="text-muted text-sm font-normal">[source]</span>
        </a>
      </div>
    </div>
  </header>
</template>

<style scoped>
.hidden { display: none; }
@media (min-width: 768px) {
  .hidden\.md\:block { display: block; }
}
.flex { display: flex; }
.items-center { align-items: center; }
.gap-4 { gap: 1rem; }
.text-sm { font-size: 0.875rem; }
.font-normal { font-weight: normal; }
</style>
