<script setup lang="ts">
const props = defineProps<{
  challenge: ChallengesPageQuery['myCurrentProject']['challenges'][number]
}>()

const { track } = useAnalytics()

const externalUrl = computed(() => {
  if (props.challenge.__typename === 'ExternalChallenge') {
    return props.challenge.url
  }
  return null
})

function onChallengeClick() {
  track(AnalyticsEvent.ChallengeLinkClicked, {
    challenge_id: props.challenge.id,
    challenge_name: props.challenge.name,
    is_external: !!externalUrl.value,
  })
}

const isCompleted = computed(() => {
  if (props.challenge.__typename === 'QuizChallenge') {
    return props.challenge.quiz.userSubmissions.some(
      (s) => s.completedAt && !s.autoSubmitted,
    )
  }
  return props.challenge.userCompletedAt !== null
})
</script>

<template>
  <div class="shadow-large rounded-card overflow-clip">
    <DesignImage
      v-if="challenge.imageObject"
      :image="challenge.imageObject"
      :alt="challenge.name"
      class="aspect-[1.25] w-full"
    />
    <div class="bg-background-raised p-default gap-default space-y-default">
      <div class="space-y-small">
        <h3 class="text-heading">{{ challenge.name }}</h3>
        <div class="text-label" v-html="challenge.description" />
      </div>
      <div v-if="challenge.buttonText" class="mt-auto grid">
        <NuxtLink
          :to="
            externalUrl || {
              name: 'challenges-challengeId',
              params: { challengeId: challenge.id },
            }
          "
          class="contents"
          @click="onChallengeClick"
        >
          <DesignButton
            size="large"
            :variant="isCompleted ? 'secondary' : 'primary'"
          >
            {{ challenge.buttonText }}
          </DesignButton>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
