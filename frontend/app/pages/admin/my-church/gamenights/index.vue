<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: $t('admin.breadcrumb.home'),
              to: { name: 'admin-my-church' },
            },
            {
              label: $t('admin.churchHome.gameNights'),
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-6 relative">
      <UButton
        color="neutral"
        variant="soft"
        size="lg"
        :to="{ name: 'admin-my-church' }"
      >
        <Icon name="lucide:arrow-left" />
        {{ $t('admin.common.back') }}
      </UButton>

      <h1 class="text-3xl font-semibold mt-12">
        {{ $t('admin.churchHome.gameNights') }}
      </h1>

      <div class="flex flex-wrap gap-12 mt-12 text-4xl">
        <NuxtLink
          v-for="gamenight in 2"
          :key="gamenight"
          :to="{
            name: 'admin-my-church-gamenights-gamenight-gamenight',
            params: { gamenight },
          }"
          class="rounded-2xl grow bg-muted p-12 gap-2 flex flex-col items-center justify-center text-center hover:bg-accented data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50"
          :data-disabled="gamenight === 2"
        >
          <Icon name="lucide:gamepad-2" class="size-8 text-dimmed shrink-0" />
          {{ $t('admin.churchHome.gameNight', { number: gamenight }) }}
        </NuxtLink>
      </div>
    </UContainer>
  </div>
</template>
