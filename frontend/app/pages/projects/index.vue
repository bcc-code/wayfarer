<script setup lang="ts">
definePageMeta({
  layout: 'admin',
})

const { data: projects, status, error } = useFetch('/api/projects')
</script>

<template>
  <UContainer>
    <h1 class="text-3xl my-8">Projects</h1>

    <ul class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
      <template v-if="status == 'success'">
        <li v-for="project in projects" :key="project.id">
          <NuxtLink :to="`/projects/${project.id}`">
            <UCard
              class="aspect-video shadow"
              :style="{ '--accent': project.color }"
              :ui="{
                root: project.color && 'ring-(--accent)/25',
                body: [project.color && 'bg-(--accent)/5', 'h-full flex gap-2'],
              }"
            >
              <div class="grow flex flex-col">
                <h3 class="font-semibold mb-2">
                  {{ project.name }}
                </h3>
                <p v-if="project.description" class="text-sm text-muted mb-2">
                  {{ project.description }}
                </p>
                <p class="text-xs font-medium text-muted mt-auto">
                  {{ formatDateRange(project.startDate, project.endDate) }}
                </p>
              </div>
              <div class="shrink-0">
                <NuxtImg
                  v-if="project.logo"
                  :src="project.logo"
                  height="32"
                  width="32"
                />
              </div>
            </UCard>
          </NuxtLink>
        </li>
      </template>
      <template v-if="status == 'pending'">
        <USkeleton v-for="i in 3" :key="i" class="aspect-video" />
      </template>
      <template v-if="status == 'error' && error">
        <p class="text-error">Error: {{ error.message }}</p>
      </template>
    </ul>
  </UContainer>
</template>
