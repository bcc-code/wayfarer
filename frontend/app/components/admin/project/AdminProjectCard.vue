<script setup lang="ts">
interface ProjectCardProps {
  id: string
  name: string
  description: string
  startDate: string
  endDate: string
  branding: {
    logoImage?: { url: string } | null
    colors: {
      light: { accent: string }
      dark: { accent: string }
    }
  }
}

const props = defineProps<{
  project: ProjectCardProps
}>()

const colorMode = useColorMode()
const accentColor = computed(() => {
  return colorMode.value === 'dark'
    ? props.project.branding.colors.dark.accent
    : props.project.branding.colors.light.accent
})
</script>

<template>
  <UCard
    class="aspect-video shadow-md"
    :style="{ '--accent': accentColor }"
    :ui="{
      root: accentColor && 'ring-(--accent)/25 hover:ring-(--accent)/50',
      body: [
        accentColor && 'bg-(--accent)/5 hover:bg-(--accent)/10',
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
      <img
        v-if="project.branding.logoImage?.url"
        :src="project.branding.logoImage.url"
        height="32"
        width="32"
        class="rounded"
      >
      <UBadge
        v-if="isWithinRange(new Date(), project.startDate, project.endDate)"
        variant="outline"
      >
        Active
      </UBadge>
    </div>
  </UCard>
</template>
