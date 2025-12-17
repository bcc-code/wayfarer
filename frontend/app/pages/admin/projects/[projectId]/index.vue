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
          descriptionPending
          descriptionCompleted
          imagePending
          imageCompleted
          points
          hidden
        }
      }
    }
    challenges(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          __typename
          id
          name
          description
          image
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
      dark: {
        accent: '',
      },
      light: {
        accent: '',
      },
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
const fallbackTab = useLocalStorage('fallback-tab', 'achievements')
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
            <img
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
            {
              value: 'achievements',
              label: 'Achievements',
              slot: 'achievements',
            },
            { value: 'challenges', label: 'Challenges', slot: 'challenges' },
          ]"
          variant="link"
        >
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
                { accessorKey: 'type', header: 'Type' },
                { id: 'actions' },
              ]"
            >
              <template #image-cell="{ row }">
                <img
                  v-if="row.original.image"
                  :src="row.original.image"
                  height="32"
                  width="32"
                  class="bg-muted size-8 rounded"
                />
              </template>
              <template #type-cell="{ row }">
                {{
                  row.original.__typename === 'ExternalChallenge'
                    ? 'External'
                    : row.original.__typename === 'QuizChallenge'
                      ? 'Quiz'
                      : 'Simple'
                }}
              </template>
              <template #actions-cell="{ row }">
                <div class="flex justify-end">
                  <UButton
                    variant="ghost"
                    size="sm"
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
                <img
                  v-if="row.original.imageCompleted"
                  :src="row.original.imageCompleted"
                  height="32"
                  width="32"
                  class="size-8 rounded aspect-square"
                />
                <img
                  v-else
                  src="/images/achievement-placeholder.png"
                  height="32"
                  width="32"
                  class="size-8 rounded aspect-square"
                />
              </template>
              <template #hidden-cell="{ row }">
                <div>
                  <UIcon v-if="row.original.hidden" name="lucide:check" />
                </div>
              </template>
              <template #actions-cell="{ row }">
                <div class="flex justify-end">
                  <UButton
                    variant="ghost"
                    size="sm"
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
