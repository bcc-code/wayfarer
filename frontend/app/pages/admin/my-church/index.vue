<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})
</script>

<template>
  <UContainer class="py-12 flex items-center justify-center h-3/4">
    <div class="space-y-12 text-4xl">
      <NuxtLink
        :to="{ name: 'admin-my-church-units' }"
        class="rounded-2xl bg-muted p-12 gap-2 flex flex-col items-center justify-center text-center hover:bg-accented"
      >
        <Icon name="lucide:users" class="size-8 text-dimmed shrink-0" />
        Units og personer
      </NuxtLink>
      <NuxtLink
        class="rounded-2xl border-2 border-dashed gap-2 border-muted p-12 flex flex-col items-center justify-center text-center"
      >
        <Icon name="lucide:gamepad-2" class="size-8 text-dimmed shrink-0" />
        Game Nights
        <p class="text-lg text-muted">Kommer snart</p>
      </NuxtLink>
    </div>
  </UContainer>
</template>
