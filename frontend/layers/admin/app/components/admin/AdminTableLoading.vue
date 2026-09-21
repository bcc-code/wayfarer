<script setup lang="ts">
/**
 * Skeleton rows for a `UTable`, for its `#loading` slot.
 *
 * `:loading` on its own draws a thin progress bar over an **empty** table, so a
 * first load reads as "there is nothing here" until the data lands — which is
 * the same thing an empty list looks like. Skeleton rows say "something is
 * coming" instead, and hold the height so the page does not jump.
 *
 * The slot renders inside a single cell spanning every column, so this stacks
 * bars rather than reproducing the column grid. That is deliberate: a fake
 * column layout that disagrees with the real one is worse than honest bars.
 */
withDefaults(defineProps<{ rows?: number }>(), { rows: 5 })
</script>

<template>
  <div class="space-y-3 py-2">
    <USkeleton v-for="row in rows" :key="row" class="h-5 w-full" />
  </div>
</template>
