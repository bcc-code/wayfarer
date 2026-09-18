<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    dismissible?: boolean
    nested?: boolean
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
        'bg-background-default rounded-t-modal h-full gradient-border ring-0 max-w-xl mx-auto mt-[max(calc(var(--spacing)*24),env(safe-area-inset-top))] max-h-[min(96%,calc(100%-calc(env(safe-area-inset-top)+0.75rem)))]',
      overlay: 'bg-black/50',
    }"
    :set-background-color-on-scale="false"
    :handle="false"
    :nested
    :dismissible
    :prevent-scroll-restoration="false"
  >
    <template #default>
      <slot />
    </template>
    <template #content>
      <TitleBar :title="title" size="small" :animate="false">
        <template #action>
          <DesignIconButton
            v-if="dismissible"
            icon="IconClose"
            @click="close"
          />
        </template>
      </TitleBar>
      <div
        class="p-list-outside grow flex flex-col gap-list-section-gap overflow-auto"
      >
        <slot name="content" :close />
      </div>
    </template>
  </UDrawer>
</template>
