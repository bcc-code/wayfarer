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

/** Pixels of scroll over which the bar hands the large title to the compact. */
const TITLE_HANDOVER_DISTANCE = 48

const { y } = useWindowScroll()

const hasLargeTitle = computed(() => Boolean(props.title || slots.title))

const titleOpacity = computed(() => {
  if (!hasLargeTitle.value) return 1
  return Math.min(Math.max(y.value / TITLE_HANDOVER_DISTANCE, 0), 1)
})
</script>

<template>
  <div class="flex min-h-full flex-col">
    <div class="sticky top-0 z-10">
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
