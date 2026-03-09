import { ref } from 'vue'
import type { Toast, ToastType } from '@/types'

const toasts = ref<Toast[]>([])
let nextId = 0

export function useToast() {
  function show(type: ToastType, message: string, duration = 3000) {
    const id = nextId++
    toasts.value.push({ id, type, message })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, duration)
  }

  function success(message: string) { show('success', message) }
  function error(message: string) { show('error', message, 5000) }
  function info(message: string) { show('info', message) }

  return { toasts, success, error, info }
}
