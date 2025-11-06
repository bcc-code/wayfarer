<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
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
    }
  }
`)

const { data, error, fetching } = useAdminProjectPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
})

const state = reactive({
  name: '',
  description: '',
  startDate: '',
  endDate: '',
  branding: {
    logo: '',
    colors: {
      primary: '',
      secondary: '',
      tertiary: '',
    },
    rounding: 0,
  },
})

watch(data, () => {
  if (data.value) {
    state.name = data.value.project.name
    state.description = data.value.project.description
    state.startDate = data.value.project.startDate
    state.endDate = data.value.project.endDate
    state.branding = data.value.project.branding
  }
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
                {{ state.name }}
              </h1>
              <p v-if="state.description" class="text-muted max-w-2xl">
                {{ state.description }}
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
                <NuxtImg :src="state.branding.logo" width="64" />
              </UFormField>
              <UFormField label="Primary Color">
                <ColorPickerField v-model="state.branding.colors.primary" />
              </UFormField>
              <UFormField label="Secondary Color">
                <ColorPickerField v-model="state.branding.colors.secondary" />
              </UFormField>
              <UFormField label="Tertiary Color">
                <ColorPickerField v-model="state.branding.colors.tertiary" />
              </UFormField>
            </div>
          </template>
          <!-- <template #events>
            <UTable :data="state.events" />
          </template>
          <template #challenges>
            <UTable :data="state.challenges" />
          </template>
          <template #streaks>
            <UTable :data="state.streaks" />
          </template>
          <template #achievements>
            <UTable
              :data="state.achievements"
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
          </template> -->
        </UTabs>
      </template>
    </UContainer>
  </div>
</template>
