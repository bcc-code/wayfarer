<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'superadmin',
})

gql(`
  query AdminAchievementsForPicker($projectId: ID!) {
    achievements(first: 200, filter: { projectId: $projectId }) {
      edges {
        node {
          __typename
          id
          name
        }
      }
    }
  }
`)

gql(`
  query AdminCheckAchievementProgress($userId: ID!, $achievementId: ID!) {
    adminCheckAchievementProgress(userId: $userId, achievementId: $achievementId) {
      achievement {
        __typename
        ... on ContentAchievement {
          id
          name
          points
        }
        ... on StreakAchievement {
          id
          name
          points
        }
      }
      alreadyAwarded
      awardedAt
      items {
        contentItem {
          id
          sortOrder
          externalContent {
            id
            title
            taskId
            contentType
          }
        }
        completed
        completedAt
        completeBy
        completedWithinDeadline
      }
      completedCount
      totalCount
    }
  }
`)

gql(`
  mutation AdminAwardAchievement($userId: ID!, $achievementId: ID!) {
    awardAchievement(userId: $userId, achievementId: $achievementId) {
      id
      name
    }
  }
`)

const route = useRoute('admin-users-userId-achievements')
const { isAuthReady } = useAuthReady()
const toast = useToast()
const { canCheckAchievements } = usePermissions()

// Get current project
const { data: currentProjectData } = useAdminUserPageCurrentProjectQuery({
  pause: computed(() => !isAuthReady.value),
})
const currentProjectId = computed(
  () => currentProjectData.value?.currentProject.id,
)

// Load user name for breadcrumb
const { data: userData } = useAdminUserPageQuery({
  variables: computed(() => ({
    id: route.params.userId,
    projectId: currentProjectId.value ?? '',
  })),
  pause: computed(() => !isAuthReady.value || !currentProjectId.value),
})

// Achievement picker
const { data: achievementsData } = useAdminAchievementsForPickerQuery({
  variables: computed(() => ({ projectId: currentProjectId.value ?? '' })),
  pause: computed(() => !isAuthReady.value || !currentProjectId.value),
})

const achievementOptions = computed(() => {
  if (!achievementsData.value) return []
  return achievementsData.value.achievements.edges
    .filter(
      (e) =>
        e.node.__typename === 'ContentAchievement' ||
        e.node.__typename === 'StreakAchievement',
    )
    .map((e) => ({
      label: `${e.node.name} (${e.node.__typename === 'ContentAchievement' ? 'Innhold' : 'Streak'})`,
      value: e.node.id,
    }))
})

const selectedAchievementId = ref<string | undefined>(undefined)

// Progress query
const {
  data: progressData,
  fetching: progressFetching,
  executeQuery: refetchProgress,
} = useAdminCheckAchievementProgressQuery({
  variables: computed(() => ({
    userId: route.params.userId,
    achievementId: selectedAchievementId.value ?? '',
  })),
  pause: computed(() => !isAuthReady.value || !selectedAchievementId.value),
  requestPolicy: 'network-only',
})

const progress = computed(
  () => progressData.value?.adminCheckAchievementProgress,
)

const isStreak = computed(
  () => progress.value?.achievement.__typename === 'StreakAchievement',
)

const achievementTypeLabel = computed(() =>
  isStreak.value ? 'Streak' : 'Innhold',
)

// Expandable event details
const expandedContentIds = ref<Set<string>>(new Set())

function toggleExpand(externalContentId: string) {
  const next = new Set(expandedContentIds.value)
  if (next.has(externalContentId)) {
    next.delete(externalContentId)
  } else {
    next.add(externalContentId)
  }
  expandedContentIds.value = next
}

// Award
const { executeMutation: awardAchievement } = useAdminAwardAchievementMutation()
const showAwardConfirm = ref(false)
const awarding = ref(false)

