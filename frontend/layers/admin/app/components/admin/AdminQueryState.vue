<script setup lang="ts">
/**
 * Loading / error / content for a page whose body depends on one query.
 *
 * Replaces the `v-if="fetching" / v-else-if="error" / v-else` chain that 16
 * pages had written by hand. Two things it standardises beyond removing the
 * repetition:
 *
 * **It does not blank on a refetch.** `v-if="fetching"` unmounts the whole page
 * body every time the query re-runs — after a mutation, on a cache refresh, on
 * a param change — so the page flashes back to a spinner even though it already
 * has content to show. Three pages had noticed and hand-rolled a
 * `hasLoadedOnce` ref against `data`; the other thirteen had not. Here the
 * spinner is shown only before the query has ever settled, so a refetch leaves
 * the content in place.
 *
 * **Error outranks a stale fetch.** Once the query has settled, an error is
 * shown rather than a spinner, so a failed refetch cannot leave the page
 * spinning forever.
 */
const props = defineProps<{
  fetching?: boolean
  // Matches `AdminErrorState`'s own prop. urql's `CombinedError` extends
  // `Error`, so a query's error ref assigns straight through.
  error?: Error | null
  /** Extra classes for the error state, e.g. reserving height in a panel. */
  errorClass?: string
}>()

/**
 * Whether the query has ever finished. Keyed on the `true → false` transition
 * rather than on `fetching` being falsy: a paused query (every admin query is
 * paused until auth is ready) starts out not fetching, and treating that as
 * "settled" would suppress the first spinner entirely.
 */
const hasSettled = ref(false)
watch(
  () => props.fetching,
  (isFetching, wasFetching) => {
    if (wasFetching && !isFetching) hasSettled.value = true
  },
)

const showLoading = computed(() => !!props.fetching && !hasSettled.value)
</script>

<template>
  <AdminLoadingState v-if="showLoading" />
  <AdminErrorState v-else-if="error" :error :class="errorClass" />
  <slot v-else />
</template>
