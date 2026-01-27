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
              label: $t('admin.churchHome.kickOff'),
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-6">
      <UButton
        color="neutral"
        variant="soft"
        size="lg"
        :to="{ name: 'admin-my-church' }"
      >
        <Icon name="lucide:arrow-left" />
        {{ $t('admin.common.back') }}
      </UButton>

      <div class="mt-12 max-w-2xl relative">
        <h2 class="text-3xl font-semibold mb-4">
          {{ $t('admin.churchHome.kickOff') }}
        </h2>
        <p>{{ $t('admin.churchHome.kickOffDescription') }}</p>

        <div
          class="relative rounded-full dark:bg-red-900 bg-red-800 aspect-square mt-20 w-64 transition duration-100 ease-100 shadow-2xl has-active:shadow-none"
        >
          <button
            class="rounded-[inherit] bg-red-500 p-8 text-black text-xl font-semibold size-full scale-105 -translate-y-6 active:translate-y-0 active:scale-100 transition duration-100 ease-out ring-2 ring-red-400 ring-inset cursor-pointer"
          >
            {{ $t('admin.churchHome.kickOffButton') }}
          </button>
        </div>
      </div>
    </UContainer>
  </div>
</template>
