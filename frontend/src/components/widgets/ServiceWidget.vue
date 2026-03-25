<script setup lang="ts">
import { computed, toRef } from 'vue'
import type { ServiceStatusData, WidgetSize } from '@/types'
import { STATUS_COLORS } from '@/constants/status'
import { getInitials } from '@/utils/initials'
import UptimeBar from '@/components/ui/UptimeBar.vue'
import DetailPanel from '@/components/ui/DetailPanel.vue'
import { useDetailMode } from '@/composables/useDetailMode'
import { useServiceHistory } from '@/composables/useServiceHistory'

const props = withDefaults(
  defineProps<{
    service: ServiceStatusData
    size?: WidgetSize
    storageEnabled?: boolean
  }>(),
  { size: 's', storageEnabled: false },
)

const sc = computed(() => STATUS_COLORS[props.service.status] ?? STATUS_COLORS.unknown)

const serviceDefault = computed(() => props.service.detailDefault ?? false)
const { showDetail } = useDetailMode(serviceDefault)
const { summaries, records, detailLoaded } = useServiceHistory(
  toRef(() => props.service.name),
  showDetail,
  toRef(() => props.storageEnabled),
)

const initials = computed(() => getInitials(props.service.name))

const displayUrl = computed(() => {
  if (props.service.url) {
    try {
      const { hostname, port } = new URL(props.service.url)
      return hostname + (port ? `:${port}` : '')
    } catch {
      return props.service.url
    }
  }
  if (props.service.container) return `container: ${props.service.container}`
  return null
})

const launchUrl = computed(() => props.service.href || props.service.url || null)

function fmtMs(ms?: number): string {
  if (ms == null) return ''
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(2)}s`
}

function timeAgo(iso: string): string {
  const secs = Math.floor((Date.now() - new Date(iso).getTime()) / 1000)
  if (secs <  60) return `${secs}s ago`
  if (secs < 3600) return `${Math.floor(secs / 60)}m ago`
  return `${Math.floor(secs / 3600)}h ago`
}
</script>

<template>
  <component
    :is="launchUrl ? 'a' : 'article'"
    :href="launchUrl ?? undefined"
    :target="launchUrl ? '_blank' : undefined"
    :rel="launchUrl ? 'noopener noreferrer' : undefined"
    class="widget-card group relative gap-3 p-4 transition-all duration-200 hover:border-gray-700 hover:shadow-lg"
    :class="launchUrl ? 'cursor-pointer' : 'cursor-default'"
  >
    <!-- Header -->
    <div class="flex items-start justify-between gap-3">
      <div class="flex min-w-0 items-center gap-3">
        <div
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg
                 bg-gray-800 text-xs font-semibold text-gray-300 ring-1 ring-gray-700"
        >
          {{ initials }}
        </div>
        <div class="min-w-0">
          <p class="truncate text-sm font-semibold text-gray-100">{{ service.name }}</p>
          <p v-if="displayUrl" class="mt-0.5 truncate text-xs text-gray-500">{{ displayUrl }}</p>
        </div>
      </div>

      <span
        class="inline-flex shrink-0 items-center gap-1.5 rounded-full px-2 py-0.5
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

    <!-- Uptime bar -->
    <UptimeBar v-if="summaries.length > 0" :summaries="summaries" />

    <!-- Detail panel (sparkline + stats) -->
    <DetailPanel
      v-if="showDetail && detailLoaded"
      :records="records"
      :summaries="summaries"
    />

    <!-- Footer row -->
    <div class="flex items-center justify-between gap-2">
      <div class="flex items-center gap-1.5">
        <span
          class="inline-flex items-center gap-1 rounded bg-gray-800 px-1.5 py-0.5
                 text-xs text-gray-500 ring-1 ring-gray-700"
        >
          <template v-if="service.type === 'docker'">
            <svg class="h-3 w-3" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <path d="M13.98 11.08h2.12v-2.1h-2.12v2.1zm-2.85 0h2.12v-2.1h-2.12v2.1zm-2.84
                0h2.12v-2.1H8.29v2.1zm-2.85 0h2.12v-2.1H5.44v2.1zm5.7-2.83h2.11V6.14h-2.11v2.11zm-2.85
                0h2.12V6.14H8.29v2.11zM5.44 8.25h2.12V6.14H5.44v2.11zM2.58 11.08H4.7v-2.1H2.58v2.1z"/>
              <path d="M23.27 11.27c-.47-.32-1.55-.44-2.4-.28a4.38 4.38 0 00-1.72-2.94l-.35-.23-.24.34c-.5.72-.64
                1.92-.22 2.82-.34.2-.69.4-1.01.53-.49.19-.97.28-1.44.28H.54l-.05.3C.3 13.17.5 14.38 1.2 15.27c.66.84
                1.66 1.47 2.9 1.78.63.16 1.29.24 1.96.24a10.3 10.3 0 004.58-1.1c.63-.3 1.18-.7 1.67-1.13a9.8 9.8 0
                001.38-1.7h.37c1.06 0 1.96-.34 2.63-.99.35-.34.63-.74.82-1.2l.06-.18-.3-.22z"/>
            </svg>
            Docker
          </template>
          <template v-else>
            <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2" aria-hidden="true">
              <circle cx="12" cy="12" r="10"/>
              <path d="M2 12h20M12 2a15.3 15.3 0 010 20M12 2a15.3 15.3 0 000 20"/>
            </svg>
            HTTP
          </template>
        </span>

        <span
          v-if="service.category"
          class="inline-flex items-center rounded bg-gray-800/60 px-1.5 py-0.5
                 text-[10px] font-medium uppercase tracking-wider text-gray-500 ring-1 ring-gray-700/50"
        >
          {{ service.category }}
        </span>
      </div>

      <div class="flex items-center gap-1.5 text-xs text-gray-500">
        <span v-if="service.responseTime" class="text-gray-400 tabular-nums">
          {{ fmtMs(service.responseTime) }}
        </span>
        <span v-if="service.responseTime" aria-hidden="true">&middot;</span>
        <span class="tabular-nums">{{ timeAgo(service.checkedAt) }}</span>
      </div>
    </div>

    <!-- Error / message banner -->
    <div class="min-h-[28px]">
      <p
        class="truncate rounded-md bg-gray-800/60 px-2 py-1 text-xs transition-opacity duration-150"
        :class="[
          sc.msg,
          service.message && service.status !== 'healthy' ? 'opacity-100' : 'opacity-0 invisible',
        ]"
      >
        {{ service.message || '&nbsp;' }}
      </p>
    </div>
  </component>
</template>
