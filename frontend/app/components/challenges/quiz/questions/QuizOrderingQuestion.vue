<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import type {
  OrderingQuestionData,
  QuestionResult,
  OrderingResponseData,
  QuizActionState,
  QuizActionHandlers,
} from '../types'
import { QuizSessionState } from '~/api/generated'

const props = defineProps<{
  question: OrderingQuestionData
  totalQuestions: number
  currentIndex: number
  submissionId: string
  // Controls whether to show correct/incorrect after answering
  revealCorrectAnswers?: boolean
  // Session-based betting props
  sessionState?: QuizSessionState
  existingResponse?: OrderingResponseData
  // Review mode props
  readonly?: boolean
  preSubmittedOrder?: string[]
  showCorrectAnswers?: boolean
  showPreviousButton?: boolean
  isLastQuestion?: boolean
  // Betting
  betAmount?: number
}>()

const emit = defineEmits<{
  answerSubmitted: [result: QuestionResult]
  betSaved: [responseId: string]
  previous: []
  next: []
}>()

const { track } = useAnalytics()
const { executeMutation: submitAnswer } = useSubmitQuizAnswerMutation()
const { executeMutation: updateAnswer } = useUpdateQuizAnswerMutation()

interface OrderingItem {
  id: string
  itemText: string
  correctOrder?: number | null
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

// Initialize items - ordered by existing response, preSubmittedOrder, or shuffled
const items = ref<OrderingItem[]>([])

// Track if bet is saved and if user is editing
const savedResponseId = ref<string | null>(props.existingResponse?.id ?? null)
const isBetSaved = ref(!!props.existingResponse)
const isEditing = ref(false)

// Determine if session is locked (only LOCKED, not FINISHED - FINISHED shows results)
const isSessionLocked = computed(() => {
  return props.sessionState === QuizSessionState.Locked
})

// Determine if session is finished (shows results)
const isSessionFinished = computed(() => {
  return props.sessionState === QuizSessionState.Finished
})

// Determine if items can be dragged
const canDrag = computed(() => {
  // Cannot drag if readonly, session locked/finished, or bet is saved and not editing
  if (props.readonly) return false
  if (isSessionLocked.value || isSessionFinished.value) return false
  if (isBetSaved.value && !isEditing.value) return false
  return true
})

// Compute item results for finished state
const itemResults = computed(() => {
  if (!isSessionFinished.value) return null
  return items.value.map((item, index) => ({
    id: item.id,
    userPosition: index + 1,
    correctPosition: item.correctOrder,
    isCorrect: item.correctOrder != null && index + 1 === item.correctOrder,
  }))
})

onMounted(() => {
  if (props.existingResponse?.submittedOrder) {
    // In betting mode with existing response, show items in submitted order
    const orderMap = new Map(
      props.existingResponse.submittedOrder.map((id, idx) => [id, idx]),
    )
    items.value = [...props.question.orderingItems].sort((a, b) => {
      const aIdx = orderMap.get(a.id) ?? 0
      const bIdx = orderMap.get(b.id) ?? 0
      return aIdx - bIdx
    })
  } else if (props.readonly && props.preSubmittedOrder) {
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
    // Shuffle items for new quiz
    items.value = shuffleArray([...props.question.orderingItems])
  }
})

// Watch for prop changes to update correctOrder values when session finishes
watch(
  () => props.question.orderingItems,
  (newItems) => {
    if (!items.value.length) return

    // Create a map of id -> correctOrder from new props
    const correctOrderMap = new Map(
      newItems.map((item) => [item.id, item.correctOrder]),
    )

    // Update correctOrder in existing items (preserves user's order)
    items.value = items.value.map((item) => ({
      ...item,
      correctOrder: correctOrderMap.get(item.id),
    }))
  },
  { deep: true },
)

const isAnswerConfirmed = ref(props.readonly ?? false)
const isSubmitting = ref(false)
const submittedResult = ref<{ isCorrect: boolean | null } | null>(null)

const containerRef = ref<HTMLElement | null>(null)
const { shake } = useShake()
const { pulse } = usePulse()

async function handleSaveBet() {
  if (isSubmitting.value) return

  isSubmitting.value = true
  const submittedOrder = items.value.map((item) => item.id)

  if (savedResponseId.value) {
    // Update existing bet
    const result = await updateAnswer({
      responseId: savedResponseId.value,
      input: {
        submittedOrder,
        betAmount: props.betAmount,
      },
    })

    if (result.data?.updateQuizAnswer) {
      isBetSaved.value = true
      isEditing.value = false
      emit('betSaved', savedResponseId.value)

      track(AnalyticsEvent.QuizAnswerSubmitted, {
        question_id: props.question.id,
        is_correct: null, // Hidden while session is OPEN
        current_question: props.currentIndex + 1,
        total_questions: props.totalQuestions,
        action: 'update_bet',
      })
    }
  } else {
    // Create new bet
    const result = await submitAnswer({
      submissionId: props.submissionId,
      input: {
        questionId: props.question.id,
        submittedOrder,
        betAmount: props.betAmount,
      },
    })

    if (result.data?.submitQuizAnswer) {
      const response = result.data.submitQuizAnswer
      savedResponseId.value = response.id
      isBetSaved.value = true
      isEditing.value = false
      emit('betSaved', response.id)

      track(AnalyticsEvent.QuizAnswerSubmitted, {
        question_id: props.question.id,
        is_correct: null, // Hidden while session is OPEN
        current_question: props.currentIndex + 1,
        total_questions: props.totalQuestions,
        action: 'save_bet',
      })
    }
  }

  isSubmitting.value = false
}

function handleChangeBet() {
  isEditing.value = true
}

// Legacy behavior for non-session quizzes
async function handleLockAnswer() {
  if (isSubmitting.value) return

  isSubmitting.value = true

  const submittedOrder = items.value.map((item) => item.id)

  const result = await submitAnswer({
    submissionId: props.submissionId,
    input: {
      questionId: props.question.id,
      submittedOrder,
      betAmount: props.betAmount ?? undefined,
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

    // Trigger animation only if revealing correct answers
    if (containerRef.value && props.revealCorrectAnswers !== false) {
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

// Determine if we're in session-based betting mode
const isSessionBettingMode = computed(() => {
  return props.sessionState === 'OPEN'
})

// Compute action mode for parent component
const actionMode = computed(() => {
  if (props.readonly) return 'review' as const
  if (isSessionFinished.value) return 'session-results' as const
  if (isSessionLocked.value) return 'session-locked' as const
  if (isSessionBettingMode.value) return 'session-betting' as const
  return 'normal' as const
})

// Action state for parent component
const actionState = computed<QuizActionState>(() => ({
  mode: actionMode.value,
  canSubmit: true, // Ordering always has a valid order
  isSubmitting: isSubmitting.value,
  isAnswerLocked: isAnswerConfirmed.value,
  isBetSaved: isBetSaved.value,
  isEditing: isEditing.value,
  showPreviousButton: props.showPreviousButton ?? false,
  isLastQuestion: props.isLastQuestion ?? false,
}))

// Handlers for parent component
const handlers: QuizActionHandlers = {
  submit: async () => {
    if (isSessionBettingMode.value) {
      await handleSaveBet()
    } else {
      await handleLockAnswer()
    }
  },
  continue: handleContinue,
  changeBet: handleChangeBet,
  previous: handlePrevious,
  next: handleNext,
}

defineExpose({ actionState, handlers })
</script>

<template>
  <div class="flex flex-col gap-default grow">
    <div class="flex flex-col justify-center gap-small p-default">
      <div ref="containerRef" class="flex flex-col justify-center">
        <VueDraggable
          v-model="items"
          ghost-class="invisible"
          drag-class="scale-105"
          :animation="200"
          :disabled="!canDrag"
          :delay="200"
          class="flex flex-col gap-small"
        >
          <div
            v-for="(item, index) in items"
            :key="item.id"
            class="flex items-center gap-medium p-medium rounded-list-inset"
            :class="{
              'ring ring-border-default ring-inset':
                !canDrag && !isSessionFinished,
              'bg-background-raised shadow-small': canDrag || isSessionFinished,
            }"
          >
            <div
              class="rounded-full aspect-square size-6 text-center shrink-0 grid place-items-center text-label"
              :class="{
                'bg-accent text-on-accent': !isSessionFinished,
                'bg-accent-positive text-on-accent':
                  isSessionFinished && itemResults?.[index]?.isCorrect,
                'bg-accent-negative text-on-accent':
                  isSessionFinished && !itemResults?.[index]?.isCorrect,
              }"
            >
              {{ index + 1 }}
            </div>
            <span class="text-label text-text-default flex-1 text-left">
              {{ item.itemText }}
            </span>
            <div
              v-if="canDrag"
              class="text-text-muted shrink-0 flex items-center cursor-default"
            >
              <UIcon name="lucide:grip-vertical" class="size-4" />
            </div>
          </div>
        </VueDraggable>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ordering-ghost {
  opacity: 0.5;
  background: var(--color-background-raised);
}
</style>
