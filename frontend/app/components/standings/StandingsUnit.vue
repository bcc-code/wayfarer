<script setup lang="ts">
gql(`
  query StandingsUnitPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter, $first: Int) {
    myCurrentProject {
      id
      leaderboard(entityType: $entityType, filter: $filter, first: $first) {
        edges {
          node {
            id
            name
            score
            image
            rank
            isMe
          }
        }
        me {
          id
          name
          score
          rank
          isMe
          image
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsUnitPageQuery({
  variables: computed(() => ({
    entityType: LeaderboardEntityType.Persons,
    filter: {},
  })),
  pause: computed(() => !isAuthReady.value),
})

const leaderboard = computed<Partial<LeaderboardEntry>[]>(() => {
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
    <template v-else-if="data">
      <div
        class="p-medium gap-medium mb-list-section-gap flex flex-col items-center"
      >
        <h2 class="text-heading">Unit Name</h2>
        <DesignButton variant="secondary" size="medium">
          {{ $t('standings.editUnit') }}
        </DesignButton>
      </div>
      <LeaderboardList
        v-if="leaderboard?.length"
        :leaderboard="leaderboard"
        :badge="(entry, index) => (index == 1 ? 'Unit Leader' : undefined)"
        hide-medals
      />
    </template>
    <EmptyState v-else />
  </div>
</template>
