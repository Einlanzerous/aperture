<script setup lang="ts">
import { computed } from 'vue'
import type { DailySummary } from '@/types'

const props = defineProps<{ summaries: DailySummary[] }>()

interface Segment {
  color: string
  tooltip: string
}

const segments = computed<Segment[]>(() => {
  // Pad to 30 days, oldest first
  const padded: (DailySummary | null)[] = Array.from({ length: 30 }, (_, i) => {
    const d = new Date()
    d.setDate(d.getDate() - (29 - i))
    const dateStr = d.toISOString().slice(0, 10)
    return props.summaries.find((s) => s.date.slice(0, 10) === dateStr) ?? null
  })

  return padded.map((s, i) => {
    const d = new Date()
    d.setDate(d.getDate() - (29 - i))
    const dateStr = d.toISOString().slice(0, 10)

    if (!s || s.totalChecks === 0) {
      return { color: 'bg-gray-700', tooltip: `${dateStr}: No data` }
    }

    const pct = s.uptimePct
    let color: string
    if (pct >= 99) color = 'bg-emerald-500'
    else if (pct >= 95) color = 'bg-amber-500'
    else color = 'bg-red-500'

    return { color, tooltip: `${dateStr}: ${pct.toFixed(1)}% uptime` }
  })
})
</script>

<template>
  <div class="flex w-full gap-px" style="height: 8px">
    <div
      v-for="(seg, i) in segments"
      :key="i"
      class="flex-1 rounded-[1px] transition-colors"
      :class="seg.color"
      :title="seg.tooltip"
    />
  </div>
</template>
