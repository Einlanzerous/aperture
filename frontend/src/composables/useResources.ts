import { isRef, type Ref } from 'vue'
import { usePollingFetch } from './usePollingFetch'
import { API } from '@/utils/api'
import { POLL_RESOURCES_MS } from '@/constants/polling'
import type { SystemResources } from '@/types'

// history may be a reactive ref so callers (e.g. the Load widget's window
// selector) can change the requested sample count at runtime; the next poll —
// or an explicit refresh() — picks up the new value.
export function useResources(
  intervalMs: number | Ref<number> = POLL_RESOURCES_MS,
  history: number | Ref<number> = 0,
) {
  const urlFn = () => {
    const n = isRef(history) ? history.value : history
    return n > 0 ? `${API.resources}?history=${n}` : API.resources
  }
  return usePollingFetch<SystemResources | null>(
    urlFn,
    intervalMs,
    { default: null },
  )
}
