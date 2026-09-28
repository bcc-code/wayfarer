<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import type {
  ExternalContentType,
  TranslationStatusFragment,
} from '~/api/generated'
import z from 'zod'

type AchievementType = 'SIMPLE' | 'CONTENT' | 'STREAK' | 'QUIZ'

interface ContentItem {
  id: string
  externalContent: {
    id: string
    title?: string | null
    contentType: ExternalContentType
    source: string
    publishedAt?: string | null
  }
}

interface InitialData {
  name: string
  descriptionPending: string
  descriptionCompleted: string
  notificationText: string
  imagePending?: string
  imageCompleted?: string
  points: number
  hidden: boolean
  awardableFrom?: string
  // Content achievement
  items?: ContentItem[]
  // Streak achievement
  streakItems?: ContentItem[]
  // Quiz achievement
  quizId?: string
  minScorePercentage?: number
  requireCompletion?: boolean
}

export interface AchievementFormData {
  name: string
  descriptionPending: string
  descriptionCompleted: string
  notificationText: string
  imagePending?: string
  imageCompleted?: string
  points: number
  hidden: boolean
  awardableFrom?: string
  achievementType: AchievementType
  // Content achievement / Streak achievement
  items?: ContentItem[]
  // Quiz achievement
  quizId?: string
  minScorePercentage?: number
  requireCompletion?: boolean
}

const props = defineProps<{
  projectId: string
  initialData?: InitialData
  achievementType?: AchievementType
  translationStatus?: TranslationStatusFragment[]
  isEditMode?: boolean
  colors?: Colors
  submitLabel: string
}>()

const emit = defineEmits<{
  submit: [data: AchievementFormData]
}>()

// Common fields schema
const schema = z.object({
  name: z.string().min(1, 'Navn er påkrevd'),
  descriptionPending: z.string().min(1, 'Beskrivelse er påkrevd'),
  descriptionCompleted: z.string().min(1, 'Beskrivelse er påkrevd'),
  notificationText: z.string().min(1, 'Varslingstekst er påkrevd'),
  imagePending: z.string().optional(),
  imageCompleted: z.string().optional(),
  points: z.number().min(0, 'Poeng kan ikke være negativt'),
  hidden: z.boolean(),
  awardableFrom: z.string().optional(),
})
type Schema = z.infer<typeof schema>

// Achievement type (defaults to SIMPLE for new, or detected type for edit)
const selectedType = ref<AchievementType>(props.achievementType ?? 'SIMPLE')

// Common state
const state = reactive<Schema>({
  name: props.initialData?.name ?? '',
  descriptionPending: props.initialData?.descriptionPending ?? '',
  descriptionCompleted: props.initialData?.descriptionCompleted ?? '',
  notificationText: props.initialData?.notificationText ?? '',
  imagePending: props.initialData?.imagePending ?? '',
  imageCompleted: props.initialData?.imageCompleted ?? '',
  points: props.initialData?.points ?? 0,
  hidden: props.initialData?.hidden ?? false,
  awardableFrom: props.initialData?.awardableFrom ?? '',
})

// Type-specific state
const contentItems = ref<ContentItem[]>(props.initialData?.items ?? [])
const streakItems = ref<ContentItem[]>(props.initialData?.streakItems ?? [])
const quizId = ref<string | undefined>(props.initialData?.quizId)
const minScorePercentage = ref<number | undefined>(
  props.initialData?.minScorePercentage,
)
const requireCompletion = ref<boolean>(
  props.initialData?.requireCompletion ?? true,
)

// Update state when initialData changes (for edit mode after data loads)
watch(
  () => props.initialData,
  (data) => {
    if (data) {
      state.name = data.name
      state.descriptionPending = data.descriptionPending
      state.descriptionCompleted = data.descriptionCompleted
      state.notificationText = data.notificationText
      state.imagePending = data.imagePending
      state.imageCompleted = data.imageCompleted
      state.points = data.points
      state.hidden = data.hidden
      state.awardableFrom = data.awardableFrom ?? ''
      // Type-specific
      contentItems.value = data.items ?? []
      streakItems.value = data.streakItems ?? []
      quizId.value = data.quizId
      minScorePercentage.value = data.minScorePercentage
      requireCompletion.value = data.requireCompletion ?? true
    }
  },
  { once: true },
)

// Update type when prop changes (for edit mode)
watch(
  () => props.achievementType,
  (type) => {
    if (type) {
      selectedType.value = type
    }
  },
  { immediate: true },
)

// Validation for type-specific fields
const typeSpecificError = computed(() => {
  switch (selectedType.value) {
    case 'STREAK':
      if (streakItems.value.length === 0) {
        return 'Minst ett innholdselement må legges til'
      }
      break
    case 'QUIZ':
      if (!quizId.value) {
        return 'En quiz må velges'
      }
      break
  }
  return null
})

function handleSubmit(event: FormSubmitEvent<Schema>) {
  if (!event.data) return

  // Check type-specific validation
  if (typeSpecificError.value) {
    return
  }

  const formData: AchievementFormData = {
    ...event.data,
    achievementType: selectedType.value,
  }

  // Add type-specific fields
  switch (selectedType.value) {
    case 'CONTENT':
      formData.items = contentItems.value
      break
    case 'STREAK':
      formData.items = streakItems.value
      break
    case 'QUIZ':
      formData.quizId = quizId.value
      formData.minScorePercentage = minScorePercentage.value
      formData.requireCompletion = requireCompletion.value
      break
  }

  emit('submit', formData)
}
</script>

