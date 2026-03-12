import { computed, type Ref } from 'vue'
import { usePollingFetch } from './usePollingFetch'
import { API } from '@/utils/api'
import type { OllamaModel, OllamaModelsResponse } from '@/types'

export function useOllama(intervalMs: number | Ref<number> = 60_000) {
  const { data, loading, error, refresh } = usePollingFetch<OllamaModelsResponse>(
    API.ollama,
    intervalMs,
    { default: { models: [] } },
  )

  const models = computed<OllamaModel[]>(() => data.value.models ?? [])

  return { models, loading, error, refresh }
}
