<script setup lang="ts">
const props = withDefaults(
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

const slots = useSlots()

/**
 * The bar is back in flow rather than `sticky`, so the scroll-driven hand-over
 * from the large title to the compact one is switched off with it: the two only
 * make sense together, and a bar that scrolls away has nothing to hand over to.
 * A page with a large title therefore pins the compact one at 0; a page without
 * one keeps the bar title visible, as before.
 *
 * Sticky is reverted rather than retuned — over real content it read as a dark
 * band with the page bleeding through it, which is a design question, not a
 * tuning one. See `notes/app-like-ux.md`.
 */
const hasLargeTitle = computed(() => Boolean(props.title || slots.title))

const titleOpacity = computed(() => (hasLargeTitle.value ? 0 : 1))
</script>

<template>
  <div class="flex min-h-full flex-col">
    <div class="top-0 z-10">
      <TitleBar :title="title" :blurred :shadow :title-opacity="titleOpacity">
        <template v-if="$slots.title" #title>
          <slot name="title" />
        </template>
        <template v-if="$slots.bar" #bar>
          <slot name="bar" />
        </template>
        <template #action>
          <slot name="action" />
        </template>
      </TitleBar>
    </div>
    <div
      :class="[
        'flex relative grow flex-col',
        { 'pb-[calc(7rem+env(safe-area-inset-bottom,0px))]': bottomPadding },
      ]"
    >
      <OfflineNotice />
      <slot />
    </div>
    <div v-if="$slots.footer" class="z-10">
      <slot name="footer" />
    </div>
  </div>
</template>
