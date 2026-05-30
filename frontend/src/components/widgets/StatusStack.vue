<script setup lang="ts">
import type { ServiceStatusData } from '@/types'
import StatusTile from '@/components/widgets/StatusTile.vue'

// Renders one or two status-only tiles stacked in a single grid slot. Each tile
// keeps its own card border — this is a layout wrapper, not a combined card. An
// odd trailing service stacks alone (the slot's lower half stays empty).
defineProps<{
  services: ServiceStatusData[]
}>()
</script>

<template>
  <div class="flex flex-col gap-4">
    <StatusTile
      v-for="service in services"
      :key="service.name"
      :service="service"
    />
  </div>
</template>
