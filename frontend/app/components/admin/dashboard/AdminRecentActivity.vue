<script setup lang="ts">
import { useTimeAgo } from '@vueuse/core'

defineProps<{
  feedbackEntries: Array<{
    id: string
    message: string
    createdAt: string
    user: { id: string; name: string }
  }>
}>()

function formatRelativeTime(dateString: string) {
  return useTimeAgo(new Date(dateString)).value
}
</script>

<template>
  <UCard>
    <template #header>
      <h3 class="font-semibold">Recent Feedback</h3>
    </template>

    <div class="space-y-3">
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
