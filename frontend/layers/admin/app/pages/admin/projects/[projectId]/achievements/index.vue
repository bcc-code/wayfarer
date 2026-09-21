<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId-achievements')
const { canEditProject } = usePermissions()
const canEdit = computed(() => canEditProject(route.params.projectId))

gql(`
  query AdminProjectAchievements($projectId: ID!) {
    achievements(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          descriptionPending
          imageCompletedObject {
            ...ImageFields
          }
          points
          hidden
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
} = useAdminProjectAchievementsQuery({
  variables: computed(() => ({ projectId: route.params.projectId })),
  pause: computed(() => !isAuthReady.value),
})

// Drag-to-reorder needs a local copy it can mutate; the query result is the
// source of truth it resets to when a reorder fails.
type AchievementNode =
  AdminProjectAchievementsQuery['achievements']['edges'][number]['node']
const achievements = ref<AchievementNode[]>([])

watch(
  () => data.value?.achievements.edges,
  (edges) => {
    if (edges) achievements.value = edges.map((e) => e.node)
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
    // Restore the server's order. This must be *this* query's executeQuery —
    // the reorder and the list it reorders have to stay together.
    refetch({ requestPolicy: 'network-only' })
    return
  }

  toast.add({ title: 'Rekkefølge lagret', color: 'success' })
}
</script>

<template>
  <div>
    <div class="mb-4 flex items-center justify-between gap-4">
      <h2 class="text-xl">Utmerkelser</h2>
      <UButton
        v-if="canEdit"
        icon="lucide:plus"
        :to="{
          name: 'admin-projects-projectId-achievements-new',
          params: { projectId: route.params.projectId },
        }"
      >
        Opprett utmerkelse
      </UButton>
    </div>

    <AdminQueryState :fetching :error>
      <UEmpty
        v-if="!achievements.length"
        icon="lucide:award"
        title="Ingen utmerkelser ennå"
        description="Opprett den første utmerkelsen for dette prosjektet."
      />
      <div v-else class="border-default rounded-lg border">
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
              :src="
                achievement.imageCompletedObject?.url ??
                '/images/achievement-placeholder.png'
              "
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
              {{ formatNumber(achievement.points) }} pts
            </div>
            <UBadge v-if="achievement.hidden" variant="soft" color="warning">
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
      </div>
    </AdminQueryState>
  </div>
</template>
