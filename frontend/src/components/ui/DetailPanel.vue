<script setup lang="ts">
import { computed } from 'vue'
import type { CheckRecord, DailySummary } from '@/types'
import ResponseSparkline from './ResponseSparkline.vue'

const props = defineProps<{
  records: CheckRecord[]
  summaries: DailySummary[]
}>()

const overallUptime = computed(() => {
  const withChecks = props.summaries.filter((s) => s.totalChecks > 0)
  if (withChecks.length === 0) return null
  const total = withChecks.reduce((acc, s) => acc + s.totalChecks, 0)
  const healthy = withChecks.reduce((acc, s) => acc + s.healthyChecks, 0)
  return ((healthy / total) * 100).toFixed(1)
})

const avgResponseTime = computed(() => {
  const valid = props.records.filter((r) => r.responseTime != null && r.responseTime > 0)
  if (valid.length === 0) return null
  const avg = valid.reduce((acc, r) => acc + r.responseTime!, 0) / valid.length
  return avg < 1000 ? `${Math.round(avg)}ms` : `${(avg / 1000).toFixed(2)}s`
})
</script>

<template>
  <div class="space-y-2">
    <ResponseSparkline :records="records" />
    <div class="flex items-center gap-3 text-xs text-gray-400">
      <span v-if="overallUptime != null" class="tabular-nums">
        {{ overallUptime }}% uptime
      </span>
      <span v-if="overallUptime != null && avgResponseTime" aria-hidden="true">&middot;</span>
      <span v-if="avgResponseTime" class="tabular-nums">
        avg {{ avgResponseTime }}
      </span>
    </div>
  </div>
</template>
