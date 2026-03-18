import { computed, type Ref } from 'vue'
import { usePollingFetch } from './usePollingFetch'
import { API } from '@/utils/api'
import type { ActionState, ActionsResponse } from '@/types'

export function useActions(intervalMs: number | Ref<number> = 30_000) {
  const { data, loading, error, refresh } = usePollingFetch<ActionsResponse>(
    API.actions,
    intervalMs,
    { default: { actions: [] } },
  )

  const actions = computed<ActionState[]>(() => data.value.actions ?? [])

  return { actions, loading, error, refresh }
}
