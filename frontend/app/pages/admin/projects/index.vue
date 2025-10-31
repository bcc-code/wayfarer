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
</script>

<template>
  <UContainer>
    <h1 class="text-3xl my-8">Projects</h1>

    <ul class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
      <template v-if="fetching">
        <USkeleton v-for="i in 3" :key="i" class="aspect-video" />
      </template>
      <template v-else-if="error">
        <p class="text-error">Error: {{ error.message }}</p>
      </template>
      <template v-else-if="data">
        <li v-for="project in data.admin.projects" :key="project.id">
          <NuxtLink
            :to="{
              name: 'admin-projects-projectId',
              params: { projectId: project.id },
            }"
          >
            <UCard
              class="aspect-video shadow"
              :style="{ '--accent': project.branding.colors.primary }"
              :ui="{
                root: project.branding.colors.primary && 'ring-(--accent)/25',
                body: [
                  project.branding.colors.primary && 'bg-(--accent)/5',
                  'h-full flex gap-2',
                ],
              }"
            >
              <div class="grow flex flex-col">
                <h3 class="font-semibold mb-2">
                  {{ project.name }}
                </h3>
                <p v-if="project.description" class="text-sm text-muted mb-2">
                  {{ project.description }}
                </p>
                <p class="text-xs font-medium text-muted mt-auto">
                  {{ formatDateRange(project.startDate, project.endDate) }}
                </p>
              </div>
              <div class="shrink-0 flex flex-col justify-between items-end">
                <NuxtImg
                  v-if="project.branding.logo"
                  :src="project.branding.logo"
                  height="32"
                  width="32"
                />
                <UBadge
                  v-if="
                    isWithinRange(
                      new Date(),
                      project.startDate,
                      project.endDate,
                    )
                  "
                  variant="outline"
                >
                  Active
                </UBadge>
              </div>
            </UCard>
          </NuxtLink>
        </li>
      </template>
    </ul>
  </UContainer>
</template>
