<script setup lang="ts">
const open = defineModel<boolean>('open')
const close = () => {
  open.value = false
}
</script>

<template>
  <UDrawer
    v-model:open="open"
    :ui="{
      content:
        'bg-background-raised rounded-t-list ring-border-default h-[95%]',
      overlay: 'bg-black/50',
      handle: 'bg-border-default!',
    }"
    should-scale-background
    :set-background-color-on-scale="false"
  >
    <template #default>
      <slot />
    </template>
    <template #content>
      <slot name="content" :close />
    </template>
  </UDrawer>
</template>
