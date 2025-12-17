<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

const props = defineProps<{
  initialData?: {
    name: string
    descriptionPending: string
    descriptionCompleted: string
    notificationText: string
    imagePending?: string
    imageCompleted?: string
    points: number
    hidden: boolean
  }
  colors?: Colors
  submitLabel: string
  onDelete?: () => void
}>()

const emit = defineEmits<{
  submit: [data: Schema]
}>()

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  descriptionPending: z.string().min(1, 'Description is required'),
  descriptionCompleted: z.string().min(1, 'Description is required'),
  notificationText: z.string().min(1, 'Notification text is required'),
  imagePending: z.string().optional(),
  imageCompleted: z.string().optional(),
  points: z.number().min(0, 'Points must be at least 0'),
  hidden: z.boolean(),
})
type Schema = z.infer<typeof schema>

const state = reactive<Schema>({
  name: props.initialData?.name ?? '',
  descriptionPending: props.initialData?.descriptionPending ?? '',
  descriptionCompleted: props.initialData?.descriptionCompleted ?? '',
  notificationText: props.initialData?.notificationText ?? '',
  imagePending: props.initialData?.imagePending ?? '',
  imageCompleted: props.initialData?.imageCompleted ?? '',
  points: props.initialData?.points ?? 0,
  hidden: props.initialData?.hidden ?? false,
})

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
    }
  },
  { once: true },
)

function handleSubmit(event: FormSubmitEvent<Schema>) {
  if (event.data) {
    emit('submit', event.data)
  }
}
</script>

<template>
  <div class="grid grid-cols-2">
    <UForm
      :state
      :schema="schema"
      loading-auto
      class="flex max-w-md flex-col gap-6"
      @submit.prevent="handleSubmit"
    >
      <UFormField name="name" label="Name">
        <UInput v-model="state.name" size="xl" required class="w-full" />
      </UFormField>
      <UFormField name="descriptionPending" label="Description (Pending)">
        <UTextarea
          v-model="state.descriptionPending"
          class="w-full"
          autoresize
          required
        />
      </UFormField>
      <UFormField name="descriptionCompleted" label="Description (Completed)">
        <UTextarea
          v-model="state.descriptionCompleted"
          class="w-full"
          autoresize
          required
        />
      </UFormField>
      <UFormField
        name="notificationText"
        label="Notification Text"
        help="Text shown in push notifications when user earns this achievement"
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
        label="Image URL (Pending)"
        hint="(optional)"
        help="URL to an image for this achievement"
      >
        <UInput v-model="state.imagePending" size="xl" class="w-full" />
      </UFormField>
      <UFormField
        name="imageCompleted"
        label="Image URL (Completed)"
        hint="(optional)"
        help="URL to an image for this achievement"
      >
        <UInput v-model="state.imageCompleted" size="xl" class="w-full" />
      </UFormField>
      <UFormField name="points" label="Points">
        <UInput
          v-model.number="state.points"
          type="number"
          size="xl"
          required
          class="w-full"
        />
      </UFormField>
      <UFormField name="hidden" label="Hidden">
        <UCheckbox
          v-model="state.hidden"
          label="Hide this achievement from users until they earn it"
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
        Delete Achievement
      </UButton>
    </UForm>
    <AdminThemedPreview :colors="colors">
      <AdminAchievementPreview :achievement="state" />
    </AdminThemedPreview>
  </div>
</template>
