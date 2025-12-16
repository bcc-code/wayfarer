<script setup lang="ts">
const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsGlobalPageQuery({
  variables: computed(() => ({
    entityType: LeaderboardEntityType.Persons,
    first: 20,
    filter: {
      ageRange: { min: 14, max: 26 },
    },
  })),
  pause: computed(() => !isAuthReady.value),
})

const leaderboard = computed<LeaderboardEntry[]>(() => {
  if (!data.value) return []

  const result = []
  result.push(
    ...data.value.myCurrentProject.leaderboard.edges.map((edge) => edge.node),
  )
  const me = data.value?.myCurrentProject.leaderboard.me
  if (me && !result.find((entry) => entry.id === me.id)) {
    result.push(me)
  }
  return result
})
</script>

<template>
  <div>
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <LeaderboardList
      v-else-if="leaderboard?.length"
      :leaderboard="leaderboard"
    />
    <EmptyState v-else :title="$t('emptyStates.standings')" />
  </div>
</template>
