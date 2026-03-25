import type { Ref } from 'vue'
import { usePollingFetch } from './usePollingFetch'
import { API } from '@/utils/api'
import { POLL_RESOURCES_MS } from '@/constants/polling'
import type { SystemResources } from '@/types'

export function useResources(intervalMs: number | Ref<number> = POLL_RESOURCES_MS) {
  return usePollingFetch<SystemResources | null>(
    API.resources,
    intervalMs,
    { default: null },
  )
}
