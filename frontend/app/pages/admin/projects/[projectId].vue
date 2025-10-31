<script setup lang="ts">
import { useAdminProjectPageQuery } from '~/api/generated'

definePageMeta({
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId')

gql(`
  query AdminProjectPage($projectId: ID!) {
    admin {
      project(id: $projectId) {
        id
        name
        description
        startDate
        endDate
        branding {
          logo
          colors {
            primary
            secondary
            tertiary
          }
          rounding
        }
      }
    }
  }
`)

const { data, error, fetching } = useAdminProjectPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
})
</script>

<template>
  <UContainer>
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <header class="my-8">
        <div class="flex flex-col gap-6 mb-8">
          <NuxtImg
            v-if="data.admin.project.branding.logo"
            :src="data.admin.project.branding.logo"
            height="64"
            width="64"
            class="shrink-0 overflow-hidden"
          />
          <div>
            <h1 class="text-3xl font-bold mb-2">
              {{ data.admin.project.name }}
            </h1>
            <p
              v-if="data.admin.project.description"
              class="text-muted max-w-2xl"
            >
              {{ data.admin.project.description }}
            </p>
          </div>
        </div>

        <div class="mt-6">
          <h2 class="text-lg font-semibold mb-3">Schedule</h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <UCard>
              <div>
                <div class="text-sm text-muted mb-1">Start Date</div>
                <div class="font-medium">
                  {{ formatDate(data.admin.project.startDate) }}
                </div>
              </div>
            </UCard>

            <UCard>
              <div>
                <div class="text-sm text-muted mb-1">End Date</div>
                <div class="font-medium">
                  {{ formatDate(data.admin.project.endDate) }}
                </div>
              </div>
            </UCard>
          </div>
        </div>
      </header>
    </template>
  </UContainer>
</template>
