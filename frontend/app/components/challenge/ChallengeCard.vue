<script setup lang="ts">
import type { ChallengesPageQuery } from '~/api/generated'

defineProps<{
  challenge: ChallengesPageQuery['user']['currentProject']['challenges'][number]
}>()
</script>

<template>
  <UCard :ui="{ body: 'flex flex-col', header: 'p-0' }">
    <template #header>
      <NuxtImg
        v-if="challenge.image"
        :src="challenge.image"
        class="w-full aspect-video object-cover"
      />
    </template>
    <template #default>
      <h3 class="mb-1 font-bold">{{ challenge.name }}</h3>
      <div class="text-sm text-muted mb-4" v-html="challenge.description" />
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
        {{ challenge.buttonText ?? 'Accept Challenge' }}
      </UButton>
    </template>
  </UCard>
</template>
