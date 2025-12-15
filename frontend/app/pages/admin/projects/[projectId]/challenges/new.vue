<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectChallengeNewPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
    }
    events(first: 100, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-challenges-new')
const toast = useToast()
const { executeMutation } = useCreateChallengeMutation()

const { isAuthReady } = useAuthReady()
const { data } = useAdminProjectChallengeNewPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
  pause: computed(() => !isAuthReady.value),
})

const schema = z.object({
  eventId: z.string().min(1, 'Event is required'),
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
  image: z.string().optional(),
  url: z.string().url('Must be a valid URL').optional().or(z.literal('')),
  buttonText: z.string().min(1, 'Button text is required'),
  endTime: z.string().optional(),
})
type Schema = z.infer<typeof schema>
const state = reactive<Schema>({
  eventId: '',
  name: '',
  description: undefined,
  image: undefined,
  url: undefined,
  buttonText: '',
  endTime: undefined,
})

async function createChallenge(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  const { eventId, ...input } = event.data

  executeMutation({
    projectId: route.params.projectId,
    eventId,
    input,
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
      description: 'Challenge created successfully',
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
              label: 'Challenges',
            },
            {
              label: 'New',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <h1 class="mb-6 text-2xl font-bold">Create Challenge</h1>
      <div class="grid grid-cols-2">
        <UForm
          :state
          :schema="schema"
          loading-auto
          class="flex max-w-md flex-col gap-6"
          @submit.prevent="createChallenge"
        >
        <UFormField name="eventId" label="Event">
          <USelect
            v-model="state.eventId"
            :items="
              data?.events.edges.map((e) => ({
                value: e.node.id,
                label: e.node.name,
              })) || []
            "
            required
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
          name="url"
          label="Challenge URL"
          hint="(optional)"
          help="External link for the challenge"
        >
          <UInput v-model="state.url" size="xl" class="w-full" />
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
        <UButton type="submit" size="lg" block>Create Challenge</UButton>
        </UForm>
        <AdminChallengeCardPreview :challenge="state" />
      </div>
    </UContainer>
  </div>
</template>
