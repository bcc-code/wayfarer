<script setup lang="ts">
withDefaults(
  defineProps<{
    dismissible?: boolean
  }>(),
  {
    dismissible: true,
  },
)

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
        'bg-background-default rounded-t-list ring-border-default h-full',
      overlay: 'bg-black/50',
      handle: 'bg-border-default!',
    }"
    should-scale-background
    :set-background-color-on-scale="false"
    :dismissible
  >
    <template #default>
      <slot />
    </template>
    <template #content>
      <slot name="content" :close />
    </template>
  </UDrawer>
</template>
