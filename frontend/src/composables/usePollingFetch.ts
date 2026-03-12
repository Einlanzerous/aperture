import { ref, watch, isRef, onMounted, onUnmounted, type Ref } from 'vue'
import { apiFetch } from '@/utils/api'
import { getErrorMessage } from '@/utils/format'

export function usePollingFetch<T>(
  url: string | (() => string),
  intervalMs: number | Ref<number>,
  options?: { transform?: (data: any) => T; default?: T },
) {
  const data    = ref<T>((options?.default ?? null) as T) as Ref<T>
  const loading = ref(true)
  const error   = ref<string | null>(null)

  let timerId: ReturnType<typeof setInterval> | null = null

  async function refresh() {
    try {
      const resolvedUrl = typeof url === 'function' ? url() : url
      const raw = await apiFetch<any>(resolvedUrl)
      data.value = options?.transform ? options.transform(raw) : raw
      error.value = null
    } catch (e) {
      error.value = getErrorMessage(e)
    } finally {
      loading.value = false
    }
  }

  function startTimer() {
    stopTimer()
    const ms = isRef(intervalMs) ? intervalMs.value : intervalMs
    timerId = setInterval(refresh, ms)
  }

  function stopTimer() {
    if (timerId !== null) {
      clearInterval(timerId)
      timerId = null
    }
  }

  onMounted(() => {
    refresh()
    startTimer()
  })

  onUnmounted(stopTimer)

  // Restart timer when interval changes (reactive support).
  if (isRef(intervalMs)) {
    watch(intervalMs, () => {
      startTimer()
    })
  }

  return { data, loading, error, refresh }
}
