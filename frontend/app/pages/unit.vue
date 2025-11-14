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
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useUnitPageQuery({
  pause: computed(() => !isAuthReady.value),
})
</script>

<template>
  <PageLayout :title="$t('pages.unit')">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <!-- <LeaderboardList
      v-else-if="data?.myCurrentProject.myTeam?.leaderboard.length"
      :leaderboard="data.myCurrentProject.myTeam.leaderboard"
      variant="expanded"
    /> -->
    <EmptyState v-else />
  </PageLayout>
</template>
