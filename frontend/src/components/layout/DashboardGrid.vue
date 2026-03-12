<script setup lang="ts">
// Grid layout — no props needed. Widgets are rendered via the default slot.
</script>

<template>
  <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
    <slot />
  </div>
</template>

<script lang="ts">
import type { WidgetSize } from '@/types'

export const widgetSizeClass = (size: WidgetSize = 's') =>
  ({
    s: 'col-span-1',
    m: 'col-span-1 md:col-span-2',
    l: 'col-span-1 md:col-span-3',
  })[size]
</script>
