<script setup lang="ts">
import type { Component } from 'vue'

// Renders one or two tiny (1-slot) tiles stacked inside a single grid cell —
// the slot-based generalization of APTR-14's status-tile pairing. Each tile
// keeps its own card border; this is a layout wrapper, not a combined card. An
// odd trailing tile stacks alone (the cell's lower half stays empty). Tiles are
// heterogeneous (CPU/Memory/GPU stat tiles or status-only service tiles), so we
// render each via <component :is>.
defineProps<{
  tiles: { id: string; component: Component; props: Record<string, unknown> }[]
}>()
</script>

<template>
  <div class="flex flex-col gap-4">
    <component
      :is="tile.component"
      v-for="tile in tiles"
      :key="tile.id"
      v-bind="tile.props"
    />
  </div>
</template>
