<script setup lang="ts">
import { computed } from 'vue'
import { useResources } from '@/composables/useResources'

const WARNING_THRESHOLD  = 60
const CRITICAL_THRESHOLD = 85

const { data: resources, loading, error } = useResources(5_000)

const cpu = computed(() => resources.value?.cpu ?? null)

function barColor(percent: number): string {
  if (percent < WARNING_THRESHOLD)  return 'bg-emerald-400'
  if (percent < CRITICAL_THRESHOLD) return 'bg-amber-400'
  return 'bg-red-400'
}

const barClass = computed(() => (cpu.value ? barColor(cpu.value.percent) : 'bg-gray-600'))
const barWidth = computed(() => `${Math.min(cpu.value?.percent ?? 0, 100)}%`)

// Left value = logical core count; right value = current utilization.
const coresLabel = computed(() => (cpu.value ? `${cpu.value.cores}c` : '—'))
const pctLabel   = computed(() => (cpu.value ? `${cpu.value.percent.toFixed(0)}%` : '—'))
</script>

<template>
  <!-- Tiny (1-slot) tile: LABEL | cores [thick bar] percent. Fixed h-16 so the
       footprint never shifts between loading / error / live states. -->
  <article class="widget-card cursor-default">
    <div class="flex h-full items-center gap-3 px-4">
      <span class="w-10 shrink-0 text-sm font-semibold text-gray-100">CPU</span>

      <template v-if="error">
        <span class="flex-1 truncate text-xs text-red-400" :title="error">{{ error }}</span>
      </template>

      <template v-else>
        <span class="w-12 shrink-0 text-right text-xs tabular-nums text-gray-400">
          {{ coresLabel }}
        </span>
        <div class="relative h-2.5 flex-1 overflow-hidden rounded-full bg-gray-800">
          <div
            class="h-full rounded-full transition-all duration-700"
            :class="loading ? 'bg-gray-700 animate-pulse' : barClass"
            :style="{ width: loading ? '40%' : barWidth }"
          />
        </div>
        <span class="w-12 shrink-0 text-right text-xs font-medium tabular-nums text-gray-200">
          {{ pctLabel }}
        </span>
      </template>
    </div>
  </article>
</template>
