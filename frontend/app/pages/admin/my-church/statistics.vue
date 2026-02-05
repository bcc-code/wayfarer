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
              label: $t('admin.churchHome.statistics'),
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-6">
      <!-- Stats here -->
    </UContainer>
  </div>
</template>
