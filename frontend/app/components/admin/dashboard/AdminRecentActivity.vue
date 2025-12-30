<script setup lang="ts">
import { useTimeAgo } from '@vueuse/core'

const props = defineProps<{
  scoreEntries: Array<{
    id: string
    points: number
    sourceType: string
    reason?: string | null
    createdAt: string
    user: { id: string; name: string; image?: string | null }
    project: { id: string; name: string }
  }>
  feedbackEntries: Array<{
    id: string
    message: string
    createdAt: string
    user: { id: string; name: string }
  }>
}>()

const activeTab = ref('scores')

function formatRelativeTime(dateString: string) {
  return useTimeAgo(new Date(dateString)).value
}
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <h3 class="font-semibold">Recent Activity</h3>
        <UTabs
          v-model="activeTab"
          :items="[
            { value: 'scores', label: 'Score Changes' },
            { value: 'feedback', label: 'Feedback' },
          ]"
          size="sm"
        />
      </div>
    </template>

    <div v-if="activeTab === 'scores'" class="space-y-3">
      <div
        v-for="entry in scoreEntries"
        :key="entry.id"
        class="flex items-center gap-3"
      >
        <UAvatar
          :src="entry.user.image ?? undefined"
          :alt="entry.user.name"
          size="sm"
        />
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm">
            <span class="font-medium">{{ entry.user.name }}</span>
            <span class="text-muted">
              {{ entry.points >= 0 ? 'earned' : 'lost' }}
            </span>
            <UBadge :color="entry.points >= 0 ? 'success' : 'error'" size="xs">
              {{ entry.points >= 0 ? '+' : '' }}{{ entry.points }}
            </UBadge>
          </p>
          <p class="text-muted text-xs">
            {{ entry.project.name }} &middot;
            {{ formatRelativeTime(entry.createdAt) }}
          </p>
        </div>
      </div>
      <p v-if="!scoreEntries.length" class="text-muted text-center text-sm">
        No recent score changes
      </p>
      <UButton variant="ghost" size="sm" to="/admin/scores" class="w-full">
        View all score changes
      </UButton>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="entry in feedbackEntries"
        :key="entry.id"
        class="border-default rounded-lg border p-3"
      >
        <p class="line-clamp-2 text-sm">{{ entry.message }}</p>
        <p class="text-muted mt-1 text-xs">
          {{ entry.user.name }} &middot;
          {{ formatRelativeTime(entry.createdAt) }}
        </p>
      </div>
      <p v-if="!feedbackEntries.length" class="text-muted text-center text-sm">
        No recent feedback
      </p>
      <UButton variant="ghost" size="sm" to="/admin/feedback" class="w-full">
        View all feedback
      </UButton>
    </div>
  </UCard>
</template>
