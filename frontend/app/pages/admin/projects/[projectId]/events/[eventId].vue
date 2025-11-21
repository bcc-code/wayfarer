<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminProjectEventPage($eventId: ID!) {
		event(id: $eventId) {
			id
			name
			description
      startDate
      endDate
      parentProject {
        id
        name
      }
		}
	}
`)

const route = useRoute('admin-projects-projectId-events-eventId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectEventPageQuery({
  variables: {
    eventId: route.params.eventId,
  },
  pause: computed(() => !isAuthReady.value),
})

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

watch(
  () => data.value,
  (d) => {
    if (d) {
      state.name = d.event.name
      state.description = d.event.description
      state.startDate = d.event.startDate
      state.endDate = d.event.endDate
    }
  },
  { once: true },
)

const { executeMutation } = useUpdateEventMutation()
const { executeMutation: executeDelete } = useDeleteEventMutation()
const toast = useToast()

async function updateEvent(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({ id: route.params.eventId, input: event.data }).then(
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
        description: 'Event updated successfully',
        color: 'success',
      })
      navigateTo({
        name: 'admin-projects-projectId',
        params: { projectId: route.params.projectId },
      })
    },
  )
}

async function deleteEvent() {
  const confirmed = confirm(
    `Are you sure you want to delete "${state.name}"? This action cannot be undone.`,
  )

  if (!confirmed) {
    return
  }

  const response = await executeDelete({ id: route.params.eventId })
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
    description: 'Event deleted successfully',
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
              label: data?.event.parentProject.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Events',
            },
            {
              label: data?.event.name ?? route.params.eventId,
              to: {
                name: 'admin-projects-projectId-events-eventId',
                params: {
                  projectId: route.params.projectId,
                  eventId: route.params.eventId,
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
          @submit.prevent="updateEvent"
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
          <UButton type="submit" size="lg" block>Save changes</UButton>
          <UButton
            color="error"
            variant="ghost"
            size="lg"
            block
            @click="deleteEvent"
          >
            Delete Event
          </UButton>
        </UForm>
      </template>
    </UContainer>
  </div>
</template>
