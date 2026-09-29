<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import type { FormSubmitEvent } from '@nuxt/ui'
import type { TranslationStatusFragment } from '~/api/generated'
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

export interface QuizAnswerFormData {
  id?: string
  /** Stable across reorders, which `id` is not for a new answer. */
  localKey?: string
  answerText: string
  isCorrect: boolean
  answerOrder: number
  translationStatus?: TranslationStatusFragment[]
}

export interface QuizQuestionFormData {
  id?: string
  /** Stable across edits and reorders, which `id` is not for a new question. */
  localKey?: string
  questionType: QuizQuestionType
  questionText: string
  questionOrder: number
  timeoutSeconds?: number
  points?: number
  allowMultipleSelection?: boolean
  predefinedAnswers?: QuizAnswerFormData[]
  minValue?: number
  maxValue?: number
  stepValue?: number
  orderingItems?: {
    id?: string
    itemText: string
    correctOrder: number
  }[]
  bettingEnabled?: boolean
  bettingMinPercentage?: number
  bettingMaxPercentage?: number
  bettingMinAbsolute?: number
  bettingMaxAbsolute?: number
  translationStatus?: TranslationStatusFragment[]
}

const props = defineProps<{
  quizData?: QuizFormData
  projectId: string
  challengeId?: string
  saving?: boolean
}>()

const emit = defineEmits<{
  save: [data: QuizFormData]
}>()

const toast = useToast()
const { confirm } = useConfirm()

let keySeq = 0
const nextKey = () => `q${++keySeq}`

const withKeys = (list: QuizQuestionFormData[]) =>
  list.map((question) => ({ ...question, localKey: nextKey() }))

// Title, description and image are not edited here — they are inherited from
// the challenge and would be a second, conflicting place to write the same
// copy. They stay in the form state so a save carries them through untouched.
const schema = z.object({
  name: z.string(),
  description: z.string(),
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

const questions = ref<QuizQuestionFormData[]>(
  withKeys(props.quizData?.questions ?? []),
)

const { markSaved } = useUnsavedChanges(() => ({
  ...state,
  questions: questions.value,
}))

/** The page owns the save, so only it knows when the form matches the server. */
defineExpose({ markSaved })

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
      questions.value = withKeys(data.questions)
      // What the server holds, not an edit.
      nextTick(markSaved)
    }
  },
  { once: true },
)

const editingQuestion = ref<QuizQuestionFormData | null>(null)
const isAddingQuestion = ref(false)

function addQuestion() {
  editingQuestion.value = {
    localKey: nextKey(),
    questionType: QuizQuestionType.Predefined,
    questionText: '',
    questionOrder: questions.value.length + 1,
    allowMultipleSelection: false,
    predefinedAnswers: [
      { answerText: '', isCorrect: true, answerOrder: 1 },
      { answerText: '', isCorrect: false, answerOrder: 2 },
    ],
    bettingEnabled: false,
  }
  isAddingQuestion.value = true
}

/** Only for the dialog heading, so it can say where you are in the quiz. */
const editingIndex = computed(() =>
  questions.value.findIndex(
    (question) => question.localKey === editingQuestion.value?.localKey,
  ),
)

function editQuestion(question: QuizQuestionFormData) {
  editingQuestion.value = { ...question }
  isAddingQuestion.value = false
}

function saveQuestion(question: QuizQuestionFormData) {
  if (isAddingQuestion.value) {
    questions.value.push(question)
  } else {
    // By `localKey`, not by `questionOrder`: order is mutable and renumbered on
    // delete and on reorder, so matching on it can edit the wrong question.
    const index = questions.value.findIndex(
      (q) => q.localKey === question.localKey,
    )
    if (index !== -1) questions.value[index] = question
  }
  renumber()
  editingQuestion.value = null
  isAddingQuestion.value = false
}

function renumber() {
  questions.value.forEach((question, index) => {
    question.questionOrder = index + 1
  })
}

function cancelEdit() {
  editingQuestion.value = null
  isAddingQuestion.value = false
}

