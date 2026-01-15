<script setup lang="ts">
import { useLocalStorage, useNow } from '@vueuse/core'

const props = defineProps<{
  projectId: string
  infoMessage: {
    markdown: string
    html: string
  } | null
  infoMessageStart?: string | null
  infoMessageEnd?: string | null
}>()

// Generate a simple hash of the content for dismissal tracking
function hashContent(content: string): string {
  let hash = 0
  for (let i = 0; i < content.length; i++) {
    const char = content.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash // Convert to 32bit integer
  }
  return Math.abs(hash).toString(36)
}

// Track dismissed messages by content hash per project
const dismissedHashes = useLocalStorage<Record<string, string>>(
  'projectInfoMessageDismissed',
  {},
)

// Update every minute to check visibility
const now = useNow({ interval: 1000 })

const contentHash = computed(() => {
  if (!props.infoMessage?.markdown) return null
  return hashContent(props.infoMessage.markdown)
})

const isDismissed = computed(() => {
  if (!contentHash.value) return true
  return dismissedHashes.value[props.projectId] === contentHash.value
})

const isWithinVisibilityWindow = computed(() => {
  const currentTime = now.value.getTime()

  // Check start time - if set and current time is before start, don't show
  if (props.infoMessageStart) {
    const startTime = new Date(props.infoMessageStart).getTime()
    if (currentTime < startTime) return false
  }

  // Check end time - if set and current time is after end, don't show
  if (props.infoMessageEnd) {
    const endTime = new Date(props.infoMessageEnd).getTime()
    if (currentTime > endTime) return false
  }

  return true
})

const shouldShow = computed(() => {
  return (
    props.infoMessage?.html &&
    !isDismissed.value &&
    isWithinVisibilityWindow.value
  )
})

function dismiss() {
  if (contentHash.value) {
    dismissedHashes.value[props.projectId] = contentHash.value
  }
}
</script>

<template>
  <div
    v-if="shouldShow"
    class="border border-border-default bg-background-default shadow-small rounded-modal"
  >
    <div class="p-default flex gap-medium">
      <IconInfo class="size-6 shrink-0 my-1.5" />
      <div
        class="my-2 text-label text-text-default grow"
        v-html="infoMessage?.html"
      />
      <DesignIconButton size="small" icon="IconClose" @click="dismiss" />
    </div>
  </div>
</template>
