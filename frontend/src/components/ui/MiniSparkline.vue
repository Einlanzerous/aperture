<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  values: number[]   // oldest->newest
  color?: string     // any CSS colour, default emerald-500
  fill?:  boolean    // stretch to the parent's height instead of a fixed 40px
}>()

const WIDTH = 200
const HEIGHT = 40
const PADDING = 2

const stroke = computed(() => props.color ?? '#10b981')  // emerald-500

const points = computed(() => {
  const valid = props.values.filter((v) => v != null && Number.isFinite(v))
  if (valid.length < 2) return null

  const min = Math.min(...valid)
  const max = Math.max(...valid)
  const range = max - min || 1

  return valid
    .map((v, i) => {
      const x = (i / (valid.length - 1)) * WIDTH
      const y = HEIGHT - PADDING - ((v - min) / range) * (HEIGHT - PADDING * 2)
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
    :class="fill ? 'h-full w-full' : 'w-full'"
    :style="fill ? undefined : 'height: 40px'"
  >
    <polygon :points="fillPoints!" :fill="stroke" fill-opacity="0.1" />
    <polyline
      :points="points"
      fill="none"
      :stroke="stroke"
      stroke-width="1.5"
      stroke-linejoin="round"
      stroke-linecap="round"
    />
  </svg>
</template>
