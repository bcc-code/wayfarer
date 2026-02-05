<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectsPage {
    projects(first: 100) {
      edges {
        node {
          id
          name
          description
          endDate
          startDate
          branding {
            logo
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
const { data, error, fetching } = useAdminProjectsPageQuery({
  pause: computed(() => !isAuthReady.value),
})

const { currentProjects, futureProjects, pastProjects } = useGroupedProjects(
  () => data.value?.projects.edges.map((edge) => edge.node),
)

const { canCreateProject } = usePermissions()
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-12 flex flex-col items-start gap-8">
      <h1 class="text-3xl">Prosjekter</h1>
      <UButton
        v-if="canCreateProject"
        icon="lucide:plus"
        :to="{ name: 'admin-projects-new' }"
      >
        Nytt prosjekt
      </UButton>
    </div>
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-12">
      <section v-if="currentProjects.length > 0">
        <h2 class="mb-4">Aktive prosjekter</h2>
        <ul class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          <li v-for="project in currentProjects" :key="project.id">
            <NuxtLink
              :to="{
                name: 'admin-projects-projectId',
                params: { projectId: project.id },
              }"
            >
              <AdminProjectCard :project />
            </NuxtLink>
          </li>
        </ul>
      </section>
      <section v-if="futureProjects.length > 0">
        <h2 class="mb-4">Kommende prosjekter</h2>
        <ul class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          <li v-for="project in futureProjects" :key="project.id">
            <NuxtLink
              :to="{
                name: 'admin-projects-projectId',
                params: { projectId: project.id },
              }"
            >
              <AdminProjectCard :project />
            </NuxtLink>
          </li>
        </ul>
      </section>
      <section v-if="pastProjects.length > 0">
        <h2 class="mb-4">Tidligere prosjekter</h2>
        <ul class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          <li
            v-for="project in pastProjects"
            :key="project.id"
            class="opacity-50 transition-opacity hover:opacity-100"
          >
            <NuxtLink
              :to="{
                name: 'admin-projects-projectId',
                params: { projectId: project.id },
              }"
            >
              <AdminProjectCard :project />
            </NuxtLink>
          </li>
        </ul>
      </section>
    </div>
  </UContainer>
</template>
