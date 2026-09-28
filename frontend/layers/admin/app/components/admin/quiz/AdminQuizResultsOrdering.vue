<script setup lang="ts">
import {
  formatPercent,
  percentageOfResponses,
} from '../../../utils/quizResultsExport'

interface Item {
  item: { id: string; itemText: string }
  correctPosition: number
  correctlyPlacedCount: number
  percentage: number
}

defineProps<{
  items: Item[]
  responseCount: number
  fullyCorrectCount: number
}>()
</script>

<template>
  <div class="space-y-3">
    <p v-if="responseCount > 0" class="text-muted text-sm">
      {{ fullyCorrectCount }} av {{ responseCount }} hadde hele rekkefølgen
      riktig
      <span class="text-dimmed">·</span>
      {{
        formatPercent(percentageOfResponses(fullyCorrectCount, responseCount))
      }}
    </p>
    <!-- Per-position accuracy shows which step people tripped on. -->
    <p class="text-dimmed text-xs">Riktig plassert</p>
    <AdminQuizResultBar
      v-for="entry in items"
      :key="entry.item.id"
      :label="`${entry.correctPosition}. ${entry.item.itemText}`"
      :count="entry.correctlyPlacedCount"
      :percentage="entry.percentage"
    />
  </div>
</template>