async function handleAward() {
  if (!selectedAchievementId.value) return
  awarding.value = true
  const result = await awardAchievement({
    userId: route.params.userId,
    achievementId: selectedAchievementId.value,
  })
  awarding.value = false

  if (result.error) {
    toast.add({
      title: 'Kunne ikke tildele prestasjon',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Prestasjon tildelt',
    description: `Tildelte "${result.data?.awardAchievement.name}"`,
    color: 'success',
  })

  showAwardConfirm.value = false
  refetchProgress({ requestPolicy: 'network-only' })
}

function statusColor(
  item: NonNullable<typeof progress.value>['items'][number],
): 'success' | 'error' | 'neutral' {
  if (item.completedWithinDeadline === false) return 'error'
  if (item.completed) return 'success'
  return 'neutral'
}

function statusLabel(
  item: NonNullable<typeof progress.value>['items'][number],
): string {
  if (item.completedWithinDeadline === false) return 'For sent'
  if (item.completed) return 'Fullfort'
  return 'Ikke fullfort'
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Brukere', to: { name: 'admin-users' } },
            {
              label: userData?.user.name ?? route.params.userId,
              to: {
                name: 'admin-users-userId',
                params: { userId: route.params.userId },
              },
            },
            { label: 'Sjekk prestasjoner' },
          ]"
        />
      </UContainer>
    </div>

    <UContainer class="py-12">
      <div v-if="!canCheckAchievements" class="text-dimmed">Ingen tilgang</div>
      <div v-else class="space-y-6">
        <h1 class="text-3xl font-bold">Sjekk prestasjoner</h1>
        <p class="text-muted">
          Bruker: {{ userData?.user.name ?? route.params.userId }}
        </p>

        <!-- Achievement selector -->
        <UCard>
          <template #header>
            <h2 class="text-xl font-semibold">Velg prestasjon</h2>
          </template>
          <UFormField label="Prestasjon">
            <USelect
              v-model="selectedAchievementId"
              :items="achievementOptions"
              value-key="value"
              placeholder="Velg en innholds- eller streak-prestasjon"
              class="w-full"
            />
          </UFormField>
        </UCard>

        <!-- Progress display -->
        <div v-if="progressFetching" class="py-8">
          <LoadingState />
        </div>

        <template v-else-if="progress">
          <!-- Status summary -->
          <UCard>
            <template #header>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <h2 class="text-xl font-semibold">Fremgang</h2>
                  <UBadge variant="subtle">{{ achievementTypeLabel }}</UBadge>
                </div>
                <div class="flex items-center gap-3">
                  <span class="text-sm font-medium">
                    {{ progress.completedCount }} av
                    {{ progress.totalCount }} fullfort
                  </span>
                  <UBadge
                    v-if="progress.alreadyAwarded"
                    color="info"
                    variant="soft"
                  >
                    Allerede tildelt
                    <template v-if="progress.awardedAt">
                      {{ formatDateTime(progress.awardedAt) }}
                    </template>
                  </UBadge>
                </div>
              </div>
            </template>

            <!-- Items table -->
            <div class="divide-default divide-y">
              <div
                v-for="item in progress.items"
                :key="item.contentItem.id"
                class="text-sm"
              >
                <!-- Item row -->
                <div
                  class="flex cursor-pointer items-center gap-3 p-3 transition-colors hover:bg-elevated"
                  @click="toggleExpand(item.contentItem.externalContent.id)"
                >
                  <span class="text-muted w-8 text-right text-xs">
                    {{ item.contentItem.sortOrder }}
                  </span>
                  <div class="min-w-0 flex-1">
                    <span class="font-medium">
                      {{
                        item.contentItem.externalContent.title ??
                        item.contentItem.externalContent.taskId
                      }}
                    </span>
                    <span class="text-muted ml-2 text-xs">
                      {{ item.contentItem.externalContent.contentType }}
                    </span>
                  </div>
                  <UBadge :color="statusColor(item)" variant="soft" size="sm">
                    {{ statusLabel(item) }}
                  </UBadge>
                  <div
                    v-if="item.completedAt"
                    class="text-muted shrink-0 text-xs"
                  >
                    {{ formatDateTime(item.completedAt) }}
                  </div>
                  <div
                    v-if="isStreak && item.completeBy"
                    class="text-muted shrink-0 text-xs"
                  >
                    Frist: {{ formatDateTime(item.completeBy) }}
                  </div>
                  <Icon
                    :name="
                      expandedContentIds.has(
                        item.contentItem.externalContent.id,
                      )
                        ? 'lucide:chevron-up'
                        : 'lucide:chevron-down'
                    "
                    class="size-4 text-dimmed shrink-0"
                  />
                </div>

                <!-- Expanded event details -->
                <div
                  v-if="
                    expandedContentIds.has(item.contentItem.externalContent.id)
                  "
                  class="border-t border-default bg-elevated px-3 py-2"
                >
                  <AdminContentEventDetails
                    :user-id="route.params.userId"
                    :external-content-id="item.contentItem.externalContent.id"
                  />
                </div>
              </div>
            </div>
          </UCard>

          <!-- Award button -->
          <div v-if="!progress.alreadyAwarded" class="flex justify-end">
            <UButton size="lg" @click="showAwardConfirm = true">
              Tildel prestasjon
            </UButton>
          </div>
        </template>
      </div>
    </UContainer>

    <!-- Award confirmation -->
    <UModal v-model:open="showAwardConfirm">
      <template #header>
        <h3 class="text-lg font-semibold">Tildel prestasjon</h3>
      </template>
      <template #body>
        <p>
          Er du sikker på at du vil tildele denne prestasjonen til brukeren?
        </p>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-3">
          <UButton
            variant="ghost"
            color="neutral"
            @click="showAwardConfirm = false"
          >
            Avbryt
          </UButton>
          <UButton :loading="awarding" @click="handleAward"> Tildel </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
