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
    Sized by its content, not by a ratio. `aspect-video` forced a 16:9 box, so a
    card in a wide column stretched to ~320px tall for three lines of text and a
    date. Cards in a row still match: the grid stretches them and `h-full`
    carries that down.

    Uniform neutral, deliberately. Tinting each card with its own project accent
    turned the grid into a patchwork where the colour carried no meaning — the
    logo already identifies the project, and the accent belongs on that
    project's own pages.
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
