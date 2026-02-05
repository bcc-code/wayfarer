<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    bottomPadding?: boolean
    shadow?: boolean
    blurred?: boolean
  }>(),
  {
    bottomPadding: true,
    shadow: true,
    blurred: true,
  },
)
</script>

<template>
  <div class="flex min-h-full flex-col">
    <!-- TODO: position sticky -->
    <div class="top-0 z-10">
      <TitleBar :title="title" :blurred :shadow :animate="false">
        <template #title>
          <slot name="title" />
        </template>
        <template #action>
          <slot name="action" />
        </template>
      </TitleBar>
    </div>
    <div :class="['flex relative grow flex-col', { 'pb-28': bottomPadding }]">
      <slot />
    </div>
    <div v-if="$slots.footer" class="z-10">
      <slot name="footer" />
    </div>
  </div>
</template>
