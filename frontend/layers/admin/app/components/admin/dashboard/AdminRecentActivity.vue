<script setup lang="ts">
defineProps<{
  feedbackEntries: Array<{
    id: string
    message: string
    createdAt: string
    tags?: string[]
    user: { id: string; name: string }
  }>
}>()
</script>

<template>
  <UCard>
    <div class="space-y-3">
      <div
        v-for="entry in feedbackEntries"
        :key="entry.id"
        class="border-default rounded-lg border p-3"
      >
        <p class="line-clamp-2 text-sm">{{ entry.message }}</p>
        <div class="mt-1 flex flex-wrap items-center gap-2">
          <p class="text-muted text-xs">
            {{ entry.user.name }} &middot;
            {{ formatRelativeTime(entry.createdAt) }}
          </p>
          <!-- Tags say what an entry is about without reading it. -->
          <UBadge
            v-for="tag in entry.tags"
            :key="tag"
            variant="subtle"
            color="neutral"
            size="sm"
          >
            {{ tag }}
          </UBadge>
        </div>
      </div>
      <UEmpty
        v-if="!feedbackEntries.length"
        icon="lucide:check"
        title="Alt behandlet"
        description="Ingen ubehandlede tilbakemeldinger."
      />
    </div>
  </UCard>
</template>
