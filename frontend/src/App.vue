<script setup lang="ts">
import { ref, onMounted } from 'vue'
import DashboardGrid, { widgetSizeClass } from '@/components/layout/DashboardGrid.vue'
import ServiceWidget  from '@/components/widgets/ServiceWidget.vue'
import OllamaWidget   from '@/components/widgets/OllamaWidget.vue'
import ResourceWidget from '@/components/widgets/ResourceWidget.vue'
import { useServices } from '@/composables/useServices'
import type { DashboardConfig, WidgetSize } from '@/types'

// ─── Dashboard config (fetched once on mount) ─────────────────────────────────

const config = ref<DashboardConfig>({
  title:         'Aperture',
  checkInterval: 30,
  ollamaEnabled: false,
  systemEnabled: false,
})

onMounted(async () => {
  try {
    const res = await fetch('/api/config')
    if (res.ok) config.value = await res.json()
  } catch { /* keep defaults */ }
})

// ─── Services ─────────────────────────────────────────────────────────────────

const { services, loading, lastUpdated, refresh } = useServices(
  // Multiply by 1000 to convert seconds → ms; fall back to 30 s.
  (config.value.checkInterval || 30) * 1000,
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

function fmtTime(d: Date | null): string {
  if (!d) return '—'
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<template>
  <div class="min-h-screen bg-gray-950 font-sans text-gray-100">

    <!-- ── Top bar ── -->
    <header class="sticky top-0 z-10 border-b border-gray-800 bg-gray-950/80 backdrop-blur">
      <div class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6">
        <div class="flex items-center gap-2.5">
          <!-- Logo mark -->
          <svg class="h-5 w-5 text-indigo-400" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="1.75" aria-hidden="true">
            <circle cx="12" cy="12" r="10"/>
            <circle cx="12" cy="12" r="4"/>
            <line x1="12" y1="2"  x2="12" y2="6"/>
            <line x1="12" y1="18" x2="12" y2="22"/>
            <line x1="2"  y1="12" x2="6"  y2="12"/>
            <line x1="18" y1="12" x2="22" y2="12"/>
          </svg>
          <span class="text-base font-semibold tracking-tight text-gray-100">
            {{ config.title }}
          </span>
        </div>

        <div class="flex items-center gap-3 text-xs text-gray-500">
          <span v-if="lastUpdated">Updated {{ fmtTime(lastUpdated) }}</span>
          <button
            class="rounded-md border border-gray-700 bg-gray-800 px-2.5 py-1
                   text-gray-300 transition-colors hover:border-gray-600 hover:bg-gray-700"
            @click="refresh"
          >
            Refresh
          </button>
        </div>
      </div>
    </header>

    <!-- ── Main content ── -->
    <main class="mx-auto max-w-7xl px-4 py-6 sm:px-6">
      <DashboardGrid>

        <!-- System Resources — full width if enabled -->
        <div v-if="config.systemEnabled" :class="widgetSizeClass('l')">
          <ResourceWidget />
        </div>

        <!-- Service widgets — sized from the config returned by the API -->
        <template v-if="loading">
          <div
            v-for="i in 6"
            :key="`skel-${i}`"
            class="col-span-1 h-28 animate-pulse rounded-xl border border-gray-800 bg-gray-900"
          />
        </template>

        <div
          v-for="service in services"
          :key="service.name"
          :class="widgetSizeClass((service.size as WidgetSize) ?? 's')"
        >
          <ServiceWidget :service="service" :size="(service.size as WidgetSize) ?? 's'" />
        </div>

        <!-- Ollama widget — medium width if enabled -->
        <div v-if="config.ollamaEnabled" :class="widgetSizeClass('m')">
          <OllamaWidget />
        </div>

      </DashboardGrid>
    </main>
  </div>
</template>
