<script setup lang="ts">
const { track } = useAnalytics()
const { t } = useI18n()

/**
 * Every board in one query. `leaderboardConfig(id)` and `leaderboardConfigs`
 * are admin-only, so a user can reach a config only through
 * `myCurrentProject.leaderboards`, which takes no arguments — there is no way
 * to fetch just the open tab's board. Tab switching is therefore free: no
 * request, the data is already here.
 */
gql(`
  query StandingsPage($first: Int) {
    myCurrentProject {
      id
      myTeam {
        id
      }
      leaderboards {
        id
        name
        leaderboard(first: $first) {
          totalCount
          edges {
            node {
              ...LeaderboardEntryWithDescriptionFields
            }
          }
          me {
            ...LeaderboardEntryWithDescriptionFields
          }
        }
      }
    }
  }
`)

/**
 * One `first` covers every board: `leaderboard` is selected once for the whole
 * list, so it cannot vary per config the way the old page used 20 for persons
 * and 500 for teams.
 */
const BOARD_SIZE = 100

/** The one tab that is not a config — see below. */
const UNIT_TAB = 'unit'

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsPageQuery({
  variables: { first: BOARD_SIZE },
  pause: computed(() => !isAuthReady.value),
})

const boards = computed(() => data.value?.myCurrentProject.leaderboards ?? [])
const hasUnit = computed(() => Boolean(data.value?.myCurrentProject.myTeam?.id))

/**
 * The unit tab is not config-driven and cannot be: it reads
 * `myTeam.memberLeaderboard`, a different field with a different shape, so no
 * `LeaderboardConfig` describes it. It stays hardcoded, and last.
 */
const tabs = computed(() => [
  ...boards.value.map((board) => ({
    key: board.id,
    label: board.name,
    value: board.id,
  })),
  ...(hasUnit.value
    ? [{ key: UNIT_TAB, label: t('standings.unit'), value: UNIT_TAB }]
    : []),
])

const params = useUrlSearchParams('history')
const savedTab = useLocalStorage('standings-tab', '')

const tab = computed({
  get() {
    const requested =
      typeof params.tab === 'string' && params.tab ? params.tab : savedTab.value
    // A stored or linked tab outlives the board it names — a deleted config, a
    // deactivated one, a different project — so an unknown value falls back to
    // the first tab rather than rendering nothing.
    const known = tabs.value.some((item) => item.value === requested)
    return known ? requested : (tabs.value[0]?.value ?? '')
  },
  set(newTab: string) {
    const oldTab = tab.value
    // Labels rather than ids: a config id says nothing in an analytics
    // dashboard, where the old values were 'global' / 'local' / 'unit'.
    track(AnalyticsEvent.LeaderboardTabChanged, {
      from: tabLabel(oldTab),
      to: tabLabel(newTab),
    })
    if (newTab === UNIT_TAB) {
      track(AnalyticsEvent.TeamLeaderboardViewed, {
        team_id: data.value?.myCurrentProject.myTeam?.id,
      })
    }
    params.tab = newTab
    savedTab.value = newTab
  },
})

function tabLabel(value: string) {
  return tabs.value.find((item) => item.value === value)?.label ?? value
}

const selectedBoard = computed(() =>
  boards.value.find((board) => board.id === tab.value),
)

// Only on the first load, so a background refetch leaves the board in place.
const isInitialLoading = computed(() => fetching.value && !data.value)
</script>

<template>
  <PageLayout :title="$t('pages.standings')">
    <div class="p-list-outside">
      <StandingsListSkeleton v-if="isInitialLoading" />
      <ErrorState v-else-if="error" :error />
      <template v-else>
        <!-- A single tab is not a choice; the board names itself in its own
             heading. -->
        <DesignTabs
          v-if="tabs.length > 1"
          v-model="tab"
          :tabs
          class="mb-default -mt-list-outside"
        />
        <StandingsBoard v-if="selectedBoard" :board="selectedBoard" />
        <StandingsUnit v-else-if="tab === UNIT_TAB" />
        <EmptyState v-else :title="$t('emptyStates.standings')" />
      </template>
    </div>
  </PageLayout>
</template>
