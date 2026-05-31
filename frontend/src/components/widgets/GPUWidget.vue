<script setup lang="ts">
import { computed } from 'vue'
import { useResources } from '@/composables/useResources'
import { fmtBytes } from '@/utils/format'

const WARNING_THRESHOLD  = 60
const CRITICAL_THRESHOLD = 85

const { data: resources, loading, error } = useResources(5_000)

const gpu = computed(() => resources.value?.gpu ?? null)

// gpu === null     => disabled in config
// gpu.available === false => no rocm-smi/nvidia-smi or probe failed
const available = computed(() => !!gpu.value && gpu.value.available)

function pct(n: number): string {
  return n.toFixed(1) + '%'
}

function barColor(percent: number): string {
  if (percent < WARNING_THRESHOLD)  return 'bg-emerald-400'
  if (percent < CRITICAL_THRESHOLD) return 'bg-amber-400'
  return 'bg-red-400'
}

const utilColor = computed(() => available.value ? barColor(gpu.value!.percent) : 'bg-gray-600')
const utilWidth = computed(() => `${Math.min(gpu.value?.percent ?? 0, 100)}%`)

// Vendor label, only shown when we actually know it.
const vendorLabel = computed(() => {
  const v = gpu.value?.vendor
  if (v === 'amd')    return 'AMD'
  if (v === 'nvidia') return 'NVIDIA'
  return ''
})

// Best-effort product name shown in the header subtitle.
const gpuName = computed(() => gpu.value?.name?.trim() || '')

// tempC may be null when the reading is unavailable.
const tempLabel = computed(() => {
  const t = gpu.value?.tempC
  return t == null ? '—' : `${Math.round(t)}°C`
})
</script>

<template>
  <article class="widget-card gap-4 p-4">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <svg class="h-4 w-4 text-gray-400" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="1.75" aria-hidden="true">
          <rect x="3" y="4" width="18" height="14" rx="1"/>
          <path d="M7 18v2M17 18v2M8 9h4M8 13h2"/>
        </svg>
        <h2 class="text-sm font-semibold text-gray-100">GPU</h2>
      </div>
      <span v-if="available && vendorLabel" class="text-xs text-gray-500">
        {{ vendorLabel }}
      </span>
    </div>

    <!-- Loading skeleton -->
    <template v-if="loading">
      <div class="space-y-1.5 animate-pulse">
        <div class="h-3 w-24 rounded bg-gray-800" />
        <div class="h-1.5 w-full rounded-full bg-gray-800" />
      </div>
    </template>

    <!-- Error state -->
    <p v-else-if="error" class="text-xs text-red-400">{{ error }}</p>

    <!-- Not available: gpu disabled (null) or no GPU detected (available === false) -->
    <p v-else-if="!available" class="text-xs text-gray-500">Not available</p>

    <!-- Stats -->
    <template v-else-if="gpu">
      <!-- Product name -->
      <p v-if="gpuName" class="-mt-2 truncate text-xs text-gray-400" :title="gpuName">
        {{ gpuName }}
      </p>

      <!-- Utilization -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between text-xs">
          <span class="font-medium text-gray-300">Utilization</span>
          <span class="tabular-nums text-gray-400">{{ pct(gpu.percent) }}</span>
        </div>
        <div class="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
          <div
            class="h-full rounded-full transition-all duration-700"
            :class="utilColor"
            :style="{ width: utilWidth }"
          />
        </div>
      </div>

      <!-- VRAM -->
      <div class="flex items-center justify-between text-xs">
        <span class="font-medium text-gray-300">VRAM</span>
        <span class="tabular-nums text-gray-400">
          {{ fmtBytes(gpu.vramUsed) }} / {{ fmtBytes(gpu.vramTotal) }}
        </span>
      </div>

      <!-- Temperature: hidden when tempC is null -->
      <div
        v-if="gpu.tempC != null"
        class="flex items-center justify-between border-t border-gray-800 pt-3 text-xs"
      >
        <span class="font-medium text-gray-400">Temp</span>
        <span class="tabular-nums text-gray-200">{{ tempLabel }}</span>
      </div>
    </template>
  </article>
</template>
