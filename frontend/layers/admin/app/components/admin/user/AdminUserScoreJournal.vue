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
   * The project everything here is filtered to. Named in the header because
   * both the total and the journal are project-scoped, and an unlabelled
   * "Poenglogg / N poeng" reads as a lifetime figure.
   */
  projectName?: string
}>()

const isTruncated = computed(() => props.totalCount > props.entries.length)

function formatSourceType(type: string) {
  return type.charAt(0) + type.slice(1).toLowerCase()
}
</script>

<template>
  <AdminSection title="Poenglogg" :count="totalCount">
    <template #actions>
      <div class="flex items-center gap-2 text-sm">
        <span v-if="projectName" class="text-muted">i {{ projectName }}</span>
        <UBadge color="neutral" variant="soft">{{ points }} poeng</UBadge>
      </div>
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
            <span class="font-medium">{{ entry.project.name }}</span>
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
