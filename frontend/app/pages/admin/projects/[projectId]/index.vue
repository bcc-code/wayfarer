<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'

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
const {
  data,
  error,
  fetching,
  executeQuery: refetch,
} = useAdminProjectPageQuery({
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

const { copy } = useClipboard()
const toast = useToast()

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

// Achievement reordering
type AchievementNode = AdminProjectPageQuery['achievements']['edges'][number]['node']
const achievements = ref<AchievementNode[]>([])

watch(
  () => data.value?.achievements.edges,
  (edges) => {
    if (edges) {
      achievements.value = edges.map((e) => e.node)
    }
  },
  { immediate: true },
)

const { executeMutation: reorderAchievements } = useReorderAchievementsMutation()
const isReordering = ref(false)

async function handleReorder() {
  if (isReordering.value) return
  isReordering.value = true

  const result = await reorderAchievements({
    projectId: route.params.projectId,
    achievementIds: achievements.value.map((a) => a.id),
  })

  isReordering.value = false

  if (result.error) {
    toast.add({
      title: 'Failed to reorder',
      description: result.error.message,
      color: 'error',
    })
    // Refetch to restore original order
    refetch({ requestPolicy: 'network-only' })
    return
  }

  toast.add({
    title: 'Order saved',
    color: 'success',
  })
}
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
            <div class="border-default rounded-lg border">
              <VueDraggable
                v-model="achievements"
                handle=".drag-handle"
                ghost-class="opacity-50"
                @end="handleReorder"
              >
                <div
                  v-for="achievement in achievements"
                  :key="achievement.id"
                  class="border-default flex items-center gap-4 border-b px-4 py-3 last:border-b-0"
                >
                  <div class="drag-handle text-muted cursor-grab active:cursor-grabbing">
                    <UIcon name="lucide:grip-vertical" class="size-5" />
                  </div>
                  <img
                    v-if="achievement.imageCompleted"
                    :src="achievement.imageCompleted"
                    height="32"
                    width="32"
                    class="size-8 shrink-0 rounded"
                  />
                  <img
                    v-else
                    src="/images/achievement-placeholder.png"
                    height="32"
                    width="32"
                    class="size-8 shrink-0 rounded"
                  />
                  <div class="min-w-0 flex-1">
                    <div class="font-medium">{{ achievement.name }}</div>
                    <div class="text-dimmed truncate text-sm">
                      {{ achievement.descriptionPending }}
                    </div>
                  </div>
                  <div class="text-muted shrink-0 text-sm">
                    {{ achievement.points }} pts
                  </div>
                  <UBadge v-if="achievement.hidden" variant="soft" color="warning">
                    Hidden
                  </UBadge>
                  <UDropdownMenu
                    :items="[
                      {
                        label: 'Edit',
                        to: {
                          name: 'admin-projects-projectId-achievements-achievementId',
                          params: {
                            projectId: route.params.projectId,
                            achievementId: achievement.id,
                          },
                        },
                      },
                      {
                        label: 'Copy ID',
                        onClick: () => {
                          copy(achievement.id)
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
              </VueDraggable>
              <div
                v-if="achievements.length === 0"
                class="text-dimmed py-8 text-center"
              >
                No achievements yet
              </div>
            </div>
          </template>
        </UTabs>
      </template>
    </UContainer>
  </div>
</template>
