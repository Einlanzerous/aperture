import type { Ref } from 'vue'
import { usePollingFetch } from './usePollingFetch'
import { API } from '@/utils/api'
import type { SystemResources } from '@/types'

export function useResources(intervalMs: number | Ref<number> = 5_000) {
  return usePollingFetch<SystemResources | null>(
    API.resources,
    intervalMs,
    { default: null },
  )
}
