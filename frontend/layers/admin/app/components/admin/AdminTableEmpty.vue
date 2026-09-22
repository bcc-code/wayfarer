<script setup lang="ts">
// For UTable's #empty slot — never as a sibling, where it renders *below* the
// table's own empty row and you get two empty states.
//
// Separates "nothing exists" from "nothing matched": the second needs a way out.
const props = withDefaults(
  defineProps<{
    filtered?: boolean
    title: string
    description?: string
    filteredTitle?: string
    clearLabel?: string
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
    <!-- No icon for a filtered miss: it is transient. -->
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
