<script setup lang="ts">
import { useAuth0 } from '@auth0/auth0-vue'
import { until } from '@vueuse/core'

definePageMeta({
  layout: false,
  middleware: [],
})

const route = useRoute()
const auth0 = useAuth0()
const { exchangeToken, me, isLoading } = useAuth()
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
    // Wait for Auth0 SDK to be ready before handling callback
    // This prevents "Invalid state" errors when SDK hasn't loaded stored state yet
    while (auth0.isLoading.value) {
      await new Promise((resolve) => setTimeout(resolve, 10))
    }

    // Check if we have Auth0 callback params (code and state)
    const hasCallbackParams = route.query.code && route.query.state

    let targetUrl = '/'

    if (hasCallbackParams) {
      // Handle the Auth0 callback and get the redirect target
      const result = await auth0.handleRedirectCallback()
      targetUrl = result.appState?.target || '/'
    }

    // Exchange Auth0 token for Wayfarer JWT if authenticated
    if (auth0.isAuthenticated.value) {
      const success = await exchangeToken()
      if (success) {
        // Wait for user data to load to check if this is a signup
        await until(isLoading).toBe(false)

        if (me.value?.createdAt) {
          const createdAt = new Date(me.value.createdAt)
          const now = new Date()
          const isNewUser = now.getTime() - createdAt.getTime() < 60_000

          if (isNewUser) {
            track(AnalyticsEvent.SignupCompleted)
          }
        }

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
    // Don't log errors about missing query params - this is expected when
    // navigating directly to /auth0-callback without OAuth flow
    const message = error instanceof Error ? error.message : String(error)
    if (!message.includes('query params')) {
      console.error('Callback error:', error)
    }
    await navigateTo('/', { replace: true })
  } finally {
    processing.value = false
  }
})

const showResetButton = ref(false)
useCountdown(10, {
  immediate: true,
  onComplete: () => (showResetButton.value = true),
})
function resetAndRetry() {
  processing.value = false
  auth0.logout({
    logoutParams: {
      returnTo: window.location.origin,
    },
  })
}
</script>

<template>
  <div class="h-dvh">
    <LoadingState>
      <DesignButton
        v-if="showResetButton"
        variant="secondary"
        size="small"
        class="grow-0"
        @click="resetAndRetry"
      >
        {{ $t('error.retry') }}
      </DesignButton>
    </LoadingState>
  </div>
</template>
