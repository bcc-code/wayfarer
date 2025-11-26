<script setup lang="ts">
gql(`
  query StandingsUnitPage {
    myCurrentProject {
      id
      myTeam {
        id
        name
        memberLeaderboard {
          id
          name
          tags
          rank
          score
        }
      }
    }
  }
`)

const { isTeamLead } = useAuth()
const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsUnitPageQuery({
  pause: computed(() => !isAuthReady.value),
})

// Update team
// const { executeMutation } = useUpdateTeamMutation()
</script>

<template>
  <div>
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <div
        v-if="data.myCurrentProject.myTeam"
        class="p-medium gap-medium mb-list-section-gap flex flex-col items-center"
      >
        <h2 class="text-heading text-balance">
          {{ data.myCurrentProject.myTeam.name }}
        </h2>
        <DesignButton v-if="isTeamLead" variant="secondary" size="medium">
          {{ $t('standings.editUnit') }}
        </DesignButton>
      </div>
      <LeaderboardList
        v-if="data.myCurrentProject.myTeam?.memberLeaderboard?.length"
        :leaderboard="data.myCurrentProject.myTeam.memberLeaderboard"
        :badge="
          (entry) =>
            entry.tags?.includes(LeaderboardEntryTag.TeamLead)
              ? $t('standings.unitLeader')
              : undefined
        "
        hide-medals
      />
    </template>
    <EmptyState v-else />
  </div>
</template>
