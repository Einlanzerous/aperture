import { computed, type Ref } from 'vue'
import { usePollingFetch } from './usePollingFetch'
import { API } from '@/utils/api'
import { POLL_SERVICES_MS } from '@/constants/polling'
import type { ServiceStatusData, ServicesResponse } from '@/types'

export function useServices(intervalMs: number | Ref<number> = POLL_SERVICES_MS) {
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
