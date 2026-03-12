<script setup lang="ts">
import { computed, onMounted } from 'vue'
import DashboardGrid, { widgetSizeClass } from '@/components/layout/DashboardGrid.vue'
import ServiceWidget  from '@/components/widgets/ServiceWidget.vue'
import OllamaWidget   from '@/components/widgets/OllamaWidget.vue'
import ResourceWidget from '@/components/widgets/ResourceWidget.vue'
import SkeletonCard   from '@/components/ui/SkeletonCard.vue'
import { useConfig }   from '@/composables/useConfig'
import { useServices } from '@/composables/useServices'

// ─── Dashboard config (fetched once on mount) ─────────────────────────────────

const { config, load: loadConfig } = useConfig()
onMounted(loadConfig)

// ─── Services (interval reacts to config changes) ────────────────────────────

const serviceInterval = computed(() => (config.value.checkInterval || 30) * 1000)
const { services, loading, lastUpdated, refresh } = useServices(serviceInterval)

// ─── Helpers ─────────────────────────────────────────────────────────────────

function fmtTime(d: Date | null): string {
  if (!d) return '—'
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<template>
  <div class="min-h-screen bg-gray-950 font-sans text-gray-100">

    <!-- Top bar -->
    <header class="sticky top-0 z-10 border-b border-gray-800 bg-gray-950/80 backdrop-blur">
      <div class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6">
        <div class="flex items-center gap-2.5">
          <img src="/aperture_logo.png" alt="Aperture" class="h-6 w-6" />
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

    <!-- Main content -->
    <main class="mx-auto max-w-7xl px-4 py-6 sm:px-6">
      <DashboardGrid>

        <!-- System Resources — full width if enabled -->
        <div v-if="config.systemEnabled" :class="widgetSizeClass('l')">
          <ResourceWidget />
        </div>

        <!-- Service skeletons while loading -->
        <template v-if="loading">
          <SkeletonCard :count="6" />
        </template>

        <div
          v-for="service in services"
          :key="service.name"
          :class="widgetSizeClass(service.size ?? 's')"
        >
          <ServiceWidget :service="service" :size="service.size ?? 's'" />
        </div>

        <!-- Ollama widget — medium width if enabled -->
        <div v-if="config.ollamaEnabled" :class="widgetSizeClass('m')">
          <OllamaWidget />
        </div>

      </DashboardGrid>
    </main>
  </div>
</template>
