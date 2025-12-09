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

const teamMembers = computed(() => {
  return data.value?.myCurrentProject.myTeam?.memberLeaderboard ?? []
})

// Update team
const form = reactive({
  name: '',
  teamLeadId: null as string | null,
})
watch(
  () => data.value,
  (d) => {
    if (d) {
      form.name = d.myCurrentProject.myTeam?.name ?? ''
      form.teamLeadId = teamLeader.value?.id ?? null
    }
  },
  { once: true },
)

const selectedTeamLeader = computed(() => {
  if (!form.teamLeadId) return null
  return teamMembers.value.find((m) => m.id === form.teamLeadId)
})

const { executeMutation } = useUpdateTeamMutation()
const { executeMutation: assignTeamLead } = useAssignTeamLeadMutation()

async function saveChanges() {
  const id = data.value?.myCurrentProject.myTeam?.id
  if (!id) return

  // Update team name
  await executeMutation({
    id,
    input: { name: form.name },
  })

  // Assign team lead if changed
  if (form.teamLeadId && form.teamLeadId !== teamLeader.value?.id) {
    await assignTeamLead({ teamId: id, userId: form.teamLeadId })
  }

  reloadNuxtApp()
}

// Team lead selector
const showLeadSelector = ref(false)

function selectTeamLead(userId: string) {
  form.teamLeadId = userId
  showLeadSelector.value = false
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
        <DesignDrawer v-if="isTeamLead" :title="$t('unit.editUnit')">
          <DesignButton variant="secondary" size="medium">
            {{ $t('unit.editUnit') }}
          </DesignButton>
          <template #content>
            <div class="gap-list-section-gap flex grow flex-col">
              <DesignInput v-model="form.name" :label="$t('unit.unitName')" />
              <DesignPanel>
                <div class="flex items-center gap-2.5 px-3 py-2">
                  <Icon name="lucide:badge-check" class="size-6" />
                  <span class="text-label">{{ $t('unit.unitLeader') }}</span>
                  <DesignButton
                    variant="secondary"
                    size="small"
                    :class="[
                      'ml-auto grow-0',
                      { 'text-text-hint': !selectedTeamLeader?.name },
                    ]"
                    @click="showLeadSelector = true"
                  >
                    {{ selectedTeamLeader?.name ?? $t('unit.noUnitLeader') }}
                  </DesignButton>
                </div>
              </DesignPanel>

              <DesignDrawer
                v-model:open="showLeadSelector"
                :title="$t('unit.selectUnitLeader')"
              >
                <template #content>
                  <div class="gap-list-section-gap flex flex-col">
                    <DesignPanel
                      v-for="member in teamMembers"
                      :key="member.id"
                      class="cursor-pointer"
                      @click="selectTeamLead(member.id)"
                    >
                      <div class="flex items-center gap-2.5 px-3 py-2">
                        <span class="text-label flex-1">
                          {{ member.name }}
                        </span>
                        <Icon
                          v-if="member.id === form.teamLeadId"
                          name="lucide:check"
                          class="text-accent size-5"
                        />
                      </div>
                    </DesignPanel>
                  </div>
                </template>
              </DesignDrawer>
              <div class="p-default flex grow flex-col justify-end">
                <DesignButton size="large" class="grow-0" @click="saveChanges">
                  {{ $t('unit.saveChanges') }}
                </DesignButton>
              </div>
            </div>
          </template>
        </DesignDrawer>
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
