<script setup lang="ts">
/**
 * The question number, an optional image and the question itself, above the
 * answer area.
 *
 * Extracted from `QuizChallenge` so the admin panel's question preview can
 * show the heading a participant actually sees rather than a copy of it.
 */
interface QuestionImage {
  url: string
  width?: number | null
  height?: number | null
  blurhash?: string | null
}

defineProps<{
  questionText: string
  /** Zero-based, as the quiz counts internally. */
  index: number
  total: number
  image?: QuestionImage | null
}>()
</script>

<template>
  <div
    class="flex flex-col items-center justify-center py-3 px-medium gap-1 text-center"
  >
    <p v-if="total > 1" class="text-caption text-text-muted">
      {{ $t('quiz.questionNumber', { current: index + 1, total }) }}
    </p>
    <!-- No fixed ratio: the image belongs to the question, and cropping it to
         a shape of ours could cut away the thing being asked about. The box
         takes the image's own ratio, so `object-cover` never has anything to
         crop. -->
    <DesignImage
      v-if="image?.url"
      :image="image"
      :alt="questionText"
      class="rounded-card w-full"
    />
    <h1
      class="text-heading limitedHeight:text-label text-text-default text-balance"
    >
      {{ questionText }}
    </h1>
  </div>
</template>
