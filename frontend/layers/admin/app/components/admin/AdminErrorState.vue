<script setup lang="ts">
/**
 * Error state for the admin panel.
 *
 * Distinct from the root `ErrorState` because that one is built on the
 * user-facing design system — `DesignButton` and the `text-title`/
 * `bg-background-indent` tokens — which reads wrong inside the dashboard
 * chrome. Keeping them apart is also what lets `design/` be user-only: root's
 * ErrorState was the single reason 30 admin files reached into it.
 *
 * Same surface as the shared one, so call sites only change the tag name.
 */
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
    class="flex size-full grow flex-col items-center justify-center gap-4 p-8 text-center"
  >
    <UIcon name="lucide:triangle-alert" class="text-error size-8" />

    <div class="max-w-sm space-y-1">
      <h3 class="text-highlighted text-lg font-semibold">
        {{ $t('error.title') }}
      </h3>
      <p class="text-muted text-sm text-balance">
        {{ isNetworkError ? $t('error.networkHint') : $t('error.hint') }}
      </p>
    </div>

    <div class="flex items-center gap-2">
      <UButton color="neutral" variant="soft" @click="handleRetry">
        {{ $t('error.retry') }}
      </UButton>
      <UButton
        color="neutral"
        variant="ghost"
        @click="
          () => {
            showDetails = !showDetails
          }
        "
      >
        {{ showDetails ? $t('error.hideDetails') : $t('error.showDetails') }}
      </UButton>
    </div>

    <p
      v-if="showDetails"
      class="text-dimmed bg-elevated max-w-sm rounded-md p-3 font-mono text-xs break-all"
    >
      {{ error.message }}
    </p>
  </div>
</template>
