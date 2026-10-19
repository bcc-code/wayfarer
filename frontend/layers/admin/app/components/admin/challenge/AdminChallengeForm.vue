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
    name: z.string().min(1, 'Tittel er påkrevd'),
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
          message: 'Må være en gyldig URL eller en lokal sti som starter med /',
        },
      )
      .optional()
      .or(z.literal('')),
    buttonText: z.string().optional(),
    publishedAt: z.string().optional(),
    endTime: z.string().optional(),
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
      message: 'Knappetekst er påkrevd',
      path: ['buttonText'],
    },
  )
  .refine(
    (data) =>
      data.type !== ChallengeType.Plugin ||
      (data.pluginChallengeId && data.pluginChallengeId.length > 0),
    {
      message: 'Plugin Challenge ID er påkrevd',
      path: ['pluginChallengeId'],
    },
  )

type Schema = z.infer<typeof schema>

const state = reactive<Schema>({
  type: props.initialData?.type ?? ChallengeType.Quiz,
  name: props.initialData?.name ?? '',
  description: props.initialData?.description,
  image: props.initialData?.image,
  url: props.initialData?.url,
  buttonText: props.initialData?.buttonText ?? '',
  publishedAt: props.initialData?.publishedAt,
  endTime: props.initialData?.endTime,
  startedAt: props.initialData?.startedAt,
  allowSelfCompletion: props.initialData?.allowSelfCompletion ?? false,
  pluginChallengeId: props.initialData?.pluginChallengeId,
  notificationText: props.initialData?.notificationText ?? '',
})

/**
 * Visibility is stored as a single nullable timestamp, which is unreadable as
 * a form field — see `challengeVisibility.ts`. Quiz challenges ignore it
 * entirely (they are shown to whoever has access to a quiz session), so the
 * control is only offered for the types it actually governs.
 *
 * A challenge being created starts visible: the stored default is an empty
 * timestamp, which means the opposite and is nobody's intent.
 */
const visibility = ref<ChallengeVisibility>(
  props.initialData
    ? visibilityFromVisibleAt(props.initialData.visibleAt)
    : 'everyone',
)
const scheduledVisibleAt = ref(
  visibility.value === 'scheduled' ? props.initialData?.visibleAt : undefined,
)

const visibilityOptions = [
  {
    value: 'everyone',
    label: 'Alle deltakere',
    description: 'Ligger i utfordringslista for alle i prosjektet.',
  },
  {
    value: 'enrolled',
    label: 'Bare de som er meldt på',
    description:
      'Skjult i lista. Deltakeren må skanne QR-koden eller få lenken.',
  },
  {
    value: 'scheduled',
    label: 'Fra et tidspunkt',
    description: 'Skjult fram til tidspunktet. Før det ser bare påmeldte den.',
  },
]

const governsVisibility = computed(() => state.type !== ChallengeType.Quiz)

const { markSaved } = useUnsavedChanges(() => ({
  ...state,
  visibility: visibility.value,
  scheduledVisibleAt: scheduledVisibleAt.value,
}))

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
      state.startedAt = data.startedAt
      visibility.value = visibilityFromVisibleAt(data.visibleAt)
      scheduledVisibleAt.value =
        visibility.value === 'scheduled' ? data.visibleAt : undefined
      // What the server holds, not an edit.
      nextTick(markSaved)
      state.allowSelfCompletion = data.allowSelfCompletion ?? false
      state.pluginChallengeId = data.pluginChallengeId
      state.notificationText = data.notificationText ?? ''
    }
  },
  { once: true },
)

// Quiz first: it is the default and by far the most common type, and the
// others read as unfamiliar jargon without the one-liner.
const challengeTypeOptions = [
  {
    value: ChallengeType.Quiz,
    label: 'Quiz',
    description: 'Spørsmål med svar.',
  },
  {
    value: ChallengeType.Simple,
    label: 'Enkel',
    description:
      'En oppgave uten innhold i appen. Deltakeren huker den av selv, eller en admin gjør det.',
  },
  {
    value: ChallengeType.External,
    label: 'Ekstern',
    description: 'Sender deltakeren videre til en annen nettside.',
  },
  {
    value: ChallengeType.Plugin,
    label: 'Plugin',
    description: 'Utfordring som styres av et annet system enn Interact.',
  },
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
    markSaved()
    emit('submit', {
      ...event.data,
      // The editor leaves an empty paragraph behind when cleared; storing it
      // would render as a gap in the app. Empty, not undefined: undefined
      // means "leave as is" on update, so a cleared description must survive
      // as a value.
      description: isBlankHtml(event.data.description)
        ? ''
        : event.data.description,
      // A quiz keeps whatever it has: the field does nothing for that type,
      // so writing to it would only churn existing data.
      visibleAt: governsVisibility.value
        ? visibleAtForVisibility(
            visibility.value,
            scheduledVisibleAt.value,
            props.initialData?.visibleAt,
          )
        : props.initialData?.visibleAt,
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
              <!-- Item descriptions are truncated to one line by default,
                   and these are sentences. -->
              <USelect
                v-model="state.type"
                :items="challengeTypeOptions"
                :disabled="isEditMode"
                :ui="{ itemDescription: 'text-clip whitespace-normal' }"
                class="w-full"
              />
            </UFormField>
            <AdminTranslatableFormField
              label="Tittel"
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
            >
              <AdminRichTextEditor
                v-model="state.description"
                content-type="html"
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

        <AdminSection title="Synlighet">
          <div class="flex flex-col gap-6">
            <UFormField
              v-if="governsVisibility"
              name="visibility"
              label="Hvem ser utfordringen"
              help="Påmeldt betyr at deltakeren har skannet QR-koden til utfordringen, åpnet lenken til den, eller blitt lagt til av en admin."
            >
              <USelect
                v-model="visibility"
                :items="visibilityOptions"
                :ui="{ itemDescription: 'text-clip whitespace-normal' }"
                class="w-full"
              />
            </UFormField>
            <UFormField
              v-if="governsVisibility && visibility === 'scheduled'"
              name="scheduledVisibleAt"
              label="Synlig fra"
            >
              <AdminDateTimeField v-model="scheduledVisibleAt" />
            </UFormField>
            <p v-if="!governsVisibility" class="text-muted text-sm">
              En quiz vises for de som har tilgang til en quiz-sesjon. Styr hvem
              som slipper til fra Sesjoner, ikke herfra.
            </p>
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
            help="Push-varselet deltakeren får når en admin melder dem på utfordringen. Tomt felt betyr at ingen varsel sendes."
          >
            <UTextarea
              v-model="state.notificationText"
              class="w-full"
              autoresize
              :rows="2"
            />
          </AdminTranslatableFormField>
        </AdminSection>

        <UButton
          :icon="isEditMode ? 'lucide:check' : 'lucide:plus'"
          type="submit"
          size="lg"
          block
          >{{ submitLabel }}</UButton
        >
      </UForm>

      <!-- Sticky: by the time you reach the timestamps the preview would
           otherwise have scrolled away. -->
      <aside class="top-6 h-fit space-y-4 @4xl:sticky">
        <AdminThemedPreview :colors="colors">
          <AdminChallengeCardPreview :challenge="state" />
        </AdminThemedPreview>

        <div class="w-[390px] max-w-full">
          <p class="text-muted mb-2 text-xs">Varselet ved påmelding</p>
          <AdminPushNotificationPreview
            :title="state.name"
            :body="state.notificationText"
            :icon="state.image"
          />
        </div>
      </aside>
    </div>
  </div>
</template>
