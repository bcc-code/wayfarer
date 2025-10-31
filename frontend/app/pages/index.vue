<script setup lang="ts">
import { useGetStandingsQuery } from '~/api/generated'

const { data, error, fetching } = useGetStandingsQuery()
</script>

<template>
  <div class="flex flex-col h-full">
    <h1 class="text-xl font-bold">Standings</h1>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <pre>{{ data }}</pre>
    </template>
  </div>
</template>
