<script setup lang="ts">
import { computed, ref, onUnmounted } from 'vue'
import type { ActionState, ActionStatus } from '@/types'
import { ACTION_STATUS_COLORS } from '@/constants/action'
import { apiPost, apiFetch, API } from '@/utils/api'

const props = defineProps<{ action: ActionState }>()

const localStatus = ref<ActionStatus>(props.action.taskStatus)
const polling = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null
let resetTimer: ReturnType<typeof setTimeout> | null = null

const sc = computed(() => ACTION_STATUS_COLORS[localStatus.value] ?? ACTION_STATUS_COLORS.idle)

const initials = computed(() =>
  props.action.name
    .split(/[\s\-_]+/)
    .map((w) => w[0] ?? '')
    .join('')
    .toUpperCase()
    .slice(0, 2),
)

const isTerminal = (s: ActionStatus) => s === 'success' || s === 'error' || s === 'stopped'
const isInFlight = computed(() => !isTerminal(localStatus.value) && localStatus.value !== 'idle')

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  polling.value = false
}

function stopReset() {
  if (resetTimer) { clearTimeout(resetTimer); resetTimer = null }
}

async function pollStatus() {
  try {
    const state = await apiFetch<ActionState>(API.actionStatus(props.action.name))
    localStatus.value = state.taskStatus
    if (isTerminal(state.taskStatus)) {
      stopPolling()
      resetTimer = setTimeout(() => { localStatus.value = 'idle' }, 5000)
    }
  } catch {
    stopPolling()
  }
}

async function trigger() {
  if (isInFlight.value) return
  stopReset()

  try {
    const state = await apiPost<ActionState>(API.actionTrigger(props.action.name))
    localStatus.value = state.taskStatus
    polling.value = true
    pollTimer = setInterval(pollStatus, 3000)
  } catch {
    localStatus.value = 'error'
    resetTimer = setTimeout(() => { localStatus.value = 'idle' }, 5000)
  }
}

onUnmounted(() => { stopPolling(); stopReset() })
</script>

<template>
  <article class="widget-card group relative gap-3 p-4 transition-all duration-200 hover:border-gray-700 hover:shadow-lg cursor-default">
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
          <p class="truncate text-sm font-semibold text-gray-100">{{ action.name }}</p>
          <p class="mt-0.5 truncate text-xs text-gray-500">Semaphore Action</p>
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

    <!-- Footer row -->
    <div class="flex items-center justify-between gap-2">
      <div class="flex items-center gap-1.5">
        <span
          class="inline-flex items-center gap-1 rounded bg-gray-800 px-1.5 py-0.5
                 text-xs text-gray-500 ring-1 ring-gray-700"
        >
          <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" aria-hidden="true">
            <path d="M14.7 6.3a1 1 0 000 1.4l1.6 1.6a1 1 0 001.4 0l3.77-3.77a6 6 0 01-7.94 7.94l-6.91
              6.91a2.12 2.12 0 01-3-3l6.91-6.91a6 6 0 017.94-7.94l-3.76 3.76z"/>
          </svg>
          Action
        </span>

        <span
          v-if="action.category"
          class="inline-flex items-center rounded bg-gray-800/60 px-1.5 py-0.5
                 text-[10px] font-medium uppercase tracking-wider text-gray-500 ring-1 ring-gray-700/50"
        >
          {{ action.category }}
        </span>
      </div>

      <button
        :disabled="isInFlight"
        class="rounded-md border px-2.5 py-1 text-xs font-medium transition-colors"
        :class="isInFlight
          ? 'border-gray-700 bg-gray-800 text-gray-500 cursor-not-allowed'
          : 'border-gray-700 bg-gray-800 text-gray-300 hover:border-gray-600 hover:bg-gray-700 hover:text-gray-100'"
        @click="trigger"
      >
        {{ isInFlight ? 'Running...' : 'Run' }}
      </button>
    </div>
  </article>
</template>
