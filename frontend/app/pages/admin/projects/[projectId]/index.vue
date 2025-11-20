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
    achievements(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
    events(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
    challenges(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
    streaks(first: 50, filter: { projectId: $projectId }) {
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

type State = Omit<AdminProjectPageQuery['project'], 'id'>

const state = reactive<State>({
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

// Tabs state management
const params = useUrlSearchParams('history')
const fallbackTab = useLocalStorage('fallback-tab', 'events')
const tab = computed({
  get() {
    if (typeof params.tab === 'string') return params.tab
    if (fallbackTab.value) return fallbackTab.value
    return 'events'
  },
  set(tab: string) {
    params.tab = tab
    fallbackTab.value = tab
  },
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
          <div class="space-y-2">
            <NuxtImg
              v-if="state.branding.logo"
              :src="state.branding.logo"
              width="64"
              class="mb-4 rounded"
            />
            <h1 class="text-3xl">
              {{ state.name }}
            </h1>
            <p v-if="state.description" class="text-muted max-w-2xl">
              {{ state.description }}
            </p>
            <div class="mt-4">
              <UButton
                variant="soft"
                icon="lucide:pencil"
                :to="{
                  name: 'admin-projects-projectId-edit',
                  params: { projectId: route.params.projectId },
                }"
              >
                Edit project
              </UButton>
            </div>
          </div>
        </header>
        <UTabs
          v-model="tab"
          :items="[
            { value: 'events', label: 'Events', slot: 'events' },
            { value: 'challenges', label: 'Challenges', slot: 'challenges' },
            { value: 'streaks', label: 'Streaks', slot: 'streaks' },
            {
              value: 'achievements',
              label: 'Achievements',
              slot: 'achievements',
            },
          ]"
          variant="link"
        >
          <template #events>
            <UTable
              :data="data.events.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'id' },
                { accessorKey: 'name' },
                { id: 'action' },
              ]"
            >
              <template #action-cell="{ row }">
                <div class="flex gap-2">
                  <UButton
                    variant="soft"
                    :to="{
                      name: 'admin-projects-projectId-events-eventId',
                      params: {
                        projectId: route.params.projectId,
                        eventId: row.original.id,
                      },
                    }"
                  >
                    Edit
                  </UButton>
                </div>
              </template>
            </UTable>
          </template>
          <template #challenges>
            <UTable
              :data="data.challenges.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'id' },
                { accessorKey: 'name' },
                { id: 'action' },
              ]"
            >
              <template #action-cell="{ row }">
                <div class="flex gap-2">
                  <UButton
                    variant="soft"
                    :to="{
                      name: 'admin-projects-projectId-challenges-challengeId',
                      params: {
                        projectId: route.params.projectId,
                        challengeId: row.original.id,
                      },
                    }"
                  >
                    Edit
                  </UButton>
                </div>
              </template>
            </UTable>
          </template>
          <template #streaks>
            <UTable
              :data="data.streaks.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'id' },
                { accessorKey: 'name' },
                { id: 'action' },
              ]"
            >
              <template #action-cell="{ row }">
                <div class="flex gap-2">
                  <UButton
                    variant="soft"
                    :to="{
                      name: 'admin-projects-projectId-streaks-streakId',
                      params: {
                        projectId: route.params.projectId,
                        streakId: row.original.id,
                      },
                    }"
                  >
                    Edit
                  </UButton>
                </div>
              </template>
            </UTable>
          </template>
          <template #achievements>
            <UTable
              :data="data.achievements.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'id' },
                { accessorKey: 'name' },
                { id: 'action' },
              ]"
            >
              <template #action-cell="{ row }">
                <div class="flex gap-2">
                  <UButton
                    variant="soft"
                    :to="{
                      name: 'admin-projects-projectId-achievements-achievementId',
                      params: {
                        projectId: route.params.projectId,
                        achievementId: row.original.id,
                      },
                    }"
                  >
                    Edit
                  </UButton>
                </div>
              </template>
            </UTable>
          </template>
        </UTabs>
      </template>
    </UContainer>
  </div>
</template>
