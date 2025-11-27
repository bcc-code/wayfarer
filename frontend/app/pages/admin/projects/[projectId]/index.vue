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
          light {
            accent
          }
          dark {
            accent
          }
        }
      }
    }
    achievements(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          description
          image
          points
          hidden
        }
      }
    }
    events(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          description
        }
      }
    }
    challenges(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          description
          image
        }
      }
    }
    streaks(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          description
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

const { copy } = useClipboard()
const toast = useToast()

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
            <div class="my-2">
              <UButton
                icon="lucide:plus"
                :to="{
                  name: 'admin-projects-projectId-events-new',
                  params: { projectId: route.params.projectId },
                }"
              >
                Create Event
              </UButton>
            </div>
            <UTable
              :data="data.events.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'name' },
                { accessorKey: 'description' },
                { id: 'actions' },
              ]"
            >
              <template #actions-cell="{ row }">
                <div class="flex justify-end gap-2">
                  <UDropdownMenu
                    :items="[
                      {
                        label: 'Edit',
                        to: {
                          name: 'admin-projects-projectId-events-eventId',
                          params: {
                            projectId: route.params.projectId,
                            eventId: row.original.id,
                          },
                        },
                      },
                      {
                        label: 'Copy ID',
                        onClick: () => {
                          copy(row.original.id)
                          toast.add({
                            title: 'Copied',
                            description: 'ID copied to clipboard',
                            color: 'success',
                          })
                        },
                      },
                    ]"
                  >
                    <UButton icon="lucide:ellipsis" variant="ghost" size="sm" />
                  </UDropdownMenu>
                </div>
              </template>
            </UTable>
          </template>
          <template #challenges>
            <div class="my-2">
              <UButton
                icon="lucide:plus"
                :to="{
                  name: 'admin-projects-projectId-challenges-new',
                  params: { projectId: route.params.projectId },
                }"
              >
                Create Challenge
              </UButton>
            </div>
            <UTable
              :data="data.challenges.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'image' },
                { accessorKey: 'name' },
                { accessorKey: 'description' },
                { id: 'actions' },
              ]"
            >
              <template #image-cell="{ row }">
                <NuxtImg
                  v-if="row.original.image"
                  :src="row.original.image"
                  height="32"
                  width="32"
                  class="bg-muted size-8 rounded"
                />
              </template>
              <template #actions-cell="{ row }">
                <div class="flex justify-end gap-2">
                  <UDropdownMenu
                    :items="[
                      {
                        label: 'Edit',
                        to: {
                          name: 'admin-projects-projectId-challenges-challengeId',
                          params: {
                            projectId: route.params.projectId,
                            challengeId: row.original.id,
                          },
                        },
                      },
                      {
                        label: 'Copy ID',
                        onClick: () => {
                          copy(row.original.id)
                          toast.add({
                            title: 'Copied',
                            description: 'ID copied to clipboard',
                            color: 'success',
                          })
                        },
                      },
                    ]"
                  >
                    <UButton icon="lucide:ellipsis" variant="ghost" size="sm" />
                  </UDropdownMenu>
                </div>
              </template>
            </UTable>
          </template>
          <template #streaks>
            <div class="my-2">
              <UButton
                icon="lucide:plus"
                :to="{
                  name: 'admin-projects-projectId-streaks-new',
                  params: { projectId: route.params.projectId },
                }"
              >
                Create Streak
              </UButton>
            </div>
            <UTable
              :data="data.streaks.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'name' },
                { accessorKey: 'description' },
                { id: 'actions' },
              ]"
            >
              <template #actions-cell="{ row }">
                <div class="flex justify-end gap-2">
                  <UDropdownMenu
                    :items="[
                      {
                        label: 'Edit',
                        to: {
                          name: 'admin-projects-projectId-streaks-streakId',
                          params: {
                            projectId: route.params.projectId,
                            streakId: row.original.id,
                          },
                        },
                      },
                      {
                        label: 'Copy ID',
                        onClick: () => {
                          copy(row.original.id)
                          toast.add({
                            title: 'Copied',
                            description: 'ID copied to clipboard',
                            color: 'success',
                          })
                        },
                      },
                    ]"
                  >
                    <UButton icon="lucide:ellipsis" variant="ghost" size="sm" />
                  </UDropdownMenu>
                </div>
              </template>
            </UTable>
          </template>
          <template #achievements>
            <div class="my-2">
              <UButton
                icon="lucide:plus"
                :to="{
                  name: 'admin-projects-projectId-achievements-new',
                  params: { projectId: route.params.projectId },
                }"
              >
                Create Achievement
              </UButton>
            </div>
            <UTable
              :data="data.achievements.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'image' },
                { accessorKey: 'name' },
                { accessorKey: 'description' },
                { accessorKey: 'points' },
                { accessorKey: 'hidden' },
                { id: 'actions' },
              ]"
            >
              <template #image-cell="{ row }">
                <NuxtImg
                  v-if="row.original.image"
                  :src="row.original.image"
                  height="32"
                  width="32"
                  class="bg-muted size-8 rounded"
                />
              </template>
              <template #hidden-cell="{ row }">
                <div>
                  <UIcon v-if="row.original.hidden" name="lucide:check" />
                </div>
              </template>
              <template #actions-cell="{ row }">
                <div class="flex justify-end gap-2">
                  <UDropdownMenu
                    :items="[
                      {
                        label: 'Edit',
                        to: {
                          name: 'admin-projects-projectId-achievements-achievementId',
                          params: {
                            projectId: route.params.projectId,
                            achievementId: row.original.id,
                          },
                        },
                      },
                      {
                        label: 'Copy ID',
                        onClick: () => {
                          copy(row.original.id)
                          toast.add({
                            title: 'Copied',
                            description: 'ID copied to clipboard',
                            color: 'success',
                          })
                        },
                      },
                    ]"
                  >
                    <UButton icon="lucide:ellipsis" variant="ghost" size="sm" />
                  </UDropdownMenu>
                </div>
              </template>
            </UTable>
          </template>
        </UTabs>
      </template>
    </UContainer>
  </div>
</template>
