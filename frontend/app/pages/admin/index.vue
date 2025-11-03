<script setup lang="ts">
import { useAdminHomePageQuery } from '~/api/generated'

definePageMeta({
  layout: 'admin',
})

gql(`
  query AdminHomePage {
    admin {
      projects {
        id
        name
        description
        endDate
        startDate
        branding {
          logo
          rounding
          colors {
            primary
            secondary
            tertiary
          }
        }
      }
    }
  }
`)

const { data } = useAdminHomePageQuery()
const currentProjects = computed(() => {
  return data.value?.admin.projects.filter(
    (project) =>
      new Date(project.startDate) < new Date() &&
      (!project.endDate || new Date(project.endDate) > new Date()),
  )
})

const { user } = useAuth()

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 18) return 'Good afternoon'
  return 'Good evening'
})
</script>

<template>
  <UContainer>
    <h1 v-if="user" class="text-3xl my-12 text-balance">
      {{ greeting }}, {{ user.name }}
    </h1>

    <section v-if="data?.admin.projects">
      <div class="flex items-baseline gap-4 mb-3">
        <h2>Current Projects</h2>
        <UButton variant="soft" size="xs" to="/admin/projects">See all</UButton>
      </div>
      <div class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4">
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
    </section>
  </UContainer>
</template>
