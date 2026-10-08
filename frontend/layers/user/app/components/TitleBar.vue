<script setup lang="ts">
/**
 * A constant height whatever the scroll: the two titles are overlaid on one row
 * and cross-faded, so only opacity moves and nothing reflows.
 *
 * The bar carries its own translucent background rather than relying on the
 * blur alone. `backdrop-filter` has nothing to sample whenever the page behind
 * it is not painting — during a page transition the outgoing page drops to
 * `opacity: 0`, which also makes it a backdrop root — and without a background
 * all that was left was the `to-shadow-default` scrim over bare body, which
 * read as a black bar flashing on every navigation.
 *
 * Not applied when `blurred` is false: `DesignDrawer` needs the corner region
 * outside `rounded-t-modal` left unpainted.
 */
withDefaults(
  defineProps<{
    title?: string
    shadow?: boolean
    blurred?: boolean
    /** `small` is a sheet bar: one always-visible title, nothing to fade. */
    size?: 'large' | 'small'
    /** How far the compact title has faded in, 0–1. */
    titleOpacity?: number
  }>(),
  {
    size: 'large',
    shadow: true,
    blurred: true,
    titleOpacity: 1,
  },
)
</script>

<template>
  <ProgressiveBlur
    direction="up"
    :class="[
      shadow && 'from-shadow-blank/0 to-shadow-default bg-linear-to-t',
      blurred && 'bg-background-default/70',
    ]"
    :enabled="blurred"
  >
    <!-- The padding sits here so the <header> is the content box, which is what
         lets the compact title centre on `top-1/2` and still line up with the
         action: against the padding box the safe-area inset would lift it. -->
    <div
      :class="[
        'px-6 pb-3',
        size === 'large'
          ? 'pt-[max(calc(env(safe-area-inset-top)+0.75rem),3rem)]'
          : 'pt-6',
      ]"
    >
      <header
        v-if="size === 'large'"
        class="relative flex min-h-11 items-center gap-4"
      >
        <div v-if="$slots.bar" class="min-w-0 grow">
          <slot name="bar" />
        </div>
        <template v-else>
          <div class="min-w-0 grow" :style="{ opacity: 1 - titleOpacity }">
            <slot name="title">
              <h1 v-if="title" class="text-text-default text-heading truncate">
                {{ title }}
              </h1>
            </slot>
          </div>
          <p
            v-if="title"
            class="text-label text-text-default pointer-events-none absolute inset-x-12 top-1/2 -translate-y-1/2 truncate text-center"
            :style="{ opacity: titleOpacity }"
          >
            {{ title }}
          </p>
        </template>
        <div class="size-11 shrink-0">
          <slot name="action" />
        </div>
      </header>

      <header v-else class="relative flex min-h-11 items-start gap-4">
        <h1 class="text-text-default text-heading min-w-0 grow truncate">
          {{ title }}
        </h1>
        <div class="size-11 shrink-0">
          <slot name="action" />
        </div>
      </header>
    </div>
  </ProgressiveBlur>
</template>
