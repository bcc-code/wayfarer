<script setup lang="ts">
const config = useRuntimeConfig()
const isStaging = computed(() => config.public.isStaging || import.meta.dev)
</script>

<template>
  <UApp>
    <NuxtPwaAssets />
    <NuxtLayout>
      <NuxtRouteAnnouncer />
      <NuxtPage />
    </NuxtLayout>
    <div
      v-if="isStaging"
      class="fixed inset-0 border-2 border-accent-negative z-50 pointer-events-none"
    >
      <div
        class="bg-accent-negative w-max px-2 pb-0.75 pt-0 text-xs text-black rounded-b-md mx-auto"
      >
        DEVELOPMENT
      </div>
    </div>
  </UApp>
</template>
