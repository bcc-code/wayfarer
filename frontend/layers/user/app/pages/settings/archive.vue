<script setup lang="ts">
/**
 * The current project is merged in explicitly rather than taken from
 * `me.projects`, which only lists projects with a `user_projects` row — a row
 * written as a side effect of joining a team, enrolling in a challenge or
 * earning a content achievement, so a user can have achievements in a project
 * and still be missing from it.
 */
const { isAuthReady } = useAuthReady()

const { data, error, fetching } = useProjectArchiveQuery({
  pause: computed(() => !isAuthReady.value),
})

const projects = computed(() => {
  const current = data.value?.myCurrentProject
  const byId = new Map(
    [...(current ? [current] : []), ...(data.value?.me.projects ?? [])].map(
      (project) => [project.id, project],
    ),
  )
  return [...byId.values()].sort(
    (a, b) => new Date(b.startDate).getTime() - new Date(a.startDate).getTime(),
  )
})

const isInitialLoading = computed(() => fetching.value && !data.value)
const isEmpty = computed(() => !!data.value && !projects.value.length)
</script>

<template>
  <PageLayout :title="$t('archive.myAchievements')">
    <template #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="IconClose" />
      </NuxtLink>
    </template>

    <div
      v-if="isInitialLoading"
      class="space-y-list-section-gap p-list-outside"
    >
      <div v-for="i in 2" :key="i">
        <div class="p-medium flex justify-center">
          <DesignSkeleton class="h-4 w-40 rounded" />
        </div>
        <div class="p-medium gap-medium grid grid-cols-4 pt-0">
          <DesignSkeleton
            v-for="j in 8"
            :key="j"
            class="aspect-square w-full rounded-full"
          />
        </div>
      </div>
    </div>
    <ErrorState v-else-if="error" :error />
    <EmptyState v-else-if="isEmpty" :title="$t('archive.empty')" />
    <div v-else class="space-y-list-section-gap p-list-outside">
      <template v-for="project in projects" :key="project.id">
        <div v-if="project.achievements.length">
          <p class="text-label text-text-hint p-medium">
            {{ project.name }}
          </p>
          <div class="p-medium gap-medium grid grid-cols-4 pt-0">
            <AchievementBadge
              v-for="achievement in project.achievements"
              :key="achievement.id"
              :achievement
            />
          </div>
        </div>
      </template>
    </div>
  </PageLayout>
</template>
