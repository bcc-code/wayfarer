<script setup lang="ts">
import { useAuth0 } from '@auth0/auth0-vue'

definePageMeta({
  layout: false,
  middleware: [],
})

const route = useRoute()
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
    // Check if we have Auth0 callback params (code and state)
    const hasCallbackParams = route.query.code && route.query.state

    let targetUrl = '/'

    if (hasCallbackParams) {
      // Handle the Auth0 callback and get the redirect target
      const result = await auth0.handleRedirectCallback()
      targetUrl = result.appState?.targetUrl || '/'
    }

    // Wait for Auth0 to be ready
    while (auth0.isLoading.value) {
      await new Promise((resolve) => setTimeout(resolve, 10))
    }

    // Exchange Auth0 token for Wayfarer JWT if authenticated
    if (auth0.isAuthenticated.value) {
      const success = await exchangeToken()
      if (success) {
        track(AnalyticsEvent.LoginCompleted)
        // Use replace to prevent back button issues
        await navigateTo(targetUrl, { replace: true })
      } else {
        console.error('Failed to exchange token')
        await navigateTo('/', { replace: true })
      }
    } else {
      // Not authenticated, redirect to home (will trigger login)
      await navigateTo('/', { replace: true })
    }
  } catch (error) {
    console.error('Callback error:', error)
    await navigateTo('/', { replace: true })
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
