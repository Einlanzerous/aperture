<script setup lang="ts">
defineProps<{
  // Callers render their widgets as named slots or default slot children.
  // The grid only controls the outer layout.
}>()
</script>

<template>
  <!--
    3-column grid. On mobile everything stacks to full width.
    Widgets declare their own size via col-span classes applied by
    the parent (App.vue) using the sizeClass helper below.
  -->
  <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
    <slot />
  </div>
</template>

<script lang="ts">
import type { WidgetSize } from '@/types'

// Exported so App.vue can map WidgetSize → Tailwind class without coupling.
export const widgetSizeClass = (size: WidgetSize = 's') =>
  ({
    s: 'col-span-1',
    m: 'col-span-1 md:col-span-2',
    l: 'col-span-1 md:col-span-3',
  })[size]
</script>
