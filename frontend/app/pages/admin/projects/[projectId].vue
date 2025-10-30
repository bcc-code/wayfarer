<script setup lang="ts">
definePageMeta({
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId')

const {
  data: project,
  status,
  error,
} = useFetch(() => `/api/projects/${route.params.projectId}`)
</script>

<template>
  <UContainer>
    <template v-if="status == 'success' && project">
      <header class="my-8">
        <div class="flex flex-col gap-6 mb-8">
          <NuxtImg
            v-if="project.logo"
            :src="project.logo"
            height="64"
            width="64"
            class="shrink-0 overflow-hidden"
          />
          <div>
            <h1 class="text-3xl font-bold mb-2">{{ project.name }}</h1>
            <p v-if="project.description" class="text-muted max-w-2xl">
              {{ project.description }}
            </p>
          </div>
        </div>

        <div class="mt-6">
          <h2 class="text-lg font-semibold mb-3">Schedule</h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <UCard>
              <div>
                <div class="text-sm text-muted mb-1">Start Date</div>
                <div class="font-medium">
                  {{ formatDate(project.startDate) }}
                </div>
              </div>
            </UCard>

            <UCard>
              <div>
                <div class="text-sm text-muted mb-1">End Date</div>
                <div class="font-medium">
                  {{ formatDate(project.endDate) }}
                </div>
              </div>
            </UCard>
          </div>
        </div>
      </header>
    </template>

    <template v-if="status == 'error' && error">
      <p class="text-error">Error: {{ error.message }}</p>
    </template>
  </UContainer>
</template>
