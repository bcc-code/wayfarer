<script setup lang="ts">
const ageRange = ref({ min: 13, max: 19 })

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsGlobalPageQuery({
  variables: computed(() => ({
    entityType: LeaderboardEntityType.Persons,
    first: 20,
    filter: {
      ageRange: ageRange.value,
    },
  })),
  pause: computed(() => !isAuthReady.value),
})

const leaderboard = computed<LeaderboardEntry[]>(() => {
  if (!data.value) return []
  return data.value.myCurrentProject.leaderboard.edges.map((edge) => edge.node)
})

const extraItems = computed<LeaderboardEntry[]>(() => {
  const me = data.value?.myCurrentProject.leaderboard.me
  if (me && !leaderboard.value.find((entry) => entry.id === me.id)) {
    return [me]
  }
  return []
})

// Only show loading state on initial load, not on refetch
const isInitialLoading = computed(() => fetching.value && !data.value)
</script>

<template>
  <div>
    <StandingsListSkeleton v-if="isInitialLoading" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="leaderboard?.length">
      <div
        class="p-medium gap-medium mb-list-section-gap flex flex-col items-center"
      >
        <h2 class="text-heading text-center text-balance">
          {{ $t('standings.top', { amount: 20 }) }}
        </h2>
      </div>
      <DesignTabs
        v-model="ageRange"
        :tabs="[
          {
            key: 'u18',
            label: $t('standings.u18'),
            value: { min: 13, max: 19 },
          },
          {
            key: 'o18',
            label: $t('standings.o18'),
            value: { min: 20, max: 37 },
          },
        ]"
        class="mb-list-section-gap"
        variant="secondary"
      >
        <template #tab="{ tab }">
          <div class="flex flex-col items-center gap-0.5">
            <IconBaby v-if="tab.key === 'u18'" class="size-7" />
            <IconSmile v-else class="size-7" />
            <span>{{ tab.label }}</span>
          </div>
        </template>
      </DesignTabs>
      <LeaderboardList :leaderboard="leaderboard" :extra-items="extraItems" />
    </template>
    <EmptyState v-else :title="$t('emptyStates.standings')" />
  </div>
</template>
