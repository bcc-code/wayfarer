<script setup lang="ts">
gql(`
	query PointHistory($last: Int) {
    myCurrentProject {
      journal(last: $last) {
        edges {
          node {
            id
            sourceType
            reason
            source {
              __typename
              ... on Achievement {
                id
                name
              }
              ... on Challenge {
                id
                name
              }
              ... on Event {
                id
                name
              }
              ... on SimpleAchievement {
                id
                name
              }
              ... on ContentAchievement {
                id
                name
              }
              ... on StreakAchievement {
                id
                name
              }
            }
            points
            createdAt
          }
        }
      }
    }
  }
`)

const { track } = useAnalytics()

const open = ref(false)

watch(open, (isOpen) => {
  if (isOpen) {
    track(AnalyticsEvent.PointsHistoryOpened)
  }
})

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = usePointHistoryQuery({
  variables: { last: 100 },
  pause: computed(() => !isAuthReady.value || !open.value),
})

const isInitialLoading = computed(() => fetching.value && !data.value)

function getScoreJournalName(
  journal: PointHistoryQuery['myCurrentProject']['journal']['edges'][number]['node'],
) {
  switch (journal.sourceType) {
    case ScoreSourceType.Achievement:
      return journal.source?.name
    case ScoreSourceType.Challenge:
      return journal.source?.name
    case ScoreSourceType.Event:
      return journal.source?.name
    case ScoreSourceType.Manual:
      return journal.reason
    default:
      return journal.reason
  }
}
</script>

<template>
  <DesignDrawer v-model:open="open" :title="$t('pages.pointHistory')">
    <slot />
    <template #content>
      <p class="text-label text-text-default p-medium pt-0 mb-list-section-gap">
        {{ $t('pointHistory.explanation') }}
      </p>

      <LoadingState v-if="isInitialLoading" />
      <ErrorState v-else-if="error" :error />
      <DesignPanel
        v-else-if="data?.myCurrentProject.journal.edges.length"
        class="space-y-list-section-inset p-list-section-inset"
      >
        <template
          v-for="(journal, index) in data.myCurrentProject.journal.edges"
          :key="journal.node.id"
        >
          <div
            class="px-3 py-2 rounded-list-inset flex gap-2.5 items-center justify-between"
          >
            <div>
              <p class="text-label">
                {{ getScoreJournalName(journal.node) }}
              </p>
              <span class="text-caption text-text-muted">
                {{ formatDateTime(journal.node.createdAt) }}
              </span>
            </div>
            <span
              :class="[
                'text-label',
                {
                  'text-accent-positive': journal.node.points >= 0,
                  'text-accent-negative': journal.node.points < 0,
                },
              ]"
            >
              {{ journal.node.points >= 0 ? '+' : ''
              }}{{ formatNumber(journal.node.points) }}
            </span>
          </div>
          <hr
            v-if="index < data.myCurrentProject.journal.edges.length - 1"
            class="mx-3 border-border-default"
          />
        </template>
      </DesignPanel>
      <EmptyState v-else :title="$t('emptyStates.pointHistory')" />
    </template>
  </DesignDrawer>
</template>
