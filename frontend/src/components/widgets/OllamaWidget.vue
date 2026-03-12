<script setup lang="ts">
import { ref, computed } from 'vue'
import { useOllama } from '@/composables/useOllama'
import { fmtBytes } from '@/utils/format'

const MAX_VISIBLE_MODELS = 5

const { models, loading, error } = useOllama(60_000)

const status = computed(() => {
  if (loading.value) return 'loading'
  if (error.value)   return 'error'
  return 'ok'
})

const expanded    = ref(false)
const visible     = computed(() => expanded.value ? models.value : models.value.slice(0, MAX_VISIBLE_MODELS))
const hasMore     = computed(() => models.value.length > MAX_VISIBLE_MODELS)
const hiddenCount = computed(() => models.value.length - MAX_VISIBLE_MODELS)
</script>

<template>
  <article class="widget-card gap-4 p-4">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <svg class="h-4 w-4 text-violet-400" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="1.75" aria-hidden="true">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z"/>
          <path d="M8 14s1.5 2 4 2 4-2 4-2"/>
          <line x1="9" y1="9" x2="9.01" y2="9"/>
          <line x1="15" y1="9" x2="15.01" y2="9"/>
        </svg>
        <h2 class="text-sm font-semibold text-gray-100">Ollama</h2>
      </div>

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

        <div class="ml-4 flex shrink-0 items-center gap-2 text-gray-500">
          <span v-if="model.details?.quantization_level" class="tabular-nums">
            {{ model.details.quantization_level }}
          </span>
          <span class="tabular-nums text-gray-400">{{ fmtBytes(model.size) }}</span>
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
