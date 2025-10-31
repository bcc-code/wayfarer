<script setup lang="ts">
import { useUnitPageQuery } from '~/api/generated'

gql(`
query UnitPage {
  user {
    currentProject {
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
}

`)

const { data, error, fetching } = useUnitPageQuery()
</script>

<template>
  <div class="h-full">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <div class="text-center p-2">
        <h1 class="text-xl font-bold">
          {{ data.user.currentProject.myTeam.name }}
        </h1>
        <p class="text-sm text-muted">
          {{ data.user.currentProject.myTeam.superTeam?.name }}
        </p>
      </div>
      <LeaderboardList
        :leaderboard="data.user.currentProject.myTeam.leaderboard"
        class="mt-4"
      />
    </template>
  </div>
</template>
