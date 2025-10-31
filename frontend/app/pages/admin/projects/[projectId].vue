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
              <NuxtImg :src="data.admin.project.branding.logo" width="64" />
            </UFormField>
            <UFormField label="Primary Color">
              <UInput
                v-model="data.admin.project.branding.colors.primary"
                type="color"
                class="w-12"
              />
            </UFormField>
            <UFormField label="Secondary Color">
              <UInput
                v-model="data.admin.project.branding.colors.secondary"
                type="color"
                class="w-12"
              />
            </UFormField>
            <UFormField label="Tertiary Color">
              <UInput
                v-model="data.admin.project.branding.colors.tertiary"
                type="color"
                class="w-12"
              />
            </UFormField>
          </div>
        </template>
        <template #events>
          <UTable :data="data.admin.project.events" />
        </template>
        <template #challenges>
          <UTable :data="data.admin.project.challenges" />
        </template>
        <template #streaks>
          <UTable :data="data.admin.project.streaks" />
        </template>
        <template #achievements>
          <UTable
            :data="data.admin.project.achievements"
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
</template>
