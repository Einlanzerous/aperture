<script setup lang="ts">
import { computed } from 'vue'
import type { ServiceStatusData } from '@/types'
import { STATUS_COLORS } from '@/constants/status'

const props = defineProps<{
  service: ServiceStatusData
}>()

const sc = computed(() => STATUS_COLORS[props.service.status] ?? STATUS_COLORS.unknown)

function timeAgo(iso: string): string {
  const secs = Math.floor((Date.now() - new Date(iso).getTime()) / 1000)
  if (secs <  60) return `${secs}s ago`
  if (secs < 3600) return `${Math.floor(secs / 60)}m ago`
  return `${Math.floor(secs / 3600)}h ago`
}

// A status-only tile is exactly one slot, so it stays a single row. Any error
// message moves into the tooltip rather than adding a second line.
const tooltip = computed(() => {
  const base = `Last checked ${timeAgo(props.service.checkedAt)}`
  return props.service.status !== 'healthy' && props.service.message
    ? `${props.service.message} — ${base}`
    : base
})
</script>

<template>
  <article
    class="widget-card cursor-default transition-all duration-200
           hover:border-gray-700 hover:shadow-lg"
    :title="tooltip"
  >
    <div class="flex h-full items-center gap-3 px-4">
      <p class="min-w-0 flex-1 truncate text-sm font-semibold text-gray-100">
        {{ service.name }}
      </p>
      <span
        v-if="service.category"
        class="shrink-0 inline-flex items-center rounded bg-gray-800/60 px-1.5 py-0.5
               text-[10px] font-medium uppercase tracking-wider text-gray-500
               ring-1 ring-gray-700/50"
      >
        {{ service.category }}
      </span>
      <span
        class="shrink-0 inline-flex items-center gap-1.5 rounded-full px-2 py-0.5
               text-xs font-medium"
        :class="sc.badge"
      >
        <span
          class="h-1.5 w-1.5 rounded-full"
          :class="[sc.dot, sc.pulse ? 'animate-pulse' : '']"
        />
        {{ sc.label }}
      </span>
    </div>
  </article>
</template>
