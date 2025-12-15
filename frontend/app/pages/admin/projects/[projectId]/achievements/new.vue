<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectAchievementsNewPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
    }
  }
`)

const route = useRoute('admin-projects-projectId-achievements-new')
const toast = useToast()
const { isAuthReady } = useAuthReady()
const { data } = useAdminProjectAchievementsNewPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
  pause: computed(() => !isAuthReady.value),
})
const { executeMutation } = useCreateSimpleAchievementMutation()

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
  name: '',
  descriptionPending: '',
  descriptionCompleted: '',
  notificationText: '',
  imagePending: '',
  imageCompleted: '',
  points: 0,
  hidden: false,
})

async function createAchievement(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({
    input: {
      ...event.data,
      projectId: route.params.projectId,
    },
  }).then((response) => {
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
      description: 'Achievement created successfully',
      color: 'success',
    })
    navigateTo({
      name: 'admin-projects-projectId',
      params: { projectId: route.params.projectId },
    })
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
              label: data?.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Achievements',
            },
            {
              label: 'New',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <h1 class="mb-6 text-2xl font-bold">Create Achievement</h1>
      <div class="grid grid-cols-2">
        <UForm
          :state
          :schema="schema"
          loading-auto
          class="flex max-w-md flex-col gap-6"
          @submit.prevent="createAchievement"
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
          <UFormField
            name="descriptionCompleted"
            label="Description (Completed)"
          >
            <UTextarea
              v-model="state.descriptionCompleted"
              class="w-full"
              autoresize
              required
            />
          </UFormField>
          <UFormField name="notificationText" label="Notification text">
            <UTextarea
              v-model="state.notificationText"
              class="w-full"
              autoresize
              required
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
          <UButton type="submit" size="lg" block>Create Achievement</UButton>
        </UForm>
        <AdminAchievementPreview :achievement="state" />
      </div>
    </UContainer>
  </div>
</template>
