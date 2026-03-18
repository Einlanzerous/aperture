import { ref, computed, type Ref } from 'vue'

// Module-scoped shared state — same ref across all component instances
const globalOverride = ref<boolean | null>(null)

export function useDetailMode(serviceDefault: Ref<boolean> | boolean = false) {
  const defaultVal = computed(() =>
    typeof serviceDefault === 'boolean' ? serviceDefault : serviceDefault.value,
  )

  const showDetail = computed(() =>
    globalOverride.value !== null ? globalOverride.value : defaultVal.value,
  )

  const isDetailMode = computed(() => globalOverride.value === true)

  function toggleGlobal() {
    globalOverride.value = globalOverride.value === true ? false : true
  }

  return { showDetail, isDetailMode, toggleGlobal }
}
