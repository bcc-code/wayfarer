<script setup lang="ts">
import { useAdminProjectPageQuery } from '~/api/generated'

definePageMeta({
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId')

gql(`
  query AdminProjectPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
      description
      startDate
      endDate
      branding {
        logo
        rounding
        colors {
          primary
          secondary
          tertiary
        }
      }
      achievements {
        id
        name
        description
        image
      }
      challenges {
        id
        name
        description
        image
        url
        buttonText
        publishedAt
        endTime
      }
      events {
        id
        name
        description
        startDate
        endDate
      }
      streaks {
        id
        name
        description
        relevantDays {
          start
          end
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
  <div>
    <div class="py-2 border-b border-default">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: 'Projects',
              to: { name: 'admin-projects' },
            },
            {
              label: data?.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error class="h-[600px]" />
      <template v-else-if="data">
        <header class="my-12">
          <div class="flex flex-col gap-6 my-8">
            <div>
              <h1 class="text-3xl mb-2">
                {{ data.project.name }}
              </h1>
              <p v-if="data.project.description" class="text-muted max-w-2xl">
                {{ data.project.description }}
              </p>
            </div>
          </div>
        </header>
        <UTabs
          :items="[
            { label: 'Branding', slot: 'branding' },
            { label: 'Events', slot: 'events' },
            { label: 'Challenges', slot: 'challenges' },
            { label: 'Streaks', slot: 'streaks' },
            { label: 'Achievements', slot: 'achievements' },
          ]"
          variant="link"
        >
          <template #branding>
            <div class="flex gap-4 mt-4 flex-col">
              <UFormField label="Logo">
                <NuxtImg :src="data.project.branding.logo" width="64" />
              </UFormField>
              <UFormField label="Primary Color">
                <UInput
                  v-model="data.project.branding.colors.primary"
                  type="color"
                  class="w-12"
                />
              </UFormField>
              <UFormField label="Secondary Color">
                <UInput
                  v-model="data.project.branding.colors.secondary"
                  type="color"
                  class="w-12"
                />
              </UFormField>
              <UFormField label="Tertiary Color">
                <UInput
                  v-model="data.project.branding.colors.tertiary"
                  type="color"
                  class="w-12"
                />
              </UFormField>
            </div>
          </template>
          <template #events>
            <UTable :data="data.project.events" />
          </template>
          <template #challenges>
            <UTable :data="data.project.challenges" />
          </template>
          <template #streaks>
            <UTable :data="data.project.streaks" />
          </template>
          <template #achievements>
            <UTable
              :data="data.project.achievements"
              :columns="[
                { accessorKey: 'image', header: 'Image' },
                { accessorKey: 'name', header: 'Name' },
                { accessorKey: 'description', header: 'Description' },
              ]"
            >
              <template #image-cell="{ row }">
                <NuxtImg
                  :src="row.getValue('image')"
                  height="64"
                  width="64"
                  class="shrink-0 overflow-hidden rounded size-10"
                />
              </template>
            </UTable>
          </template>
        </UTabs>
      </template>
    </UContainer>
  </div>
</template>
