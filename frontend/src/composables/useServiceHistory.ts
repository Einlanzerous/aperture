import { ref, watch, type Ref } from 'vue'
import { apiFetch } from '@/utils/api'
import { API } from '@/utils/api'
import type { CheckRecord, DailySummary, HistoryResponse, UptimeResponse } from '@/types'

export function useServiceHistory(
  serviceName: Ref<string>,
  showDetail: Ref<boolean>,
  enabled: Ref<boolean>,
) {
  const summaries = ref<DailySummary[]>([])
  const records = ref<CheckRecord[]>([])
  const detailLoaded = ref(false)

  let uptimeFetched = false
  let detailFetched = false

  async function fetchUptime() {
    if (uptimeFetched || !enabled.value) return
    try {
      const data = await apiFetch<UptimeResponse>(API.serviceUptime(serviceName.value, 30))
      summaries.value = data.summaries ?? []
      uptimeFetched = true
    } catch {
      // silent — history sections hidden via v-if
    }
  }

  async function fetchDetail() {
    if (detailFetched || !enabled.value) return
    try {
      const data = await apiFetch<HistoryResponse>(API.serviceHistory(serviceName.value, '24h'))
      records.value = data.records ?? []
      detailFetched = true
      detailLoaded.value = true
    } catch {
      // silent
    }
  }

  // Fetch uptime immediately when enabled
  watch(enabled, (val) => {
    if (val && !uptimeFetched) fetchUptime()
  }, { immediate: true })

  // Lazily fetch detail on first showDetail=true
  watch(showDetail, (val) => {
    if (val && !detailFetched && enabled.value) fetchDetail()
  }, { immediate: true })

  return { summaries, records, detailLoaded }
}
