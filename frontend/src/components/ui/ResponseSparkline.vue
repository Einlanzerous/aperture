<script setup lang="ts">
import { computed } from 'vue'
import type { CheckRecord } from '@/types'
import MiniSparkline from './MiniSparkline.vue'

const props = defineProps<{ records: CheckRecord[] }>()

// Map response-time records to a plain number series for MiniSparkline.
const values = computed(() =>
  props.records
    .filter((r) => r.responseTime != null && r.responseTime > 0)
    .map((r) => r.responseTime!),
)
</script>

<template>
  <MiniSparkline :values="values" />
</template>
