<script setup lang="ts">
definePageMeta({
  layout: 'admin',
})

gql(`
  query AdminProjectsPage {
    projects {
      id
      name
      description
      endDate
      startDate
      branding {
        logo
        colors {
          primary
        }
      }
    }
  }
`)

const { data, error, fetching } = useAdminProjectsPageQuery()

const { currentProjects, futureProjects, pastProjects } = useGroupedProjects(
  () => data.value?.projects,
)
</script>

<template>
  <UContainer>
    <div class="flex flex-col items-start gap-8 my-12">
      <h1 class="text-3xl">Projects</h1>
      <UButton icon="lucide:plus" :to="{ name: 'admin-projects-new' }">
        New Project
      </UButton>
    </div>
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-12">
      <section v-if="currentProjects.length > 0">
        <h2 class="mb-4">Current Projects</h2>
        <ul class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
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
        <h2 class="mb-4">Upcoming Projects</h2>
        <ul class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
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
        <h2 class="mb-4">Past Projects</h2>
        <ul class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          <li
            v-for="project in pastProjects"
            :key="project.id"
            class="opacity-50 hover:opacity-100 transition-opacity"
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
