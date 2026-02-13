<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import type { QuizQuestionFormData } from './AdminQuizForm.vue'

const props = defineProps<{
  question: QuizQuestionFormData
  questionTypeOptions: { value: QuizQuestionType; label: string }[]
}>()

const emit = defineEmits<{
  save: [question: QuizQuestionFormData]
  cancel: []
}>()

const localQuestion = reactive<QuizQuestionFormData>({
  ...props.question,
  predefinedAnswers: props.question.predefinedAnswers?.map((a) => ({ ...a })),
  orderingItems: props.question.orderingItems?.map((item) => ({ ...item })),
})

// Initialize ordering items when switching to ORDERING type
watch(
  () => localQuestion.questionType,
  (newType) => {
    if (
      newType === QuizQuestionType.Ordering &&
      !localQuestion.orderingItems?.length
    ) {
      localQuestion.orderingItems = [
        { itemText: '', correctOrder: 1 },
        { itemText: '', correctOrder: 2 },
      ]
    }
  },
  { immediate: true },
)

function addAnswer() {
  if (!localQuestion.predefinedAnswers) {
    localQuestion.predefinedAnswers = []
  }
  localQuestion.predefinedAnswers.push({
    answerText: '',
    isCorrect: false,
    answerOrder: localQuestion.predefinedAnswers.length + 1,
  })
}

function removeAnswer(index: number) {
  localQuestion.predefinedAnswers?.splice(index, 1)
  // Reorder
  localQuestion.predefinedAnswers?.forEach((a, i) => {
    a.answerOrder = i + 1
  })
}

function addOrderingItem() {
  if (!localQuestion.orderingItems) {
    localQuestion.orderingItems = []
  }
  localQuestion.orderingItems.push({
    itemText: '',
    correctOrder: localQuestion.orderingItems.length + 1,
  })
}

function removeOrderingItem(index: number) {
  localQuestion.orderingItems?.splice(index, 1)
  // Reorder
  localQuestion.orderingItems?.forEach((item, i) => {
    item.correctOrder = i + 1
  })
}

function handleOrderingReorder() {
  // Update correctOrder based on new positions
  localQuestion.orderingItems?.forEach((item, i) => {
    item.correctOrder = i + 1
  })
}

function handleSave() {
  // Validate
  if (!localQuestion.questionText.trim()) {
    return
  }

  if (localQuestion.questionType === QuizQuestionType.Predefined) {
    if (
      !localQuestion.predefinedAnswers ||
      localQuestion.predefinedAnswers.length < 2
    ) {
      return
    }
    const hasCorrect = localQuestion.predefinedAnswers.some((a) => a.isCorrect)
    if (!hasCorrect) {
      return
    }
  }

  if (localQuestion.questionType === QuizQuestionType.Ordering) {
    if (
      !localQuestion.orderingItems ||
      localQuestion.orderingItems.length < 2
    ) {
      return
    }
    const hasEmptyItems = localQuestion.orderingItems.some(
      (item) => !item.itemText.trim(),
    )
    if (hasEmptyItems) {
      return
    }
  }

  // Helper to convert empty/NaN/0 to undefined for optional number fields
  const toOptionalNumber = (value: number | undefined): number | undefined => {
    if (value === undefined || value === null) return undefined
    if (typeof value === 'number' && isNaN(value)) return undefined
    return value
  }

  // Clean up data based on question type - only include relevant fields
  const cleanedQuestion: QuizQuestionFormData = {
    id: localQuestion.id,
    questionType: localQuestion.questionType,
    questionText: localQuestion.questionText,
    questionOrder: localQuestion.questionOrder,
    timeoutSeconds: toOptionalNumber(localQuestion.timeoutSeconds),
    points: localQuestion.points,
    bettingEnabled: localQuestion.bettingEnabled,
    bettingMinPercentage: toOptionalNumber(localQuestion.bettingMinPercentage),
    bettingMaxPercentage: toOptionalNumber(localQuestion.bettingMaxPercentage),
    bettingMinAbsolute: toOptionalNumber(localQuestion.bettingMinAbsolute),
    bettingMaxAbsolute: toOptionalNumber(localQuestion.bettingMaxAbsolute),
  }

  if (localQuestion.questionType === QuizQuestionType.Predefined) {
    cleanedQuestion.allowMultipleSelection =
      localQuestion.allowMultipleSelection
    cleanedQuestion.predefinedAnswers = localQuestion.predefinedAnswers
  } else if (localQuestion.questionType === QuizQuestionType.Number) {
    cleanedQuestion.minValue = localQuestion.minValue
    cleanedQuestion.maxValue = localQuestion.maxValue
    cleanedQuestion.stepValue = localQuestion.stepValue
  } else if (localQuestion.questionType === QuizQuestionType.Ordering) {
    cleanedQuestion.orderingItems = localQuestion.orderingItems
  }

  emit('save', cleanedQuestion)
}
</script>