async function deleteQuestion(index: number) {
  const question = questions.value[index]
  if (!question) return

  const confirmed = await confirm({
    title: 'Slette spørsmålet?',
    description: question.questionText
      ? `«${question.questionText}» fjernes fra quizen når du lagrer.`
      : 'Spørsmålet fjernes fra quizen når du lagrer.',
  })
  if (!confirmed) return

  questions.value.splice(index, 1)
  renumber()
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

// What each type is; what it is worth belongs with the points field, which
// sits in the same dialog.
const questionTypeOptions = [
  {
    value: QuizQuestionType.Predefined,
    label: 'Flervalg',
    description: 'Du lager svaralternativene, og deltakeren velger.',
  },
  {
    value: QuizQuestionType.FreeText,
    label: 'Fritekst',
    description: 'Deltakeren skriver svaret med egne ord.',
  },
  {
    value: QuizQuestionType.Number,
    label: 'Tall',
    description: 'Deltakeren svarer med et tall.',
  },
  {
    value: QuizQuestionType.Ordering,
    label: 'Rekkefølge',
    description:
      'Du legger inn elementene i riktig rekkefølge, og deltakeren drar dem på plass.',
  },
]

// One source for the labels; the list used to repeat them in a ternary chain.
const questionTypeLabel = (type: QuizQuestionType) =>
  questionTypeOptions.find((option) => option.value === type)?.label ?? 'Ukjent'

const questionPoints = computed(() =>
  questions.value.reduce((sum, question) => sum + (question.points ?? 0), 0),
)
</script>

<template>
  <UForm
    :state
    :schema="schema"
    class="max-w-3xl space-y-8"
    @submit.prevent="handleSubmit"
  >
    <AdminSection title="Spørsmål" :count="questions.length">
      <template #actions>
        <!-- Soft: "Lagre quiz" is this page's primary action, and two solid
             buttons compete for it. -->
        <UButton
          size="sm"
          variant="soft"
          icon="lucide:plus"
          @click="addQuestion"
        >
          Legg til spørsmål
        </UButton>
      </template>

      <div class="space-y-3">
        <!-- What the quiz is worth, which neither number alone answers. -->
        <p v-if="questions.length" class="text-muted text-sm">
          {{ formatNumber(questionPoints) }} poeng fra spørsmål
          <template v-if="state.completionPoints">
            + {{ formatNumber(state.completionPoints) }} for fullføring =
            <span class="text-default font-medium">
              {{ formatNumber(questionPoints + state.completionPoints) }} poeng
            </span>
          </template>
        </p>

        <VueDraggable
          v-if="questions.length"
          v-model="questions"
          handle=".drag-handle"
          ghost-class="opacity-50"
          :animation="200"
          class="divide-default divide-y"
          @end="renumber"
        >
          <div
            v-for="(question, index) in questions"
            :key="question.localKey"
            class="flex items-start gap-3 py-3"
          >
            <div
              class="drag-handle text-muted mt-1 cursor-grab active:cursor-grabbing"
            >
              <UIcon name="lucide:grip-vertical" class="size-5" />
            </div>

            <div class="min-w-0 grow">
              <p class="text-muted text-xs">
                {{ index + 1 }}. {{ questionTypeLabel(question.questionType) }}
                <template v-if="question.points">
                  &middot; {{ formatNumber(question.points) }} poeng
                </template>
              </p>
              <p class="font-medium">
                {{ question.questionText || '(Ingen spørsmålstekst)' }}
              </p>
            </div>

            <div class="flex shrink-0 gap-2">
              <UButton
                size="xs"
                variant="ghost"
                icon="lucide:pencil"
                @click="editQuestion(question)"
              >
                Rediger
              </UButton>
              <UButton
                size="xs"
                variant="ghost"
                color="error"
                icon="lucide:trash-2"
                @click="deleteQuestion(index)"
              />
            </div>
          </div>
        </VueDraggable>

        <p v-else class="text-dimmed text-sm">
          Ingen spørsmål ennå. Klikk «Legg til spørsmål» for å opprette ett.
        </p>
      </div>
    </AdminSection>

    <AdminSection title="Innstillinger">
      <div class="flex flex-col gap-6">
        <UFormField
          name="completionPoints"
          label="Fullføringspoeng"
          help="Gis til alle som fullfører quizen, uansett hvor mange svar som er riktige. Kommer i tillegg til poengene for de enkelte spørsmålene."
        >
          <UInput
            v-model.number="state.completionPoints"
            type="number"
            size="xl"
            required
            class="w-full"
          />
        </UFormField>

        <UFormField
          name="timeoutSeconds"
          label="Tidsbegrensning (sekunder)"
          hint="(valgfritt)"
          help="Gjelder hele quizen. Spørsmål kan i tillegg ha sin egen tidsbegrensning, og den strengeste av de to gjelder."
        >
          <UInput
            v-model.number="state.timeoutSeconds"
            type="number"
            size="xl"
            class="w-full"
          />
        </UFormField>

        <div class="space-y-4">
          <UFormField name="randomizeQuestions">
            <UCheckbox
              v-model="state.randomizeQuestions"
              label="Tilfeldig spørsmålsrekkefølge"
              description="Hver deltaker får sin egen rekkefølge, trukket når de starter quizen. Uten dette får alle rekkefølgen ovenfor."
            />
          </UFormField>

          <UFormField name="revealCorrectAnswers">
            <UCheckbox
              v-model="state.revealCorrectAnswers"
              label="Vis riktige svar"
              description="Brukeren ser hva som var riktig – underveis, i resultatet og når quizen leses om igjen. Uten dette får de bare en bekreftelse på at svarene er levert."
            />
          </UFormField>

          <UFormField name="allowRetakes">
            <UCheckbox
              v-model="state.allowRetakes"
              label="Tillat brukere å ta quizen på nytt"
              description="Uten dette kan hver bruker levere quizen bare én gang."
            />
          </UFormField>
        </div>
      </div>
    </AdminSection>

    <UButton type="submit" size="lg" block :loading="saving">
      Lagre quiz
    </UButton>

    <!-- A dialog, per the panel's page-or-dialog rule: editing one question is
         an action on the list you are looking at, and the list stays visible
         behind it. -->
    <!-- Not dismissible: a click outside or an Esc would throw away a
         half-written question. Avbryt and Lagre are the ways out. -->
    <UModal
      :open="!!editingQuestion"
      :dismissible="false"
      :ui="{ content: 'max-w-3xl' }"
      @update:open="(open) => !open && cancelEdit()"
    >
      <template #header>
        <h3 class="text-lg font-semibold">
          {{
            isAddingQuestion
              ? `Nytt spørsmål (${questions.length + 1} av ${questions.length + 1})`
              : `Spørsmål ${editingIndex + 1} av ${questions.length}`
          }}
        </h3>
      </template>
      <template #body>
        <AdminQuizQuestionEditor
          v-if="editingQuestion"
          :question="editingQuestion"
          :question-type-options="questionTypeOptions"
          @save="saveQuestion"
          @cancel="cancelEdit"
        />
      </template>
    </UModal>
  </UForm>
</template>
