<script setup lang="ts">
interface Team {
  id: string
  name: string
  parentProject: { id: string; name: string }
}

defineProps<{ teams: Team[] }>()
</script>

<template>
  <UCard>
    <template #header>
      <h2 class="text-xl font-semibold">
        Lag
        <span v-if="teams.length" class="text-dimmed text-sm font-normal">
          ({{ teams.length }})
        </span>
      </h2>
    </template>

    <div v-if="teams.length" class="space-y-2">
      <NuxtLink
        v-for="team in teams"
        :key="team.id"
        :to="{
          name: 'admin-projects-projectId-teams-teamId',
          params: { projectId: team.parentProject.id, teamId: team.id },
        }"
        class="border-default hover:bg-elevated flex items-center justify-between rounded-md border p-3 transition-colors"
      >
        <div>
          <span class="font-medium">{{ team.name }}</span>
          <div class="text-muted text-xs">{{ team.parentProject.name }}</div>
        </div>
        <UIcon name="lucide:chevron-right" class="text-dimmed size-4" />
      </NuxtLink>
    </div>
    <div v-else class="text-dimmed">Ikke med i noen lag</div>
  </UCard>
</template>
