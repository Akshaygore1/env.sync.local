<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { toasts } = useToast()
</script>

<template>
  <div class="toast-container">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="toast"
      :class="`toast-${toast.type}`"
    >
      <span class="toast-icon">
        {{ toast.type === 'success' ? '✓' : toast.type === 'error' ? '✗' : 'ℹ' }}
      </span>
      <span>{{ toast.message }}</span>
    </div>
  </div>
</template>

<style scoped>
.toast-icon {
  font-weight: bold;
}
</style>
