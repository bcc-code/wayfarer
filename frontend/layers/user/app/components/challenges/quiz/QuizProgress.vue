<script setup lang="ts">
import { cva } from 'cva'
import type { QuestionResult } from './types'

const props = withDefaults(
  defineProps<{
    currentIndex: number
    totalQuestions: number
    results: QuestionResult[]
    size?: 'medium' | 'large'
    revealCorrectAnswers?: boolean
  }>(),
  {
    size: 'medium',
    revealCorrectAnswers: true,
  },
)

type DotState = 'pending' | 'current' | 'correct' | 'wrong' | 'answered'

const badgeClass = cva(
  'bg-background-raised relative gradient-border flex items-center rounded-full',
  {
    variants: {
      size: {
        medium: 'px-5 py-2 h-11 gap-medium',
        large: 'px-7 py-6 gap-default',
      },
    },
  },
)

const dotClass = cva('w-2 h-2 rounded-full transition-colors', {
  variants: {
    state: {
      pending: 'bg-border-default',
      current: 'bg-accent-contrast',
      correct: 'bg-accent-positive',
      wrong: 'bg-accent-negative',
      answered: 'bg-accent-contrast',
    },
    size: {
      medium: 'w-2 h-2',
      large: 'w-3 h-3',
    },
  },
})

function getDotState(index: number): DotState {
  if (index === props.currentIndex) {
    return 'current'
  }

  const result = props.results[index]
  if (result) {
    // When not revealing correct answers, just show as answered
    if (!props.revealCorrectAnswers) {
      return 'answered'
    }
    if (result.isCorrect === true) {
      return 'correct'
    }
    if (result.isCorrect === false) {
      return 'wrong'
    }
  }

  return 'pending'
}
</script>

<template>
  <div :class="badgeClass({ size })">
    <div
      v-for="index in totalQuestions"
      :key="index"
      :class="dotClass({ state: getDotState(index - 1), size })"
    />
  </div>
</template>
