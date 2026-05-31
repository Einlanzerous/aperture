import type { Ref } from 'vue'
import { usePollingFetch } from './usePollingFetch'
import { API } from '@/utils/api'
import { POLL_RESOURCES_MS } from '@/constants/polling'
import type { SystemResources } from '@/types'

export function useResources(
  intervalMs: number | Ref<number> = POLL_RESOURCES_MS,
  history = 0,
) {
  const url = history > 0
    ? `${API.resources}?history=${history}`
    : API.resources
  return usePollingFetch<SystemResources | null>(
    url,
    intervalMs,
    { default: null },
  )
}
