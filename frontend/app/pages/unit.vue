<script setup lang="ts">
gql(`
  query UnitPage {
    myCurrentProject {
      id
      myTeam {
        id
        name
        superTeam {
          id
          name
        }
        leaderboard(type: TOTAL) {
          name
          description
          score
          image
        }
      }
    }
  }
`)

const { data, error, fetching } = useUnitPageQuery()
</script>

<template>
  <PageLayout title="Your unit">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <LeaderboardList
      v-else-if="data?.myCurrentProject.myTeam?.leaderboard.length"
      :leaderboard="data.myCurrentProject.myTeam.leaderboard"
      variant="expanded"
    />
    <EmptyState v-else />
  </PageLayout>
</template>
