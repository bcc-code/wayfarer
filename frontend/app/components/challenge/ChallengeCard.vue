<script setup lang="ts">
defineProps<{
  challenge: ChallengesPageQuery['myCurrentProject']['challenges'][number]
}>()
</script>

<template>
  <div class="shadow-large rounded-card overflow-clip">
    <NuxtImg
      v-if="challenge.image"
      :src="challenge.image"
      class="bg-accent aspect-[1.25] w-full object-cover"
    />
    <div class="bg-background-raised p-default gap-default space-y-default">
      <div class="space-y-small">
        <h3 class="text-heading">{{ challenge.name }}</h3>
        <div class="text-label" v-html="challenge.description" />
      </div>
      <div class="gap-small mt-auto grid grid-cols-2">
        <DesignButton size="large" variant="secondary"> Decline </DesignButton>
        <NuxtLink
          :to="
            challenge.url ?? {
              name: 'challenges-challengeId',
              params: { challengeId: challenge.id },
            }
          "
          class="contents"
        >
          <DesignButton size="large">Accept</DesignButton>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
