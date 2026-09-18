<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import { z } from 'zod'

definePageMeta({
  permission: 'scores:view',
  layout: 'admin',
})

gql(`
  mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
    createScoreAdjustment(input: $input) {
      id
      points
      reason
    }
  }
`)

const route = useRoute('admin-projects-projectId-scores-new')

const schema = z.object({
  userId: z.string().min(1, 'Bruker ID er påkrevd'),
  points: z.number().int('Poengjustering må være heltall'),
  reason: z.string().optional(),
})

type Schema = z.infer<typeof schema>

const state = reactive<Schema>({
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
      // From the route rather than a picker: this form is project-scoped now.
      projectId: route.params.projectId,
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

  navigateTo({
    name: 'admin-projects-projectId-scores',
    params: { projectId: route.params.projectId },
  })
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <div>
        <UBreadcrumb
          :items="[
            {
              label: 'Poenglogg',
              to: {
                name: 'admin-projects-projectId-scores',
                params: { projectId: route.params.projectId },
              },
            },
            { label: 'Ny justering' },
          ]"
        />
      </div>
    </div>
    <div>
      <div class="max-w-2xl">
        <h1 class="mb-6 text-3xl font-bold">Opprett poengjustering</h1>

        <UForm :state :schema @submit="handleSubmit">
          <div class="space-y-6">
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
              <UButton
                variant="ghost"
                :to="{
                  name: 'admin-projects-projectId-scores',
                  params: { projectId: route.params.projectId },
                }"
              >
                Avbryt
              </UButton>
              <UButton type="submit">Opprett justering</UButton>
            </div>
          </div>
        </UForm>
      </div>
    </div>
  </div>
</template>
