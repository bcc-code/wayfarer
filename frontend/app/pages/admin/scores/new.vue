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

const schema = z.object({
  projectId: z.string().min(1, 'Prosjekt er påkrevd'),
  userId: z.string().min(1, 'Bruker ID er påkrevd'),
  points: z.number().int('Poengjustering må være heltall'),
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
      title: 'Kunne ikke opprette poengjustering',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Poengjustering opprettet',
    description: `${event.data.points >= 0 ? '+' : ''}${event.data.points} poeng`,
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
            { label: 'Poenglogg', to: { name: 'admin-scores' } },
            { label: 'Ny justering' },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <div class="max-w-2xl">
        <h1 class="mb-6 text-3xl font-bold">Opprett poengjustering</h1>

        <LoadingState v-if="fetching" />
        <UForm v-else :state :schema @submit="handleSubmit">
          <div class="space-y-6">
            <UFormField name="projectId" label="Prosjekt" required>
              <USelect
                v-model="state.projectId"
                :items="projects.map((p) => ({ label: p.name, value: p.id }))"
                placeholder="Velg et prosjekt"
                class="w-full"
              />
            </UFormField>

            <UFormField name="userId" label="Bruker ID" required>
              <UInput
                v-model="state.userId"
                placeholder="Skriv bruker ID her..."
                class="w-full"
              />
            </UFormField>

            <UFormField name="points" label="Poeng" required>
              <UInput
                v-model.number="state.points"
                type="number"
                class="w-full"
                placeholder="100"
              />
              <template #description>
                Bruk negative verdier for å trekke fra poeng
              </template>
            </UFormField>

            <UFormField name="reason" label="Grunn">
              <UTextarea
                v-model="state.reason"
                class="w-full"
                autoresize
                placeholder="Grunn for denne justeringen..."
              />
            </UFormField>

            <div class="flex justify-end gap-3 pt-4">
              <UButton variant="ghost" :to="{ name: 'admin-scores' }">
                Avbryt
              </UButton>
              <UButton type="submit">Opprett justering</UButton>
            </div>
          </div>
        </UForm>
      </div>
    </UContainer>
  </div>
</template>
