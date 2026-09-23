<script setup lang="ts">
import type { StandingsPageQuery } from '~/api/generated'
import { getExtraItems } from '~/utils/leaderboard'

/**
 * One admin-configured board, already computed by the server. The page fetches
 * every board in one query — `leaderboardConfig(id)` is admin-only, so a user
 * can only reach configs through `myCurrentProject.leaderboards` — so this
 * component only renders; it has no query of its own.
 */
type Board = StandingsPageQuery['myCurrentProject']['leaderboards'][number]

const props = defineProps<{ board: Board }>()

const entries = computed(() =>
  props.board.leaderboard.edges.map((edge) => edge.node),
)

const extraItems = computed(() =>
  getExtraItems(entries.value, props.board.leaderboard.me),
)
</script>

<template>
  <div>
    <div
      class="p-medium gap-medium mb-list-section-gap flex flex-col items-center"
    >
      <h2 class="text-heading text-center text-balance">{{ board.name }}</h2>
    </div>
    <LeaderboardList
      v-if="entries.length"
      :leaderboard="entries"
      :extra-items="extraItems"
    />
    <EmptyState v-else :title="$t('emptyStates.standings')" />
  </div>
</template>
