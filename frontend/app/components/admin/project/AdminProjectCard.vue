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
    <div class="flex grow flex-col">
      <h3 class="mb-2 font-semibold">
        {{ project.name }}
      </h3>
      <p
        v-if="project.description"
        class="text-muted mb-2 line-clamp-3 truncate text-sm whitespace-normal"
      >
        {{ project.description }}
      </p>
      <p class="text-muted mt-auto text-xs font-medium">
        {{ formatDateRange(project.startDate, project.endDate) }}
      </p>
    </div>
    <div class="flex shrink-0 flex-col items-end justify-between">
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
