<script setup lang="ts">
const entityType = ref(LeaderboardEntityType.Persons)

const { me } = useAuth()
const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsLocalPageQuery({
  variables: computed(() => ({
    filter: {
      churchId: me.value?.church.id,
      ageRange: { min: 13, max: 37 },
    },
  })),
  pause: computed(() => !isAuthReady.value || !me.value?.church.id),
})

const personLeaderboard = computed<Partial<LeaderboardEntry>[]>(() => {
  if (!data.value) return []
  const result = []
  result.push(
    ...data.value.myCurrentProject.personLeaderboard.edges.map(
      (edge) => edge.node,
    ),
  )
  const me = data.value?.myCurrentProject.personLeaderboard.me
  if (me && !result.find((entry) => entry.id === me.id)) {
    result.push(me)
  }
  return result
})

const unitLeaderboard = computed<Partial<LeaderboardEntry>[]>(() => {
  if (!data.value) return []
  const result = []
  result.push(
    ...data.value.myCurrentProject.unitLeaderboard.edges.map(
      (edge) => edge.node,
    ),
  )
  const me = data.value?.myCurrentProject.unitLeaderboard.me
  if (me && !result.find((entry) => entry.id === me.id)) {
    result.push(me)
  }
  return result
})

const debouncedFetching = useDebounce(fetching, 200)

const totalPersons = computed(() => {
  const totalCount = data.value?.myCurrentProject.personLeaderboard.totalCount

  if (!totalCount) return 0
  if (totalCount >= 50) return 20
  if (totalCount > 20) return 10
  return 3
})
</script>

<template>
  <div>
    <LoadingState v-if="debouncedFetching" />
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
        v-model="entityType"
        :tabs="[
          {
            key: 'persons',
            label: $t('standings.top', { amount: totalPersons }),
            value: LeaderboardEntityType.Persons,
            icon: 'IconUser',
          },
          {
            key: 'units',
            label: $t('standings.units'),
            value: LeaderboardEntityType.Teams,
            icon: 'IconUsers',
          },
        ]"
        class="mb-list-section-gap"
        variant="secondary"
      >
        <template #tab="{ tab }">
          <div class="flex flex-col items-center gap-0.5">
            <Icon :name="tab.icon" class="size-7" />
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
        hide-medals
      />
      <LeaderboardList
        v-if="
          unitLeaderboard?.length && entityType === LeaderboardEntityType.Teams
        "
        :leaderboard="unitLeaderboard"
        hide-medals
      />
    </template>
    <EmptyState v-else :title="$t('emptyStates.standings')" />
  </div>
</template>
