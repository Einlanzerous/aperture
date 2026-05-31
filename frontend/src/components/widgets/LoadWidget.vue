<script setup lang="ts">
import { computed, ref } from 'vue'
import { useResources } from '@/composables/useResources'
import MiniSparkline from '@/components/ui/MiniSparkline.vue'

// ~20 min of history at the 5s poll cadence (20 * 60 / 5 = 240 samples).
const HISTORY_SAMPLES = 240

const { data: resources, loading, error } = useResources(5_000, HISTORY_SAMPLES)

const expanded = ref(false)

// load1 history, oldest->newest; null/empty when history is absent or too short.
const load1History = computed(() => resources.value?.history?.load1 ?? [])
const hasSparkline = computed(() => load1History.value.length >= 2)
</script>

<template>
  <article class="widget-card gap-4 p-4">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <svg class="h-4 w-4 text-gray-400" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="1.75" aria-hidden="true">
          <path d="M3 3v18h18"/>
          <path d="M19 9l-5 5-4-4-3 3"/>
        </svg>
        <h2 class="text-sm font-semibold text-gray-100">Load Average</h2>
      </div>
      <button
        v-if="resources?.load"
        type="button"
        class="rounded p-0.5 text-gray-500 transition-colors hover:text-gray-300"
        :aria-expanded="expanded"
        :aria-label="expanded ? 'Hide history' : 'Show history'"
        @click="expanded = !expanded"
      >
        <svg
          class="h-4 w-4 transition-transform duration-200"
          :class="expanded ? 'rotate-180' : ''"
          viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="1.75" aria-hidden="true"
        >
          <path d="M6 9l6 6 6-6"/>
        </svg>
      </button>
    </div>

    <!-- Loading skeleton -->
    <template v-if="loading">
      <div class="space-y-1.5 animate-pulse">
        <div class="h-3 w-32 rounded bg-gray-800" />
      </div>
    </template>

    <!-- Error state -->
    <p v-else-if="error" class="text-xs text-red-400">{{ error }}</p>

    <!-- Stats -->
    <template v-else-if="resources?.load">
      <!-- Collapsed: 1m/5m/15m numbers -->
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

      <!-- Expanded: load1 sparkline over the last ~20 min -->
      <div v-if="expanded" class="border-t border-gray-800 pt-3">
        <p class="mb-1.5 text-xs font-medium text-gray-400">1m load &middot; last ~20 min</p>
        <MiniSparkline v-if="hasSparkline" :values="load1History" />
        <p v-else class="text-xs text-gray-600">Not enough history yet.</p>
      </div>
    </template>

    <!-- Load unavailable (disabled in config / null) -->
    <p v-else class="text-xs text-gray-600">Load average unavailable.</p>
  </article>
</template>
