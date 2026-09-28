<script setup lang="ts">
// For UTable's #loading slot. `:loading` alone draws a bar over an empty table,
// which looks like an empty list. The slot renders in one cell spanning all
// columns, hence bars rather than a column grid.
withDefaults(defineProps<{ rows?: number }>(), { rows: 5 })
</script>

<template>
  <div class="space-y-3 py-2">
    <USkeleton v-for="row in rows" :key="row" class="h-5 w-full" />
  </div>
</template>
