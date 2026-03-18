<script setup lang="ts">
import { computed } from 'vue'
import type { CheckRecord } from '@/types'

const props = defineProps<{ records: CheckRecord[] }>()

const WIDTH = 200
const HEIGHT = 40
const PADDING = 2

const points = computed(() => {
  const valid = props.records.filter((r) => r.responseTime != null && r.responseTime > 0)
  if (valid.length < 2) return null

  const times = valid.map((r) => r.responseTime!)
  const min = Math.min(...times)
  const max = Math.max(...times)
  const range = max - min || 1

  return valid
    .map((_, i) => {
      const x = (i / (valid.length - 1)) * WIDTH
      const y = HEIGHT - PADDING - ((times[i] - min) / range) * (HEIGHT - PADDING * 2)
      return `${x},${y}`
    })
    .join(' ')
})

const fillPoints = computed(() => {
  if (!points.value) return null
  return `0,${HEIGHT} ${points.value} ${WIDTH},${HEIGHT}`
})
</script>

<template>
  <svg
    v-if="points"
    :viewBox="`0 0 ${WIDTH} ${HEIGHT}`"
    preserveAspectRatio="none"
    class="w-full"
    style="height: 40px"
  >
    <polygon :points="fillPoints!" class="fill-emerald-500/10" />
    <polyline
      :points="points"
      fill="none"
      class="stroke-emerald-500"
      stroke-width="1.5"
      stroke-linejoin="round"
      stroke-linecap="round"
    />
  </svg>
</template>
