<script setup lang="ts">
const props = defineProps<{
  challenge: ChallengesPageQuery['myCurrentProject']['challenges'][number]
}>()

const { track } = useAnalytics()

function onChallengeClick() {
  track(AnalyticsEvent.ChallengeLinkClicked, {
    challenge_id: props.challenge.id,
    challenge_name: props.challenge.name,
    is_external: !!props.challenge.url,
  })
}
</script>

<template>
  <div class="shadow-large rounded-card overflow-clip">
    <img
      v-if="challenge.image"
      :src="challenge.image"
      loading="lazy"
      class="bg-accent aspect-[1.25] w-full object-cover"
    />
    <div class="bg-background-raised p-default gap-default space-y-default">
      <div class="space-y-small">
        <h3 class="text-heading">{{ challenge.name }}</h3>
        <div class="text-label" v-html="challenge.description" />
      </div>
      <div class="mt-auto grid">
        <NuxtLink
          :to="
            challenge.url || {
              name: 'challenges-challengeId',
              params: { challengeId: challenge.id },
            }
          "
          class="contents"
          @click="onChallengeClick"
        >
          <DesignButton size="large">
            {{ challenge.buttonText }}
          </DesignButton>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
