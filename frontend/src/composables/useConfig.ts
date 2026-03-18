import { ref } from 'vue'
import { apiFetch } from '@/utils/api'
import { API } from '@/utils/api'
import { getErrorMessage } from '@/utils/format'
import type { DashboardConfig } from '@/types'

const defaultConfig: DashboardConfig = {
  title:          'Aperture',
  checkInterval:  30,
  ollamaEnabled:  false,
  systemEnabled:  false,
  actionsEnabled: false,
}

export function useConfig() {
  const config  = ref<DashboardConfig>({ ...defaultConfig })
  const loading = ref(true)
  const error   = ref<string | null>(null)

  async function load() {
    try {
      config.value = await apiFetch<DashboardConfig>(API.config)
      error.value = null
    } catch (e) {
      error.value = getErrorMessage(e)
    } finally {
      loading.value = false
    }
  }

  return { config, loading, error, load }
}
