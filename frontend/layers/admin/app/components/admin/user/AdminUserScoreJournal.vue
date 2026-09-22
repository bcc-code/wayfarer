<script setup lang="ts">
interface ScoreEntry {
  id: string
  points: number
  sourceType: string
  reason?: string | null
  createdAt: string
  project: { id: string; name: string }
}

const props = defineProps<{
  entries: ScoreEntry[]
  totalCount: number
  /** The user's total in the scoped project. */
  points: number
  /**
   * The project the **points total** is scoped to — not the journal.
   *
   * `points(projectId:)` is per project; `adminScoreJournal(filter: { userId })`
   * is not, so its rows span every project the user has scored in. Labelling
   * the whole section "Poenglogg i <project>" was wrong: the heading claimed a
   * scope the rows below it did not have. The name belongs on the badge.
   */
  projectName?: string
  /** Lets each row link to that project's score journal for this user. */
  userId: string
}>()

const isTruncated = computed(() => props.totalCount > props.entries.length)

function formatSourceType(type: string) {
  return type.charAt(0) + type.slice(1).toLowerCase()
}
</script>

<template>
  <AdminSection title="Poenglogg" :count="totalCount">
    <template #actions>
      <UBadge color="neutral" variant="soft">
        {{ points }} poeng<template v-if="projectName">
          i {{ projectName }}</template
        >
      </UBadge>
    </template>

    <!--
      Dividers only in the long lists, and faint: they guide a scan down 24
      rows, they are not structure.
    -->
    <div v-if="entries.length" class="divide-default/60 divide-y">
      <div
        v-for="entry in entries"
        :key="entry.id"
        class="flex items-center justify-between gap-4 py-3"
      >
        <div class="flex items-center gap-3">
          <UBadge
            :color="entry.points >= 0 ? 'success' : 'error'"
            variant="soft"
          >
            {{ entry.points >= 0 ? '+' : '' }}{{ formatNumber(entry.points) }}
          </UBadge>
          <div>
            <!--
              The project name links to that project's score journal filtered
              to this user — the "see everything" target for a log that spans
              projects, where one section-level link could not point anywhere
              meaningful.
            -->
            <NuxtLink
              :to="{
                name: 'admin-projects-projectId-scores',
                params: { projectId: entry.project.id },
                query: { userId },
              }"
              class="font-medium hover:underline"
            >
              {{ entry.project.name }}
            </NuxtLink>
            <UBadge variant="subtle" size="xs" class="ml-2">
              {{ formatSourceType(entry.sourceType) }}
            </UBadge>
          </div>
        </div>
        <div class="text-right">
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
      <p v-if="isTruncated" class="text-dimmed pt-3 text-center text-sm">
        Viser {{ entries.length }} av {{ totalCount }} oppføringer
      </p>
    </div>
    <p v-else class="text-dimmed text-sm">Ingen poengoppføringer</p>
  </AdminSection>
</template>
