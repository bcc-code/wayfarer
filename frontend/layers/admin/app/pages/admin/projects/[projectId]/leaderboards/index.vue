<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import type { LeaderboardConfigFieldsFragment } from '~/api/generated'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId-leaderboards')
const { canEditProject } = usePermissions()
const canEdit = computed(() => canEditProject(route.params.projectId))

/**
 * `Project.leaderboards` rather than the paginated `leaderboardConfigs` query:
 * it is ordered by `sortOrder` (the list query orders by `createdAt DESC`) and
 * it already returns inactive configs to admins. It also returns the project's
 * event-scoped configs, which is what this page wants to show.
 */
gql(`
  query AdminProjectLeaderboards($projectId: ID!) {
    project(id: $projectId) {
      id
      leaderboards {
        ...LeaderboardConfigFields
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
} = useAdminProjectLeaderboardsQuery({
  variables: computed(() => ({ projectId: route.params.projectId })),
  pause: computed(() => !isAuthReady.value),
})

// Drag-to-reorder needs a local copy it can mutate; the query result is the
// source of truth it resets to when a reorder fails.
const configs = ref<LeaderboardConfigFieldsFragment[]>([])

watch(
  () => data.value?.project.leaderboards,
  (list) => {
    if (list) configs.value = [...list]
  },
  { immediate: true },
)

const entityTypeLabels = LEADERBOARD_ENTITY_TYPE_LABELS
const describeFilter = summarizeLeaderboardFilter

const { executeMutation: updateConfig } = useUpdateLeaderboardConfigMutation()
const isReordering = ref(false)
const toast = useToast()

/**
 * There is no bulk reorder mutation, so a drag becomes one full-replace update
 * per row that actually moved — hence `leaderboardFilterViewToInput`: the
 * filter this page is not changing still has to be resent.
 */
async function handleReorder() {
  if (isReordering.value) return
  isReordering.value = true

  const moved = configs.value
    .map((config, index) => ({ config, index }))
    .filter(({ config, index }) => config.sortOrder !== index)

  const results = await Promise.all(
    moved.map(({ config, index }) =>
      updateConfig({
        id: config.id,
        input: {
          name: config.name,
          entityType: config.entityType,
          filter: leaderboardFilterViewToInput(config.filter),
          maxEntries: config.maxEntries ?? null,
          sortOrder: index,
          isActive: config.isActive,
        },
      }),
    ),
  )

  isReordering.value = false

  const failed = results.find((result) => result.error)
  if (failed?.error) {
    toast.add({
      title: 'Kunne ikke endre rekkefølge',
      description: failed.error.message,
      color: 'error',
    })
    // Restore the server's order. This must be *this* query's executeQuery —
    // the reorder and the list it reorders have to stay together.
    refetch({ requestPolicy: 'network-only' })
    return
  }

  if (moved.length) toast.add({ title: 'Rekkefølge lagret', color: 'success' })
}
</script>

<template>
  <div>
    <div class="mb-4 flex items-center justify-between gap-4">
      <h2 class="text-xl">Ledertavler</h2>
      <UButton
        v-if="canEdit"
        icon="lucide:plus"
        :to="{
          name: 'admin-projects-projectId-leaderboards-new',
          params: { projectId: route.params.projectId },
        }"
      >
        Opprett ledertavle
      </UButton>
    </div>

    <AdminQueryState :fetching :error>
      <UEmpty
        v-if="!configs.length"
        icon="lucide:list-ordered"
        title="Ingen ledertavler ennå"
        description="Opprett den første ledertavlen for dette prosjektet."
      />
      <div v-else class="@container border-default rounded-lg border">
        <VueDraggable
          v-model="configs"
          handle=".drag-handle"
          ghost-class="opacity-50"
          :animation="200"
          @end="handleReorder"
        >
          <div
            v-for="config in configs"
            :key="config.id"
            class="border-default flex items-center gap-4 border-b px-4 py-3 last:border-b-0"
          >
            <div
              class="drag-handle text-muted cursor-grab active:cursor-grabbing"
            >
              <UIcon name="lucide:grip-vertical" class="size-5" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="font-medium">{{ config.name }}</div>
              <div class="text-dimmed flex flex-wrap gap-x-2 text-sm">
                <span>{{ entityTypeLabels[config.entityType] }}</span>
                <span v-if="config.maxEntries"
                  >· Topp {{ config.maxEntries }}</span
                >
                <span v-for="part in describeFilter(config.filter)" :key="part">
                  · {{ part }}
                </span>
              </div>
            </div>
            <UBadge v-if="config.event" variant="soft" color="neutral">
              {{ config.event.name }}
            </UBadge>
            <UBadge v-if="!config.isActive" variant="soft" color="warning">
              Inaktiv
            </UBadge>
            <UButton
              variant="ghost"
              size="sm"
              :to="{
                name: 'admin-projects-projectId-leaderboards-leaderboardId',
                params: {
                  projectId: route.params.projectId,
                  leaderboardId: config.id,
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
