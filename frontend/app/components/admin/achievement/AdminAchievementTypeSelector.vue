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
    value: 'SIMPLE' as AchievementType,
    label: 'Simple',
    description: 'Basic achievement with name and points',
  },
  {
    value: 'CONTENT' as AchievementType,
    label: 'Content',
    description: 'Requires completing content items (articles, episodes, etc.)',
  },
  // {
  //   value: 'STREAK' as AchievementType,
  //   label: 'Streak',
  //   description: 'Requires maintaining activity for a number of days',
  // },
  {
    value: 'QUIZ' as AchievementType,
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
  <URadioGroup
    :items="typeOptions"
    variant="table"
    size="lg"
    :model-value="modelValue"
    @update:model-value="selectType"
  />
</template>
