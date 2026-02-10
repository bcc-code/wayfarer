<script setup lang="ts">
import { cva } from 'cva'

const props = defineProps<{
  text: string
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
  'border-2 relative text-label rounded-list p-medium flex gap-list-section-inset text-center justify-center items-center min-h-22 transition-colors',
  {
    variants: {
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
    :class="classes({ highlighted, confirmed, wrong, correct, disabled })"
    :disabled="disabled"
  >
    {{ text }}

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
      {{ $t('quiz.yourAnswer') }}
    </span>
  </button>
</template>
