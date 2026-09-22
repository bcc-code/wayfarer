<script setup lang="ts">
const props = defineProps<{
  projectId: string
  /** Participants, the denominator for every share here. */
  participants?: number
}>()

// `first: 50` matches the achievements list page's own limit. Projects run well
// under that; see cross-cutting item 1 in the note.
gql(`
  query AdminProjectEngagement($projectId: ID!) {
    challenges(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          completionCount
        }
      }
    }
    achievements(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          awardedUserCount
          imageCompletedObject {
            ...ImageFields
          }
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectEngagementQuery({
  variables: computed(() => ({ projectId: props.projectId })),
  pause: computed(() => !isAuthReady.value),
})

const share = (count: number) =>
  props.participants
    ? ` · ${Math.round((count / props.participants) * 100)} %`
    : ''

// Most-completed first: the tail is then the answer to "what is nobody doing".
const challenges = computed(() =>
  (data.value?.challenges.edges ?? [])
    .map((edge) => edge.node)
    .sort((a, b) => b.completionCount - a.completionCount)
    .map((challenge) => ({
      ...challenge,
      barWidth: props.participants
        ? `${Math.min((challenge.completionCount / props.participants) * 100, 100)}%`
        : '0%',
      tooltip: `${challenge.name}: ${formatNumber(challenge.completionCount)} av ${formatNumber(props.participants ?? 0)}${share(challenge.completionCount)}`,
    })),
)

const achievements = computed(() =>
  (data.value?.achievements.edges ?? [])
    .map((edge) => edge.node)
    .sort((a, b) => b.awardedUserCount - a.awardedUserCount)
    .map((achievement) => ({
      ...achievement,
      tooltip: `${achievement.name}: ${formatNumber(achievement.awardedUserCount)} av ${formatNumber(props.participants ?? 0)}${share(achievement.awardedUserCount)}`,
    })),
)
</script>

<template>
  <div class="@container">
    <AdminErrorState v-if="error" :error />
    <AdminLoadingState v-else-if="fetching && !data" />
    <div v-else class="space-y-5">
      <!-- The denominator once, rather than on every row. -->
      <p v-if="participants" class="text-dimmed text-xs">
        Andel av {{ formatNumber(participants) }} deltakere.
      </p>

      <div>
        <div
          class="text-muted mb-2 flex items-baseline justify-between text-xs"
        >
          <span>Utmerkelser</span>
          <NuxtLink
            :to="{
              name: 'admin-projects-projectId-achievements',
              params: { projectId },
            }"
            class="hover:text-default hover:underline"
          >
            Se alle
          </NuxtLink>
        </div>

        <p v-if="!achievements.length" class="text-dimmed text-sm">
          Ingen utmerkelser ennå.
        </p>
        <ul v-else class="flex flex-wrap gap-x-4 gap-y-3">
          <li v-for="achievement in achievements" :key="achievement.id">
            <UTooltip :text="achievement.tooltip" :delay-duration="200">
              <NuxtLink
                :to="{
                  name: 'admin-projects-projectId-achievements-achievementId',
                  params: { projectId, achievementId: achievement.id },
                }"
                class="block rounded"
              >
                <AdminEngagementRing
                  :count="achievement.awardedUserCount"
                  :total="participants"
                  :image="achievement.imageCompletedObject?.url"
                />
                <span class="sr-only">{{ achievement.name }}</span>
              </NuxtLink>
            </UTooltip>
          </li>
        </ul>
      </div>

      <div>
        <div
          class="text-muted mb-2 flex items-baseline justify-between text-xs"
        >
          <span>Utfordringer</span>
          <NuxtLink
            :to="{
              name: 'admin-projects-projectId-challenges',
              params: { projectId },
            }"
            class="hover:text-default hover:underline"
          >
            Se alle
          </NuxtLink>
        </div>

        <p v-if="!challenges.length" class="text-dimmed text-sm">
          Ingen utfordringer ennå.
        </p>
        <ul v-else class="grid gap-x-8 @3xl:grid-cols-2">
          <li v-for="challenge in challenges" :key="challenge.id">
            <UTooltip :text="challenge.tooltip" :delay-duration="200">
              <NuxtLink
                :to="{
                  name: 'admin-projects-projectId-challenges-challengeId',
                  params: { projectId, challengeId: challenge.id },
                }"
                class="hover:bg-elevated/60 -mx-1 flex items-center gap-3 rounded px-1 py-1"
              >
                <span class="min-w-0 grow truncate text-sm">
                  {{ challenge.name }}
                </span>
                <span class="w-6 shrink-0 text-right text-sm tabular-nums">
                  {{ formatNumber(challenge.completionCount) }}
                </span>
                <span
                  class="bg-elevated h-1.5 w-12 shrink-0 overflow-hidden rounded-full"
                >
                  <span
                    class="bg-primary block h-full rounded-full"
                    :style="{ width: challenge.barWidth }"
                  />
                </span>
              </NuxtLink>
            </UTooltip>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
