<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    bottomPadding?: boolean
  }>(),
  {
    bottomPadding: true,
  },
)

const { y } = useWindowScroll()
const hasScrolled = computed(() => y.value > 25)
</script>

<template>
  <div class="flex min-h-full flex-col">
    <div class="sticky top-0 z-10">
      <ProgressiveBlur
        direction="up"
        class="from-shadow-blank/0 to-shadow-default bg-linear-to-t"
      >
        <header
          :class="[
            'relative flex items-center justify-between gap-4 px-6 pt-12 pb-3',
            { 'min-h-20': hasScrolled, 'min-h-24': !hasScrolled },
          ]"
        >
          <h1
            :class="[
              'text-text-default absolute transition-all duration-300 ease-out',
              {
                'text-heading bottom-3 left-6 translate-x-0 translate-y-0':
                  !hasScrolled,
                'text-label top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2':
                  hasScrolled,
              },
            ]"
          >
            {{ title }}
          </h1>
          <div
            :class="[
              'absolute right-6 bottom-3 size-11 transition-all duration-300 ease-out',
              {
                'bottom-3': !hasScrolled,
                'top-1/2 -translate-y-1/2': hasScrolled,
              },
            ]"
          >
            <slot name="action" />
          </div>
        </header>
      </ProgressiveBlur>
    </div>
    <div
      :class="['p-list-outside flex grow flex-col', { 'pb-28': bottomPadding }]"
    >
      <slot />
    </div>
  </div>
</template>
