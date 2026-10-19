<script setup lang="ts">
import StandingsBoard from '#layers/user/app/components/standings/StandingsBoard.vue'

/**
 * The board as it stands, rendered by the participant app's own
 * `StandingsBoard`. The filter fields say what was configured; only the list
 * says who that lets through.
 *
 * It shows the **saved** configuration: the server computes the board from the
 * stored filter, and there is no query that takes an unsaved one. Edits appear
 * once they are saved, which the label above the preview says outright.
 */
const props = defineProps<{ configId: string }>()

gql(`
  query AdminLeaderboardPreview($id: ID!) {
    leaderboardConfig(id: $id) {
      id
      name
      leaderboard {
        totalCount
        edges {
          node {
            ...LeaderboardEntryWithDescriptionFields
          }
        }
        me {
          ...LeaderboardEntryWithDescriptionFields
        }
        nearestChurchRivals {
          ...LeaderboardEntryWithDescriptionFields
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminLeaderboardPreviewQuery({
  variables: computed(() => ({ id: props.configId })),
  pause: computed(() => !isAuthReady.value),
})

const board = computed(() => data.value?.leaderboardConfig)
</script>

<template>
  <div
    class="bg-background-default aspect-[9/19.5] w-full overflow-y-auto rounded-xl py-4"
  >
    <AdminLoadingState v-if="fetching && !board" />
    <AdminErrorState v-else-if="error" :error />
    <StandingsBoard
      v-else-if="board"
      :board="board as InstanceType<typeof StandingsBoard>['$props']['board']"
    />
  </div>
</template>
