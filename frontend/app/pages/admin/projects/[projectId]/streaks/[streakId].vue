<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminProjectStreakPage($streakId: ID!) {
    streak(id: $streakId) {
      id
      name
      description
      status
      relevantDays {
        start
        end
      }
      translationStatus {
        ...TranslationStatus
      }
      project {
        id
        name
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-streaks-streakId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectStreakPageQuery({
  variables: {
    streakId: route.params.streakId,
  },
  pause: computed(() => !isAuthReady.value),
})

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
  relevantDays: [],
})

watch(
  () => data.value,
  (d) => {
    if (d) {
      state.name = d.streak.name
      state.description = d.streak.description
      state.relevantDays = d.streak.relevantDays.map((range) => ({
        start: range.start,
        end: range.end,
      }))
    }
  },
  { once: true },
)

function addDateRange() {
  state.relevantDays.push({ start: '', end: '' })
}

function removeDateRange(index: number) {
  state.relevantDays.splice(index, 1)
}

const { executeMutation } = useUpdateStreakMutation()
const { executeMutation: executeDelete } = useDeleteStreakMutation()
const toast = useToast()

async function updateStreak(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({ id: route.params.streakId, input: event.data }).then(
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
        title: 'Suksess',
        description: 'Streak oppdatert',
        color: 'success',
      })
      navigateTo({
        name: 'admin-projects-projectId',
        params: { projectId: route.params.projectId },
      })
    },
  )
}

async function deleteStreak() {
  const confirmed = confirm(
    `Er du sikker på at du vil slette "${state.name}"? Denne handlingen kan ikke angres.`,
  )

  if (!confirmed) {
    return
  }

  const response = await executeDelete({ id: route.params.streakId })
  if (response.error) {
    toast.add({
      title: response.error.name,
      description: response.error.message,
      color: 'error',
    })
    return
  }
  toast.add({
    title: 'Suksess',
    description: 'Streak slettet',
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
              label: 'Prosjekter',
              to: { name: 'admin-projects' },
            },
            {
              label: data?.streak.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Streaker',
            },
            {
              label: data?.streak.name ?? route.params.streakId,
              to: {
                name: 'admin-projects-projectId-streaks-streakId',
                params: {
                  projectId: route.params.projectId,
                  streakId: route.params.streakId,
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
          @submit.prevent="updateStreak"
        >
          <AdminTranslatableFormField label="Navn" :translation-status="data?.streak.translationStatus" name="name">
            <UInput v-model="state.name" size="xl" required class="w-full" />
          </AdminTranslatableFormField>
          <AdminTranslatableFormField label="Beskrivelse" :translation-status="data?.streak.translationStatus" name="description">
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
                <span class="text-sm font-medium"
                  >Datoperiode {{ index + 1 }}</span
                >
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
          <UButton type="submit" size="lg" block>Lagre endringer</UButton>
          <UButton
            color="error"
            variant="ghost"
            size="lg"
            block
            @click="deleteStreak"
          >
            Slett streak
          </UButton>
        </UForm>
      </template>
    </UContainer>
  </div>
</template>
