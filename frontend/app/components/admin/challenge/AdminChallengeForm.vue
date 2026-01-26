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
    pluginChallengeId?: string
  }
  quizData?: QuizFormData
  projectId?: string
  colors?: Colors
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
  buttonText?: string
  endTime?: string
  visibleAt?: string
  startedAt?: string
  allowSelfCompletion?: boolean
  pluginChallengeId?: string
  quiz?: QuizFormData
}

const schema = z
  .object({
    type: z.nativeEnum(ChallengeType),
    name: z.string().min(1, 'Name is required'),
    description: z.string().optional(),
    image: z.string().optional(),
    url: z.string().url('Must be a valid URL').optional().or(z.literal('')),
    buttonText: z.string().optional(),
    endTime: z.string().optional(),
    visibleAt: z.string().optional(),
    startedAt: z.string().optional(),
    allowSelfCompletion: z.boolean().optional(),
    pluginChallengeId: z.string().optional(),
  })
  .refine(
    (data) =>
      data.type === ChallengeType.Plugin ||
      (data.buttonText && data.buttonText.length > 0),
    {
      message: 'Button text is required',
      path: ['buttonText'],
    },
  )
  .refine(
    (data) =>
      data.type !== ChallengeType.Plugin ||
      (data.pluginChallengeId && data.pluginChallengeId.length > 0),
    {
      message: 'Plugin Challenge ID is required',
      path: ['pluginChallengeId'],
    },
  )

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
  pluginChallengeId: props.initialData?.pluginChallengeId,
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
      state.pluginChallengeId = data.pluginChallengeId
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
  { value: ChallengeType.Simple, label: 'Enkel' },
  { value: ChallengeType.External, label: 'Ekstern' },
  { value: ChallengeType.Quiz, label: 'Quiz' },
  { value: ChallengeType.Plugin, label: 'Plugin' },
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
        <UFormField name="type" label="Utfordringstype">
          <USelect
            v-model="state.type"
            :items="challengeTypeOptions"
            :disabled="isEditMode"
            class="w-full"
          />
        </UFormField>
        <UFormField name="name" label="Navn">
          <UInput v-model="state.name" size="xl" required class="w-full" />
        </UFormField>
        <UFormField
          name="description"
          label="Beskrivelse"
          hint="(valgfritt)"
          help="Støtter HTML-formatering"
        >
          <UTextarea v-model="state.description" class="w-full" autoresize />
        </UFormField>
        <UFormField name="image" label="Bilde" hint="(valgfritt)">
          <AdminFileUpload v-model="state.image" />
        </UFormField>
        <UFormField
          v-if="state.type === ChallengeType.External"
          name="url"
          label="Ekstern URL"
          help="URL-en brukere vil bli sendt til"
        >
          <UInput v-model="state.url" size="xl" required class="w-full" />
        </UFormField>
        <UFormField
          v-if="state.type === ChallengeType.Simple"
          name="allowSelfCompletion"
          label="Selvfullføring"
        >
          <UCheckbox
            v-model="state.allowSelfCompletion"
            label="Tillat brukere å markere denne utfordringen som fullført"
          />
        </UFormField>
        <UFormField
          v-if="state.type === ChallengeType.Plugin"
          name="pluginChallengeId"
          label="Plugin Challenge ID"
          help="Unik identifikator for plugin-utfordringen"
        >
          <UInput
            v-model="state.pluginChallengeId"
            size="xl"
            required
            class="w-full"
          />
        </UFormField>
        <UFormField
          name="buttonText"
          label="Knappetekst"
          :hint="
            state.type === ChallengeType.Plugin ? '(valgfritt)' : undefined
          "
        >
          <UInput
            v-model="state.buttonText"
            size="xl"
            :required="state.type !== ChallengeType.Plugin"
            class="w-full"
          />
        </UFormField>
        <UFormField
          v-if="isEditMode"
          name="visibleAt"
          label="Synlig fra"
          hint="(valgfritt)"
          help="Når denne utfordringen blir synlig for brukere"
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
          label="Startet"
          hint="(valgfritt)"
          help="Når denne utfordringen startet"
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
          label="Sluttid"
          hint="(valgfritt)"
          help="Når denne utfordringen utløper"
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
          Slett utfordring
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

    <AdminThemedPreview :colors="colors">
      <AdminChallengeCardPreview :challenge="state" />
    </AdminThemedPreview>
  </div>
</template>
