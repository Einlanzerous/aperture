<script setup lang="ts">
import { computed } from 'vue'
import { useResources } from '@/composables/useResources'
import { fmtBytes } from '@/utils/format'

const WARNING_THRESHOLD  = 60
const CRITICAL_THRESHOLD = 85

const { data: resources, loading, error } = useResources(5_000)

const mem = computed(() => resources.value?.memory ?? null)

function barColor(percent: number): string {
  if (percent < WARNING_THRESHOLD)  return 'bg-emerald-400'
  if (percent < CRITICAL_THRESHOLD) return 'bg-amber-400'
  return 'bg-red-400'
}

const barClass = computed(() => (mem.value ? barColor(mem.value.percent) : 'bg-gray-600'))
const barWidth = computed(() => `${Math.min(mem.value?.percent ?? 0, 100)}%`)

// Left value = currently used; right value = total (the max).
const usedLabel  = computed(() => (mem.value ? fmtBytes(mem.value.used) : '—'))
const totalLabel = computed(() => (mem.value ? fmtBytes(mem.value.total) : '—'))
</script>

<template>
  <!-- Tiny (1-slot) tile: RAM | used [thick bar] total. -->
  <article class="widget-card cursor-default">
    <div class="flex h-full items-center gap-3 px-4">
      <span class="w-10 shrink-0 text-sm font-semibold text-gray-100">RAM</span>

      <template v-if="error">
        <span class="flex-1 truncate text-xs text-red-400" :title="error">{{ error }}</span>
      </template>

      <template v-else>
        <span class="w-16 shrink-0 text-right text-xs tabular-nums text-gray-200">
          {{ usedLabel }}
        </span>
        <div class="relative h-2.5 flex-1 overflow-hidden rounded-full bg-gray-800">
          <div
            class="h-full rounded-full transition-all duration-700"
            :class="loading ? 'bg-gray-700 animate-pulse' : barClass"
            :style="{ width: loading ? '40%' : barWidth }"
          />
        </div>
        <span class="w-16 shrink-0 text-right text-xs tabular-nums text-gray-400">
          {{ totalLabel }}
        </span>
      </template>
    </div>
  </article>
</template>
