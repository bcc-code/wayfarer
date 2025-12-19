<script setup lang="ts">
type Platform = 'ios' | 'android'

// Auto-detect platform from user agent
function detectPlatform(): Platform {
  if (typeof navigator === 'undefined') return 'ios'
  const ua = navigator.userAgent
  if (/iPhone|iPad|iPod/i.test(ua)) return 'ios'
  if (/Android/i.test(ua)) return 'android'
  return 'ios' // Default for desktop
}

const selectedPlatform = ref<Platform>(detectPlatform())

const tabs = [
  { key: 'ios', label: 'iOS', value: 'ios' as Platform },
  { key: 'android', label: 'Android', value: 'android' as Platform },
]

const images = computed(() => {
  const prefix = `/images/installation/${selectedPlatform.value}`
  return [`${prefix}-1.png`, `${prefix}-2.png`, `${prefix}-3.png`]
})
</script>

<template>
  <PageLayout :title="$t('pages.addToHomeScreen')">
    <template #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="IconClose" />
      </NuxtLink>
    </template>

    <div class="space-y-medium p-default">
      <DesignTabs v-model="selectedPlatform" :tabs="tabs" />

      <div class="space-y-medium">
        <img
          v-for="(src, index) in images"
          :key="src"
          :src="src"
          :alt="`Step ${index + 1}`"
          class="w-full rounded-lg"
        />
      </div>
    </div>
  </PageLayout>
</template>
