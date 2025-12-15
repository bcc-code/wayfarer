<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminProjectChallengePage($challengeId: ID!) {
    challenge(id: $challengeId) {
      id
      name
      description
      image
      buttonText
      publishedAt
      endTime
      project {
        id
        name
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-challenges-challengeId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectChallengePageQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
})

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
  image: z.string().optional(),
  url: z.url('Must be a valid URL').optional().or(z.literal('')),
  buttonText: z.string().min(1, 'Button text is required'),
  endTime: z.string().optional(),
  publishedAt: z.string().optional(),
})
type Schema = z.infer<typeof schema>
const state = reactive<Schema>({
  name: '',
  description: undefined,
  image: undefined,
  url: undefined,
  buttonText: '',
  endTime: undefined,
  publishedAt: undefined,
})

watch(
  () => data.value,
  (d) => {
    if (d) {
      state.name = d.challenge.name
      state.description = d.challenge.description ?? undefined
      state.image = d.challenge.image ?? undefined
      state.url = d.challenge.url ?? undefined
      state.buttonText = d.challenge.buttonText
      state.endTime = d.challenge.endTime?.split('T')[0] ?? undefined
      state.publishedAt = d.challenge.publishedAt?.split('T')[0] ?? undefined
    }
  },
  { once: true },
)

const { executeMutation } = useUpdateChallengeMutation()
const { executeMutation: executeDelete } = useDeleteChallengeMutation()
const toast = useToast()

async function updateChallenge(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({ id: route.params.challengeId, input: event.data }).then(
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
        description: 'Challenge updated successfully',
        color: 'success',
      })
      navigateTo({
        name: 'admin-projects-projectId',
        params: { projectId: route.params.projectId },
      })
    },
  )
}

async function deleteChallenge() {
  const confirmed = confirm(
    `Are you sure you want to delete "${state.name}"? This action cannot be undone.`,
  )

  if (!confirmed) {
    return
  }

  const response = await executeDelete({ id: route.params.challengeId })
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
    description: 'Challenge deleted successfully',
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
              label: data?.challenge.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Challenges',
            },
            {
              label: data?.challenge.name ?? route.params.challengeId,
              to: {
                name: 'admin-projects-projectId-challenges-challengeId',
                params: {
                  projectId: route.params.projectId,
                  challengeId: route.params.challengeId,
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
        <div class="grid grid-cols-2">
          <UForm
            :state
            :schema="schema"
            loading-auto
            class="flex max-w-md flex-col gap-6"
            @submit.prevent="updateChallenge"
          >
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
            name="publishedAt"
            label="Published At"
            hint="(optional)"
            help="When this challenge is published"
          >
            <UInput
              v-model="state.publishedAt"
              type="date"
              size="xl"
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
              type="date"
              size="xl"
              class="w-full"
            />
          </UFormField>
          <UButton type="submit" size="lg" block>Save changes</UButton>
          <UButton
            color="error"
            variant="ghost"
            size="lg"
            block
            @click="deleteChallenge"
          >
            Delete Challenge
          </UButton>
        </UForm>
          <AdminChallengeCardPreview :challenge="state" />
        </div>
      </template>
    </UContainer>
  </div>
</template>
