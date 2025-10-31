<script setup lang="ts">
const { data: challenges, status, error } = useFetch('/api/challenges')
</script>

<template>
  <div class="h-full">
    <h1 class="text-xl font-bold mb-4">Challenges</h1>

    <LoadingState v-if="status == 'pending'" />
    <ErrorState v-else-if="status == 'error'" :error="error?.data" />
    <div v-else-if="status == 'success' && challenges" class="space-y-4">
      <ChallengeCard
        v-for="challenge in challenges"
        :key="challenge.id"
        :challenge
      />
    </div>
  </div>
</template>
