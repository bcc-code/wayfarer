<script setup lang="ts">
definePageMeta({
  layout: 'admin',
})

const route = useRoute()

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
        <h1 class="text-3xl text-balance mb-4">{{ project.name }}</h1>
        <p
          v-if="project.description"
          class="text-muted text-pretty max-w-md text-sm"
        >
          {{ project.description }}
        </p>
      </header>
    </template>

    <template v-if="status == 'error' && error">
      <p class="text-red-500">Error: {{ error.message }}</p>
    </template>
  </UContainer>
</template>
