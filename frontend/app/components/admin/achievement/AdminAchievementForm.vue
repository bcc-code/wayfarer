<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import type { ExternalContentType } from '~/api/generated'
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
  streakId?: string
  neededStreak?: number
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
  // Content achievement
  items?: ContentItem[]
  // Streak achievement
  streakId?: string
  neededStreak?: number
  // Quiz achievement
  quizId?: string
  minScorePercentage?: number
  requireCompletion?: boolean
}

const props = defineProps<{
  projectId: string
  initialData?: InitialData
  achievementType?: AchievementType
  isEditMode?: boolean
  colors?: Colors
  submitLabel: string
  onDelete?: () => void
}>()

const emit = defineEmits<{
  submit: [data: AchievementFormData]
}>()

// Common fields schema
const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  descriptionPending: z.string().min(1, 'Description is required'),
  descriptionCompleted: z.string().min(1, 'Description is required'),
  notificationText: z.string().min(1, 'Notification text is required'),
  imagePending: z.string().optional(),
  imageCompleted: z.string().optional(),
  points: z.number().min(0, 'Points must be at least 0'),
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
const streakId = ref<string | undefined>(props.initialData?.streakId)
const neededStreak = ref<number>(props.initialData?.neededStreak ?? 7)
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
      streakId.value = data.streakId
      neededStreak.value = data.neededStreak ?? 7
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
      if (!streakId.value) {
        return 'En streak må velges'
      }
      if (neededStreak.value < 1) {
        return 'Antall påkrevde dager må være minst 1'
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
      formData.streakId = streakId.value
      formData.neededStreak = neededStreak.value
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
  <div class="flex gap-8">
    <UForm
      :state
      :schema="schema"
      loading-auto
      class="flex flex-col gap-8 grow"
      @submit.prevent="handleSubmit"
    >
      <!-- Type Selector (only in create mode) -->
      <UFormField v-if="!isEditMode" name="type" label="Utmerkelsestype">
        <AdminAchievementTypeSelector
          v-model="selectedType"
          :disabled="isEditMode"
        />
      </UFormField>

      <!-- Type indicator in edit mode -->
      <div v-else class="text-muted text-sm">
        <span class="font-medium">Type:</span>
        {{
          selectedType === 'SIMPLE'
            ? 'Enkel'
            : selectedType === 'CONTENT'
              ? 'Innhold'
              : selectedType === 'STREAK'
                ? 'Streak'
                : 'Quiz'
        }}
        utmerkelse
      </div>

      <!-- Common Fields -->
      <UFormField name="name" label="Navn">
        <UInput v-model="state.name" size="xl" required class="w-full" />
      </UFormField>

      <UFormField name="descriptionPending" label="Beskrivelse (ikke oppnådd)">
        <UTextarea
          v-model="state.descriptionPending"
          class="w-full"
          autoresize
          required
        />
      </UFormField>

      <UFormField name="descriptionCompleted" label="Beskrivelse (oppnådd)">
        <UTextarea
          v-model="state.descriptionCompleted"
          class="w-full"
          autoresize
          required
        />
      </UFormField>

      <UFormField
        name="notificationText"
        label="Varslingstekst"
        help="Tekst som vises i push-varsler når brukere oppnår denne utmerkelsen"
      >
        <UInput
          v-model="state.notificationText"
          size="xl"
          required
          class="w-full"
        />
      </UFormField>

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

      <UFormField name="points" label="Poeng for utmerkelsen">
        <UInput
          v-model.number="state.points"
          type="number"
          size="xl"
          required
          class="w-full"
        />
      </UFormField>

      <UFormField name="hidden" label="Skjult">
        <UCheckbox
          v-model="state.hidden"
          label="Skjul denne utmerkelsen fra brukere frem til de oppnår den"
        />
      </UFormField>

      <UFormField
        name="awardableFrom"
        label="Tidligste tildelings-tidspunkt"
        hint="(valgfritt)"
        description="Utmerkelsen kan tidligst tildeles fra dette tidspunktet"
      >
        <UInput
          v-model="state.awardableFrom"
          type="datetime-local"
          size="xl"
          class="w-full"
        />
      </UFormField>

      <!-- Type-specific sections -->
      <template v-if="selectedType === 'CONTENT'">
        <div class="border-default border-t pt-6">
          <h3 class="mb-4 font-medium">Innholdselementer</h3>
          <AdminContentItemSelector v-model="contentItems" />
        </div>
      </template>

      <template v-else-if="selectedType === 'STREAK'">
        <div class="border-default border-t pt-6">
          <h3 class="mb-4 font-medium">Streak-konfigurasjon</h3>
          <AdminStreakSelector
            :project-id="projectId"
            :streak-id="streakId"
            :needed-streak="neededStreak"
            @update:streak-id="(v) => (streakId = v)"
            @update:needed-streak="(v) => (neededStreak = v)"
          />
        </div>
      </template>

      <template v-else-if="selectedType === 'QUIZ'">
        <div class="border-default border-t pt-6">
          <h3 class="mb-4 font-medium">Quiz-konfigurasjon</h3>
          <AdminQuizSelector
            :project-id="projectId"
            :quiz-id="quizId"
            :min-score-percentage="minScorePercentage"
            :require-completion="requireCompletion"
            @update:quiz-id="(v) => (quizId = v)"
            @update:min-score-percentage="(v) => (minScorePercentage = v)"
            @update:require-completion="(v) => (requireCompletion = v)"
          />
        </div>
      </template>

      <!-- Type-specific validation error -->
      <div v-if="typeSpecificError" class="text-error text-sm">
        {{ typeSpecificError }}
      </div>

      <UButton type="submit" size="lg" block>{{ submitLabel }}</UButton>
      <UButton
        v-if="onDelete"
        color="error"
        variant="ghost"
        size="lg"
        block
        @click="onDelete"
      >
        Slett utmerkelse
      </UButton>
    </UForm>

    <AdminThemedPreview :colors="colors">
      <AdminAchievementPreview :achievement="state" />
    </AdminThemedPreview>
  </div>
</template>
