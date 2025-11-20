<script setup lang="ts">
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
        <pre>{{ data.event }}</pre>
      </template>
    </UContainer>
  </div>
</template>
