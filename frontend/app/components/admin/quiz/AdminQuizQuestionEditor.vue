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
      {{ question.id ? 'Edit Question' : 'New Question' }}
    </h4>

    <UFormField name="questionType" label="Question Type">
      <USelect
        v-model="localQuestion.questionType"
        :items="questionTypeOptions"
        :disabled="!!question.id"
        class="w-full"
      />
    </UFormField>

    <UFormField name="questionText" label="Question Text">
      <UTextarea
        v-model="localQuestion.questionText"
        class="w-full"
        autoresize
        required
      />
    </UFormField>

    <div class="grid grid-cols-2 gap-4">
      <UFormField name="points" label="Points">
        <UInput
          v-model.number="localQuestion.points"
          type="number"
          size="xl"
          class="w-full"
        />
      </UFormField>

      <UFormField
        name="timeoutSeconds"
        label="Timeout (seconds)"
        hint="(optional)"
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
          label="Allow multiple answers"
        />
      </UFormField>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium">Answers</label>
          <UButton size="xs" variant="ghost" @click="addAnswer">
            Add Answer
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
            placeholder="Answer text"
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
            Remove
          </UButton>
        </div>

        <p class="text-xs text-text-muted">
          Check the box next to correct answer(s)
        </p>
      </div>
    </template>

    <!-- Number Question Options -->
    <template v-if="localQuestion.questionType === QuizQuestionType.Number">
      <div class="grid grid-cols-3 gap-4">
        <UFormField name="minValue" label="Min Value">
          <UInput
            v-model.number="localQuestion.minValue"
            type="number"
            size="xl"
            class="w-full"
          />
        </UFormField>

        <UFormField name="maxValue" label="Max Value">
          <UInput
            v-model.number="localQuestion.maxValue"
            type="number"
            size="xl"
            class="w-full"
          />
        </UFormField>

        <UFormField name="stepValue" label="Step">
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
        {{ question.id ? 'Update' : 'Add' }} Question
      </UButton>
      <UButton variant="ghost" @click="emit('cancel')">Cancel</UButton>
    </div>
  </div>
</template>
