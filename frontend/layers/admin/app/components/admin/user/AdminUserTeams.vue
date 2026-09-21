<script setup lang="ts">
interface Team {
  id: string
  name: string
  parentProject: { id: string; name: string }
}

defineProps<{ teams: Team[] }>()
</script>

<template>
  <AdminSection title="Lag" :count="teams.length">
    <!--
      Divided rows, not boxed ones: each row used to carry its own border
      inside a bordered card, so every item was a box in a box.
    -->
    <div v-if="teams.length" class="space-y-1">
      <NuxtLink
        v-for="team in teams"
        :key="team.id"
        :to="{
          name: 'admin-projects-projectId-teams-teamId',
          params: { projectId: team.parentProject.id, teamId: team.id },
        }"
        class="hover:bg-elevated -mx-2 flex items-center justify-between rounded px-2 py-2 transition-colors"
      >
        <div>
          <span class="font-medium">{{ team.name }}</span>
          <div class="text-muted text-xs">{{ team.parentProject.name }}</div>
        </div>
        <UIcon name="lucide:chevron-right" class="text-dimmed size-4" />
      </NuxtLink>
    </div>
    <p v-else class="text-dimmed text-sm">Ikke med i noen lag</p>
  </AdminSection>
</template>
