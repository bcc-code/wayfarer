<script setup lang="ts">
definePageMeta({
  layout: 'admin',
})
</script>

<template>
  <UContainer>
    <h1 class="text-2xl my-8">Home</h1>
    <div class="flex gap-2 mt-4">
      <UButton variant="solid">Solid Button</UButton>
      <UButton variant="soft">Soft Button</UButton>
      <UButton variant="ghost">Ghost Button</UButton>
    </div>
  </UContainer>
</template>
