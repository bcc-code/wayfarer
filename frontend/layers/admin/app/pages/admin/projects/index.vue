<script setup lang="ts">
definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

gql(`
  query AdminProjectsPage {
    projects(first: 100, filter: { archived: false }) {
      edges {
        node {
          id
          name
          description
          endDate
          startDate
          branding {
            logoImage {
              url
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

/**
 * `auto-fill`, not `auto-fit`: a section holding a single project would
 * otherwise collapse the empty tracks and stretch that one card across the
 * whole row. `min()` keeps the track from overflowing a container narrower
 * than the 18rem floor.
 */
const PROJECT_GRID =
  'grid grid-cols-[repeat(auto-fill,minmax(min(300px,100%),1fr))] gap-4'
</script>

<template>
  <div>
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
    <AdminQueryState :fetching :error>
      <div v-if="data" class="space-y-12">
        <section v-if="currentProjects.length > 0">
          <h2 class="mb-4">Aktive prosjekter</h2>
          <ul :class="PROJECT_GRID">
            <li v-for="project in currentProjects" :key="project.id">
              <NuxtLink
                class="block h-full"
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
          <ul :class="PROJECT_GRID">
            <li v-for="project in futureProjects" :key="project.id">
              <NuxtLink
                class="block h-full"
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
          <ul :class="PROJECT_GRID">
            <li
              v-for="project in pastProjects"
              :key="project.id"
              class="opacity-50 transition-opacity hover:opacity-100"
            >
              <NuxtLink
                class="block h-full"
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
    </AdminQueryState>
  </div>
</template>
