<script setup lang="ts">
type AchievementType = 'SIMPLE' | 'CONTENT' | 'STREAK' | 'QUIZ'

const props = defineProps<{
  modelValue: AchievementType
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: AchievementType]
}>()

const typeOptions = [
  {
    value: 'SIMPLE' as const,
    label: 'Simple',
    description: 'Basic achievement with name and points',
  },
  {
    value: 'CONTENT' as const,
    label: 'Content',
    description: 'Requires completing content items (articles, episodes, etc.)',
  },
  {
    value: 'STREAK' as const,
    label: 'Streak',
    description: 'Requires maintaining activity for a number of days',
  },
  {
    value: 'QUIZ' as const,
    label: 'Quiz',
    description: 'Requires completing a quiz with optional score requirement',
  },
]

function selectType(type: AchievementType) {
  if (!props.disabled) {
    emit('update:modelValue', type)
  }
}
</script>

<template>
  <div class="flex flex-col gap-2">
    <div class="grid grid-cols-2 gap-3">
      <button
        v-for="option in typeOptions"
        :key="option.value"
        type="button"
        :disabled="disabled"
        class="rounded-lg border-2 p-4 text-left transition-colors"
        :class="[
          modelValue === option.value
            ? 'border-primary bg-primary/5'
            : 'border-default hover:border-primary/50',
          disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer',
        ]"
        @click="selectType(option.value)"
      >
        <div class="font-medium">{{ option.label }}</div>
        <div class="text-muted mt-1 text-sm">{{ option.description }}</div>
      </button>
    </div>
  </div>
</template>
