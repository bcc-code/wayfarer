<script setup lang="ts">
const route = useRoute('callback')

definePageMeta({
  layout: false,
})

const config = useRuntimeConfig()
const { setAccessToken } = useAuth()

onBeforeMount(async () => {
  const { token, redirect } = route.query
  if (token) {
    try {
      const response = await $fetch<{ token: string }>(
        config.public.callbackUrl,
        { method: 'GET' },
      )
      if (response && response.token) {
        setAccessToken(response.token)

        if (redirect && typeof redirect === 'string') {
          navigateTo(redirect)
        } else {
          navigateTo('/')
        }
      }
    } catch (e) {
      console.error(e)
    }
  }
})
</script>

<template>
  <div class="h-dvh">
    <LoadingState />
  </div>
</template>
