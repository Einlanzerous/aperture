<script setup lang="ts">
import type { WidgetSize } from '@/types'

// Grid uses 3 columns; each widget declares its own col-span via the size prop.
const SIZE_CLASS: Record<WidgetSize, string> = {
  s: 'col-span-1',              // 1/3 width
  m: 'col-span-1 md:col-span-2', // 2/3 width
  l: 'col-span-1 md:col-span-3', // full width
}

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
// Exported so App.vue can map WidgetSize → Tailwind class without coupling.
export const widgetSizeClass = (size: WidgetSize = 's') =>
  ({
    s: 'col-span-1',
    m: 'col-span-1 md:col-span-2',
    l: 'col-span-1 md:col-span-3',
  })[size]
</script>
