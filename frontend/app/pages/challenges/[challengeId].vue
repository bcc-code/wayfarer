<script setup lang="ts">
const route = useRoute('challenges-challengeId')

const { data: challenge, status } = useFetch(
  () => `/api/challenges/${route.params.challengeId}`,
)
</script>

<template>
  <div>
    <template v-if="status == 'success' && challenge">
      <h1 class="text-xl font-bold mb-4">{{ challenge.name }}</h1>
      <ChallengeCard :challenge="challenge" />
    </template>
  </div>
</template>
