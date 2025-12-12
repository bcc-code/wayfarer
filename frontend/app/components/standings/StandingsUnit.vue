<script setup lang="ts">
const { isTeamLead } = useAuth()
const { isAuthReady } = useAuthReady()
const {
  data,
  error,
  fetching,
  executeQuery: refetch,
} = useStandingsUnitPageQuery({
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

  refetch()
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
                nested
              >
                <template #content>
                  <DesignPanel class="gap-list-section-inset flex flex-col">
                    <template
                      v-for="(member, index) in teamMembers"
                      :key="member.id"
                    >
                      <hr v-if="index > 0" class="border-border-default mx-3" />
                      <button
                        class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
                        @click="selectTeamLead(member.id)"
                      >
                        <p class="text-label">{{ member.name }}</p>
                        <Icon
                          v-if="member.id === form.teamLeadId"
                          name="lucide:check"
                          class="size-6"
                        />
                      </button>
                    </template>
                  </DesignPanel>
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
