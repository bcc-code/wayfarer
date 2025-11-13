<script setup lang="ts">
gql(`
  query StandingsPage {
    myCurrentProject {
      id
      leaderboard(entityType: PERSONS) {
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

const { data, error, fetching } = useStandingsPageQuery()

const leaderboard = computed<LeaderboardEntry[]>(() => {
  if (!data.value) return []

  const result = []
  result.push(
    ...data.value.myCurrentProject.leaderboard.edges.map((edge) => edge.node),
  )
  const me = data.value?.myCurrentProject.leaderboard.me
  if (me) {
    result.push(me)
  }
  return [...new Set(result)]
})
</script>

<template>
  <PageLayout :title="$t('pages.standings')">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <LeaderboardList
      v-else-if="leaderboard?.length"
      :leaderboard="leaderboard"
      variant="expanded"
    />
    <EmptyState v-else />
  </PageLayout>
</template>
