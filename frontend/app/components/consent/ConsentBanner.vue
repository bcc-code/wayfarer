<script setup lang="ts">
defineEmits<{
  close: []
}>()
</script>

<template>
  <DesignBanner :title="$t('consent.bannerTitle')" @close="$emit('close')">
    <template #action>
      <NuxtLink :to="{ name: 'settings-consent' }">
        <DesignButton size="large" class="w-full">
          {{ $t('consent.bannerButton') }}
        </DesignButton>
      </NuxtLink>
    </template>
  </DesignBanner>
</template>
