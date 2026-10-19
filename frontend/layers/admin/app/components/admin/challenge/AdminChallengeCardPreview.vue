<script setup lang="ts">
import ChallengeCard from '#layers/user/app/components/challenges/ChallengeCard.vue'

/**
 * The challenge as a participant will see it — the user layer's own
 * `ChallengeCard`, not a copy of it. A replica drifts: this panel had one, and
 * it had already lost the card's completed state and its empty-image
 * behaviour.
 *
 * What is left here is the adapter. The card reads a challenge off the query
 * that feeds the app; the form holds a draft, so the missing half is filled in
 * with what a freshly created challenge would have.
 */
const props = defineProps<{
  challenge: {
    type?: ChallengeType
    name?: string
    description?: string
    image?: string
    url?: string
    buttonText?: string
  }
}>()

const typenames: Record<ChallengeType, string> = {
  [ChallengeType.Simple]: 'SimpleChallenge',
  [ChallengeType.External]: 'ExternalChallenge',
  [ChallengeType.Quiz]: 'QuizChallenge',
  [ChallengeType.Plugin]: 'PluginChallenge',
}

// Shaped like the query result the card is typed against. The cast is the
// point of the adapter: a draft has no id, no completion and no submissions,
// and the card only reads them.
const preview = computed(
  () =>
    ({
      __typename: typenames[props.challenge.type ?? ChallengeType.Simple],
      id: 'preview',
      name: props.challenge.name ?? '',
      description: props.challenge.description ?? '',
      buttonText: props.challenge.buttonText ?? '',
      imageObject: props.challenge.image
        ? { url: props.challenge.image }
        : null,
      url: props.challenge.url ?? '',
      userCompletedAt: null,
      quiz: { userSubmissions: [] },
    }) as unknown as InstanceType<typeof ChallengeCard>['$props']['challenge'],
)
</script>

<template>
  <ChallengeCard :challenge="preview" />
</template>
