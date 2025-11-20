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
      title: 'Success',
      description: 'Streak created successfully',
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
              label: route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Streaks',
            },
            {
              label: 'New',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <h1 class="mb-6 text-2xl font-bold">Create Streak</h1>
      <UForm
        :state
        :schema="schema"
        loading-auto
        class="flex max-w-md flex-col gap-6"
        @submit.prevent="createStreak"
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
              <span class="text-sm font-medium">
                Date Range {{ index + 1 }}
              </span>
              <UButton
                v-if="state.relevantDays.length > 1"
                color="error"
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
        <UButton type="submit" size="lg" block>Create Streak</UButton>
      </UForm>
    </UContainer>
  </div>
</template>
