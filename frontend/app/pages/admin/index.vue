<script setup lang="ts">
import { useAdminHomePageQuery } from '~/api/generated'

definePageMeta({
  layout: 'admin',
})

gql(`
  query AdminHomePage {
    me {
      id
      name
    }
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
`)

const { data, fetching, error } = useAdminHomePageQuery()
const { currentProjects } = useGroupedProjects(() => data.value?.projects)

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 18) return 'Good afternoon'
  return 'Good evening'
})
</script>

<template>
  <UContainer>
    <h1 v-if="data?.me" class="text-3xl my-12 text-balance">
      {{ greeting }}, {{ data.me.name }}
    </h1>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />

    <section v-else-if="data">
      <div class="flex items-baseline gap-4 mb-3">
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
  </UContainer>
</template>
