<script setup lang="ts">
import type { Challenge } from '#shared/types'

defineProps<{
  challenge: Challenge
}>()
</script>

<template>
  <UCard :ui="{ body: 'flex flex-col aspect-video' }">
    <template #default>
      <h3 class="mb-1">{{ challenge.name }}</h3>
      <p class="text-sm text-muted mb-4">{{ challenge.description }}</p>
      <UButton
        :to="
          challenge.url ?? {
            name: 'challenges-challengeId',
            params: { challengeId: challenge.id },
          }
        "
        block
        class="mt-auto"
      >
        Solve
      </UButton>
    </template>
  </UCard>
</template>
