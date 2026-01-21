<script setup lang="ts">
import { getExtraItems } from '~/utils/leaderboard'

const AGE_RANGE_YOUNG = { min: 13, max: 18 } as const
const AGE_RANGE_ADULT = { min: 19, max: 37 } as const

const { me } = useAuth()

function getAgeRangeForAge(age: number | null | undefined) {
  if (age && age >= AGE_RANGE_ADULT.min && age <= AGE_RANGE_ADULT.max) {
    return AGE_RANGE_ADULT
  }
  return AGE_RANGE_YOUNG
}

const ageRange = ref(getAgeRangeForAge(me.value?.age))

// Update age range when user data loads (only once)
let stopAgeWatch: (() => void) | undefined
// eslint-disable-next-line prefer-const
stopAgeWatch = watch(
  () => me.value?.age,
  (age) => {
    if (age !== undefined) {
      ageRange.value = getAgeRangeForAge(age)
      stopAgeWatch?.()
    }
  },
  { immediate: true },
)

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
  return getExtraItems(
    leaderboard.value,
    data.value?.myCurrentProject.leaderboard.me,
  )
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
            label: '13 - 18',
            value: AGE_RANGE_YOUNG,
          },
          {
            key: 'o18',
            label: '19 - 36',
            value: AGE_RANGE_ADULT,
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
