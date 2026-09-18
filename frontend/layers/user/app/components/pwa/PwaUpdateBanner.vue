<script setup lang="ts">
const { $pwa } = useNuxtApp()
</script>

<template>
  <div class="px-navigation-outside z-20">
    <DesignPanel
      class="gradient-border relative shadow-large! mx-auto w-full max-w-xl"
    >
      <div class="p-medium flex flex-col gap-medium">
        <p class="text-title text-text-default text-center">
          {{ $t('pwa.updateAvailableTitle') }}
        </p>
        <DesignButton
          variant="primary"
          size="large"
          @click="$pwa?.updateServiceWorker(true)"
        >
          {{ $t('pwa.updateAvailableButton') }}
        </DesignButton>
      </div>
    </DesignPanel>
  </div>
</template>
