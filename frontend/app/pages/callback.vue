<script setup lang="ts">
import { useAuth0 } from '@auth0/auth0-vue'

definePageMeta({
  layout: false,
  middleware: [],
})

const auth0 = useAuth0()
const { exchangeToken } = useAuth()
const { track } = useAnalytics()

// Use a global state to prevent multiple concurrent callbacks across component re-mounts
const processing = useState('callback-processing', () => false)

onMounted(async () => {
  // Only process once globally
  if (processing.value) {
    return
  }

  processing.value = true

  try {
    // Handle the Auth0 callback and get the redirect target
    const result = await auth0.handleRedirectCallback()
    const targetUrl = result.appState?.targetUrl || '/'

    // Exchange Auth0 token for Wayfarer JWT
    const success = await exchangeToken()
    if (success) {
      track(AnalyticsEvent.LoginCompleted)
      // Use replace to prevent back button issues
      await navigateTo(targetUrl, { replace: true })
    } else {
      console.error('Failed to exchange token')
    }
  } catch (error) {
    console.error('Callback error:', error)
  } finally {
    processing.value = false
  }
})
</script>

<template>
  <div class="h-dvh">
    <LoadingState />
  </div>
</template>
