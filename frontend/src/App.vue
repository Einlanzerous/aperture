<script setup lang="ts">
import { computed, onMounted, type Component } from 'vue'
import DraggableGrid from '@/components/layout/DraggableGrid.vue'
import ServiceWidget  from '@/components/widgets/ServiceWidget.vue'
import StatusStack    from '@/components/widgets/StatusStack.vue'
import ActionWidget   from '@/components/widgets/ActionWidget.vue'
import OllamaWidget   from '@/components/widgets/OllamaWidget.vue'
import CPUWidget      from '@/components/widgets/CPUWidget.vue'
import MemoryWidget   from '@/components/widgets/MemoryWidget.vue'
import LoadWidget     from '@/components/widgets/LoadWidget.vue'
import GPUWidget      from '@/components/widgets/GPUWidget.vue'
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

  // System metrics render as individual S tiles, each gated on its own config
  // flag, so the dashboard only shows the metrics the backend is collecting.
  if (config.value.system.cpu) {
    list.push({
      id:        'system:cpu',
      kind:      'resource',
      size:      's',
      component: CPUWidget,
      props:     {},
    })
  }

  if (config.value.system.memory) {
    list.push({
      id:        'system:memory',
      kind:      'resource',
      size:      's',
      component: MemoryWidget,
      props:     {},
    })
  }

  if (config.value.system.load) {
    list.push({
      id:        'system:load',
      kind:      'resource',
      size:      's',
      component: LoadWidget,
      props:     {},
    })
  }

  if (config.value.system.gpu) {
    list.push({
      id:        'system:gpu',
      kind:      'resource',
      size:      's',
      component: GPUWidget,
      props:     {},
    })
  }

  // Normal services render one-per-slot. Status-only services are collected and
  // paired two-per-slot into a StatusStack, so a thin tile no longer eats a full
  // S slot. The backend sorts services alphabetically, so status-only services
  // arrive scattered — pairing them here keeps the stacks contiguous regardless.
  const statusOnly: typeof services.value = []

  for (const service of services.value) {
    if (service.statusOnly) {
      statusOnly.push(service)
      continue
    }
    const size = service.size ?? 's'
    list.push({
      id:        `service:${service.name}`,
      kind:      'service',
      size,
      component: ServiceWidget,
      props:     { service, size, storageEnabled: config.value.storageEnabled },
    })
  }

  for (let i = 0; i < statusOnly.length; i += 2) {
    const pair = statusOnly.slice(i, i + 2)
    list.push({
      id:        `status-pair:${pair[0].name}`,
      kind:      'service',
      size:      's',
      component: StatusStack,
      props:     { services: pair },
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

// Whether any system metric tile is present — drives where loading skeletons sit.
const systemWidgetIds = computed(() =>
  widgets.value.filter(w => w.kind === 'resource').map(w => w.id),
)
const hasSystemWidgets = computed(() => systemWidgetIds.value.length > 0)
const lastSystemWidgetId = computed(() =>
  systemWidgetIds.value[systemWidgetIds.value.length - 1] ?? null,
)

// ─── Layout (persisted order + size overrides) ───────────────────────────────

const dashboardTitle = computed(() => config.value.title)
const { applyLayout, setOrder, reset: resetLayout, isCustomized } = useLayout(dashboardTitle)
const orderedWidgets = computed(() => applyLayout(widgets.value))

function onResetLayout(): void {
  if (isCustomized.value && !window.confirm('Reset dashboard layout to defaults?')) return
  resetLayout()
}

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
                   text-gray-300 transition-colors hover:border-gray-600 hover:bg-gray-700
                   disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:border-gray-700
                   disabled:hover:bg-gray-800"
            :disabled="!isCustomized"
            :title="isCustomized ? 'Reset layout to defaults' : 'Layout is already at defaults'"
            @click="onResetLayout"
          >
            Reset layout
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
        <template v-if="loading && !hasSystemWidgets" #before>
          <SkeletonCard :count="6" />
        </template>

        <template #default="{ item }">
          <component :is="item.component" v-bind="item.props" />
        </template>

        <!-- Skeleton bridges system → services while loading -->
        <template #after-item="{ item }">
          <SkeletonCard v-if="loading && item.id === lastSystemWidgetId" :count="6" />
        </template>

      </DraggableGrid>
    </main>
  </div>
</template>
