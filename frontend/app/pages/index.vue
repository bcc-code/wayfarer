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
  <PageLayout :title="$t('pages.standings')">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <LeaderboardList
      v-else-if="data?.myCurrentProject.leaderboard.length"
      :leaderboard="data.myCurrentProject.leaderboard"
      variant="expanded"
    />
    <EmptyState v-else />
  </PageLayout>
</template>
