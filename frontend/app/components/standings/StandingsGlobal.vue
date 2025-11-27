<script setup lang="ts">
gql(`
  query StandingsGlobalPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter, $first: Int) {
    myCurrentProject {
      id
      leaderboard(entityType: $entityType, filter: $filter, first: $first) {
        edges {
          node {
            id
            name
            description
            score
            rank
            tags
          }
        }
        me {
          id
          name
          description
          score
          rank
          tags
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsGlobalPageQuery({
  variables: computed(() => ({
    entityType: LeaderboardEntityType.Persons,
    first: 20,
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
    <EmptyState v-else />
  </div>
</template>