<template>
  <div class="border border-default rounded-lg p-4 space-y-4">
    <h4 class="font-medium">
      {{ question.id ? 'Rediger spørsmål' : 'Nytt spørsmål' }}
    </h4>

    <UFormField name="questionType" label="Spørsmålstype">
      <USelect
        v-model="localQuestion.questionType"
        :items="questionTypeOptions"
        :disabled="!!question.id"
        class="w-full"
      />
    </UFormField>

    <UFormField name="questionText" label="Spørsmålstekst">
      <UTextarea
        v-model="localQuestion.questionText"
        class="w-full"
        autoresize
        required
      />
    </UFormField>

    <div class="grid grid-cols-2 gap-4">
      <UFormField name="points" label="Poeng">
        <UInput
          v-model.number="localQuestion.points"
          type="number"
          size="xl"
          class="w-full"
        />
      </UFormField>

      <UFormField
        name="timeoutSeconds"
        label="Tidsbegrensning (sekunder)"
        hint="(valgfritt)"
      >
        <UInput
          v-model.number="localQuestion.timeoutSeconds"
          type="number"
          size="xl"
          class="w-full"
        />
      </UFormField>
    </div>

    <!-- Betting Settings -->
    <div class="space-y-4 border border-default rounded-lg p-4">
      <UFormField name="bettingEnabled">
        <UCheckbox
          v-model="localQuestion.bettingEnabled"
          label="Aktiver betting"
        />
      </UFormField>

      <template v-if="localQuestion.bettingEnabled">
        <div class="space-y-4 pl-6">
          <div>
            <label class="text-sm font-medium">
              Prosent-grenser (valgfritt)
            </label>
            <div class="grid grid-cols-2 gap-4 mt-2">
              <UFormField name="bettingMinPercentage" label="Min %">
                <UInput
                  v-model.number="localQuestion.bettingMinPercentage"
                  type="number"
                  :min="0"
                  :max="100"
                  size="xl"
                  class="w-full"
                />
              </UFormField>

              <UFormField name="bettingMaxPercentage" label="Maks %">
                <UInput
                  v-model.number="localQuestion.bettingMaxPercentage"
                  type="number"
                  :min="0"
                  :max="100"
                  size="xl"
                  class="w-full"
                />
              </UFormField>
            </div>
          </div>

          <div>
            <label class="text-sm font-medium">
              Absolutte grenser (valgfritt)
            </label>
            <div class="grid grid-cols-2 gap-4 mt-2">
              <UFormField name="bettingMinAbsolute" label="Min poeng">
                <UInput
                  v-model.number="localQuestion.bettingMinAbsolute"
                  type="number"
                  :min="0"
                  size="xl"
                  class="w-full"
                />
              </UFormField>

              <UFormField name="bettingMaxAbsolute" label="Maks poeng">
                <UInput
                  v-model.number="localQuestion.bettingMaxAbsolute"
                  type="number"
                  :min="0"
                  size="xl"
                  class="w-full"
                />
              </UFormField>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Predefined Question Options -->
    <template v-if="localQuestion.questionType === QuizQuestionType.Predefined">
      <UFormField name="allowMultipleSelection">
        <UCheckbox
          v-model="localQuestion.allowMultipleSelection"
          label="Tillat flere svar"
        />
      </UFormField>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium">Svaralternativer</label>
          <UButton size="xs" variant="ghost" @click="addAnswer">
            Legg til svar
          </UButton>
        </div>

        <div
          v-for="(answer, index) in localQuestion.predefinedAnswers"
          :key="index"
          class="flex items-center gap-3"
        >
          <UCheckbox v-model="answer.isCorrect" />
          <UInput
            v-model="answer.answerText"
            placeholder="Svartekst"
            size="xl"
            class="flex-1"
          />
          <UButton
            size="xs"
            variant="ghost"
            color="error"
            :disabled="(localQuestion.predefinedAnswers?.length ?? 0) <= 2"
            @click="removeAnswer(index)"
          >
            Fjern
          </UButton>
        </div>

        <p class="text-xs text-text-muted">Kryss av for riktig(e) svar</p>
      </div>
    </template>

    <!-- Number Question Options -->
    <template v-if="localQuestion.questionType === QuizQuestionType.Number">
      <div class="grid grid-cols-3 gap-4">
        <UFormField name="minValue" label="Minimumsverdi">
          <UInput
            v-model.number="localQuestion.minValue"
            type="number"
            size="xl"
            class="w-full"
          />
        </UFormField>

        <UFormField name="maxValue" label="Maksimumsverdi">
          <UInput
            v-model.number="localQuestion.maxValue"
            type="number"
            size="xl"
            class="w-full"
          />
        </UFormField>

        <UFormField name="stepValue" label="Steg">
          <UInput
            v-model.number="localQuestion.stepValue"
            type="number"
            size="xl"
            class="w-full"
          />
        </UFormField>
      </div>
    </template>

    <!-- Ordering Question Options -->
    <template v-if="localQuestion.questionType === QuizQuestionType.Ordering">
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium">
            Elementer (i riktig rekkefølge)
          </label>
          <UButton size="xs" variant="ghost" @click="addOrderingItem">
            Legg til element
          </UButton>
        </div>

        <p class="text-xs text-text-muted">
          Dra for å endre rekkefølge. Rekkefølgen i listen er den korrekte
          rekkefølgen.
        </p>

        <VueDraggable
          v-model="localQuestion.orderingItems!"
          handle=".drag-handle"
          ghost-class="opacity-50"
          :animation="200"
          @end="handleOrderingReorder"
        >
          <div
            v-for="(item, index) in localQuestion.orderingItems"
            :key="index"
            class="flex items-center gap-3 mb-2"
          >
            <div
              class="drag-handle text-text-muted cursor-grab active:cursor-grabbing"
            >
              <UIcon name="lucide:grip-vertical" class="size-5" />
            </div>
            <span class="text-sm text-text-muted w-6">{{ index + 1 }}.</span>
            <UInput
              v-model="item.itemText"
              placeholder="Elementtekst"
              size="xl"
              class="flex-1"
            />
            <UButton
              size="xs"
              variant="ghost"
              color="error"
              :disabled="(localQuestion.orderingItems?.length ?? 0) <= 2"
              @click="removeOrderingItem(index)"
            >
              Fjern
            </UButton>
          </div>
        </VueDraggable>
      </div>
    </template>

    <div class="flex gap-3 pt-2">
      <UButton @click="handleSave">
        {{ question.id ? 'Oppdater' : 'Legg til' }} spørsmål
      </UButton>
      <UButton variant="ghost" @click="emit('cancel')">Avbryt</UButton>
    </div>
  </div>
</template>
