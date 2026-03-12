<script setup lang="ts">
withDefaults(defineProps<{ count?: number }>(), { count: 1 })
</script>

<template>
  <div
    v-for="i in count"
    :key="i"
    class="col-span-1 h-28 animate-pulse rounded-xl border border-gray-800 bg-gray-900"
  />
</template>
