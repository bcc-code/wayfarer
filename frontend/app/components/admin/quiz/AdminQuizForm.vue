<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

export interface QuizFormData {
  id?: string
  name: string
  description: string
  image?: string
  timeoutSeconds?: number
  randomizeQuestions: boolean
  revealCorrectAnswers: boolean
  allowRetakes: boolean
  completionPoints: number
  questions: QuizQuestionFormData[]
}

export interface QuizQuestionFormData {
  id?: string
  questionType: QuizQuestionType
  questionText: string
  questionOrder: number
  timeoutSeconds?: number
  points?: number
  allowMultipleSelection?: boolean
  predefinedAnswers?: {
    id?: string
    answerText: string
    isCorrect: boolean
    answerOrder: number
  }[]
  minValue?: number
  maxValue?: number
  stepValue?: number
}

const props = defineProps<{
  quizData?: QuizFormData
  projectId: string
  challengeId?: string
}>()

const emit = defineEmits<{
  save: [data: QuizFormData]
}>()

const toast = useToast()

const schema = z.object({
  name: z.string().min(1, 'Navn er påkrevd'),
  description: z.string().min(1, 'Beskrivelse er påkrevd'),
  image: z.string().optional(),
  timeoutSeconds: z.number().optional(),
  randomizeQuestions: z.boolean(),
  revealCorrectAnswers: z.boolean(),
  allowRetakes: z.boolean(),
  completionPoints: z.number().min(0),
})

type Schema = z.infer<typeof schema>

const state = reactive<Schema>({
  name: props.quizData?.name ?? '',
  description: props.quizData?.description ?? '',
  image: props.quizData?.image,
  timeoutSeconds: props.quizData?.timeoutSeconds,
  randomizeQuestions: props.quizData?.randomizeQuestions ?? false,
  revealCorrectAnswers: props.quizData?.revealCorrectAnswers ?? true,
  allowRetakes: props.quizData?.allowRetakes ?? false,
  completionPoints: props.quizData?.completionPoints ?? 0,
})

const questions = ref<QuizQuestionFormData[]>(props.quizData?.questions ?? [])

watch(
  () => props.quizData,
  (data) => {
    if (data) {
      state.name = data.name
      state.description = data.description
      state.image = data.image
      state.timeoutSeconds = data.timeoutSeconds
      state.randomizeQuestions = data.randomizeQuestions
      state.revealCorrectAnswers = data.revealCorrectAnswers
      state.allowRetakes = data.allowRetakes
      state.completionPoints = data.completionPoints
      questions.value = data.questions
    }
  },
  { once: true },
)

const editingQuestion = ref<QuizQuestionFormData | null>(null)
const isAddingQuestion = ref(false)

function addQuestion() {
  editingQuestion.value = {
    questionType: QuizQuestionType.Predefined,
    questionText: '',
    questionOrder: questions.value.length + 1,
    points: 1,
    allowMultipleSelection: false,
    predefinedAnswers: [
      { answerText: '', isCorrect: true, answerOrder: 1 },
      { answerText: '', isCorrect: false, answerOrder: 2 },
    ],
  }
  isAddingQuestion.value = true
}

function editQuestion(question: QuizQuestionFormData) {
  editingQuestion.value = { ...question }
  isAddingQuestion.value = false
}

function saveQuestion(question: QuizQuestionFormData) {
  if (isAddingQuestion.value) {
    questions.value.push(question)
  } else {
    const index = questions.value.findIndex(
      (q) => q.questionOrder === question.questionOrder,
    )
    if (index !== -1) {
      questions.value[index] = question
    }
  }
  editingQuestion.value = null
  isAddingQuestion.value = false
}

function cancelEdit() {
  editingQuestion.value = null
  isAddingQuestion.value = false
}

function deleteQuestion(index: number) {
  questions.value.splice(index, 1)
  // Reorder remaining questions
  questions.value.forEach((q, i) => {
    q.questionOrder = i + 1
  })
}

function handleSubmit(event: FormSubmitEvent<Schema>) {
  if (!event.data) return

  if (questions.value.length === 0) {
    toast.add({
      title: 'Feil',
      description: 'Legg til minst ett spørsmål',
      color: 'error',
    })
    return
  }

  emit('save', {
    id: props.quizData?.id,
    ...event.data,
    questions: questions.value,
  })
}

