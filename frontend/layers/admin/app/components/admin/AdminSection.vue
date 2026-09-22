<script setup lang="ts">
// Use instead of UCard for page sections. UCard is for discrete objects on a
// surface — cards in a grid, stat tiles.
defineProps<{
  title: string
  count?: number
}>()
</script>

<template>
  <section>
    <div
      class="mb-2 flex flex-wrap items-baseline justify-between gap-x-4 gap-y-2"
    >
      <h2 class="text-lg font-semibold">
        {{ title }}
        <span
          v-if="count !== undefined"
          class="text-dimmed text-sm font-normal"
        >
          ({{ count }})
        </span>
      </h2>
      <slot name="actions" />
    </div>
    <!-- White on light, tinted on dark: the shell's page ground sits one step
         behind the surfaces in both modes. -->
    <div class="bg-default dark:bg-elevated/40 rounded-lg px-4 py-3">
      <slot />
    </div>
  </section>
</template>
