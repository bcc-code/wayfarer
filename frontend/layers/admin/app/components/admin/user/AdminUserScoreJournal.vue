<script setup lang="ts">
/**
 * A user's points: per-project totals, plus the handful of most recent entries.
 *
 * It used to be a flat 100-row journal with a project column that was identical
 * on every row, under a heading naming the *current* project while the rows
 * spanned all of them — `points(projectId:)` is project-scoped and
 * `adminScoreJournal(filter: { userId })` is not, so the number and the list
 * described different things.
 *
 * Two shapes instead, because the page is used for support:
 *
 * - **Totals per project** answer "where does this person have points, and how
 *   many" — and each links to that project's journal filtered to this user,
 *   where the full, paginated log lives.
 * - **The last few entries** answer "what just happened", which is where a
 *   support conversation usually starts.
 *
 * There is deliberately **no cross-project total**. Projects differ in length
 * and scoring scale, so a lifetime sum is a number nobody can act on — the same
 * vanity metric that came off the home dashboard.
 */
interface ProjectPoints {
  projectId: string
  projectName: string
  points: number
}

interface ScoreEntry {
  id: string
  points: number
  sourceType: string
  reason?: string | null
  createdAt: string
  project: { id: string; name: string }
}

const props = defineProps<{
  /** Per-project totals, most recently active project first. */
  pointsByProject: ProjectPoints[]
  /** The most recent entries across all projects. */
  recent: ScoreEntry[]
  /** Total rows in the journal, so the window can say what it is showing. */
  totalCount: number
  /** Lets every link point at this user's entries. */
  userId: string
  /** Marked in the list so the project being worked on is findable. */
  currentProjectId?: string
}>()

const isTruncated = computed(() => props.totalCount > props.recent.length)

function formatSourceType(type: string) {
  return type.charAt(0) + type.slice(1).toLowerCase()
}

const scoresRoute = (projectId: string) => ({
  name: 'admin-projects-projectId-scores' as const,
  params: { projectId },
  query: { userId: props.userId },
})
</script>

<template>
  <AdminSection title="Poeng">
    <div v-if="pointsByProject.length || recent.length" class="space-y-5">
      <!-- Totals per project -->
      <div v-if="pointsByProject.length" class="space-y-1">
        <NuxtLink
          v-for="entry in pointsByProject"
          :key="entry.projectId"
          :to="scoresRoute(entry.projectId)"
          class="hover:bg-elevated -mx-2 flex items-center justify-between gap-4 rounded px-2 py-2 transition-colors"
        >
          <span class="flex min-w-0 items-center gap-2">
            <span class="truncate font-medium">{{ entry.projectName }}</span>
            <UBadge
              v-if="entry.projectId === currentProjectId"
              variant="subtle"
              size="xs"
            >
              Aktivt
            </UBadge>
          </span>
          <span class="flex shrink-0 items-center gap-2">
            <span class="font-medium">
              {{ formatNumber(entry.points) }} poeng
            </span>
            <UIcon name="lucide:chevron-right" class="text-dimmed size-4" />
          </span>
        </NuxtLink>
      </div>

      <!-- Recent activity across projects -->
      <div v-if="recent.length">
        <h3 class="text-muted mb-1 text-sm font-medium">Siste aktivitet</h3>
        <div class="divide-default/60 divide-y">
          <div
            v-for="entry in recent"
            :key="entry.id"
            class="flex items-center justify-between gap-4 py-2"
          >
            <div class="flex min-w-0 items-center gap-3">
              <UBadge
                :color="entry.points >= 0 ? 'success' : 'error'"
                variant="soft"
              >
                {{ entry.points >= 0 ? '+' : ''
                }}{{ formatNumber(entry.points) }}
              </UBadge>
              <div class="min-w-0">
                <!--
                  The project is shown here because this list *does* span
                  projects — unlike the old flat log, where the same name
                  repeated on every row.
                -->
                <NuxtLink
                  :to="scoresRoute(entry.project.id)"
                  class="truncate font-medium hover:underline"
                >
                  {{ entry.project.name }}
                </NuxtLink>
                <UBadge variant="subtle" size="xs" class="ml-2">
                  {{ formatSourceType(entry.sourceType) }}
                </UBadge>
              </div>
            </div>
            <div class="shrink-0 text-right">
              <div
                v-if="entry.reason"
                class="text-dimmed max-w-xs truncate text-sm"
              >
                {{ entry.reason }}
              </div>
              <div class="text-dimmed text-xs">
                {{ formatDateTime(entry.createdAt) }}
              </div>
            </div>
          </div>
        </div>
        <p v-if="isTruncated" class="text-dimmed pt-2 text-xs">
          Viser de {{ recent.length }} nyeste av
          {{ formatNumber(totalCount) }} oppføringer — åpne et prosjekt over for
          hele loggen.
        </p>
      </div>
    </div>
    <p v-else class="text-dimmed text-sm">Ingen poengoppføringer</p>
  </AdminSection>
</template>
