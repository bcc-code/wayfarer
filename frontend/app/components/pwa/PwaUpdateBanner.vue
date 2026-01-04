<script setup lang="ts">
const { $pwa } = useNuxtApp()
</script>

<template>
  <div class="px-navigation-outside z-20">
    <DesignPanel class="gradient-border relative shadow-large!">
      <div class="p-medium flex flex-col">
        <p class="text-title text-text-default">
          {{ $t('pwa.updateAvailableTitle') }}
        </p>
        <p class="text-label text-text-muted mb-medium mt-small">
          {{ $t('pwa.updateAvailableDescription') }}
        </p>
        <DesignButton
          variant="secondary"
          size="large"
          @click="$pwa?.updateServiceWorker(true)"
        >
          {{ $t('pwa.updateAvailableButton') }}
        </DesignButton>
      </div>
    </DesignPanel>
  </div>
</template>
