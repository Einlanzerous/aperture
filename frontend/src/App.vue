<script setup lang="ts">
import { computed, onMounted, type Component } from 'vue'
import DraggableGrid from '@/components/layout/DraggableGrid.vue'
import ServiceWidget  from '@/components/widgets/ServiceWidget.vue'
import ActionWidget   from '@/components/widgets/ActionWidget.vue'
import OllamaWidget   from '@/components/widgets/OllamaWidget.vue'
import ResourceWidget from '@/components/widgets/ResourceWidget.vue'
import SkeletonCard   from '@/components/ui/SkeletonCard.vue'
import type { WidgetSize } from '@/types'
import { useConfig }     from '@/composables/useConfig'
import { useServices }   from '@/composables/useServices'
import { useActions }    from '@/composables/useActions'
import { useDetailMode } from '@/composables/useDetailMode'
import { useLayout }     from '@/composables/useLayout'

// ─── Dashboard config (fetched once on mount) ─────────────────────────────────

const { config, load: loadConfig } = useConfig()
onMounted(loadConfig)

// ─── Services (interval reacts to config changes) ────────────────────────────

const serviceInterval = computed(() => (config.value.checkInterval || 30) * 1000)
const { services, loading, lastUpdated, refresh } = useServices(serviceInterval)

// ─── Actions (interval reacts to config changes) ──────────────────────────

const { actions } = useActions(serviceInterval)

// ─── Detail mode toggle ─────────────────────────────────────────────────────

const { isDetailMode, toggleGlobal } = useDetailMode()

// ─── Widget registry ─────────────────────────────────────────────────────────

type WidgetKind = 'resource' | 'service' | 'action' | 'ollama'

interface Widget {
  id:        string
  kind:      WidgetKind
  size:      WidgetSize
  component: Component
  props:     Record<string, unknown>
}

const widgets = computed<Widget[]>(() => {
  const list: Widget[] = []

  if (config.value.systemEnabled) {
    list.push({
      id:        'system',
      kind:      'resource',
      size:      'l',
      component: ResourceWidget,
      props:     {},
    })
  }

  for (const service of services.value) {
    const size = service.size ?? 's'
    list.push({
      id:        `service:${service.name}`,
      kind:      'service',
      size,
      component: ServiceWidget,
      props:     { service, size, storageEnabled: config.value.storageEnabled },
    })
  }

  for (const action of actions.value) {
    list.push({
      id:        `action:${action.name}`,
      kind:      'action',
      size:      action.size ?? 's',
      component: ActionWidget,
      props:     { action },
    })
  }

  if (config.value.ollamaEnabled) {
    list.push({
      id:        'ollama',
      kind:      'ollama',
      size:      'm',
      component: OllamaWidget,
      props:     {},
    })
  }

  return list
})

// ─── Layout (persisted order + size overrides) ───────────────────────────────

const dashboardTitle = computed(() => config.value.title)
const { applyLayout, setOrder } = useLayout(dashboardTitle)
const orderedWidgets = computed(() => applyLayout(widgets.value))

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
            v-if="config.storageEnabled"
            class="rounded-md border border-gray-700 bg-gray-800 p-1.5
                   transition-colors hover:border-gray-600 hover:bg-gray-700"
            :class="isDetailMode ? 'text-emerald-400' : 'text-gray-300'"
            title="Toggle detailed history"
            @click="toggleGlobal"
          >
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
            </svg>
          </button>
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
      <DraggableGrid :items="orderedWidgets" @reorder="setOrder">

        <!-- Service skeletons while loading (rendered at top when no system widget precedes) -->
        <template v-if="loading && !config.systemEnabled" #before>
          <SkeletonCard :count="6" />
        </template>

        <template #default="{ item }">
          <component :is="item.component" v-bind="item.props" />
        </template>

        <!-- Skeleton bridges system → services while loading -->
        <template #after-item="{ item }">
          <SkeletonCard v-if="loading && item.kind === 'resource'" :count="6" />
        </template>

      </DraggableGrid>
    </main>
  </div>
</template>
