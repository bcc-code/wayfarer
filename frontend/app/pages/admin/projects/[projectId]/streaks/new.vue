<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

const route = useRoute('admin-projects-projectId-streaks-new')
const toast = useToast()
const { executeMutation } = useCreateStreakMutation()

const dateRangeSchema = z.object({
  start: z.string().min(1, 'Startdato er påkrevd'),
  end: z.string().min(1, 'Sluttdato er påkrevd'),
})

const schema = z.object({
  name: z.string().min(1, 'Navn er påkrevd'),
  description: z.string().min(1, 'Beskrivelse er påkrevd'),
  relevantDays: z
    .array(dateRangeSchema)
    .min(1, 'Minst én datoperiode er påkrevd'),
})
type Schema = z.infer<typeof schema>
const state = reactive<Schema>({
  name: '',
  description: '',
  relevantDays: [{ start: '', end: '' }],
})

function addDateRange() {
  state.relevantDays.push({ start: '', end: '' })
}

function removeDateRange(index: number) {
  state.relevantDays.splice(index, 1)
}

async function createStreak(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  // Strip time portion from dates (backend expects Date scalar, not DateTime)
  const input = {
    name: state.name,
    description: state.description,
    projectId: route.params.projectId,
    relevantDays: state.relevantDays.map((range) => ({
      start: range.start.split('T')[0],
      end: range.end.split('T')[0],
    })),
  }

  executeMutation({ input }).then((response) => {
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
      title: 'Suksess',
      description: 'Streak opprettet',
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
              label: 'Prosjekter',
              to: { name: 'admin-projects' },
            },
            {
              label: route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Streaker',
            },
            {
              label: 'Ny',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <h1 class="mb-6 text-2xl font-bold">Opprett streak</h1>
      <UForm
        :state
        :schema="schema"
        loading-auto
        class="flex max-w-md flex-col gap-6"
        @submit.prevent="createStreak"
      >
        <AdminTranslatableFormField name="name" label="Navn">
          <UInput v-model="state.name" size="xl" required class="w-full" />
        </AdminTranslatableFormField>
        <AdminTranslatableFormField name="description" label="Beskrivelse">
          <UTextarea
            v-model="state.description"
            class="w-full"
            autoresize
            required
          />
        </AdminTranslatableFormField>
        <div class="flex flex-col gap-4">
          <label class="text-sm font-medium">Relevante dager</label>
          <div
            v-for="(range, index) in state.relevantDays"
            :key="index"
            class="flex flex-col gap-2 rounded-lg border border-gray-200 p-4"
          >
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">
                Datoperiode {{ index + 1 }}
              </span>
              <UButton
                v-if="state.relevantDays.length > 1"
                color="error"
                variant="ghost"
                size="xs"
                @click="removeDateRange(index)"
              >
                Fjern
              </UButton>
            </div>
            <DateRangeField
              v-model:start="range.start"
              v-model:end="range.end"
            />
          </div>
          <UButton type="button" variant="outline" @click="addDateRange">
            Legg til datoperiode
          </UButton>
        </div>
        <UButton type="submit" size="lg" block>Opprett streak</UButton>
      </UForm>
    </UContainer>
  </div>
</template>
