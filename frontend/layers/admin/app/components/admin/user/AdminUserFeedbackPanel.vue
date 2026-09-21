<script setup lang="ts">
/**
 * Named `…Panel` because `AdminUserFeedback` is already taken — that is the
 * "Gi oss tilbakemelding" widget in the admin shell. Components are registered
 * in one flat namespace (`pathPrefix: false`), so the directory does not
 * disambiguate them.
 */
interface FeedbackEntry {
  id: string
  message: string
  createdAt: string
  canContactMe: boolean
  platform?: string | null
  screenWidth?: number | null
  screenHeight?: number | null
  appVersion?: string | null
}

const props = defineProps<{
  entries: FeedbackEntry[]
  totalCount: number
}>()

/**
 * The panel shows at most what the query asked for, so the "showing N of M"
 * line is derived from the rows actually present rather than hardcoded — it
 * previously claimed 10 while rendering up to 100.
 */
const isTruncated = computed(() => props.totalCount > props.entries.length)
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h2 class="text-xl font-semibold">
          Tilbakemeldinger
          <span v-if="totalCount" class="text-dimmed text-sm font-normal">
            ({{ totalCount }}
            {{ totalCount === 1 ? 'oppføring' : 'oppføringer' }})
          </span>
        </h2>
        <UButton variant="ghost" size="sm" :to="{ name: 'admin-feedback' }">
          Vis alle
        </UButton>
      </div>
    </template>

    <div v-if="entries.length" class="space-y-3">
      <div
        v-for="entry in entries"
        :key="entry.id"
        class="border-default rounded-md border p-3"
      >
        <div class="flex items-start justify-between gap-4">
          <p class="text-sm whitespace-pre-wrap">{{ entry.message }}</p>
          <UBadge
            :color="entry.canContactMe ? 'success' : 'neutral'"
            variant="soft"
            class="shrink-0"
          >
            {{ entry.canContactMe ? 'Kan kontaktes' : 'Ikke kontakt' }}
          </UBadge>
        </div>
        <div
          class="text-dimmed mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs"
        >
          <span>{{ formatDateTime(entry.createdAt) }}</span>
          <span v-if="entry.platform">{{ entry.platform }}</span>
          <span v-if="entry.screenWidth && entry.screenHeight">
            {{ entry.screenWidth }}x{{ entry.screenHeight }}
          </span>
          <code v-if="entry.appVersion">v{{ entry.appVersion }}</code>
        </div>
      </div>
      <div v-if="isTruncated" class="text-dimmed pt-2 text-center text-sm">
        Viser {{ entries.length }} av {{ totalCount }} oppføringer
      </div>
    </div>
    <div v-else class="text-dimmed">Ingen tilbakemeldinger</div>
  </UCard>
</template>
