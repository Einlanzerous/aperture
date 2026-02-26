import { ref, onMounted, onUnmounted } from 'vue'
import type { SystemResources } from '@/types'

export function useResources(intervalMs = 5_000) {
  const resources = ref<SystemResources | null>(null)
  const loading   = ref(true)
  const error     = ref<string | null>(null)

  let timerId: ReturnType<typeof setInterval> | null = null

  async function fetchResources() {
    try {
      const res = await fetch('/api/system/resources')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      resources.value = await res.json()
      error.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    fetchResources()
    timerId = setInterval(fetchResources, intervalMs)
  })

  onUnmounted(() => {
    if (timerId !== null) clearInterval(timerId)
  })

  return { resources, loading, error, refresh: fetchResources }
}
