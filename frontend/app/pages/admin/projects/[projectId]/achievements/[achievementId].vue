<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminProjectAchievementPage($achievementId: ID!) {
    achievement(id: $achievementId) {
      id
      name
      descriptionPending
      descriptionCompleted
      imagePending
      imageCompleted
      achievedAt
      points
      hidden
      project {
        id
        name
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-achievements-achievementId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectAchievementPageQuery({
  variables: {
    achievementId: route.params.achievementId,
  },
  pause: computed(() => !isAuthReady.value),
})

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().min(1, 'Description is required'),
  image: z.string().optional(),
  points: z.number().min(0, 'Points must be at least 0'),
  hidden: z.boolean(),
})
type Schema = z.infer<typeof schema>
const state = reactive<Schema>({
  name: '',
  description: '',
  image: undefined,
  points: 0,
  hidden: false,
})

watch(
  () => data.value,
  (d) => {
    if (d) {
      state.name = d.achievement.name
      state.description = d.achievement.description
      state.image = d.achievement.image ?? undefined
      state.points = d.achievement.points
      state.hidden = d.achievement.hidden
    }
  },
  { once: true },
)

const { executeMutation } = useUpdateAchievementMutation()
const { executeMutation: executeDelete } = useDeleteAchievementMutation()
const toast = useToast()

async function updateAchievement(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({ id: route.params.achievementId, input: event.data }).then(
    (response) => {
      if (response.error) {
        toast.add({
          title: response.error.name,
          description: response.error.message,
          color: 'error',
        })
        return
      }
      if (!response.data) {
        return
      }
      toast.add({
        title: 'Success',
        description: 'Achievement updated successfully',
        color: 'success',
      })
      navigateTo({
        name: 'admin-projects-projectId',
        params: { projectId: route.params.projectId },
      })
    },
  )
}

async function deleteAchievement() {
  const confirmed = confirm(
    `Are you sure you want to delete "${state.name}"? This action cannot be undone.`,
  )

  if (!confirmed) {
    return
  }

  const response = await executeDelete({ id: route.params.achievementId })
  if (response.error) {
    toast.add({
      title: response.error.name,
      description: response.error.message,
      color: 'error',
    })
    return
  }
  toast.add({
    title: 'Success',
    description: 'Achievement deleted successfully',
    color: 'success',
  })
  navigateTo({
    name: 'admin-projects-projectId',
    params: { projectId: route.params.projectId },
  })
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: 'Projects',
              to: { name: 'admin-projects' },
            },
            {
              label: data?.achievement.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Achievements',
            },
            {
              label: data?.achievement.name ?? route.params.achievementId,
              to: {
                name: 'admin-projects-projectId-achievements-achievementId',
                params: {
                  projectId: route.params.projectId,
                  achievementId: route.params.achievementId,
                },
              },
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <template v-else-if="data">
        <UForm
          :state
          :schema="schema"
          loading-auto
          class="flex max-w-md flex-col gap-6"
          @submit.prevent="updateAchievement"
        >
          <UFormField name="name" label="Name">
            <UInput v-model="state.name" size="xl" required class="w-full" />
          </UFormField>
          <UFormField name="description" label="Description">
            <UTextarea
              v-model="state.description"
              class="w-full"
              autoresize
              required
            />
          </UFormField>
          <UFormField
            name="image"
            label="Image URL"
            hint="(optional)"
            help="URL to an image for this achievement"
          >
            <UInput v-model="state.image" size="xl" class="w-full" />
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
          <UButton type="submit" size="lg" block>Save changes</UButton>
          <UButton
            color="error"
            variant="ghost"
            size="lg"
            block
            @click="deleteAchievement"
          >
            Delete Achievement
          </UButton>
        </UForm>
      </template>
    </UContainer>
  </div>
</template>
