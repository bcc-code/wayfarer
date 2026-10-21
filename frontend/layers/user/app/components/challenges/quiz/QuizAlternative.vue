<script setup lang="ts">
import { cva } from 'cva'

const props = defineProps<{
  text: string
  /** Derived from the answer's position; null when the question is unlabelled. */
  label?: string | null
  description?: string | null
  highlighted?: boolean
  confirmed?: boolean
  wrong?: boolean
  correct?: boolean
  disabled?: boolean
  selected?: boolean
}>()

const buttonRef = ref<HTMLButtonElement | null>(null)
const { shake } = useShake()
const { pulse } = usePulse()

// Watch for confirmation to trigger animations
watch(
  () => props.confirmed,
  (isConfirmed) => {
    if (!isConfirmed || !buttonRef.value) return

    if (props.wrong && props.selected) {
      shake(buttonRef.value)
    } else if (props.correct) {
      pulse(buttonRef.value)
    }
  },
)

const classes = cva(
  'border-2 relative text-label rounded-list p-medium flex gap-list-section-inset min-h-22 transition-colors',
  {
    variants: {
      /**
       * A title with an explanation under it is a block of copy, and two
       * centred wrapped lines give the eye no edge to come back to. A label
       * needs the same edge, so the letters line up as a column rather than
       * drifting with each title's width. A bare alternative is a few words,
       * so it stays centred.
       */
      alignStart: {
        true: 'text-left justify-start',
        false: 'text-center justify-center',
      },
      /**
       * A label sits with the title it belongs to. Against a single line
       * that means the middle, but against a title with an explanation under
       * it the whole row rises to the top so the label meets the title
       * rather than the paragraph's centre.
       */
      withDescription: {
        true: 'items-start',
        false: 'items-center',
      },
      highlighted: {
        true: '',
        false: '',
      },
      confirmed: {
        true: '',
        false: '',
      },
      wrong: {
        true: '',
        false: '',
      },
      correct: {
        true: '',
        false: '',
      },
      disabled: {
        true: 'cursor-default',
        false: '',
      },
    },
    compoundVariants: [
      {
        highlighted: false,
        confirmed: false,
        class: 'border-border-default',
      },
      {
        highlighted: true,
        confirmed: false,
        class: 'border-accent-contrast text-accent-contrast',
      },
      {
        confirmed: true,
        wrong: true,
        class: 'border-accent-negative',
      },
      {
        confirmed: true,
        correct: true,
        class: 'border-accent-positive',
      },
      // Selected answer when not showing correct/wrong
      {
        confirmed: true,
        highlighted: true,
        wrong: false,
        correct: false,
        class: 'border-accent-contrast text-accent-contrast',
      },
      // Non-selected answers when not showing correct/wrong
      {
        confirmed: true,
        highlighted: false,
        wrong: false,
        correct: false,
        class: 'border-border-default opacity-50',
      },
    ],
  },
)
</script>

<template>
  <button
    ref="buttonRef"
    :class="
      classes({
        alignStart: !!description || !!label,
        withDescription: !!description,
        highlighted,
        confirmed,
        wrong,
        correct,
        disabled,
      })
    "
    :disabled="disabled"
  >
    <!-- Its own column, not a prefix on the text: glued to the front of the
         title the letters drift with each title's width instead of lining up
         as something the eye can run down. `min-w` keeps "9." and "10."
         in the same column. -->
    <span v-if="label" class="min-w-[2ch] shrink-0 tabular-nums opacity-70">
      {{ label }}
    </span>
    <span v-if="description" class="flex w-full flex-col gap-1">
      <span>{{ text }}</span>
      <span class="text-caption opacity-70">{{ description }}</span>
    </span>
    <span v-else>{{ text }}</span>

    <span
      v-if="confirmed && wrong && selected"
      class="absolute top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 text-label text-on-accent bg-accent-negative rounded-full pl-2 pr-3 py-1 flex gap-1 items-center"
    >
      <IconClose class="size-6" />
      {{ $t('quiz.yourAnswer') }}
    </span>
    <span
      v-else-if="confirmed && wrong && !selected"
      class="absolute top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 text-label text-on-accent bg-accent-negative rounded-full pl-2 pr-3 py-1 flex gap-1 items-center"
    >
      <IconClose class="size-6" />
      {{ $t('quiz.wrongAnswer') }}
    </span>
    <span
      v-if="confirmed && correct"
      class="absolute top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 text-label text-on-accent bg-accent-positive rounded-full pl-2 pr-3 py-1 flex gap-1 items-center"
    >
      <IconCheck class="size-6" />
      {{ $t('quiz.correctAnswer') }}
    </span>
    <span
      v-if="confirmed && selected && !correct && !wrong"
      class="absolute top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 text-label text-on-accent bg-accent-contrast rounded-full px-3 py-1 flex gap-1 items-center"
    >
      <IconLock class="size-6" />
      {{ $t('quiz.yourAnswer') }}
    </span>
  </button>
</template>
