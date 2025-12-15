<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import type { QuizFormData } from '../quiz/AdminQuizForm.vue'
import z from 'zod'

const props = defineProps<{
  initialData?: {
    type: ChallengeType
    name: string
    description?: string
    image?: string
    url?: string
    buttonText: string
    endTime?: string
    visibleAt?: string
    startedAt?: string
    allowSelfCompletion?: boolean
  }
  quizData?: QuizFormData
  projectId?: string
  submitLabel: string
  isEditMode?: boolean
  onDelete?: () => void
}>()

const emit = defineEmits<{
  submit: [data: ChallengeFormData]
}>()

export interface ChallengeFormData {
  type: ChallengeType
  name: string
  description?: string
  image?: string
  url?: string
  buttonText: string
  endTime?: string
  visibleAt?: string
  startedAt?: string
  allowSelfCompletion?: boolean
  quiz?: QuizFormData
}

const schema = z.object({
  type: z.nativeEnum(ChallengeType),
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
  image: z.string().optional(),
  url: z.string().url('Must be a valid URL').optional().or(z.literal('')),
  buttonText: z.string().min(1, 'Button text is required'),
  endTime: z.string().optional(),
  visibleAt: z.string().optional(),
  startedAt: z.string().optional(),
  allowSelfCompletion: z.boolean().optional(),
})

type Schema = z.infer<typeof schema>

const state = reactive<Schema>({
  type: props.initialData?.type ?? ChallengeType.Simple,
  name: props.initialData?.name ?? '',
  description: props.initialData?.description,
  image: props.initialData?.image,
  url: props.initialData?.url,
  buttonText: props.initialData?.buttonText ?? '',
  endTime: props.initialData?.endTime,
  visibleAt: props.initialData?.visibleAt,
  startedAt: props.initialData?.startedAt,
  allowSelfCompletion: props.initialData?.allowSelfCompletion ?? false,
})

const quizFormData = ref<QuizFormData | undefined>(props.quizData)

// Update state when initialData changes (for edit mode after data loads)
watch(
  () => props.initialData,
  (data) => {
    if (data) {
      state.type = data.type
      state.name = data.name
      state.description = data.description
      state.image = data.image
      state.url = data.url
      state.buttonText = data.buttonText
      state.endTime = data.endTime
      state.visibleAt = data.visibleAt
      state.startedAt = data.startedAt
      state.allowSelfCompletion = data.allowSelfCompletion ?? false
    }
  },
  { once: true },
)

watch(
  () => props.quizData,
  (data) => {
    if (data) {
      quizFormData.value = data
    }
  },
  { once: true },
)

const challengeTypeOptions = [
  { value: ChallengeType.Simple, label: 'Simple' },
  { value: ChallengeType.External, label: 'External' },
  { value: ChallengeType.Quiz, label: 'Quiz' },
]

function handleQuizSave(data: QuizFormData) {
  quizFormData.value = data
}

function handleSubmit(event: FormSubmitEvent<Schema>) {
  if (event.data) {
    emit('submit', {
      ...event.data,
      quiz: state.type === ChallengeType.Quiz ? quizFormData.value : undefined,
    })
  }
}
</script>

<template>
  <div class="grid grid-cols-2 gap-8">
    <div class="space-y-6">
      <UForm
        :state
        :schema="schema"
        loading-auto
        class="flex max-w-md flex-col gap-6"
        @submit.prevent="handleSubmit"
      >
        <slot name="before-type" />
        <UFormField name="type" label="Challenge Type">
          <USelect
            v-model="state.type"
            :items="challengeTypeOptions"
            :disabled="isEditMode"
            class="w-full"
          />
        </UFormField>
        <UFormField name="name" label="Name">
          <UInput v-model="state.name" size="xl" required class="w-full" />
        </UFormField>
        <UFormField
          name="description"
          label="Description"
          hint="(optional)"
          help="Supports HTML formatting"
        >
          <UTextarea v-model="state.description" class="w-full" autoresize />
        </UFormField>
        <UFormField
          name="image"
          label="Image URL"
          hint="(optional)"
          help="URL to an image for this challenge"
        >
          <UInput v-model="state.image" size="xl" class="w-full" />
        </UFormField>
        <UFormField
          v-if="state.type === ChallengeType.External"
          name="url"
          label="External URL"
          help="The URL users will be redirected to"
        >
          <UInput v-model="state.url" size="xl" required class="w-full" />
        </UFormField>
        <UFormField
          v-if="state.type === ChallengeType.Simple"
          name="allowSelfCompletion"
          label="Self Completion"
        >
          <UCheckbox
            v-model="state.allowSelfCompletion"
            label="Allow users to mark this challenge as completed"
          />
        </UFormField>
        <UFormField name="buttonText" label="Button Text">
          <UInput
            v-model="state.buttonText"
            size="xl"
            required
            class="w-full"
          />
        </UFormField>
        <UFormField
          v-if="isEditMode"
          name="visibleAt"
          label="Visible At"
          hint="(optional)"
          help="When this challenge becomes visible to users"
        >
          <UInput
            v-model="state.visibleAt"
            type="datetime-local"
            size="xl"
            class="w-full"
          />
        </UFormField>
        <UFormField
          v-if="isEditMode"
          name="startedAt"
          label="Started At"
          hint="(optional)"
          help="When this challenge started"
        >
          <UInput
            v-model="state.startedAt"
            type="datetime-local"
            size="xl"
            class="w-full"
          />
        </UFormField>
        <UFormField
          name="endTime"
          label="End Time"
          hint="(optional)"
          help="When this challenge expires"
        >
          <UInput
            v-model="state.endTime"
            type="datetime-local"
            size="xl"
            class="w-full"
          />
        </UFormField>
        <UButton type="submit" size="lg" block>{{ submitLabel }}</UButton>
        <UButton
          v-if="onDelete"
          color="error"
          variant="ghost"
          size="lg"
          block
          @click="onDelete"
        >
          Delete Challenge
        </UButton>
      </UForm>

      <!-- Quiz Form (shown when Quiz type is selected) -->
      <div v-if="state.type === ChallengeType.Quiz" class="border-t pt-6">
        <AdminQuizForm
          :quiz-data="quizFormData"
          :project-id="projectId ?? ''"
          @save="handleQuizSave"
        />
      </div>
    </div>

    <AdminChallengeCardPreview :challenge="state" />
  </div>
</template>
