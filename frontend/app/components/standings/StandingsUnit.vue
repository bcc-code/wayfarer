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

const teamLeader = computed(() => {
  return data.value?.myCurrentProject.myTeam?.memberLeaderboard.find((entry) =>
    entry.tags?.includes(LeaderboardEntryTag.TeamLead),
  )
})

// Update team
const { executeMutation } = useUpdateTeamMutation()
function saveChanges() {
  const id = data.value?.myCurrentProject.myTeam?.id
  if (!id) return

  // executeMutation({
  //   id,
  //   input: { name: data.value?.myCurrentProject.myTeam?.name },
  // })
}
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
        <UModal
          v-if="isTeamLead"
          :ui="{ content: 'bg-background-default' }"
          :transition="false"
          modal
          fullscreen
        >
          <DesignButton variant="secondary" size="medium">
            {{ $t('unit.editUnit') }}
          </DesignButton>
          <template #content="{ close }">
            <PageLayout :title="$t('unit.editUnit')">
              <template #action>
                <DesignIconButton icon="lucide:x" @click="close" />
              </template>
              <div class="gap-list-section-gap flex h-full flex-col">
                <DesignInput
                  v-model="data.myCurrentProject.myTeam.name"
                  :label="$t('unit.unitName')"
                />
                <DesignPanel>
                  <div class="flex items-center gap-2.5 px-3 py-2">
                    <Icon name="lucide:badge-check" class="size-6" />
                    <span class="text-label">{{ $t('unit.unitLeader') }}</span>
                    <DesignButton
                      variant="secondary"
                      size="small"
                      :class="[
                        'ml-auto grow-0',
                        { 'text-text-hint': !teamLeader?.name },
                      ]"
                    >
                      {{ teamLeader?.name ?? $t('unit.noUnitLeader') }}
                    </DesignButton>
                  </div>
                </DesignPanel>
                <div class="p-default flex h-full flex-col justify-end">
                  <DesignButton
                    size="large"
                    class="grow-0"
                    @click="saveChanges"
                  >
                    {{ $t('unit.saveChanges') }}
                  </DesignButton>
                </div>
              </div>
            </PageLayout>
          </template>
        </UModal>
      </div>
      <LeaderboardList
        v-if="data.myCurrentProject.myTeam?.memberLeaderboard?.length"
        :leaderboard="data.myCurrentProject.myTeam.memberLeaderboard"
        :badge="
          (entry) =>
            entry.tags?.includes(LeaderboardEntryTag.TeamLead)
              ? $t('unit.unitLeader')
              : undefined
        "
        hide-medals
      />
    </template>
    <EmptyState v-else />
  </div>
</template>
