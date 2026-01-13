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
    label: 'Enkel',
    description: 'Enkel utmerkelse med navn og poeng',
  },
  {
    value: 'CONTENT' as AchievementType,
    label: 'Innhold',
    description:
      'Krever fullføring av innholdselementer (artikler, episoder, etc.)',
  },
  // {
  //   value: 'STREAK' as AchievementType,
  //   label: 'Streak',
  //   description: 'Requires maintaining activity for a number of days',
  // },
  {
    value: 'QUIZ' as AchievementType,
    label: 'Quiz',
    description: 'Krever fullføring av en quiz med valgfritt poengkrav',
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
