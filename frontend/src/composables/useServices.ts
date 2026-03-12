import { computed, type Ref } from 'vue'
import { usePollingFetch } from './usePollingFetch'
import { API } from '@/utils/api'
import type { ServiceStatusData, ServicesResponse } from '@/types'

export function useServices(intervalMs: number | Ref<number> = 30_000) {
  const { data, loading, error, refresh } = usePollingFetch<ServicesResponse>(
    API.services,
    intervalMs,
    { default: { services: [], updatedAt: '' } },
  )

  const services    = computed<ServiceStatusData[]>(() => data.value.services ?? [])
  const lastUpdated = computed<Date | null>(() =>
    data.value.updatedAt ? new Date(data.value.updatedAt) : null,
  )

  return { services, loading, error, lastUpdated, refresh }
}
