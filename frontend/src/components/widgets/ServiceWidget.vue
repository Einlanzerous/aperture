<script setup lang="ts">
import type { ServiceStatusData, WidgetSize } from '@/types'
import ServiceTile from '@/components/widgets/ServiceTile.vue'
import StatusTile  from '@/components/widgets/StatusTile.vue'

withDefaults(
  defineProps<{
    service: ServiceStatusData
    size?: WidgetSize
    storageEnabled?: boolean
  }>(),
  { size: 's', storageEnabled: false },
)
</script>

<template>
  <StatusTile v-if="service.statusOnly" :service="service" />
  <ServiceTile
    v-else
    :service="service"
    :size="size"
    :storage-enabled="storageEnabled"
  />
</template>
