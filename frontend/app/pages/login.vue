<script setup lang="ts">
definePageMeta({
  layout: false,
  middleware: [],
})

const { loginWithRedirect } = useAuth()
</script>

<template>
  <div
    class="bg-background-default h-dvh flex flex-col items-center justify-end p-default"
  >
    <DesignButton size="large" class="grow-0 w-full" @click="loginWithRedirect">
      {{ $t('auth.loginButton') }}
    </DesignButton>
  </div>
</template>
