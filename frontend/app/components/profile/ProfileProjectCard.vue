<script setup lang="ts">
defineProps<{
  projectName: string
  score?: number
  rank?: number | null
  achievements?: Partial<Achievement>[]
}>()
</script>

<template>
  <DesignCard>
    <div class="p-default gap-medium flex flex-col">
      <p class="text-label text-center">{{ projectName }}</p>
      <div class="divide-border-default grid grid-cols-2 divide-x py-2">
        <div class="flex flex-col items-center">
          <p class="title-text">{{ score ?? 0 }}</p>
          <p class="text-label text-text-hint">{{ $t('points') }}</p>
        </div>
        <div class="flex flex-col items-center">
          <p class="title-text">{{ rank || '-' }}</p>
          <p class="text-label text-text-hint">{{ $t('place') }}</p>
        </div>
      </div>
      <div class="gap-medium grid grid-cols-2">
        <ProfilePointHistory>
          <DesignButton variant="secondary">
            {{ $t('pointHistory.pointHistoryButton') }}
          </DesignButton>
        </ProfilePointHistory>
        <ProfileGetPoints>
          <DesignButton variant="secondary">
            {{ $t('standings.getPointsButton') }}
          </DesignButton>
        </ProfileGetPoints>
      </div>
    </div>
    <slot />
    <div class="p-medium gap-medium grid grid-cols-4">
      <AchievementBadge
        v-for="achievement in achievements"
        :key="achievement.id"
        :achievement
      />
    </div>
  </DesignCard>
</template>
