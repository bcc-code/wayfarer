<script setup lang="ts">
defineProps<{
  projectId: string
  projectName: string
  entries: Array<{
    id: string
    name: string
    score: number
    rank?: number | null
    image?: string | null
  }>
  totalCount: number
}>()
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <h3 class="font-semibold">Toppliste</h3>
        <UBadge variant="subtle">{{ projectName }}</UBadge>
      </div>
    </template>

    <div class="space-y-2">
      <div
        v-for="(entry, index) in entries"
        :key="entry.id"
        class="flex items-center gap-3"
      >
        <span
          :class="[
            'flex size-6 items-center justify-center rounded-full text-xs font-bold',
            index === 0 &&
              'bg-yellow-500/20 text-yellow-600 dark:text-yellow-400',
            index === 1 && 'bg-gray-500/20 text-gray-600 dark:text-gray-400',
            index === 2 &&
              'bg-orange-500/20 text-orange-600 dark:text-orange-400',
            index > 2 && 'bg-muted text-muted',
          ]"
        >
          {{ entry.rank ?? index + 1 }}
        </span>
        <span class="min-w-0 flex-1 truncate text-sm">{{ entry.name }}</span>
        <span class="text-muted text-sm font-medium">
          {{ formatNumber(entry.score) }} p
        </span>
      </div>
      <p v-if="!entries.length" class="text-muted text-center text-sm">
        Ingen deltakere ennå
      </p>
    </div>

    <template #footer>
      <div class="flex items-center justify-between text-sm">
        <span class="text-muted">{{ totalCount }} deltakere</span>
        <NuxtLink
          :to="{ name: 'admin-projects-projectId', params: { projectId } }"
          class="text-primary hover:underline"
        >
          Se prosjekt
        </NuxtLink>
      </div>
    </template>
  </UCard>
</template>
