<script setup lang="ts">
const props = defineProps<{
  error: Error
}>()

const emit = defineEmits<{
  retry: []
}>()

const showDetails = ref(false)

const isNetworkError = computed(() => {
  const message = props.error.message.toLowerCase()
  return (
    message.includes('network') ||
    message.includes('fetch') ||
    message.includes('connection') ||
    message.includes('offline')
  )
})

function handleRetry() {
  emit('retry')
  window.location.reload()
}
</script>

<template>
  <div
    class="flex size-full grow flex-col items-center justify-center gap-spacing-small rounded-card p-spacing-default text-center gap-default"
  >
    <div
      class="bg-accent-negative/15 text-accent-negative mb-spacing-small flex size-14 items-center justify-center rounded-full"
    >
      <Icon name="lucide:triangle-alert" class="size-7" />
    </div>
    <div class="space-y-small max-w-sm">
      <h3 class="text-title text-text-default">{{ $t('error.title') }}</h3>
      <p class="text-caption text-text-muted text-balance">
        {{ isNetworkError ? $t('error.networkHint') : $t('error.hint') }}
      </p>
    </div>
    <DesignButton variant="secondary" class="grow-0" @click="handleRetry">
      {{ $t('error.retry') }}
    </DesignButton>
    <button
      class="text-caption text-text-hint underline"
      @click="showDetails = !showDetails"
    >
      {{ showDetails ? $t('error.hideDetails') : $t('error.showDetails') }}
    </button>
    <p
      v-if="showDetails"
      class="text-caption text-text-hint bg-background-indent max-w-sm break-all rounded-list p-medium font-mono"
    >
      {{ error.message }}
    </p>
  </div>
</template>
