<script setup lang="ts">
gql(`
  query StandingsLocalPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter) {
    me {
      church {
        id
        name
      }
    }
    myCurrentProject {
      id
      leaderboard(entityType: $entityType, filter: $filter) {
        edges {
          node {
            id
            name
            score
            rank
            tags
          }
        }
        me {
          id
          name
          score
          rank
          tags
        }
      }
    }
  }
`)

const entityType = ref(LeaderboardEntityType.Persons)

const { me } = useAuth()
const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useStandingsLocalPageQuery({
  variables: computed(() => ({
    entityType: entityType.value,
    filter: {
      churchId: me.value?.church.id,
      ageRange: { min: 14, max: 36 },
    },
  })),
  pause: computed(() => !isAuthReady.value || !me.value?.church.id),
})

const leaderboard = computed<Partial<LeaderboardEntry>[]>(() => {
  if (!data.value) return []

  const result = []
  result.push(
    ...data.value.myCurrentProject.leaderboard.edges.map((edge) => edge.node),
  )
  const me = data.value?.myCurrentProject.leaderboard.me
  if (me && !result.find((entry) => entry.id === me.id)) {
    result.push(me)
  }
  return result
})
</script>

<template>
  <div>
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <div
        v-if="data.me.church"
        class="p-medium gap-medium mb-list-section-gap flex flex-col items-center"
      >
        <h2 class="text-heading text-center text-balance">
          {{ data.me.church.name }}
        </h2>
      </div>
      <DesignTabs
        v-model="entityType"
        :tabs="[
          {
            label: $t('standings.personal'),
            value: LeaderboardEntityType.Persons,
            icon: 'IconUser',
          },
          {
            label: $t('standings.units'),
            value: LeaderboardEntityType.Teams,
            icon: 'IconUsers',
          },
        ]"
        class="mb-list-section-gap"
        variant="secondary"
      >
        <template #tab="{ tab }">
          <div class="flex flex-col items-center gap-0.5">
            <Icon :name="tab.icon" class="size-7" />
            <span>{{ tab.label }}</span>
          </div>
        </template>
      </DesignTabs>
      <LeaderboardList
        v-if="leaderboard?.length"
        :leaderboard="leaderboard"
        hide-medals
      />
    </template>
    <EmptyState v-else />
  </div>
</template>
