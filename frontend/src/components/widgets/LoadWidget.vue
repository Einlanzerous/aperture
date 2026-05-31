<script setup lang="ts">
import { computed, ref } from 'vue'
import { useResources } from '@/composables/useResources'
import MiniSparkline from '@/components/ui/MiniSparkline.vue'

// History windows, expressed in samples at the 5s sampler cadence. The backend
// ring buffer holds 3h (2160 samples), so the longest window reads it whole.
const WINDOWS = [
  { label: '20m', samples: 240 },
  { label: '60m', samples: 720 },
  { label: '3h',  samples: 2160 },
] as const

const windowSamples = ref<number>(WINDOWS[0].samples)

// The requested history count is reactive, so changing the window re-points the
// polling URL; refresh() pulls the new window immediately rather than waiting
// for the next tick.
const { data: resources, loading, error, refresh } = useResources(5_000, windowSamples)

function onWindowChange(e: Event): void {
  windowSamples.value = Number((e.target as HTMLSelectElement).value)
  refresh()
}

const activeLabel = computed(
  () => WINDOWS.find(w => w.samples === windowSamples.value)?.label ?? '',
)

// load1 history, oldest->newest; empty until at least two samples land.
const load1History = computed(() => resources.value?.history?.load1 ?? [])
const hasSparkline = computed(() => load1History.value.length >= 2)
</script>

<template>
  <article class="widget-card gap-3 p-4">

    <!-- Header: title + history-window selector -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <svg class="h-4 w-4 text-gray-400" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="1.75" aria-hidden="true">
          <path d="M3 3v18h18"/>
          <path d="M19 9l-5 5-4-4-3 3"/>
        </svg>
        <h2 class="text-sm font-semibold text-gray-100">Load Average</h2>
      </div>
      <select
        v-if="resources?.load"
        class="rounded border border-gray-700 bg-gray-800 px-1.5 py-0.5 text-xs
               text-gray-300 transition-colors hover:border-gray-600
               focus:outline-none focus:ring-1 focus:ring-emerald-500/40"
        :value="windowSamples"
        aria-label="History window"
        @change="onWindowChange"
      >
        <option v-for="w in WINDOWS" :key="w.samples" :value="w.samples">{{ w.label }}</option>
      </select>
    </div>

    <!-- Loading skeleton -->
    <template v-if="loading">
      <div class="space-y-2 animate-pulse">
        <div class="h-3 w-32 rounded bg-gray-800" />
        <div class="h-10 w-full rounded bg-gray-800" />
      </div>
    </template>

    <!-- Error state -->
    <p v-else-if="error" class="text-xs text-red-400">{{ error }}</p>

    <!-- Graph-first: 1m/5m/15m numbers above the always-on sparkline -->
    <template v-else-if="resources?.load">
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

      <div>
        <p class="mb-1.5 text-xs font-medium text-gray-400">1m load &middot; last {{ activeLabel }}</p>
        <MiniSparkline v-if="hasSparkline" :values="load1History" />
        <p v-else class="text-xs text-gray-600">Collecting history…</p>
      </div>
    </template>

    <!-- Load unavailable (disabled in config / null) -->
    <p v-else class="text-xs text-gray-600">Load average unavailable.</p>
  </article>
</template>
