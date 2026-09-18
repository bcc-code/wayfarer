<script setup lang="ts">
definePageMeta({
  layout: 'admin',
})

gql(`
  query AdminHomePage($now: DateTime!) {
    me {
      id
      name
    }
    adminDashboardStats {
      totalUsers
      totalPointsAwarded
      newUsersLast7Days
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

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'God morgen'
  if (hour < 18) return 'God ettermiddag'
  return 'God kveld'
})
</script>

<template>
  <div>
    <h1 v-if="data?.me" class="my-8 text-3xl text-balance">
      {{ greeting }}, {{ data.me.name }}
    </h1>

    <AdminLoadingState v-if="fetching" />
    <AdminErrorState v-else-if="error" :error />

    <!-- Container query, not viewport: the sidebar has already taken ~300px. -->
    <div v-else-if="data" class="@container space-y-8">
      <AdminDashboardStats :stats="data.adminDashboardStats" />

      <div class="grid gap-6 @4xl:grid-cols-3">
        <div class="@4xl:col-span-2">
          <AdminRecentActivity :feedback-entries="feedbackEntries" />
        </div>
      </div>

      <section>
        <div class="mb-3 flex items-baseline gap-4">
          <h2>Aktive prosjekter</h2>
          <UButton
            v-if="currentProjects.length"
            variant="soft"
            size="xs"
            to="/admin/projects"
          >
            Se alle
          </UButton>
        </div>
        <div
          v-if="currentProjects.length"
          class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4"
        >
          <NuxtLink
            v-for="project in currentProjects"
            :key="project.id"
            class="block h-full"
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
          title="Her er det tomt"
          description="Det er ingen aktive prosjekter for øyeblikket"
          :actions="[{ label: 'Se alle prosjekter', to: '/admin/projects' }]"
        />
      </section>
    </div>
  </div>
</template>
