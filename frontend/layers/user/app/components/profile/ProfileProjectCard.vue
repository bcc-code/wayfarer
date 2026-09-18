<script setup lang="ts">
import { gsap } from 'gsap'

type ProjectCardAchievement =
  ProfilePageQuery['myCurrentProject']['achievements'][number]

interface BannerImage {
  url: string
  width?: number | null
  height?: number | null
  blurhash?: string | null
}

const props = defineProps<{
  projectName: string
  banner?: BannerImage | null
  score?: number
  rank?: number | null
  achievements?: ProjectCardAchievement[]
}>()

// Animated values for counting effect
const animatedScore = ref(0)
const animatedRank = ref(0)

onMounted(() => {
  const prefersReducedMotion = window.matchMedia(
    '(prefers-reduced-motion: reduce)',
  ).matches

  if (prefersReducedMotion) {
    animatedScore.value = props.score ?? 0
    animatedRank.value = props.rank ?? 0
  } else {
    // Animate score
    if (props.score) {
      gsap.to(animatedScore, {
        value: props.score,
        duration: 0.8,
        ease: 'power2.out',
        onUpdate: () => {
          animatedScore.value = Math.round(animatedScore.value)
        },
      })
    }

    // Animate rank
    if (props.rank) {
      gsap.to(animatedRank, {
        value: props.rank,
        duration: 0.6,
        delay: 0.2,
        ease: 'power2.out',
        onUpdate: () => {
          animatedRank.value = Math.round(animatedRank.value)
        },
      })
    }
  }
})
</script>

<template>
  <DesignCard class="overflow-clip">
    <DesignImage v-if="banner" :image="banner" class="w-full h-50" />
    <div class="p-default gap-medium flex flex-col">
      <p v-if="!banner" class="text-label text-center">{{ projectName }}</p>
      <div class="divide-border-default grid grid-cols-2 divide-x py-2">
        <div class="flex flex-col items-center">
          <p class="title-text tabular-nums">
            {{ formatNumber(animatedScore) }}
          </p>
          <p class="text-label text-text-hint">{{ $t('points') }}</p>
        </div>
        <div class="flex flex-col items-center">
          <p class="title-text tabular-nums">{{ animatedRank || '–' }}</p>
          <p class="text-label text-text-hint">{{ $t('place') }}</p>
        </div>
      </div>
      <div class="gap-medium grid grid-cols-2">
        <ProfilePointHistory>
          <DesignButton variant="secondary">
            {{ $t('pointHistory.pointHistoryButton') }}
          </DesignButton>
        </ProfilePointHistory>
        <ProjectRules>
          <DesignButton variant="secondary">
            {{ $t('standings.rules') }}
          </DesignButton>
        </ProjectRules>
      </div>
    </div>
    <slot />
    <div
      v-if="achievements?.length"
      class="p-medium gap-medium grid grid-cols-4"
    >
      <AchievementBadge
        v-for="achievement in achievements"
        :key="achievement.id"
        :achievement
      />
    </div>
  </DesignCard>
</template>
