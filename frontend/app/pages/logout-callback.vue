<script setup lang="ts">
definePageMeta({
  layout: false,
  middleware: [],
})

onMounted(() => {
  // Clear the Wayfarer session after Auth0 logout
  clearSessionTokens()
  // Redirect to home (will trigger login)
  navigateTo('/', { replace: true })
})
</script>

<template>
  <div class="h-dvh">
    <LoadingState />
  </div>
</template>
