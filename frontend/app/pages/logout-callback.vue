<script setup lang="ts">
definePageMeta({
  layout: false,
  middleware: [],
})

const wayfarerToken = useLocalStorage<string>('token', () => null)

onMounted(() => {
  // Clear the Wayfarer token after Auth0 logout
  wayfarerToken.value = null
  // Redirect to home (will trigger login)
  navigateTo('/', { replace: true })
})
</script>

<template>
  <div class="h-dvh">
    <LoadingState />
  </div>
</template>
