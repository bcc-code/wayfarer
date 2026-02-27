<script setup lang="ts">
import { AGE_RANGE_ADULT, AGE_RANGE_YOUNG } from '~/utils/constants'
import { getExtraItems } from '~/utils/leaderboard'

const entityType = useLocalStorage<LeaderboardEntityType>(
  'standings-local:entity-type',
  LeaderboardEntityType.Persons,
)

const { me } = useAuth()
const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsLocalPageQuery({
  variables: computed(() => ({
    first: entityType.value === LeaderboardEntityType.Persons ? 20 : 500,
    filter: {
      churchId: me.value?.church.id,
      ageRange: { min: AGE_RANGE_YOUNG.min, max: AGE_RANGE_ADULT.max },
    },
  })),
  pause: computed(() => !isAuthReady.value || !me.value?.church.id),
})

const personLeaderboard = computed<Partial<LeaderboardEntry>[]>(() => {
  if (!data.value) return []
  return data.value.myCurrentProject.personLeaderboard.edges.map(
    (edge) => edge.node,
  )
})

const personExtraItems = computed<Partial<LeaderboardEntry>[]>(() => {
  return getExtraItems(
    personLeaderboard.value,
    data.value?.myCurrentProject.personLeaderboard.me,
  )
})

const unitLeaderboard = computed<Partial<LeaderboardEntry>[]>(() => {
  if (!data.value) return []
  return data.value.myCurrentProject.unitLeaderboard.edges.map(
    (edge) => edge.node,
  )
})

const unitExtraItems = computed<Partial<LeaderboardEntry>[]>(() => {
  return getExtraItems(
    unitLeaderboard.value,
    data.value?.myCurrentProject.unitLeaderboard.me,
  )
})

const debouncedFetching = useDebounce(fetching, 200)

const totalPersons = computed(() => {
  const totalCount = data.value?.myCurrentProject.personLeaderboard.totalCount

  if (!totalCount) return 20
  if (totalCount >= 50) return 20
  if (totalCount > 20) return 10
  return 3
})
</script>

<template>
  <div>
    <StandingsListSkeleton v-if="debouncedFetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <div
        v-if="data.me.church"
        class="p-medium gap-medium mb-list-section-gap flex flex-col items-center"
      >
        <h2 class="text-heading text-center text-balance">
          {{ data.me.church.name }}
        </h2>
      </div>
      <DesignTabs
        v-if="unitLeaderboard.length"
        v-model="entityType"
        :tabs="[
          {
            key: 'persons',
            label: $t('standings.top', { amount: totalPersons }),
            value: LeaderboardEntityType.Persons,
          },
          {
            key: 'units',
            label: $t('standings.units'),
            value: LeaderboardEntityType.Teams,
          },
        ]"
        class="mb-list-section-gap"
        variant="secondary"
      >
        <template #tab="{ tab }">
          <div class="flex flex-col items-center gap-0.5">
            <IconUser v-if="tab.key === 'persons'" class="size-7" />
            <IconUsers v-else class="size-7" />
            <span>{{ tab.label }}</span>
          </div>
        </template>
      </DesignTabs>
      <LeaderboardList
        v-if="
          personLeaderboard?.length &&
          entityType === LeaderboardEntityType.Persons
        "
        :leaderboard="personLeaderboard"
        :extra-items="personExtraItems"
      />
      <LeaderboardList
        v-if="
          unitLeaderboard?.length && entityType === LeaderboardEntityType.Teams
        "
        :leaderboard="unitLeaderboard"
        :extra-items="unitExtraItems"
      />
    </template>
    <EmptyState v-else :title="$t('emptyStates.standings')" />
  </div>
</template>
