<script setup lang="ts">
gql(`
  query UnitPage {
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
`)

const { data, error, fetching } = useUnitPageQuery()
</script>

<template>
  <PageLayout title="Your unit">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <div class="text-center p-2">
        <h1 class="text-xl font-bold">
          {{ data.currentProject.myTeam.name }}
        </h1>
        <p class="text-sm text-muted">
          {{ data.currentProject.myTeam.superTeam?.name }}
        </p>
      </div>
      <LeaderboardList
        :leaderboard="data.currentProject.myTeam.leaderboard"
        class="mt-4"
      />
    </template>
  </PageLayout>
</template>
