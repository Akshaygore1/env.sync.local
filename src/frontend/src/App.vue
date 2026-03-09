<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from './components/layout/AppLayout.vue'
import FirstRunWizard from './components/common/FirstRunWizard.vue'
import { useSettingsStore } from './stores/settings'

const settings = useSettingsStore()
const ready = ref(false)

onMounted(async () => {
  await settings.checkInitialized()
  ready.value = true
})
</script>

<template>
  <div v-if="ready">
    <FirstRunWizard v-if="!settings.initialized" />
    <AppLayout v-else />
  </div>
</template>
