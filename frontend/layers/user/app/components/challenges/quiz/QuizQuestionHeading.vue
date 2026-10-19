<script setup lang="ts">
/**
 * The question number and the question itself, above the answer area.
 *
 * Extracted from `QuizChallenge` so the admin panel's question preview can
 * show the heading a participant actually sees rather than a copy of it.
 */
defineProps<{
  questionText: string
  /** Zero-based, as the quiz counts internally. */
  index: number
  total: number
}>()
</script>

<template>
  <div
    class="flex flex-col items-center justify-center py-3 px-medium gap-1 text-center"
  >
    <p v-if="total > 1" class="text-caption text-text-muted">
      {{ $t('quiz.questionNumber', { current: index + 1, total }) }}
    </p>
    <h1
      class="text-heading limitedHeight:text-label text-text-default text-balance"
    >
      {{ questionText }}
    </h1>
  </div>
</template>
