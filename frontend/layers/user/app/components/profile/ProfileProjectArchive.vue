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

const { data, error } = useProjectArchiveQuery({
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

// The query only runs once expanded, so "nothing yet and no error" is the
// loading state — more reliable than `fetching`, which is briefly false
// between unpausing and the request going out.
const isInitialLoading = computed(() => !data.value && !error.value)
const isEmpty = computed(() => !!data.value && !projects.value.length)

const listRef = ref<HTMLElement | null>(null)
const { animate } = useStaggeredEntrance({ totalDuration: 0.6 })

watch(projects, (list) => {
  if (!list.length) return
  nextTick(() => {
    const items = listRef.value?.querySelectorAll('.archive-project')
    if (items?.length) animate(items)
  })
})
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
        class="text-text-hint size-5 transition-transform duration-300"
        :class="expanded && 'rotate-90'"
      />
    </button>

    <Transition
      enter-active-class="grid transition-all duration-300 ease-out"
      enter-from-class="grid-rows-[0fr] opacity-0"
      enter-to-class="grid-rows-[1fr] opacity-100"
      leave-active-class="grid transition-all duration-200 ease-in"
      leave-from-class="grid-rows-[1fr] opacity-100"
      leave-to-class="grid-rows-[0fr] opacity-0"
    >
      <div v-if="expanded" class="grid grid-rows-[1fr]">
        <div class="min-h-0 overflow-hidden">
          <div v-if="isInitialLoading" class="space-y-list-section-gap">
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
          <p
            v-else-if="isEmpty"
            class="text-caption text-text-hint p-medium text-center"
          >
            {{ $t('archive.empty') }}
          </p>
          <div v-else ref="listRef" class="space-y-list-section-gap">
            <div
              v-for="project in projects"
              :key="project.id"
              class="archive-project"
            >
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
          </div>
        </div>
      </div>
    </Transition>
  </section>
</template>
