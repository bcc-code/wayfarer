<script setup lang="ts">
/**
 * A titled section of a detail page.
 *
 * Replaces `UCard` for page content. A card draws a box, and the rows inside
 * these panels were *themselves* boxed — so every item wore chrome twice and
 * five stacked cards read as five competing panels rather than one page. A
 * heading plus a rule separates them with far less ink.
 *
 * Cards still belong where something genuinely is a discrete object on a
 * surface — a project card in a grid, a stat tile. Not for "this page has
 * sections".
 *
 * **A tinted surface holds the content; the heading sits outside it.** Four
 * versions got here, and the last step needed light mode to see:
 *
 * 1. `UCard` per section with boxed rows inside — every item wore chrome twice.
 * 2. A rule above each heading plus a divider per row: ~35 horizontal lines on
 *    a user with a full points journal. Line chrome for box chrome.
 * 3. No lines at all. Acceptable in dark mode, where surfaces carry some tonal
 *    separation of their own — and completely flat in **light** mode, where
 *    everything sits on white and nothing groups a section. I had been judging
 *    every iteration in dark mode only.
 * 4. A hairline under each heading: better, still thin grouping on white.
 *
 * A soft surface groups a section's rows in both modes without a border, a
 * shadow, or a box per row. It is not the card it replaced: no ring, no
 * elevation, and the heading stays outside so the page still reads as sections
 * rather than as a stack of panels.
 *
 * **The surface is white in light mode, not a grey tint.** The shell sets the
 * page ground to `bg-neutral-100 dark:bg-neutral-950` — deliberately one step
 * *behind* the surfaces in both modes (see the restructure note). So a raised
 * surface on light is `bg-default` (white); a grey tint there is *darker* than
 * the page and reads as sunk into it rather than lifted off it. Dark mode keeps
 * the lighter `bg-elevated` tint, which is the same relationship in the other
 * direction.
 */
defineProps<{
  title: string
  /** Shown next to the title when there is something to count. */
  count?: number
}>()
</script>

<template>
  <section>
    <div
      class="mb-3 flex flex-wrap items-baseline justify-between gap-x-4 gap-y-2"
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
    <div class="bg-default dark:bg-elevated/40 rounded-lg px-4 py-3">
      <slot />
    </div>
  </section>
</template>
