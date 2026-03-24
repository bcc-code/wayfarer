<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: ['admin', 'superadmin'],
})

const tools = [
  {
    title: 'Fix innholdsfremgang',
    description:
      'Finner og fikser brukere som har fullfort eksterne innholdsoppgaver, men som mangler fremgangsregistreringer.',
    icon: 'lucide:file-check',
    to: '/admin/maintenance/fix-content-progress',
  },
  {
    title: 'Check points journal',
    description:
      'Viser brukere som mangler poengjournal-oppforinger for fullfort innhold.',
    icon: 'lucide:calculator',
    to: '/admin/maintenance/check-points-journal',
  },
  {
    title: 'Fix streakfremgang',
    description:
      'Finner og fikser brukere som har fullfort eksterne innholdsoppgaver for streak-prestasjoner, men som mangler fremgangsregistreringer.',
    icon: 'lucide:flame',
    to: '/admin/maintenance/fix-streak-progress',
  },
  {
    title: 'Massejobber',
    description:
      'Se oversikt over asynkrone massejobber og deres status.',
    icon: 'lucide:layers',
    to: '/admin/maintenance/bulk-jobs',
  },
]
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-8">
      <h1 class="text-3xl">Vedlikehold</h1>
      <p class="text-muted mt-2">
        Verktoy for systemadministrasjon og vedlikehold.
      </p>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <NuxtLink
        v-for="tool in tools"
        :key="tool.title"
        :to="tool.to"
        class="block"
      >
        <UCard class="hover:border-primary h-full transition-colors">
          <div class="flex h-full flex-col gap-3">
            <div class="flex items-center gap-3">
              <UIcon :name="tool.icon" class="text-primary h-6 w-6" />
              <h3 class="font-medium">{{ tool.title }}</h3>
            </div>
            <p class="text-muted text-sm">{{ tool.description }}</p>
          </div>
        </UCard>
      </NuxtLink>
    </div>
  </UContainer>
</template>
