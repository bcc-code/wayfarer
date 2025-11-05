<script setup lang="ts">
defineProps<{
  project: AdminProjectsPageQuery['projects'][number]
}>()
</script>

<template>
  <UCard
    class="aspect-video shadow-md"
    :style="{ '--accent': project.branding.colors.primary }"
    :ui="{
      root:
        project.branding.colors.primary &&
        'ring-(--accent)/25 hover:ring-(--accent)/50',
      body: [
        project.branding.colors.primary &&
          'bg-(--accent)/5 hover:bg-(--accent)/10',
        'h-full flex gap-2',
      ],
    }"
  >
    <div class="grow flex flex-col">
      <h3 class="font-semibold mb-2">
        {{ project.name }}
      </h3>
      <p
        v-if="project.description"
        class="text-sm text-muted mb-2 truncate line-clamp-3 whitespace-normal"
      >
        {{ project.description }}
      </p>
      <p class="text-xs font-medium text-muted mt-auto">
        {{ formatDateRange(project.startDate, project.endDate) }}
      </p>
    </div>
    <div class="shrink-0 flex flex-col justify-between items-end">
      <NuxtImg
        v-if="project.branding.logo"
        :src="project.branding.logo"
        height="32"
        width="32"
        class="rounded"
      />
      <UBadge
        v-if="isWithinRange(new Date(), project.startDate, project.endDate)"
        variant="outline"
      >
        Active
      </UBadge>
    </div>
  </UCard>
</template>
