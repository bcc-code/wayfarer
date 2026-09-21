<script setup lang="ts">
/**
 * Empty state for a `UTable`, for its `#empty` slot.
 *
 * Belongs in the slot, never beside the table. Rendered as a sibling it appears
 * *below* the table's own empty row, so an empty list shows two empty states —
 * a mistake made independently on six pages in this codebase before this
 * component existed.
 *
 * It also separates the two cases a list actually has, which the sibling
 * version could not: nothing exists yet, versus nothing matched. The second
 * needs a way back out, and telling a user "no challenges yet" when they have
 * simply filtered them all away is actively misleading.
 */
const props = withDefaults(
  defineProps<{
    /** True when a search or filter is narrowing the list. */
    filtered?: boolean
    /** Shown when the list is genuinely empty. */
    title: string
    description?: string
    /** Shown when a filter hid everything. Falls back to a generic line. */
    filteredTitle?: string
    /** Label for the escape hatch out of a filtered miss. */
    clearLabel?: string
    /** Optional icon for the genuinely-empty case. */
    icon?: string
  }>(),
  {
    filtered: false,
    filteredTitle: undefined,
    description: undefined,
    icon: undefined,
    clearLabel: 'Nullstill søk og filtre',
  },
)

const emit = defineEmits<{ clear: [] }>()

const heading = computed(() =>
  props.filtered
    ? (props.filteredTitle ?? 'Ingenting passer søket')
    : props.title,
)
</script>

<template>
  <div class="py-6 text-center">
    <!-- Icon only for the "nothing here yet" case; a filtered miss is a
         transient state, and decorating it overstates it. -->
    <UIcon
      v-if="icon && !filtered"
      :name="icon"
      class="text-dimmed mb-2 size-6"
    />
    <p class="text-muted text-sm">{{ heading }}</p>
    <p v-if="!filtered && description" class="text-dimmed mt-1 text-xs">
      {{ description }}
    </p>
    <UButton
      v-if="filtered"
      variant="link"
      size="sm"
      class="mt-1"
      @click="emit('clear')"
    >
      {{ clearLabel }}
    </UButton>
  </div>
</template>
