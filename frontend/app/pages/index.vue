<script setup lang="ts">
gql(`
  query StandingsPage {
    myCurrentProject {
      id
      leaderboard(type: TOTAL) {
        name
        description
        score
        image
      }
    }
  }
`)

const { data, error, fetching } = useStandingsPageQuery()
</script>

<template>
  <PageLayout title="Standings">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <LeaderboardList
      v-else-if="data"
      :leaderboard="data.myCurrentProject.leaderboard"
    />
  </PageLayout>
</template>
