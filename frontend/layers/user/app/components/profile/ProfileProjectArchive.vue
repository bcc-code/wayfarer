<script setup lang="ts">
/**
 * "Earlier projects" — the user's past projects and the achievements they
 * earned in them, below the current project card on the home tab.
 *
 * Owns its own query, which stays paused until the section is expanded: the
 * home tab's critical path is `ProfilePage`, and the archive must not compete
 * with it on first paint. urql caches the result, so collapsing and
 * re-expanding does not refetch.
 */
const props = defineProps<{
  /** Dropped from the list — it already has its own card above. */
  currentProjectId?: string
}>()

const { isAuthReady } = useAuthReady()

const expanded = ref(false)

const { data, error, fetching } = useProjectArchiveQuery({
  pause: computed(() => !isAuthReady.value || !expanded.value),
})

const projects = computed(() => {
  const all = data.value?.me.projects ?? []
  return all
    .filter((project) => project.id !== props.currentProjectId)
    .sort(
      (a, b) =>
        new Date(b.startDate).getTime() - new Date(a.startDate).getTime(),
    )
})

const isInitialLoading = computed(() => fetching.value && !data.value)
const isEmpty = computed(
  () => !fetching.value && !error.value && !projects.value.length,
)
</script>

<template>
  <section>
    <button
      class="flex w-full items-center justify-between gap-small p-medium"
      :aria-expanded="expanded"
      @click="expanded = !expanded"
    >
      <span class="text-label text-text-muted">
        {{ $t('archive.earlierProjects') }}
      </span>
      <IconChevronRight
        class="text-text-hint size-5 transition-transform duration-200"
        :class="expanded && 'rotate-90'"
      />
    </button>

    <div v-if="expanded" class="space-y-list-section-gap">
      <div v-if="isInitialLoading" class="p-medium gap-medium grid grid-cols-4">
        <DesignSkeleton
          v-for="i in 8"
          :key="i"
          class="aspect-square w-full rounded-full"
        />
      </div>
      <ErrorState v-else-if="error" :error />
      <p
        v-else-if="isEmpty"
        class="text-caption text-text-hint p-medium text-center"
      >
        {{ $t('archive.empty') }}
      </p>
      <template v-else>
        <div v-for="project in projects" :key="project.id">
          <p class="text-label text-text-hint p-medium text-center">
            {{ project.name }}
          </p>
          <div
            v-if="project.achievements.length"
            class="p-medium gap-medium grid grid-cols-4 pt-0"
          >
            <AchievementBadge
              v-for="achievement in project.achievements"
              :key="achievement.id"
              :achievement
            />
          </div>
        </div>
      </template>
    </div>
  </section>
</template>
