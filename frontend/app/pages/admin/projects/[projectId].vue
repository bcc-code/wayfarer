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
        }
      }
    }
    achievements(filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
    events(filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
    challenges(filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
    streaks(filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useAdminProjectPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
  pause: computed(() => !isAuthReady.value),
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
    <div class="border-default border-b py-2">
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
          <div class="my-8 flex flex-col gap-6">
            <div>
              <h1 class="mb-2 text-3xl">
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
            <div class="mt-4 flex flex-col gap-4">
              <UFormField label="Logo">
                <NuxtImg :src="state.branding.logo" width="64" />
              </UFormField>
              <UFormField label="Accent Color">
                <ColorPickerInput v-model="state.branding.colors.primary" />
              </UFormField>
            </div>
          </template>
          <template #events>
            <UTable :data="data.events.edges.map((e) => e.node)" />
          </template>
          <template #challenges>
            <UTable :data="data.challenges.edges.map((e) => e.node)" />
          </template>
          <template #streaks>
            <UTable :data="data.streaks.edges.map((e) => e.node)" />
          </template>
          <template #achievements>
            <UTable :data="data.achievements.edges.map((e) => e.node)" />
          </template>
        </UTabs>
      </template>
    </UContainer>
  </div>
</template>
