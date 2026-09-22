<script setup lang="ts">
type Badge = {
  id: string
  name: string
  awardedUserCount: number
  imageCompletedObject?: { url: string } | null
}

const props = defineProps<{
  achievements: Badge[]
  projectId: string
  /** Participants, the denominator for each ring. */
  participants?: number
  size?: number
}>()

// Most-awarded first, so the tail answers "what is nobody earning".
const badges = computed(() =>
  [...props.achievements]
    .sort((a, b) => b.awardedUserCount - a.awardedUserCount)
    .map((achievement) => ({
      ...achievement,
      tooltip: props.participants
        ? `${achievement.name}: ${formatNumber(achievement.awardedUserCount)} av ${formatNumber(props.participants)} · ${Math.round((achievement.awardedUserCount / props.participants) * 100)} %`
        : `${achievement.name}: ${formatNumber(achievement.awardedUserCount)}`,
    })),
)
</script>

<template>
  <ul class="flex flex-wrap gap-x-4 gap-y-3">
    <li v-for="badge in badges" :key="badge.id">
      <UTooltip :text="badge.tooltip" :delay-duration="200">
        <NuxtLink
          :to="{
            name: 'admin-projects-projectId-achievements-achievementId',
            params: { projectId, achievementId: badge.id },
          }"
          class="block rounded"
        >
          <AdminEngagementRing
            :count="badge.awardedUserCount"
            :total="participants"
            :image="badge.imageCompletedObject?.url"
            :size="size"
          />
          <span class="sr-only">{{ badge.name }}</span>
        </NuxtLink>
      </UTooltip>
    </li>
  </ul>
</template>
