<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import { z } from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

// Query for projects and users dropdowns
gql(`
  query AdminScoresNewPage {
    projects(first: 100) {
      edges {
        node {
          id
          name
        }
      }
    }
    users(first: 100) {
      edges {
        node {
          id
          name
          email
        }
      }
    }
  }
`)

gql(`
  mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
    createScoreAdjustment(input: $input) {
      id
      points
      reason
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, fetching } = useAdminScoresNewPageQuery({
  pause: computed(() => !isAuthReady.value),
})

const projects = computed(
  () => data.value?.projects.edges.map((edge) => edge.node) ?? [],
)

const users = computed(
  () => data.value?.users.edges.map((edge) => edge.node) ?? [],
)

const schema = z.object({
  projectId: z.string().min(1, 'Project is required'),
  userId: z.string().min(1, 'User is required'),
  points: z.number().int('Points must be a whole number'),
  reason: z.string().optional(),
})

type Schema = z.infer<typeof schema>

const state = reactive<Schema>({
  projectId: '',
  userId: '',
  points: 0,
  reason: '',
})

const { executeMutation: createAdjustment } = useCreateScoreAdjustmentMutation()
const toast = useToast()

async function handleSubmit(event: FormSubmitEvent<Schema>) {
  if (!event.data) return

  const result = await createAdjustment({
    input: {
      projectId: event.data.projectId,
      userId: event.data.userId,
      points: event.data.points,
      reason: event.data.reason || undefined,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Failed to create score adjustment',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Score adjustment created',
    description: `${event.data.points >= 0 ? '+' : ''}${event.data.points} points`,
    color: 'success',
  })

  navigateTo({ name: 'admin-scores' })
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Score Journal', to: { name: 'admin-scores' } },
            { label: 'New Adjustment' },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <div class="max-w-2xl">
        <h1 class="mb-6 text-3xl font-bold">Create Score Adjustment</h1>

        <LoadingState v-if="fetching" />
        <UForm v-else :state :schema @submit="handleSubmit">
          <div class="space-y-6">
            <UFormField name="projectId" label="Project" required>
              <USelect
                v-model="state.projectId"
                :items="
                  projects.map((p) => ({ label: p.name, value: p.id }))
                "
                placeholder="Select a project"
                class="w-full"
              />
            </UFormField>

            <UFormField name="userId" label="User" required>
              <USelect
                v-model="state.userId"
                :items="
                  users.map((u) => ({
                    label: `${u.name} (${u.email})`,
                    value: u.id,
                  }))
                "
                placeholder="Select a user"
                class="w-full"
              />
            </UFormField>

            <UFormField name="points" label="Points" required>
              <UInput
                v-model.number="state.points"
                type="number"
                class="w-full"
                placeholder="100"
              />
              <template #description>
                Use negative values to deduct points
              </template>
            </UFormField>

            <UFormField name="reason" label="Reason">
              <UTextarea
                v-model="state.reason"
                class="w-full"
                autoresize
                placeholder="Reason for this adjustment..."
              />
            </UFormField>

            <div class="flex justify-end gap-3 pt-4">
              <UButton variant="ghost" :to="{ name: 'admin-scores' }">
                Cancel
              </UButton>
              <UButton type="submit"> Create Adjustment </UButton>
            </div>
          </div>
        </UForm>
      </div>
    </UContainer>
  </div>
</template>
