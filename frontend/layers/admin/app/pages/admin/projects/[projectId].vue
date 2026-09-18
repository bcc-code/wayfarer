<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  permission: 'projects:view',
})

// Deliberately thin. The project data lives in `useCurrentProject()` so the
// layout's sidebar, switcher and navbar can read it too — they are ancestors of
// this page and could not see anything it provided.
//
// Note: this route is *unnamed*. `unrouting` drops the name from a parent whose
// children include one at path '' (unrouting/dist/index.mjs:271), so
// `admin-projects-projectId` still belongs to `[projectId]/index.vue`. Use a
// bare `useRoute()` here, never the typed overload for that name.
const { error } = useCurrentProject()
</script>

<template>
  <ErrorState v-if="error" :error />
  <NuxtPage v-else />
</template>
