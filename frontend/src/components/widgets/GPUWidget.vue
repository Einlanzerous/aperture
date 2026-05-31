<script setup lang="ts">
import { computed } from 'vue'
import { useResources } from '@/composables/useResources'
import { fmtGB } from '@/utils/format'

const WARNING_THRESHOLD  = 60
const CRITICAL_THRESHOLD = 85

const { data: resources, loading, error } = useResources(5_000)

const gpu = computed(() => resources.value?.gpu ?? null)

// gpu === null            => disabled in config
// gpu.available === false => no smi binary on PATH or probe failed
const available = computed(() => !!gpu.value && gpu.value.available)

function barColor(percent: number): string {
  if (percent < WARNING_THRESHOLD)  return 'bg-emerald-500'
  if (percent < CRITICAL_THRESHOLD) return 'bg-amber-500'
  return 'bg-red-500'
}

// Bar tracks VRAM usage, mirroring RAM: used inside, total on the right.
const vramPercent = computed(() => {
  const g = gpu.value
  if (!available.value || !g || g.vramTotal === 0) return 0
  return (g.vramUsed / g.vramTotal) * 100
})

const barClass = computed(() => (available.value ? barColor(vramPercent.value) : 'bg-gray-700'))
const barWidth = computed(() => `${Math.min(vramPercent.value, 100)}%`)

const insideText = computed(() => (available.value ? fmtGB(gpu.value!.vramUsed) : 'Not available'))
const rightText  = computed(() => (available.value ? fmtGB(gpu.value!.vramTotal) : ''))
</script>

<template>
  <!-- Tiny (1-slot) tile: GPU [chunky bar, VRAM used inside] total. The
       unavailable state keeps the exact same footprint (label + empty bar). -->
  <article class="widget-card cursor-default">
    <div class="flex h-full items-center gap-3 px-4">
      <span class="w-9 shrink-0 text-sm font-semibold text-gray-100">GPU</span>

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
                   [text-shadow:_0_1px_2px_rgb(0_0_0_/_0.6)]"
            :class="available ? 'text-white' : 'text-gray-500'"
          >
            {{ insideText }}
          </span>
        </div>
        <span class="w-14 shrink-0 text-right text-xs tabular-nums text-gray-400">
          {{ rightText || '—' }}
        </span>
      </template>
    </div>
  </article>
</template>
