<script setup lang="ts">
import { useStandingsPageQuery } from '~/api/generated'

gql(`
query StandingsPage {
  user {
    currentProject {
      id
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

const { data, error, fetching } = useStandingsPageQuery()
</script>

<template>
  <div class="h-full">
    <h1 class="text-xl font-bold mb-4">Standings</h1>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <LeaderboardList
      v-else-if="data"
      :leaderboard="data.user.currentProject.leaderboard"
    />
  </div>
</template>
