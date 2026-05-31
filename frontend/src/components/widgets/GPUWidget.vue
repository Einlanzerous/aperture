<script setup lang="ts">
import { computed } from 'vue'
import { useResources } from '@/composables/useResources'

const WARNING_THRESHOLD  = 60
const CRITICAL_THRESHOLD = 85
const BYTES_PER_GB       = 1024 ** 3

const { data: resources, loading, error } = useResources(5_000)

const gpu = computed(() => resources.value?.gpu ?? null)

// gpu === null            => disabled in config
// gpu.available === false => no smi binary on PATH or probe failed
const available = computed(() => !!gpu.value && gpu.value.available)

function barColor(percent: number): string {
  if (percent < WARNING_THRESHOLD)  return 'bg-emerald-400'
  if (percent < CRITICAL_THRESHOLD) return 'bg-amber-400'
  return 'bg-red-400'
}

const barClass = computed(() => (available.value ? barColor(gpu.value!.percent) : 'bg-gray-600'))
const barWidth = computed(() => `${Math.min(gpu.value?.percent ?? 0, 100)}%`)

const pctLabel = computed(() => (available.value ? `${gpu.value!.percent.toFixed(0)}%` : '—'))

// Compact VRAM, whole GiB to fit the tiny row: "18/32 GB".
const vramLabel = computed(() => {
  if (!available.value) return '—'
  const used  = Math.round(gpu.value!.vramUsed / BYTES_PER_GB)
  const total = Math.round(gpu.value!.vramTotal / BYTES_PER_GB)
  return `${used}/${total} GB`
})
</script>

<template>
  <!-- Tiny (1-slot) tile: GPU | util [thick bar] VRAM. The "Not available" state
       keeps the exact same h-16 footprint so the grid never shifts. -->
  <article class="widget-card cursor-default">
    <div class="flex h-16 items-center gap-3 px-4">
      <span class="w-10 shrink-0 text-sm font-semibold text-gray-100">GPU</span>

      <template v-if="error">
        <span class="flex-1 truncate text-xs text-red-400" :title="error">{{ error }}</span>
      </template>

      <!-- Disabled / no adapter: muted track + label, identical footprint. -->
      <template v-else-if="!loading && !available">
        <div class="h-2.5 flex-1 overflow-hidden rounded-full bg-gray-800" />
        <span class="shrink-0 text-xs text-gray-500">Not available</span>
      </template>

      <template v-else>
        <span class="w-12 shrink-0 text-right text-xs tabular-nums text-gray-200">
          {{ pctLabel }}
        </span>
        <div class="relative h-2.5 flex-1 overflow-hidden rounded-full bg-gray-800">
          <div
            class="h-full rounded-full transition-all duration-700"
            :class="loading ? 'bg-gray-700 animate-pulse' : barClass"
            :style="{ width: loading ? '40%' : barWidth }"
          />
        </div>
        <span class="w-16 shrink-0 text-right text-xs tabular-nums text-gray-400">
          {{ vramLabel }}
        </span>
      </template>
    </div>
  </article>
</template>