const questionTypeOptions = [
  { value: QuizQuestionType.Predefined, label: 'Flervalg' },
  { value: QuizQuestionType.FreeText, label: 'Fritekst' },
  { value: QuizQuestionType.Number, label: 'Tall' },
]
</script>

<template>
  <div class="space-y-8">
    <UForm
      :state
      :schema="schema"
      class="flex max-w-md flex-col gap-6"
      @submit.prevent="handleSubmit"
    >
      <h3 class="text-lg font-semibold">Quiz-innstillinger</h3>

      <UFormField name="name" label="Quiz-navn">
        <UInput v-model="state.name" size="xl" required class="w-full" />
      </UFormField>

      <UFormField name="description" label="Beskrivelse">
        <UTextarea
          v-model="state.description"
          class="w-full"
          autoresize
          required
        />
      </UFormField>

      <UFormField name="image" label="Bilde" hint="(valgfritt)">
        <AdminFileUpload v-model="state.image" />
      </UFormField>

      <UFormField
        name="timeoutSeconds"
        label="Tidsbegrensning (sekunder)"
        hint="(valgfritt)"
        help="Total tidsbegrensning for hele quizen"
      >
        <UInput
          v-model.number="state.timeoutSeconds"
          type="number"
          size="xl"
          class="w-full"
        />
      </UFormField>

      <UFormField name="completionPoints" label="Fullføringspoeng">
        <UInput
          v-model.number="state.completionPoints"
          type="number"
          size="xl"
          required
          class="w-full"
        />
      </UFormField>

      <div class="space-y-3">
        <UFormField name="randomizeQuestions">
          <UCheckbox
            v-model="state.randomizeQuestions"
            label="Tilfeldig spørsmålsrekkefølge"
          />
        </UFormField>

        <UFormField name="revealCorrectAnswers">
          <UCheckbox
            v-model="state.revealCorrectAnswers"
            label="Vis riktige svar etter fullføring"
          />
        </UFormField>

        <UFormField name="allowRetakes">
          <UCheckbox
            v-model="state.allowRetakes"
            label="Tillat brukere å ta quizen på nytt"
          />
        </UFormField>
      </div>

      <!-- Questions Section -->
      <div class="border-t pt-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold">
            Spørsmål ({{ questions.length }})
          </h3>
          <UButton v-if="!editingQuestion" size="sm" @click="addQuestion">
            Legg til spørsmål
          </UButton>
        </div>

        <!-- Question List -->
        <div v-if="!editingQuestion" class="space-y-3">
          <div
            v-for="(question, index) in questions"
            :key="question.questionOrder"
            class="border border-default rounded-lg p-4"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="flex-1">
                <div class="text-sm text-text-muted mb-1">
                  Spørsmål {{ index + 1 }} -
                  {{
                    question.questionType === QuizQuestionType.Predefined
                      ? 'Flervalg'
                      : question.questionType === QuizQuestionType.Number
                        ? 'Tall'
                        : 'Fritekst'
                  }}
                </div>
                <div class="font-medium">
                  {{ question.questionText || '(Ingen spørsmålstekst)' }}
                </div>
                <div
                  v-if="question.points"
                  class="text-sm text-text-muted mt-1"
                >
                  {{ formatNumber(question.points) }} poeng
                </div>
              </div>
              <div class="flex gap-2">
                <UButton
                  size="xs"
                  variant="ghost"
                  @click="editQuestion(question)"
                >
                  Rediger
                </UButton>
                <UButton
                  size="xs"
                  variant="ghost"
                  color="error"
                  @click="deleteQuestion(index)"
                >
                  Slett
                </UButton>
              </div>
            </div>
          </div>

          <div
            v-if="questions.length === 0"
            class="text-center py-8 text-text-muted border border-dashed border-default rounded-lg"
          >
            Ingen spørsmål ennå. Klikk «Legg til spørsmål» for å opprette ett.
          </div>
        </div>

        <!-- Question Editor -->
        <AdminQuizQuestionEditor
          v-if="editingQuestion"
          :question="editingQuestion"
          :question-type-options="questionTypeOptions"
          @save="saveQuestion"
          @cancel="cancelEdit"
        />
      </div>

      <UButton v-if="!editingQuestion" type="submit" size="lg" block>
        Lagre quiz
      </UButton>
    </UForm>
  </div>
</template>
