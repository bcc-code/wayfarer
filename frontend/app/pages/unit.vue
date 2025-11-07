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
    <template v-else-if="data">
      <LeaderboardList
        v-if="data.myCurrentProject.myTeam"
        :leaderboard="data.myCurrentProject.myTeam.leaderboard"
        variant="expanded"
      />
    </template>
  </PageLayout>
</template>
