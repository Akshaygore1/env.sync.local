<script setup lang="ts">
defineProps<{
  size?: 'sm' | 'md' | 'lg'
  label?: string
}>()
</script>

<template>
  <div class="loading-wrapper">
    <div class="spinner" :class="`spinner-${size || 'md'}`"></div>
    <span v-if="label" class="loading-label text-muted">{{ label }}</span>
  </div>
</template>

<style scoped>
.loading-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 24px;
}

.spinner-sm {
  width: 16px;
  height: 16px;
}

.spinner-md {
  width: 24px;
  height: 24px;
}

.spinner-lg {
  width: 40px;
  height: 40px;
}

.loading-label {
  font-size: 13px;
}
</style>
