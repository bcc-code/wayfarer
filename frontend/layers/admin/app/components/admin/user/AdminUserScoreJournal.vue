<script setup lang="ts">
// Totals per project plus the most recent entries, for support cases: "where
// does this person have points" and "what just happened". No cross-project
// total — projects differ in length and scale, so the sum means nothing.
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
  /** Most recently active project first. */
  pointsByProject: ProjectPoints[]
  recent: ScoreEntry[]
  totalCount: number
  userId: string
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
