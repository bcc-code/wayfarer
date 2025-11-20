<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminProjectStreakPage($projectId: ID!, $streakId: ID!) {
    streak(id: $streakId) {
      id
      name
      description
      status
      relevantDays {
        start
        end
      }
    }
    project(id: $projectId) {
      id
      name
    }
  }
`)

const route = useRoute('admin-projects-projectId-streaks-streakId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectStreakPageQuery({
  variables: {
    projectId: route.params.projectId,
    streakId: route.params.streakId,
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
              label: data?.project.name ?? route.params.projectId,
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
        <pre>{{ data.streak }}</pre>
      </template>
    </UContainer>
  </div>
</template>
