<script setup lang="ts">
import { computed } from 'vue'
import { useResources } from '@/composables/useResources'

const { resources, loading, error } = useResources(5_000)

// ─── Helpers ─────────────────────────────────────────────────────────────────

function pct(n: number): string {
  return n.toFixed(1) + '%'
}

function fmtBytes(bytes: number): string {
  const gb = bytes / 1073741824
  return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / 1048576).toFixed(0)} MB`
}

/** Return a Tailwind colour class based on usage percentage. */
function barColor(percent: number): string {
  if (percent < 60)  return 'bg-emerald-400'
  if (percent < 85)  return 'bg-amber-400'
  return 'bg-red-400'
}

const cpuColor  = computed(() => resources.value ? barColor(resources.value.cpu.percent)    : 'bg-gray-600')
const ramColor  = computed(() => resources.value ? barColor(resources.value.memory.percent) : 'bg-gray-600')
const cpuWidth  = computed(() => `${Math.min(resources.value?.cpu.percent    ?? 0, 100)}%`)
const ramWidth  = computed(() => `${Math.min(resources.value?.memory.percent ?? 0, 100)}%`)
</script>

<template>
  <article class="flex flex-col gap-4 rounded-xl border border-gray-800 bg-gray-900 p-5 shadow-md">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <!-- Server icon -->
        <svg class="h-4 w-4 text-gray-400" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="1.75" aria-hidden="true">
          <rect x="2" y="3" width="20" height="6" rx="1"/>
          <rect x="2" y="12" width="20" height="6" rx="1"/>
          <path d="M6 6h.01M6 15h.01"/>
        </svg>
        <h2 class="text-sm font-semibold text-gray-100">System Resources</h2>
      </div>
      <span v-if="resources" class="text-xs text-gray-500 tabular-nums">
        {{ resources.cpu.cores }} cores
      </span>
    </div>

    <!-- Loading skeleton -->
    <template v-if="loading">
      <div v-for="i in 2" :key="i" class="space-y-1.5 animate-pulse">
        <div class="h-3 w-24 rounded bg-gray-800" />
        <div class="h-1.5 w-full rounded-full bg-gray-800" />
      </div>
    </template>

    <!-- Error state -->
    <p v-else-if="error" class="text-xs text-red-400">{{ error }}</p>

    <!-- Stats -->
    <template v-else-if="resources">
      <!-- CPU -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between text-xs">
          <span class="font-medium text-gray-300">CPU</span>
          <span class="tabular-nums text-gray-400">{{ pct(resources.cpu.percent) }}</span>
        </div>
        <div class="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
          <div
            class="h-full rounded-full transition-all duration-700"
            :class="cpuColor"
            :style="{ width: cpuWidth }"
          />
        </div>
      </div>

      <!-- Memory -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between text-xs">
          <span class="font-medium text-gray-300">Memory</span>
          <span class="tabular-nums text-gray-400">
            {{ fmtBytes(resources.memory.used) }} / {{ fmtBytes(resources.memory.total) }}
            &nbsp;({{ pct(resources.memory.percent) }})
          </span>
        </div>
        <div class="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
          <div
            class="h-full rounded-full transition-all duration-700"
            :class="ramColor"
            :style="{ width: ramWidth }"
          />
        </div>
      </div>

      <!-- Load average -->
      <div class="flex items-center gap-3 border-t border-gray-800 pt-3">
        <span class="text-xs font-medium text-gray-400">Load avg</span>
        <div class="flex gap-4 text-xs tabular-nums">
          <span>
            <span class="text-gray-500">1m </span>
            <span class="text-gray-200">{{ resources.load.load1.toFixed(2) }}</span>
          </span>
          <span>
            <span class="text-gray-500">5m </span>
            <span class="text-gray-200">{{ resources.load.load5.toFixed(2) }}</span>
          </span>
          <span>
            <span class="text-gray-500">15m </span>
            <span class="text-gray-200">{{ resources.load.load15.toFixed(2) }}</span>
          </span>
        </div>
      </div>
    </template>
  </article>
</template>
