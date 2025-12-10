<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
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
        'bg-background-default rounded-t-modal h-full gradient-border ring-0',
      overlay: 'bg-black/50',
    }"
    should-scale-background
    :set-background-color-on-scale="false"
    :handle="false"
    :dismissible
  >
    <template #default>
      <slot />
    </template>
    <template #content>
      <TitleBar :title="title" size="small">
        <template #action>
          <DesignIconButton v-if="dismissible" icon="lucide:x" @click="close" />
        </template>
      </TitleBar>
      <div
        class="p-list-outside grow flex flex-col gap-list-section-gap overflow-auto"
      >
        <slot name="content" />
      </div>
    </template>
  </UDrawer>
</template>
