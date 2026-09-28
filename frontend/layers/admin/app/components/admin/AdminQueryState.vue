<script setup lang="ts">
const props = defineProps<{
  fetching?: boolean
  error?: Error | null
  errorClass?: string
}>()

// Spinner only before the query has ever settled, so a refetch leaves content
// in place instead of blanking the page.
//
// Keyed on the true → false transition: admin queries are paused until auth is
// ready, so `fetching` starts false and treating that as settled would suppress
// the first spinner.
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
