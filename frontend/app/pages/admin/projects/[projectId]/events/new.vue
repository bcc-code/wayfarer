<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

const route = useRoute('admin-projects-projectId-events-new')
const toast = useToast()
const { executeMutation } = useCreateEventMutation()

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().min(1, 'Description is required'),
  startDate: z.string().min(1, 'Start date is required'),
  endDate: z.string().min(1, 'End date is required'),
})
type Schema = z.infer<typeof schema>
const state = reactive<Schema>({
  name: '',
  description: '',
  startDate: '',
  endDate: '',
})

async function createEvent(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({
    projectId: route.params.projectId,
    input: event.data,
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
      description: 'Event created successfully',
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
              label: 'Events',
            },
            {
              label: 'New',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <h1 class="mb-6 text-2xl font-bold">Create Event</h1>
      <UForm
        :state
        :schema="schema"
        loading-auto
        class="flex max-w-md flex-col gap-6"
        @submit.prevent="createEvent"
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
        <DateRangeField
          v-model:start="state.startDate"
          v-model:end="state.endDate"
        />
        <UButton type="submit" size="lg" block>Create Event</UButton>
      </UForm>
    </UContainer>
  </div>
</template>
