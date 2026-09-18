<script setup lang="ts">
import { gsap } from 'gsap'
import { vConfetti } from '@neoconfetti/vue'
import type { QuestionResult } from './types'

const props = defineProps<{
  score: number
  maxScore: number
  pointsAwarded: number
  results: QuestionResult[]
  canReview?: boolean
  completedAt?: string | null
  revealCorrectAnswers?: boolean
}>()

const emit = defineEmits<{
  startReview: []
}>()

const { t } = useI18n()

const containerRef = ref<HTMLElement | null>(null)

const isPerfectScore = computed(
  () => props.score === props.maxScore && props.maxScore > 0,
)

// Animated points value for counting effect
const animatedPoints = ref(0)

const resultText = computed(() => {
  if (isPerfectScore.value) {
    return t('quiz.result.perfect')
  }
  if (props.score > 0) {
    return t('quiz.result.wellDone')
  }
  return t('quiz.result.betterLuckNextTime')
})

// Use computed to reactively update the translated text with animated value
const pointsText = computed(() => {
  if (props.pointsAwarded === 0) {
    return t('quiz.result.receivedNoPoints')
  }
  return t('quiz.result.receivedPoints', { points: animatedPoints.value })
})

const showConfetti = ref(false)
onMounted(() => {
  // Animate points counting up
  if (props.pointsAwarded > 0) {
    // Check for reduced motion preference
    const prefersReducedMotion = window.matchMedia(
      '(prefers-reduced-motion: reduce)',
    ).matches
    if (prefersReducedMotion) {
      animatedPoints.value = props.pointsAwarded
    } else {
      gsap.to(animatedPoints, {
        value: props.pointsAwarded,
        duration: 0.8,
        ease: 'power2.out',
        onUpdate: () => {
          animatedPoints.value = Math.round(animatedPoints.value)
        },
      })
    }
  }

  // Confetti for perfect scores
  if (isPerfectScore.value && containerRef.value) {
    // Small delay for the confetti to feel more natural after the result appears
    setTimeout(() => {
      showConfetti.value = true
    }, 300)
  }
})
</script>

<template>
  <div
    ref="containerRef"
    class="text-center p-default flex flex-col gap-large grow relative overflow-hidden"
  >
    <!-- Results hidden - show thank you message -->
    <template v-if="revealCorrectAnswers === false">
      <div class="grow flex flex-col items-center justify-center gap-default">
        <div
          class="rounded-full bg-accent-positive/10 p-6 flex items-center justify-center"
        >
          <UIcon name="lucide:check" class="size-12 text-accent-positive" />
        </div>
        <h1 class="text-heading text-text-default">
          {{ $t('quiz.result.thanksForAnswers') }}
        </h1>
      </div>

      <div class="flex flex-col gap-small">
        <DesignButton
          v-if="canReview"
          size="large"
          variant="secondary"
          class="w-full"
          @click="emit('startReview')"
        >
          {{ $t('quiz.reviewAnswers') }}
        </DesignButton>
        <NuxtLink :to="{ name: 'challenges' }">
          <DesignButton size="large" class="w-full">
            {{ $t('quiz.done') }}
          </DesignButton>
        </NuxtLink>
      </div>
    </template>

    <!-- Normal results view -->
    <template v-else>
      <div v-if="showConfetti" v-confetti />
      <div class="grow flex flex-col items-center justify-center gap-default">
        <QuizProgress
          size="large"
          :total-questions="results.length"
          :current-index="results.length"
          :results
        />

        <h1 class="text-heading text-text-default tabular-nums">
          {{ resultText }}
          <br />
          {{ pointsText }}
        </h1>
      </div>

      <div class="flex flex-col gap-small">
        <DesignButton
          v-if="canReview"
          size="large"
          variant="secondary"
          class="w-full"
          @click="emit('startReview')"
        >
          {{ $t('quiz.reviewAnswers') }}
        </DesignButton>

        <NuxtLink :to="{ name: 'challenges' }">
          <DesignButton size="large" class="w-full">
            {{ $t('quiz.done') }}
          </DesignButton>
        </NuxtLink>
      </div>
    </template>
  </div>
</template>
