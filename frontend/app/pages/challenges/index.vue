<script setup lang="ts">
const { data: challenges, status } = useFetch('/api/challenges')
</script>

<template>
  <div>
    <h1 class="text-xl font-bold mb-4">Challenges</h1>
    <div v-if="status == 'success' && challenges" class="space-y-4">
      <ChallengeCard
        v-for="challenge in challenges"
        :key="challenge.id"
        :challenge
      />
    </div>
  </div>
</template>
