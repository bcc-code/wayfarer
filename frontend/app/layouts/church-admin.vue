<script setup lang="ts">
import '~/assets/styles/admin.css'

// Force Norwegian locale in admin
const { setLocale } = useI18n()
setLocale('nb')

useHead({
  title: 'Interact Admin',
})
</script>

<template>
  <div class="bg-default h-full">
    <header class="border-default border-b">
      <UContainer class="py-3">
        <UColorModeImage
          light="/images/logo/logo.svg"
          dark="/images/logo/logo-light.svg"
          class="h-6"
        />
      </UContainer>
    </header>
    <slot />
  </div>
</template>