<template>
  <!-- `@container`, not a fixed two-column flex: the preview only earns a
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
        <!-- Only on create; in edit mode the type is a badge in the page
             header, where it cannot be mistaken for something editable. -->
        <AdminSection v-if="!isEditMode" title="Utmerkelsestype">
          <UFormField name="type">
            <AdminAchievementTypeSelector v-model="selectedType" />
          </UFormField>
        </AdminSection>

        <AdminSection title="Innhold">
          <div class="flex flex-col gap-6">
            <AdminTranslatableFormField
              label="Navn"
              :translation-status="translationStatus"
              name="name"
            >
              <UInput v-model="state.name" size="xl" required class="w-full" />
            </AdminTranslatableFormField>

            <AdminTranslatableFormField
              label="Beskrivelse (ikke oppnådd)"
              :translation-status="translationStatus"
              name="descriptionPending"
              help="Vises mens utmerkelsen står åpen."
            >
              <UTextarea
                v-model="state.descriptionPending"
                class="w-full"
                autoresize
                required
              />
            </AdminTranslatableFormField>

            <AdminTranslatableFormField
              label="Beskrivelse (oppnådd)"
              :translation-status="translationStatus"
              name="descriptionCompleted"
              help="Vises etter at deltakeren har fått utmerkelsen."
            >
              <UTextarea
                v-model="state.descriptionCompleted"
                class="w-full"
                autoresize
                required
              />
            </AdminTranslatableFormField>

            <!-- Side by side: they are two states of one image, and seeing
                 them apart made it easy to upload the same file twice. -->
            <div class="grid gap-4 @lg:grid-cols-2">
              <UFormField
                name="imagePending"
                label="Bilde (ikke oppnådd)"
                hint="(valgfritt)"
              >
                <AdminFileUpload v-model="state.imagePending" />
              </UFormField>

              <UFormField
                name="imageCompleted"
                label="Bilde (oppnådd)"
                hint="(valgfritt)"
              >
                <AdminFileUpload v-model="state.imageCompleted" />
              </UFormField>
            </div>
          </div>
        </AdminSection>

        <AdminSection title="Varsling">
          <AdminTranslatableFormField
            label="Varslingstekst"
            :translation-status="translationStatus"
            name="notificationText"
            help="Tekst som vises i push-varsler når deltakeren oppnår utmerkelsen."
          >
            <UTextarea
              v-model="state.notificationText"
              class="w-full"
              autoresize
              :rows="2"
              required
            />
          </AdminTranslatableFormField>
        </AdminSection>

        <AdminSection title="Poeng og synlighet">
          <div class="flex flex-col gap-6">
            <UFormField name="points" label="Poeng for utmerkelsen">
              <UInput
                v-model.number="state.points"
                type="number"
                size="xl"
                required
                class="w-full"
              />
            </UFormField>

            <UFormField name="hidden">
              <UCheckbox
                v-model="state.hidden"
                label="Skjult"
                description="Deltakeren ser ikke utmerkelsen før de har oppnådd den."
              />
            </UFormField>

            <UFormField
              name="awardableFrom"
              label="Tidligste tildelings-tidspunkt"
              hint="(valgfritt)"
              help="Utmerkelsen kan ikke oppnås før dette tidspunktet. La feltet stå tomt for ingen sperre."
            >
              <AdminDateTimeField v-model="state.awardableFrom" />
            </UFormField>
          </div>
        </AdminSection>

        <AdminSection
          v-if="selectedType === 'CONTENT'"
          title="Innholdselementer"
        >
          <AdminContentItemSelector v-model="contentItems" />
        </AdminSection>

        <AdminSection
          v-else-if="selectedType === 'STREAK'"
          title="Innholdselementer (med frist)"
        >
          <AdminContentItemSelector v-model="streakItems" />
          <p v-if="typeSpecificError" class="text-error mt-3 text-sm">
            {{ typeSpecificError }}
          </p>
        </AdminSection>

        <AdminSection v-else-if="selectedType === 'QUIZ'" title="Quiz">
          <AdminQuizSelector
            :project-id="projectId"
            :quiz-id="quizId"
            :min-score-percentage="minScorePercentage"
            :require-completion="requireCompletion"
            @update:quiz-id="(v) => (quizId = v)"
            @update:min-score-percentage="(v) => (minScorePercentage = v)"
            @update:require-completion="(v) => (requireCompletion = v)"
          />
          <p v-if="typeSpecificError" class="text-error mt-3 text-sm">
            {{ typeSpecificError }}
          </p>
        </AdminSection>

        <UButton type="submit" size="lg" block>{{ submitLabel }}</UButton>
      </UForm>

      <!-- Sticky: it used to scroll away before you reached the points and
           visibility fields. -->
      <AdminThemedPreview :colors="colors" class="top-6 h-fit @4xl:sticky">
        <AdminAchievementPreview :achievement="state" />
      </AdminThemedPreview>
    </div>
  </div>
</template>
