<script setup lang="ts">
import QuizQuestionHeading from '#layers/user/app/components/challenges/quiz/QuizQuestionHeading.vue'
import QuizPredefinedQuestion from '#layers/user/app/components/challenges/quiz/questions/QuizPredefinedQuestion.vue'
import QuizFreeTextQuestion from '#layers/user/app/components/challenges/quiz/questions/QuizFreeTextQuestion.vue'
import QuizNumberQuestion from '#layers/user/app/components/challenges/quiz/questions/QuizNumberQuestion.vue'
import QuizOrderingQuestion from '#layers/user/app/components/challenges/quiz/questions/QuizOrderingQuestion.vue'
import type { QuizQuestionFormData } from './AdminQuizForm.vue'

/**
 * The question as a participant meets it, rendered by the participant app's
 * own components. Writing a question in a dialog of form fields says nothing
 * about how long alternatives wrap or what an ordering question looks like
 * once it is a list you drag.
 *
 * Every component is given `readonly`, so it renders the question without
 * offering to answer it — and the frame around the preview is `inert`, so
 * nothing here can be submitted.
 */
const props = defineProps<{
  question: QuizQuestionFormData
  /** Where the question sits in the quiz, for the "n av m" line. */
  index?: number
  total?: number
}>()

/**
 * Shaped like the query the question components are typed against. A draft has
 * no ids until it is saved, so the answers get positional ones: the components
 * key on them, and two empty strings would collide.
 */
const preview = computed(() => ({
  id: props.question.id ?? 'preview',
  questionText: props.question.questionText,
  allowMultipleSelection: props.question.allowMultipleSelection ?? false,
  predefinedAnswers: (props.question.predefinedAnswers ?? []).map(
    (answer, i) => ({
      id: answer.id ?? `answer-${i}`,
      answerText: answer.answerText,
      answerOrder: answer.answerOrder,
      isCorrect: answer.isCorrect,
    }),
  ),
  orderingItems: (props.question.orderingItems ?? []).map((item, i) => ({
    id: item.id ?? `item-${i}`,
    itemText: item.itemText,
    correctOrder: item.correctOrder,
  })),
  minValue: props.question.minValue,
  maxValue: props.question.maxValue,
  stepValue: props.question.stepValue,
  bettingEnabled: false,
}))

// The components are typed against the query result; a draft is not one.
const asQuestion = computed(() => preview.value as never)

const common = computed(() => ({
  question: asQuestion.value,
  totalQuestions: props.total ?? 1,
  currentIndex: props.index ?? 0,
  submissionId: '',
  readonly: true,
}))
</script>

<template>
  <!-- A phone screen's shape: this preview is a whole screen, so judging it
       in a box of some other proportion says little about how much of it the
       question fills. Taller content scrolls, as it does on the phone. -->
  <div
    class="bg-background-default text-text-default aspect-[9/19.5] overflow-y-auto rounded-xl p-4"
  >
    <QuizQuestionHeading
      :question-text="question.questionText || 'Spørsmålstekst'"
      :index="index ?? 0"
      :total="total ?? 1"
    />

    <QuizPredefinedQuestion
      v-if="question.questionType === QuizQuestionType.Predefined"
      v-bind="common"
    />
    <QuizOrderingQuestion
      v-else-if="question.questionType === QuizQuestionType.Ordering"
      v-bind="common"
    />
    <QuizNumberQuestion
      v-else-if="question.questionType === QuizQuestionType.Number"
      v-bind="common"
    />
    <QuizFreeTextQuestion
      v-else-if="question.questionType === QuizQuestionType.FreeText"
      v-bind="common"
    />
  </div>
</template>
