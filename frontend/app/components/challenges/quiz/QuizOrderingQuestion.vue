<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import type { OrderingQuestionData, QuestionResult } from './types'

const props = defineProps<{
  question: OrderingQuestionData
  totalQuestions: number
  currentIndex: number
  submissionId: string
  // Review mode props
  readonly?: boolean
  preSubmittedOrder?: string[]
  showCorrectAnswers?: boolean
  showPreviousButton?: boolean
  isLastQuestion?: boolean
}>()

const emit = defineEmits<{
  answerSubmitted: [result: QuestionResult]
  previous: []
  next: []
}>()

const { track } = useAnalytics()
const { executeMutation: submitAnswer } = useSubmitQuizAnswerMutation()

interface OrderingItem {
  id: string
  itemText: string
}

// Shuffle items on mount (Fisher-Yates shuffle)
function shuffleArray<T>(array: T[]): T[] {
  const shuffled = [...array]
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    const temp = shuffled[i]!
    shuffled[i] = shuffled[j]!
    shuffled[j] = temp
  }
  return shuffled
}

// Initialize items - shuffled for normal mode, ordered by submission for review mode
const items = ref<OrderingItem[]>([])

onMounted(() => {
  if (props.readonly && props.preSubmittedOrder) {
    // In review mode, show items in the order they were submitted
    const orderMap = new Map(
      props.preSubmittedOrder.map((id, idx) => [id, idx]),
    )
    items.value = [...props.question.orderingItems].sort((a, b) => {
      const aIdx = orderMap.get(a.id) ?? 0
      const bIdx = orderMap.get(b.id) ?? 0
      return aIdx - bIdx
    })
  } else {
    // Shuffle items for quiz mode
    items.value = shuffleArray([...props.question.orderingItems])
  }
})

const isAnswerConfirmed = ref(props.readonly ?? false)
const isSubmitting = ref(false)
const submittedResult = ref<{ isCorrect: boolean | null } | null>(null)

const containerRef = ref<HTMLElement | null>(null)
const { shake } = useShake()
const { pulse } = usePulse()

async function handleLockAnswer() {
  if (isSubmitting.value) return

  isSubmitting.value = true

  const submittedOrder = items.value.map((item) => item.id)

  const result = await submitAnswer({
    submissionId: props.submissionId,
    input: {
      questionId: props.question.id,
      submittedOrder,
    },
  })

  if (result.data?.submitQuizAnswer) {
    const response = result.data.submitQuizAnswer
    const isCorrect =
      response.__typename === 'OrderingResponse'
        ? (response.isCorrect ?? null)
        : null

    submittedResult.value = { isCorrect }
    isAnswerConfirmed.value = true

    // Trigger animation
    if (containerRef.value) {
      if (isCorrect === true) {
        pulse(containerRef.value)
      } else if (isCorrect === false) {
        shake(containerRef.value)
      }
    }

    track(AnalyticsEvent.QuizAnswerSubmitted, {
      question_id: props.question.id,
      is_correct: isCorrect,
      current_question: props.currentIndex + 1,
      total_questions: props.totalQuestions,
    })
  }

  isSubmitting.value = false
}

function handleContinue() {
  emit('answerSubmitted', {
    questionId: props.question.id,
    isCorrect: submittedResult.value?.isCorrect ?? null,
  })
}

const { t } = useI18n()
const continueText = computed(() => {
  if (props.currentIndex === props.totalQuestions - 1) {
    return t('quiz.continue')
  }
  return t('quiz.nextQuestion')
})

// In readonly mode, use isLastQuestion prop to determine the next button text
const nextButtonText = computed(() => {
  if (props.readonly) {
    return props.isLastQuestion
      ? t('quiz.finishReview')
      : t('quiz.nextQuestion')
  }
  return continueText.value
})

function handlePrevious() {
  emit('previous')
}

function handleNext() {
  emit('next')
}
</script>

<template>
  <div class="text-center p-default flex flex-col gap-small grow">
    <p class="text-caption text-text-hint text-center">
      {{ question.questionText }}
    </p>
    <div ref="containerRef">
      <VueDraggable
        v-model="items"
        ghost-class="opacity-50"
        :animation="200"
        :disabled="isAnswerConfirmed"
        class="flex flex-col gap-small"
      >
        <div
          v-for="(item, index) in items"
          :key="item.id"
          class="flex items-center gap-medium p-medium bg-background-raised rounded-list-inset shadow-small"
          :class="{ 'opacity-50': isAnswerConfirmed && !readonly }"
        >
          <div
            class="bg-accent rounded-full aspect-square size-6 text-center shrink-0 grid place-items-center text-label text-on-accent"
          >
            <span class="">
              {{ index + 1 }}
            </span>
          </div>
          <span class="text-label text-text-default flex-1 text-left">
            {{ item.itemText }}
          </span>
          <div
            class="text-text-hint shrink-0 flex items-center"
            :class="
              isAnswerConfirmed
                ? 'cursor-default'
                : 'cursor-grab active:cursor-grabbing'
            "
          >
            <UIcon name="lucide:grip-vertical" class="size-4" />
          </div>
        </div>
      </VueDraggable>
    </div>

    <!-- Result badge -->
    <div v-if="isAnswerConfirmed && !readonly" class="flex justify-center">
      <span
        v-if="submittedResult?.isCorrect === true"
        class="text-label text-on-accent bg-accent-positive rounded-full pl-2 pr-3 py-1 flex gap-1 items-center"
      >
        <IconCheck class="size-6" />
        {{ $t('quiz.correctAnswer') }}
      </span>
      <span
        v-else-if="submittedResult?.isCorrect === false"
        class="text-label text-on-accent bg-accent-negative rounded-full pl-2 pr-3 py-1 flex gap-1 items-center"
      >
        <IconClose class="size-6" />
        {{ $t('quiz.wrongAnswer') }}
      </span>
    </div>

    <!-- Normal mode: Lock answer / Continue buttons -->
    <template v-if="!readonly">
      <DesignButton
        v-if="!isAnswerConfirmed"
        size="large"
        class="grow-0"
        :disabled="isSubmitting"
        :loading="isSubmitting"
        @click="handleLockAnswer"
      >
        {{ $t('quiz.lockAnswer') }}
      </DesignButton>

      <DesignButton v-else size="large" class="grow-0" @click="handleContinue">
        {{ continueText }}
      </DesignButton>
    </template>

    <!-- Readonly/review mode: Previous / Next navigation -->
    <template v-else>
      <div class="flex gap-small grow-0">
        <DesignButton
          v-if="showPreviousButton"
          size="large"
          variant="secondary"
          class="flex-1"
          @click="handlePrevious"
        >
          {{ $t('quiz.previousQuestion') }}
        </DesignButton>
        <DesignButton
          size="large"
          :class="showPreviousButton ? 'flex-1' : 'w-full'"
          @click="handleNext"
        >
          {{ nextButtonText }}
        </DesignButton>
      </div>
    </template>
  </div>
</template>
