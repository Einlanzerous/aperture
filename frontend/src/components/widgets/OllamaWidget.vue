<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { OllamaModel } from '@/types'

// ─── State ────────────────────────────────────────────────────────────────────

const models  = ref<OllamaModel[]>([])
const loading = ref(true)
const error   = ref<string | null>(null)

let timerId: ReturnType<typeof setInterval> | null = null

async function fetchModels() {
  try {
    const res = await fetch('/api/ollama/models')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    models.value = data.models ?? []
    error.value  = null
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Unknown error'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchModels()
  timerId = setInterval(fetchModels, 60_000)
})
onUnmounted(() => { if (timerId !== null) clearInterval(timerId) })

// ─── Derived ─────────────────────────────────────────────────────────────────

const status = computed(() => {
  if (loading.value) return 'loading'
  if (error.value)   return 'error'
  return 'ok'
})

/** Show first 5; user can toggle to see all. */
const expanded   = ref(false)
const visible    = computed(() => expanded.value ? models.value : models.value.slice(0, 5))
const hasMore    = computed(() => models.value.length > 5)
const hiddenCount = computed(() => models.value.length - 5)

function fmtSize(bytes: number): string {
  const gb = bytes / 1073741824
  return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / 1048576).toFixed(0)} MB`
}

</script>

<template>
  <article class="flex flex-col gap-4 rounded-xl border border-gray-800 bg-gray-900 p-5 shadow-md">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <!-- Brain/AI icon -->
        <svg class="h-4 w-4 text-violet-400" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="1.75" aria-hidden="true">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z"/>
          <path d="M8 14s1.5 2 4 2 4-2 4-2"/>
          <line x1="9" y1="9" x2="9.01" y2="9"/>
          <line x1="15" y1="9" x2="15.01" y2="9"/>
        </svg>
        <h2 class="text-sm font-semibold text-gray-100">Ollama</h2>
      </div>

      <!-- Connection status -->
      <span
        class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
        :class="{
          'bg-emerald-400/10 text-emerald-400 ring-1 ring-emerald-400/20': status === 'ok',
          'bg-red-400/10    text-red-400    ring-1 ring-red-400/20':    status === 'error',
          'bg-gray-500/10   text-gray-400   ring-1 ring-gray-500/20':   status === 'loading',
        }"
      >
        <span
          class="h-1.5 w-1.5 rounded-full"
          :class="{
            'bg-emerald-400 animate-pulse': status === 'ok',
            'bg-red-400':                   status === 'error',
            'bg-gray-500':                  status === 'loading',
          }"
        />
        <template v-if="status === 'ok'">{{ models.length }} model{{ models.length !== 1 ? 's' : '' }}</template>
        <template v-else-if="status === 'error'">Unreachable</template>
        <template v-else>Checking…</template>
      </span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-2 animate-pulse">
      <div v-for="i in 3" :key="i"
           class="flex items-center justify-between rounded-lg bg-gray-800 px-3 py-2">
        <div class="h-3 w-32 rounded bg-gray-700" />
        <div class="h-3 w-12 rounded bg-gray-700" />
      </div>
    </div>

    <!-- Error -->
    <p v-else-if="error" class="text-xs text-red-400">{{ error }}</p>

    <!-- Empty -->
    <p v-else-if="models.length === 0" class="text-xs text-gray-500">No models pulled yet.</p>

    <!-- Model list -->
    <ul v-else class="space-y-1.5">
      <li
        v-for="model in visible"
        :key="model.name"
        class="flex items-center justify-between rounded-lg bg-gray-800/60 px-3 py-2
               text-xs transition-colors hover:bg-gray-800"
      >
        <!-- Name + family -->
        <div class="flex items-center gap-2 min-w-0">
          <span class="truncate font-medium text-gray-200">{{ model.name }}</span>
          <span
            v-if="model.details?.parameter_size"
            class="shrink-0 rounded bg-violet-400/10 px-1.5 py-0.5
                   text-[10px] font-medium text-violet-400 ring-1 ring-violet-400/20"
          >
            {{ model.details.parameter_size }}
          </span>
        </div>

        <!-- Right: quantisation + size -->
        <div class="ml-4 flex shrink-0 items-center gap-2 text-gray-500">
          <span v-if="model.details?.quantization_level" class="tabular-nums">
            {{ model.details.quantization_level }}
          </span>
          <span class="tabular-nums text-gray-400">{{ fmtSize(model.size) }}</span>
        </div>
      </li>
    </ul>

    <!-- Expand / collapse -->
    <button
      v-if="hasMore"
      class="text-xs text-gray-500 hover:text-gray-300 transition-colors text-left"
      @click="expanded = !expanded"
    >
      {{ expanded ? '↑ Show fewer' : `+ ${hiddenCount} more model${hiddenCount !== 1 ? 's' : ''}` }}
    </button>
  </article>
</template>
