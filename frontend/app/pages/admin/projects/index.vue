<script setup lang="ts">
import { useAdminProjectsPageQuery } from '~/api/generated'

definePageMeta({
  layout: 'admin',
})

gql(`
query AdminProjectsPage {
  admin {
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
}
`)

const { data, error, fetching } = useAdminProjectsPageQuery()

const currentProjects = computed(() => {
  if (!data.value?.admin.projects) return []
  return data.value.admin.projects.filter((project) =>
    isWithinRange(new Date(), project.startDate, project.endDate),
  )
})

const futureProjects = computed(() => {
  if (!data.value?.admin.projects) return []
  return data.value.admin.projects.filter((project) => {
    const now = new Date()
    const startDate = new Date(project.startDate)
    return startDate > now
  })
})

const pastProjects = computed(() => {
  if (!data.value?.admin.projects) return []
  return data.value.admin.projects.filter((project) => {
    const now = new Date()
    const endDate = new Date(project.endDate)
    return (
      endDate < now && !isWithinRange(now, project.startDate, project.endDate)
    )
  })
})
</script>

<template>
  <UContainer>
    <h1 class="text-3xl my-8">Projects</h1>

    <template v-if="fetching">
      <ul class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        <USkeleton v-for="i in 3" :key="i" class="aspect-video" />
      </ul>
    </template>
    <template v-else-if="error">
      <p class="text-error">Error: {{ error.message }}</p>
    </template>
    <div v-else-if="data" class="space-y-12">
      <!-- Current Projects -->
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

      <!-- Future Projects -->
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

      <!-- Past Projects -->
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
