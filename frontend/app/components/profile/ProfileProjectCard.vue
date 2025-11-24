<script setup lang="ts">
import { cva } from 'cva'

defineProps<{
  projectName: string
  points: number
  standing: number
  achievements?: Partial<Achievement>[]
  highlighted?: boolean
}>()

const classes = cva('rounded-list', {
  variants: {
    highlighted: {
      true: 'bg-background-raised',
      false: '',
    },
  },
})
</script>

<template>
  <div :class="classes({ highlighted })">
    <div class="p-medium gap-medium flex flex-col">
      <p class="text-text-hint text-label text-center">{{ projectName }}</p>
      <div class="divide-border-default grid grid-cols-2 divide-x">
        <div class="flex flex-col items-center">
          <p class="text-title">{{ points }}</p>
          <p class="text-label text-text-hint">{{ $t('points') }}</p>
        </div>
        <div class="flex flex-col items-center">
          <p class="text-title">{{ standing }}</p>
          <p class="text-label text-text-hint">{{ $t('place') }}</p>
        </div>
      </div>
    </div>
    <div class="p-medium gap-medium grid grid-cols-4">
      <AchievementBadge
        v-for="achievement in achievements"
        :key="achievement.id"
        :achievement
      />
    </div>
  </div>
</template>
