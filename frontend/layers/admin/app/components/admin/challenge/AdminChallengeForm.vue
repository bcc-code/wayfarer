<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import type { TranslationStatusFragment } from '~/api/generated'
import z from 'zod'

const props = defineProps<{
  initialData?: {
    type: ChallengeType
    name: string
    description?: string
    image?: string
    url?: string
    buttonText: string
    publishedAt?: string
    endTime?: string
    visibleAt?: string
    startedAt?: string
    allowSelfCompletion?: boolean
    pluginChallengeId?: string
    notificationText?: string
  }
  projectId?: string
  challengeId?: string
  translationStatus?: TranslationStatusFragment[]
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
  publishedAt?: string
  endTime?: string
  visibleAt?: string
  startedAt?: string
  allowSelfCompletion?: boolean
  pluginChallengeId?: string
  notificationText?: string
}

const schema = z
  .object({
    type: z.nativeEnum(ChallengeType),
    name: z.string().min(1, 'Name is required'),
    description: z.string().optional(),
    image: z.string().optional(),
    url: z
      .string()
      .refine((val) => val === '' || val.startsWith('/') || z.string().url().safeParse(val).success, {
        message: 'Must be a valid URL or a local path starting with /',
      })
      .optional()
      .or(z.literal('')),
    buttonText: z.string().optional(),
    publishedAt: z.string().optional(),
    endTime: z.string().optional(),
    visibleAt: z.string().optional(),
    startedAt: z.string().optional(),
    allowSelfCompletion: z.boolean().optional(),
    pluginChallengeId: z.string().optional(),
    notificationText: z.string().optional(),
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
  publishedAt: props.initialData?.publishedAt,
  endTime: props.initialData?.endTime,
  visibleAt: props.initialData?.visibleAt,
  startedAt: props.initialData?.startedAt,
  allowSelfCompletion: props.initialData?.allowSelfCompletion ?? false,
  pluginChallengeId: props.initialData?.pluginChallengeId,
  notificationText: props.initialData?.notificationText ?? '',
})


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
      state.publishedAt = data.publishedAt
      state.endTime = data.endTime
      state.visibleAt = data.visibleAt
      state.startedAt = data.startedAt
      state.allowSelfCompletion = data.allowSelfCompletion ?? false
      state.pluginChallengeId = data.pluginChallengeId
      state.notificationText = data.notificationText ?? ''
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

function handleSubmit(event: FormSubmitEvent<Schema>) {
  if (event.data) {
    emit('submit', {
      ...event.data,
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
        <AdminTranslatableFormField label="Navn" :translation-status="translationStatus" name="name">
          <UInput v-model="state.name" size="xl" required class="w-full" />
        </AdminTranslatableFormField>
        <AdminTranslatableFormField label="Beskrivelse" :translation-status="translationStatus" name="description" hint="(valgfritt)" help="Støtter HTML-formatering">
          <UTextarea v-model="state.description" class="w-full" autoresize />
        </AdminTranslatableFormField>
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
        <AdminTranslatableFormField
          label="Knappetekst"
          :translation-status="translationStatus"
          name="buttonText"
          :hint="state.type === ChallengeType.Plugin ? '(valgfritt)' : undefined"
        >
          <UInput
            v-model="state.buttonText"
            size="xl"
            :required="state.type !== ChallengeType.Plugin"
            class="w-full"
          />
        </AdminTranslatableFormField>
        <AdminTranslatableFormField
          label="Varslingstekst"
          :translation-status="translationStatus"
          name="notificationText"
          hint="(valgfritt)"
          help="Tekst som vises i push-varsler når admin melder bruker på utfordringen. La feltet stå tomt for ingen varsling."
        >
          <UInput
            v-model="state.notificationText"
            size="xl"
            class="w-full"
          />
        </AdminTranslatableFormField>
        <UFormField
          name="publishedAt"
          label="Publiseringstidspunkt"
          hint="(valgfritt - standard: nå)"
          help="Når utfordringen blir tilgjengelig for brukere"
        >
          <UInput
            v-model="state.publishedAt"
            type="datetime-local"
            size="xl"
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
        <!-- <UFormField
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
        </UFormField> -->
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

      <!-- Quiz section (shown when Quiz type is selected) -->
      <div v-if="state.type === ChallengeType.Quiz" class="border-t pt-6">
        <h3 class="text-lg font-semibold mb-4">Quiz</h3>
        <template v-if="isEditMode && projectId && challengeId">
          <div class="border border-default rounded-lg p-4 space-y-3">
            <p class="text-text-muted">
              Konfigurer quiz-innstillinger og spørsmål.
            </p>
            <UButton
              :to="{
                name: 'admin-projects-projectId-challenges-challengeId-quiz',
                params: { projectId, challengeId },
              }"
            >
              Rediger quiz
            </UButton>
          </div>
        </template>
        <div
          v-else
          class="border border-dashed border-default rounded-lg p-4 text-center"
        >
          <p class="text-text-muted">
            Quiz-innstillinger blir tilgjengelig etter at utfordringen er
            opprettet.
          </p>
        </div>
      </div>
    </div>

    <AdminThemedPreview :colors="colors">
      <AdminChallengeCardPreview :challenge="state" />
    </AdminThemedPreview>
  </div>
</template>
