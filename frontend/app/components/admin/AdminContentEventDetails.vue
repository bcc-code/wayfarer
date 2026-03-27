<script setup lang="ts">
gql(`
  query AdminExternalContentEvents($userId: ID!, $externalContentId: ID!) {
    adminExternalContentEvents(
      userId: $userId
      externalContentId: $externalContentId
    ) {
      id
      taskId
      planId
      source
      receivedAt
      consumedAt
      contentProgress
    }
  }
`)

const props = defineProps<{
  userId: string
  externalContentId: string
}>()

const { isAuthReady } = useAuthReady()

const { data, fetching } = useAdminExternalContentEventsQuery({
  variables: computed(() => ({
    userId: props.userId,
    externalContentId: props.externalContentId,
  })),
  pause: computed(() => !isAuthReady.value),
})

const events = computed(
  () => data.value?.adminExternalContentEvents ?? [],
)
</script>

<template>
  <div>
    <div v-if="fetching" class="py-2 text-sm text-muted">Laster...</div>
    <div v-else-if="events.length === 0" class="py-2 text-sm text-muted">
      Ingen hendelser funnet
    </div>
    <div v-else class="space-y-1">
      <div class="text-xs font-medium text-muted mb-1">
        Hendelser ({{ events.length }})
      </div>
      <div
        v-for="event in events"
        :key="event.id"
        class="flex flex-wrap items-center gap-x-4 gap-y-1 rounded bg-default px-2 py-1.5 text-xs"
      >
        <span class="font-mono text-muted">{{ event.id.slice(0, 12) }}...</span>
        <span>
          <span class="text-muted">task:</span> {{ event.taskId }}
        </span>
        <span v-if="event.planId">
          <span class="text-muted">plan:</span> {{ event.planId }}
        </span>
        <span>
          <span class="text-muted">kilde:</span> {{ event.source }}
        </span>
        <span>
          <span class="text-muted">mottatt:</span>
          {{ formatDateTime(event.receivedAt) }}
        </span>
        <span v-if="event.consumedAt">
          <span class="text-muted">konsumert:</span>
          {{ formatDateTime(event.consumedAt) }}
        </span>
        <span v-if="event.contentProgress != null">
          <span class="text-muted">fremgang:</span>
          {{ Math.round(event.contentProgress * 100) }}%
        </span>
      </div>
    </div>
  </div>
</template>
