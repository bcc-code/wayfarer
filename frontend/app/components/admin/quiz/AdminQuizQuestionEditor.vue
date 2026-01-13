<script setup lang="ts">
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
})

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

  emit('save', { ...localQuestion })
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

        <p class="text-xs text-text-muted">
          Kryss av for riktig(e) svar
        </p>
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

    <div class="flex gap-3 pt-2">
      <UButton @click="handleSave">
        {{ question.id ? 'Oppdater' : 'Legg til' }} spørsmål
      </UButton>
      <UButton variant="ghost" @click="emit('cancel')">Avbryt</UButton>
    </div>
  </div>
</template>
