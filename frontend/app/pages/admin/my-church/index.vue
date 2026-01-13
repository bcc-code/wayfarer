<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

const { canManageChurchAdmins } = usePermissions()
</script>

<template>
  <UContainer class="py-12 flex items-center justify-center h-3/4">
    <div class="grid xl:grid-cols-3 gap-12 text-4xl">
      <NuxtLink
        v-if="canManageChurchAdmins"
        :to="{ name: 'admin-my-church-admins' }"
        class="rounded-2xl bg-muted p-12 gap-2 flex flex-col items-center justify-center text-center hover:bg-accented"
      >
        <Icon name="lucide:badge-check" class="size-8 text-dimmed shrink-0" />
        {{ $t('admin.churchHome.administrators') }}
      </NuxtLink>
      <NuxtLink
        :to="{ name: 'admin-my-church-units' }"
        class="rounded-2xl bg-muted p-12 gap-2 flex flex-col items-center justify-center text-center hover:bg-accented"
      >
        <Icon name="lucide:users" class="size-8 text-dimmed shrink-0" />
        {{ $t('admin.churchHome.unitsAndPeople') }}
      </NuxtLink>
      <NuxtLink
        class="rounded-2xl border-2 border-dashed gap-2 border-muted p-12 flex flex-col items-center justify-center text-center"
      >
        <Icon name="lucide:gamepad-2" class="size-8 text-dimmed shrink-0" />
        {{ $t('admin.churchHome.gameNights') }}
        <p class="text-lg text-muted">
          {{ $t('admin.churchHome.comingSoon') }}
        </p>
      </NuxtLink>
    </div>
  </UContainer>
</template>
