<script setup lang="ts">
import '~/assets/styles/admin.css'

// Force Norwegian locale in admin
const { setLocale } = useI18n()
setLocale('nb')

useHead({
  title: 'Interact Admin',
})

const { me, logout } = useAuth()
</script>

<template>
  <div class="bg-default h-full">
    <header class="border-default border-b">
      <UContainer class="py-3 flex items-center gap-4">
        <NuxtLink :to="{ name: 'admin-my-church' }">
          <UColorModeImage
            light="/images/logo/logo.svg"
            dark="/images/logo/logo-light.svg"
            class="h-6"
          />
        </NuxtLink>
        <div class="ml-auto flex gap-4 items-center">
          <AdminUserFeedback />
          <div class="text-end flex flex-col">
            <span class="text-sm">{{ me?.name }}</span>
          </div>
          <UButton variant="soft" color="neutral" @click="logout">
            {{ $t('auth.logoutButton') }}
          </UButton>
        </div>
      </UContainer>
    </header>
    <slot />
  </div>
</template>
