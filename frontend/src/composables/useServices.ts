import { ref, onMounted, onUnmounted } from 'vue'
import type { ServiceStatusData } from '@/types'

export function useServices(intervalMs = 30_000) {
  const services    = ref<ServiceStatusData[]>([])
  const loading     = ref(true)
  const error       = ref<string | null>(null)
  const lastUpdated = ref<Date | null>(null)

  let timerId: ReturnType<typeof setInterval> | null = null

  async function fetchServices() {
    try {
      const res = await fetch('/api/services')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      services.value  = data.services ?? []
      lastUpdated.value = new Date()
      error.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    fetchServices()
    timerId = setInterval(fetchServices, intervalMs)
  })

  onUnmounted(() => {
    if (timerId !== null) clearInterval(timerId)
  })

  return { services, loading, error, lastUpdated, refresh: fetchServices }
}
