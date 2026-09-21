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
  <UCard>
    <template #header>
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-xl font-semibold">
          Poenglogg
          <span v-if="projectName" class="text-muted text-sm font-normal">
            i {{ projectName }}
          </span>
          <UBadge color="neutral" variant="soft">{{ points }} poeng</UBadge>
          <span v-if="totalCount" class="text-dimmed text-sm font-normal">
            ({{ totalCount }} oppføringer)
          </span>
        </h2>
      </div>
    </template>

    <div v-if="entries.length" class="space-y-2">
      <div
        v-for="entry in entries"
        :key="entry.id"
        class="border-default flex items-center justify-between rounded-md border p-3"
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
      <div v-if="isTruncated" class="text-dimmed pt-2 text-center text-sm">
        Viser {{ entries.length }} av {{ totalCount }} oppføringer
      </div>
    </div>
    <div v-else class="text-dimmed">Ingen poengoppføringer</div>
  </UCard>
</template>
