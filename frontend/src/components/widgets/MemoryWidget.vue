<script setup lang="ts">
import { computed } from 'vue'
import { useResources } from '@/composables/useResources'
import { fmtGB } from '@/utils/format'

const WARNING_THRESHOLD  = 60
const CRITICAL_THRESHOLD = 85

const { data: resources, loading, error } = useResources(5_000)

const mem = computed(() => resources.value?.memory ?? null)

function barColor(percent: number): string {
  if (percent < WARNING_THRESHOLD)  return 'bg-emerald-500'
  if (percent < CRITICAL_THRESHOLD) return 'bg-amber-500'
  return 'bg-red-500'
}

const barClass = computed(() => (mem.value ? barColor(mem.value.percent) : 'bg-gray-700'))
const barWidth = computed(() => `${Math.min(mem.value?.percent ?? 0, 100)}%`)

// Inside the bar: currently used. Right of the bar: total (the max).
const insideText = computed(() => (mem.value ? fmtGB(mem.value.used) : '—'))
const rightText  = computed(() => (mem.value ? fmtGB(mem.value.total) : '—'))
</script>

<template>
  <!-- Tiny (1-slot) tile: RAM [chunky bar, used inside] total. -->
  <article class="widget-card cursor-default">
    <div class="flex h-full items-center gap-3 px-4">
      <span class="w-9 shrink-0 text-sm font-semibold text-gray-100">RAM</span>

      <template v-if="error">
        <span class="flex-1 truncate text-xs text-red-400" :title="error">{{ error }}</span>
      </template>

      <template v-else>
        <div class="relative h-7 flex-1 overflow-hidden rounded-md bg-gray-800">
          <div
            class="h-full rounded-md transition-all duration-700"
            :class="loading ? 'bg-gray-700 animate-pulse' : barClass"
            :style="{ width: loading ? '40%' : barWidth }"
          />
          <span
            class="absolute inset-0 flex items-center px-2 text-[11px] font-semibold
                   text-white [text-shadow:_0_1px_2px_rgb(0_0_0_/_0.6)]"
          >
            {{ insideText }}
          </span>
        </div>
        <span class="w-16 shrink-0 text-right text-xs tabular-nums text-gray-400">
          {{ rightText }}
        </span>
      </template>
    </div>
  </article>
</template>
