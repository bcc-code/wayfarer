<script setup lang="ts">
defineEmits<{
  close: []
}>()
</script>

<template>
  <DesignBanner
    :title="$t('consent.bannerTitle')"
    :dismissable="false"
    @close="$emit('close')"
  >
    <template #action>
      <NuxtLink :to="{ name: 'settings-consent' }">
        <DesignButton size="large" class="w-full">
          {{ $t('consent.bannerButton') }}
          <Icon name="lucide:arrow-right" />
        </DesignButton>
      </NuxtLink>
    </template>
  </DesignBanner>
</template>
