<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

const route = useRoute('admin-projects-projectId')

const { canEditProject } = usePermissions()
const { isSuperAdmin } = useAuth()
const canEdit = computed(() => canEditProject(route.params.projectId))

gql(`
  query AdminProjectPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
      description
      startDate
      endDate
      branding {
        logoImage {
          ...ImageFields
        }
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
          imagePendingObject {
            ...ImageFields
          }
          imageCompletedObject {
            ...ImageFields
          }
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
          imageObject {
            ...ImageFields
          }
        }
      }
    }
    superteams(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          description
          color
          imageObject {
            ...ImageFields
          }
          teams {
            id
          }
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
    logoImage: null,
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

// Achievement reordering
type AchievementNode =
  AdminProjectPageQuery['achievements']['edges'][number]['node']
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

const { executeMutation: reorderAchievements } =
  useReorderAchievementsMutation()
const isReordering = ref(false)

const toast = useToast()

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
      title: 'Kunne ikke endre rekkefølge',
      description: result.error.message,
      color: 'error',
    })
    // Refetch to restore original order
    refetch({ requestPolicy: 'network-only' })
    return
  }

  toast.add({
    title: 'Rekkefølge lagret',
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
              label: 'Prosjekter',
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
      <ErrorState v-else-if="error" :error class="h-150" />
      <template v-else-if="data">
        <header class="my-12">
          <div class="space-y-2">
            <img
              v-if="state.branding.logoImage?.url"
              :src="state.branding.logoImage.url"
              width="64"
              class="mb-4 rounded"
            >
            <h1 class="text-3xl">
              {{ state.name }}
            </h1>
            <p v-if="state.description" class="text-muted max-w-2xl">
              {{ state.description }}
            </p>
            <div v-if="canEdit" class="mt-4 flex gap-2">
              <UButton
                variant="soft"
                icon="lucide:pencil"
                :to="{
                  name: 'admin-projects-projectId-edit',
                  params: { projectId: route.params.projectId },
                }"
              >
                Rediger prosjekt
              </UButton>
              <UButton
                v-if="isSuperAdmin"
                variant="soft"
                icon="lucide:users"
                :to="{
                  name: 'admin-projects-projectId-superteams',
                  params: { projectId: route.params.projectId },
                }"
              >
                LADD Superteams
              </UButton>
            </div>
          </div>
        </header>
        <UTabs
          v-model="tab"
          :items="[
            {
              value: 'achievements',
              label: 'Utmerkelser',
              slot: 'achievements',
            },
            { value: 'challenges', label: 'Utfordringer', slot: 'challenges' },
            {
              value: 'superteams',
              label: 'Superteams',
              slot: 'superteams',
            },
          ]"
          variant="link"
        >
          <template #challenges>
            <div v-if="canEdit" class="my-2">
              <UButton
                icon="lucide:plus"
                :to="{
                  name: 'admin-projects-projectId-challenges-new',
                  params: { projectId: route.params.projectId },
                }"
              >
                Opprett utfordring
              </UButton>
            </div>
            <UTable
              :data="data.challenges.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'imageObject' },
                { accessorKey: 'name' },
                { accessorKey: 'description' },
                { accessorKey: 'type', header: 'Type' },
                { id: 'actions' },
              ]"
            >
              <template #imageObject-cell="{ row }">
                <img
                  v-if="row.original.imageObject?.url"
                  :src="row.original.imageObject.url"
                  height="32"
                  width="32"
                  class="bg-muted size-8 rounded"
                >
              </template>
              <template #type-cell="{ row }">
                {{
                  row.original.__typename === 'ExternalChallenge'
                    ? 'Ekstern'
                    : row.original.__typename === 'QuizChallenge'
                      ? 'Quiz'
                      : row.original.__typename === 'PluginChallenge'
                        ? 'Plugin'
                        : 'Enkel'
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
                    Rediger
                  </UButton>
                </div>
              </template>
            </UTable>
          </template>
          <template #superteams>
            <div v-if="canEdit" class="my-2">
              <UButton
                icon="lucide:plus"
                :to="{
                  name: 'admin-projects-projectId-superteams-new',
                  params: { projectId: route.params.projectId },
                }"
              >
                Opprett superteam
              </UButton>
            </div>
            <UTable
              :data="data.superteams.edges.map((e) => e.node)"
              :columns="[
                { accessorKey: 'color', header: 'Farge' },
                { accessorKey: 'imageObject', header: 'Bilde' },
                { accessorKey: 'name', header: 'Navn' },
                { accessorKey: 'teams', header: 'Lag' },
                { id: 'actions' },
              ]"
            >
              <template #color-cell="{ row }">
                <div
                  v-if="row.original.color"
                  class="size-6 rounded-full border"
                  :style="{ backgroundColor: row.original.color }"
                />
              </template>
              <template #imageObject-cell="{ row }">
                <img
                  v-if="row.original.imageObject?.url"
                  :src="row.original.imageObject.url"
                  height="32"
                  width="32"
                  class="bg-muted size-8 rounded"
                >
              </template>
              <template #teams-cell="{ row }">
                {{ row.original.teams.length }} lag
              </template>
              <template #actions-cell="{ row }">
                <div class="flex justify-end">
                  <UButton
                    variant="ghost"
                    size="sm"
                    :to="{
                      name: 'admin-projects-projectId-superteams-superTeamId',
                      params: {
                        projectId: route.params.projectId,
                        superTeamId: row.original.id,
                      },
                    }"
                  >
                    Rediger
                  </UButton>
                </div>
              </template>
            </UTable>
            <div
              v-if="data.superteams.edges.length === 0"
              class="text-dimmed py-8 text-center"
            >
              Ingen superteams ennå
            </div>
          </template>
          <template #achievements>
            <div v-if="canEdit" class="mt-2 mb-4">
              <UButton
                icon="lucide:plus"
                :to="{
                  name: 'admin-projects-projectId-achievements-new',
                  params: { projectId: route.params.projectId },
                }"
              >
                Opprett utmerkelse
              </UButton>
            </div>
            <div class="border-default rounded-lg border">
              <VueDraggable
                v-model="achievements"
                handle=".drag-handle"
                ghost-class="opacity-50"
                :animation="200"
                @end="handleReorder"
              >
                <div
                  v-for="achievement in achievements"
                  :key="achievement.id"
                  class="border-default flex items-center gap-4 border-b px-4 py-3 last:border-b-0"
                >
                  <div
                    class="drag-handle text-muted cursor-grab active:cursor-grabbing"
                  >
                    <UIcon name="lucide:grip-vertical" class="size-5" />
                  </div>
                  <img
                    v-if="achievement.imageCompletedObject?.url"
                    :src="achievement.imageCompletedObject.url"
                    height="32"
                    width="32"
                    class="size-8 shrink-0 rounded"
                  >
                  <img
                    v-else
                    src="/images/achievement-placeholder.png"
                    height="32"
                    width="32"
                    class="size-8 shrink-0 rounded"
                  >
                  <div class="min-w-0 flex-1">
                    <div class="font-medium">{{ achievement.name }}</div>
                    <div class="text-dimmed truncate text-sm">
                      {{ achievement.descriptionPending }}
                    </div>
                  </div>
                  <div class="text-muted shrink-0 text-sm">
                    {{ formatNumber(achievement.points) }} pts
                  </div>
                  <UBadge
                    v-if="achievement.hidden"
                    variant="soft"
                    color="warning"
                  >
                    Skjult
                  </UBadge>
                  <UButton
                    variant="ghost"
                    size="sm"
                    :to="{
                      name: 'admin-projects-projectId-achievements-achievementId',
                      params: {
                        projectId: route.params.projectId,
                        achievementId: achievement.id,
                      },
                    }"
                  >
                    Rediger
                  </UButton>
                </div>
              </VueDraggable>
              <div
                v-if="achievements.length === 0"
                class="text-dimmed py-8 text-center"
              >
                Ingen utmerkelser ennå
              </div>
            </div>
          </template>
        </UTabs>
      </template>
    </UContainer>
  </div>
</template>
