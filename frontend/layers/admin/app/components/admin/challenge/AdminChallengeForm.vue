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
}>()

const emit = defineEmits<{
  submit: [data: ChallengeFormData]
}>()

/** Lets the page react to the type, which decides what else it renders. */
const selectedType = defineModel<ChallengeType>('type')

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
      .refine(
        (val) =>
          val === '' ||
          val.startsWith('/') ||
          z.string().url().safeParse(val).success,
        {
          message: 'Must be a valid URL or a local path starting with /',
        },
      )
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

watch(
  () => state.type,
  (type) => {
    selectedType.value = type
  },
  { immediate: true },
)

const challengeTypeLabel = computed(
  () =>
    challengeTypeOptions
      .find((option) => option.value === state.type)
      ?.label.toLowerCase() ?? 'utfordringen',
)

function handleSubmit(event: FormSubmitEvent<Schema>) {
  if (event.data) {
    emit('submit', {
      ...event.data,
    })
  }
}
</script>

<template>
  <!-- `@container`, not a fixed two-column grid: the preview only earns a
       column of its own once the panel is wide enough for both. -->
  <div class="@container">
    <div class="grid gap-8 @4xl:grid-cols-[minmax(0,32rem)_minmax(0,1fr)]">
      <UForm
        :state
        :schema="schema"
        loading-auto
        class="space-y-8"
        @submit.prevent="handleSubmit"
      >
        <AdminSection title="Innhold">
          <div class="flex flex-col gap-6">
            <slot name="before-type" />
            <UFormField
              name="type"
              label="Utfordringstype"
              :help="isEditMode ? 'Typen kan ikke endres etterpå' : undefined"
            >
              <USelect
                v-model="state.type"
                :items="challengeTypeOptions"
                :disabled="isEditMode"
                class="w-full"
              />
            </UFormField>
            <AdminTranslatableFormField
              label="Navn"
              :translation-status="translationStatus"
              name="name"
            >
              <UInput v-model="state.name" size="xl" required class="w-full" />
            </AdminTranslatableFormField>
            <AdminTranslatableFormField
              label="Beskrivelse"
              :translation-status="translationStatus"
              name="description"
              hint="(valgfritt)"
              help="Støtter HTML-formatering"
            >
              <UTextarea
                v-model="state.description"
                class="w-full"
                autoresize
              />
            </AdminTranslatableFormField>
            <UFormField name="image" label="Bilde" hint="(valgfritt)">
              <AdminFileUpload v-model="state.image" />
            </UFormField>
            <AdminTranslatableFormField
              label="Knappetekst"
              :translation-status="translationStatus"
              name="buttonText"
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
            </AdminTranslatableFormField>
          </div>
        </AdminSection>

        <!-- Only ever one of these renders; the section is skipped for the
             types that have no extra field. -->
        <AdminSection
          v-if="state.type !== ChallengeType.Quiz"
          :title="`Innstillinger for ${challengeTypeLabel}`"
        >
          <UFormField
            v-if="state.type === ChallengeType.External"
            name="url"
            label="Ekstern URL"
            help="URL-en brukere vil bli sendt til"
          >
            <UInput v-model="state.url" size="xl" required class="w-full" />
          </UFormField>
          <UFormField
            v-else-if="state.type === ChallengeType.Simple"
            name="allowSelfCompletion"
            label="Selvfullføring"
          >
            <UCheckbox
              v-model="state.allowSelfCompletion"
              label="Tillat brukere å markere denne utfordringen som fullført"
            />
          </UFormField>
          <UFormField
            v-else-if="state.type === ChallengeType.Plugin"
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
        </AdminSection>

        <AdminSection title="Tidspunkt">
          <div class="flex flex-col gap-6">
            <UFormField
              name="publishedAt"
              label="Publiseringstidspunkt"
              hint="(valgfritt - standard: nå)"
              help="Når utfordringen blir tilgjengelig for brukere"
            >
              <AdminDateTimeField v-model="state.publishedAt" />
            </UFormField>
            <UFormField
              name="visibleAt"
              label="Synlig fra"
              hint="(valgfritt)"
              help="Når utfordringen blir synlig for brukere. Før dette ser bare påmeldte den."
            >
              <AdminDateTimeField v-model="state.visibleAt" />
            </UFormField>
            <UFormField
              name="endTime"
              label="Sluttid"
              hint="(valgfritt)"
              help="Når utfordringen utløper. Uten sluttid varer den ut prosjektet."
            >
              <AdminDateTimeField v-model="state.endTime" />
            </UFormField>
          </div>
        </AdminSection>

        <AdminSection title="Varsling">
          <AdminTranslatableFormField
            label="Varslingstekst"
            :translation-status="translationStatus"
            name="notificationText"
            hint="(valgfritt)"
            help="Tekst som vises i push-varsler når admin melder bruker på utfordringen. La feltet stå tomt for ingen varsling."
          >
            <UInput v-model="state.notificationText" size="xl" class="w-full" />
          </AdminTranslatableFormField>
        </AdminSection>

        <UButton type="submit" size="lg" block>{{ submitLabel }}</UButton>
      </UForm>

      <!-- Sticky: by the time you reach the timestamps the preview would
           otherwise have scrolled away. -->
      <AdminThemedPreview :colors="colors" class="top-6 h-fit @4xl:sticky">
        <AdminChallengeCardPreview :challenge="state" />
      </AdminThemedPreview>
    </div>
  </div>
</template>
