<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})
</script>

<template>
  <UContainer class="py-12 flex items-center justify-center h-3/4">
    <div class="grid lg:grid-cols-2 gap-8 text-4xl">
      <NuxtLink
        :to="{ name: 'admin-my-church-units' }"
        class="rounded-2xl bg-muted p-8 gap-2 aspect-video h-64 flex flex-col items-center justify-center text-center hover:bg-accented"
      >
        <Icon name="lucide:users" class="size-8 text-dimmed" />
        Units
      </NuxtLink>
      <NuxtLink
        class="rounded-2xl border-2 border-dashed gap-2 border-muted p-8 aspect-video h-64 flex flex-col items-center justify-center text-center"
      >
        <Icon name="lucide:gamepad-2" class="size-8 text-dimmed" />
        Game Nights
        <p class="text-lg text-muted">Kommer snart</p>
      </NuxtLink>
    </div>
  </UContainer>
</template>
