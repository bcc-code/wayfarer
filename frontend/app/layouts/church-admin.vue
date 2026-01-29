<script setup lang="ts">
import '~/assets/styles/admin.css'

useHead({
  title: 'Interact Admin',
})

const { me, logout } = useAuth()

// Initialize Firestore sync for realtime updates
const { initialize: initFirestoreSync } = useFirestoreSync()
onMounted(() => {
  initFirestoreSync()
})

// PWA update notification
const { $pwa } = useNuxtApp()
const toast = useToast()

watch(
  () => $pwa?.needRefresh,
  (needRefresh) => {
    if (needRefresh) {
      toast.add({
        id: 'pwa-update',
        title: 'Oppdatering tilgjengelig',
        description: 'En ny versjon av appen er klar.',
        icon: 'lucide:download',
        close: false,
        duration: 0,
        color: 'neutral',
        actions: [
          {
            label: 'Oppdater nå',
            color: 'neutral',
            onClick: () => $pwa?.updateServiceWorker(true),
          },
        ],
      })
    }
  },
  { immediate: true },
)
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
        <div class="ml-auto flex gap-2 items-center">
          <AdminUserFeedback />
          <div class="text-end">
            <span class="text-sm">{{ me?.name }}</span>
          </div>
          <AdminColorModeSelector />
          <AdminLocaleSelector />
          <UButton variant="soft" color="neutral" @click="logout">
            {{ $t('auth.logoutButton') }}
          </UButton>
        </div>
      </UContainer>
    </header>
    <slot />
  </div>
</template>
