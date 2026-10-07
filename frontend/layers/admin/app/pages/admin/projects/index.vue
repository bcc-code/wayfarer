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
    currentProject {
      id
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

// The project end users see. Shown here so the one that is live is findable
// without opening each project in turn.
const currentProjectId = computed(() => data.value?.currentProject.id)

// `auto-fill`, not `auto-fit`: the latter stretches a lone card across the row.
// `min()` stops the track overflowing a narrower container.
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
                <AdminProjectCard
                  :project
                  :is-current="project.id === currentProjectId"
                />
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
                <AdminProjectCard
                  :project
                  :is-current="project.id === currentProjectId"
                />
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
                <AdminProjectCard
                  :project
                  :is-current="project.id === currentProjectId"
                />
              </NuxtLink>
            </li>
          </ul>
        </section>
      </div>
    </AdminQueryState>
  </div>
</template>
