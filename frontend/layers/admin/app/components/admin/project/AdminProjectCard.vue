<script setup lang="ts">
interface ProjectCardProps {
  id: string
  name: string
  description: string
  startDate: string
  endDate: string
  branding: {
    logoImage?: { url: string } | null
  }
}

defineProps<{
  project: ProjectCardProps
}>()
</script>

<template>
  <!--
    Content-sized, not `aspect-video`: that forced a 16:9 box and stretched cards
    in wide columns. `h-full` keeps a row matched.

    Neutral on purpose — tinting each card with its project accent made the grid
    a patchwork where colour carried no meaning.
  -->
  <UCard
    class="hover:ring-accented h-full shadow-md transition"
    :ui="{ body: 'h-full flex gap-2' }"
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
      <img
        v-if="project.branding.logoImage?.url"
        :src="project.branding.logoImage.url"
        height="32"
        class="h-8 w-auto max-w-24 rounded object-contain"
        alt=""
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
