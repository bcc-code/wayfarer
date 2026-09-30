<script setup lang="ts">
interface Group {
  text: string
  count: number
  percentage: number
}

const props = withDefaults(
  defineProps<{
    groups: Group[]
    responses: string[]
    distinctCount: number
    /** Groups shown before the list is collapsed. */
    visibleGroups?: number
  }>(),
  { visibleGroups: 8 },
)

// Free text has a long tail of one-off answers; the rest stay one click away.
const expanded = ref(false)
const visible = computed(() =>
  expanded.value ? props.groups : props.groups.slice(0, props.visibleGroups),
)
const hiddenCount = computed(() =>
  Math.max(0, props.groups.length - props.visibleGroups),
)

const showingRaw = ref(false)
</script>

<template>
  <div class="space-y-3">
    <p class="text-muted text-sm">
      {{ responses.length }} svar
      <span class="text-dimmed">·</span>
      {{ distinctCount }} unike
    </p>

    <AdminQuizResultBar
      v-for="group in visible"
      :key="group.text"
      :label="group.text"
      :count="group.count"
      :percentage="group.percentage"
    />

    <div class="flex flex-wrap gap-2 pt-1">
      <UButton
        v-if="hiddenCount > 0"
        variant="ghost"
        size="xs"
        :icon="expanded ? 'lucide:chevron-up' : 'lucide:chevron-down'"
        @click="expanded = !expanded"
      >
        {{ expanded ? 'Vis færre' : `Vis ${hiddenCount} til` }}
      </UButton>
      <UButton
        v-if="responses.length"
        variant="ghost"
        size="xs"
        icon="lucide:list"
        @click="showingRaw = !showingRaw"
      >
        {{
          showingRaw ? 'Skjul alle svar' : `Vis alle ${responses.length} svar`
        }}
      </UButton>
    </div>

    <!-- Every answer as submitted, including the one-offs grouping hides. -->
    <ul
      v-if="showingRaw"
      class="bg-elevated max-h-80 space-y-1 overflow-y-auto rounded-md p-3 text-sm"
    >
      <li
        v-for="(response, index) in responses"
        :key="index"
        class="text-muted"
      >
        {{ response }}
      </li>
    </ul>
  </div>
</template>
