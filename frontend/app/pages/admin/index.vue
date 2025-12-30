<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminHomePage($now: DateTime!) {
    me {
      id
      name
    }
    adminDashboardStats {
      totalUsers
      totalProjects
      totalChallenges
      totalPointsAwarded
      newUsersLast7Days
      activeProjectsCount
    }
    feedback(first: 5) {
      edges {
        node {
          id
          message
          createdAt
          user {
            id
            name
          }
        }
      }
    }
    projects(filter: { endDateAfter: $now }) {
      edges {
        node {
          id
          name
          description
          endDate
          startDate
          branding {
            logo
            rounding
            colors {
              light {
                accent
              }
              dark {
                accent
              }
            }
          }
          leaderboard(entityType: PERSONS, first: 5) {
            totalCount
            edges {
              node {
                id
                name
                score
                rank
                image
              }
            }
          }
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminHomePageQuery({
  variables: { now: new Date().toISOString() },
  pause: computed(() => !isAuthReady.value),
})
const { currentProjects } = useGroupedProjects(() =>
  data.value?.projects.edges.map((edge) => edge.node),
)

const feedbackEntries = computed(
  () => data.value?.feedback.edges.map((edge) => edge.node) ?? [],
)

const topProject = computed(() => currentProjects.value[0])

const leaderboardEntries = computed(
  () => topProject.value?.leaderboard.edges.map((edge) => edge.node) ?? [],
)

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 18) return 'Good afternoon'
  return 'Good evening'
})
</script>

<template>
  <UContainer>
    <h1 v-if="data?.me" class="my-8 text-3xl text-balance">
      {{ greeting }}, {{ data.me.name }}
    </h1>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />

    <div v-else-if="data" class="space-y-8">
      <AdminDashboardStats :stats="data.adminDashboardStats" />

      <div class="grid gap-6 lg:grid-cols-3">
        <div class="lg:col-span-2">
          <AdminRecentActivity :feedback-entries="feedbackEntries" />
        </div>

        <div v-if="topProject">
          <AdminTopPerformers
            :project-id="topProject.id"
            :project-name="topProject.name"
            :entries="leaderboardEntries"
            :total-count="topProject.leaderboard.totalCount"
          />
        </div>
      </div>

      <section>
        <div class="mb-3 flex items-baseline gap-4">
          <h2>Current Projects</h2>
          <UButton
            v-if="currentProjects.length"
            variant="soft"
            size="xs"
            to="/admin/projects"
          >
            See all
          </UButton>
        </div>
        <div
          v-if="currentProjects.length"
          class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4"
        >
          <NuxtLink
            v-for="project in currentProjects"
            :key="project.id"
            :to="{
              name: 'admin-projects-projectId',
              params: { projectId: project.id },
            }"
          >
            <AdminProjectCard :project />
          </NuxtLink>
        </div>
        <UEmpty
          v-else
          icon="lucide:square-dashed-mouse-pointer"
          title="A whole lotta nothin'"
          description="There are no projects currently running"
          :actions="[{ label: 'See all projects', to: '/admin/projects' }]"
        />
      </section>
    </div>
  </UContainer>
</template>
