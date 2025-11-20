<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminProjectAchievementPage($projectId: ID!, $achievementId: ID!) {
    achievement(id: $achievementId) {
      id
      name
      description
    }
    project(id: $projectId) {
      id
      name
    }
  }
`)

const route = useRoute('admin-projects-projectId-achievements-achievementId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectAchievementPageQuery({
  variables: {
    projectId: route.params.projectId,
    achievementId: route.params.achievementId,
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
              label: 'Achievements',
            },
            {
              label: data?.achievement.name ?? route.params.achievementId,
              to: {
                name: 'admin-projects-projectId-achievements-achievementId',
                params: {
                  projectId: route.params.projectId,
                  achievementId: route.params.achievementId,
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
        <pre>{{ data.achievement }}</pre>
      </template>
    </UContainer>
  </div>
</template>
