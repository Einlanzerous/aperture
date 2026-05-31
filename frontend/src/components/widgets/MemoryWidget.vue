<script setup lang="ts">
import { computed } from 'vue'
import { useResources } from '@/composables/useResources'
import { fmtBytes } from '@/utils/format'

const WARNING_THRESHOLD  = 60
const CRITICAL_THRESHOLD = 85

const { data: resources, loading, error } = useResources(5_000)

function pct(n: number): string {
  return n.toFixed(1) + '%'
}

function barColor(percent: number): string {
  if (percent < WARNING_THRESHOLD)  return 'bg-emerald-400'
  if (percent < CRITICAL_THRESHOLD) return 'bg-amber-400'
  return 'bg-red-400'
}

const ramColor = computed(() => resources.value?.memory ? barColor(resources.value.memory.percent) : 'bg-gray-600')
const ramWidth = computed(() => `${Math.min(resources.value?.memory?.percent ?? 0, 100)}%`)
</script>

<template>
  <article class="widget-card gap-4 p-4">

    <!-- Header -->
    <div class="flex items-center gap-2">
      <svg class="h-4 w-4 text-gray-400" viewBox="0 0 24 24" fill="none"
           stroke="currentColor" stroke-width="1.75" aria-hidden="true">
        <rect x="2" y="3" width="20" height="6" rx="1"/>
        <rect x="2" y="12" width="20" height="6" rx="1"/>
        <path d="M6 6h.01M6 15h.01"/>
      </svg>
      <h2 class="text-sm font-semibold text-gray-100">Memory</h2>
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

    <!-- Stats -->
    <div v-else-if="resources?.memory" class="space-y-1.5">
      <div class="flex items-center justify-between text-xs">
        <span class="font-medium text-gray-300">Used</span>
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

    <!-- No data -->
    <p v-else class="text-xs text-gray-500">Memory metrics unavailable</p>
  </article>
</template>
