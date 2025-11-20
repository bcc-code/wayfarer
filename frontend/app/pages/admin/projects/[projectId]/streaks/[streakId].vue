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
  start: z.string().min(1, 'Start date is required'),
  end: z.string().min(1, 'End date is required'),
})

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().min(1, 'Description is required'),
  relevantDays: z
    .array(dateRangeSchema)
    .min(1, 'At least one date range is required'),
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
        title: 'Success',
        description: 'Streak updated successfully',
        color: 'success',
      })
    },
  )
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
              label: data?.streak.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Streaks',
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
          <div class="flex flex-col gap-4">
            <label class="text-sm font-medium">Relevant Days</label>
            <div
              v-for="(range, index) in state.relevantDays"
              :key="index"
              class="flex flex-col gap-2 rounded-lg border border-gray-200 p-4"
            >
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium"
                  >Date Range {{ index + 1 }}</span
                >
                <UButton
                  v-if="state.relevantDays.length > 1"
                  color="red"
                  variant="ghost"
                  size="xs"
                  @click="removeDateRange(index)"
                >
                  Remove
                </UButton>
              </div>
              <DateRangeField
                v-model:start="range.start"
                v-model:end="range.end"
              />
            </div>
            <UButton type="button" variant="outline" @click="addDateRange">
              Add Date Range
            </UButton>
          </div>
          <UButton type="submit" size="lg" block>Save changes</UButton>
        </UForm>
      </template>
    </UContainer>
  </div>
</template>
