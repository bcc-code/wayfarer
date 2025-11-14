<script setup lang="ts">
const route = useRoute('callback')

definePageMeta({
  layout: false,
  middleware: [],
})

const config = useRuntimeConfig()
const { setAccessToken } = useAuth()

// Use a global state to prevent multiple concurrent callbacks across component re-mounts
const processing = useState('callback-processing', () => false)

onMounted(async () => {
  // Only process once globally
  if (processing.value) {
    return
  }

  processing.value = true

  const { token, redirect } = route.query
  if (!token || typeof token !== 'string') {
    processing.value = false
    return
  }

  try {
    const response = await $fetch<{ token: string }>(
      `${config.public.callbackUrl}?token=${token}`,
      { method: 'GET' },
    )

    if (response && response.token) {
      setAccessToken(response.token)

      const redirectPath =
        redirect && typeof redirect === 'string' ? redirect : '/'

      // Use replace to prevent back button issues
      await navigateTo(redirectPath, { replace: true })
      processing.value = false
    } else {
      processing.value = false
    }
  } catch {
    processing.value = false
  }
})
</script>

<template>
  <div class="h-dvh">
    <LoadingState />
  </div>
</template>
