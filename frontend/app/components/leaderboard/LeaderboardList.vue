<script setup lang="ts" generic="Entry extends Partial<LeaderboardEntry>">
const props = defineProps<{
  leaderboard: Entry[]
  hideImages?: boolean
  badge?: (item: Entry, index: number) => string | undefined
  hideMedals?: boolean
}>()

const containerRef = ref<HTMLElement | null>(null)
const { animate } = useStaggeredEntrance()
const hasAnimated = ref(false)

function runAnimation() {
  if (hasAnimated.value) return
  if (props.leaderboard.length > 0 && containerRef.value) {
    hasAnimated.value = true
    nextTick(() => {
      const items = containerRef.value?.querySelectorAll('.leaderboard-item')
      if (items) {
        animate(items)
      }
    })
  }
}

// Animate on data change (for initial load when data arrives after mount)
watch(() => props.leaderboard, runAnimation)

// Animate on mount (for v-if toggling when data already exists)
onMounted(runAnimation)
</script>

<template>
  <div ref="containerRef" class="space-y-list-section-gap">
    <DesignPanel
      v-for="(item, index) in leaderboard"
      :key="index"
      class="leaderboard-item"
    >
      <LeaderboardItem
        :item
        :hide-image="hideImages"
        :badge="badge ? badge(item, index) : undefined"
        :hide-medal="hideMedals"
        :is-me="item.tags?.includes(LeaderboardEntryTag.Me)"
      />
    </DesignPanel>
  </div>
</template>
