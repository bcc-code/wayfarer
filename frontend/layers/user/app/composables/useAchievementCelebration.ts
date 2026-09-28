import type { ProfilePageQuery } from '~/api/generated'

type Achievement = ProfilePageQuery['myCurrentProject']['achievements'][number]

export function useAchievementCelebration(
  achievements: Ref<Achievement[] | undefined>,
) {
  const { openAchievementSheet, setCelebrating, openAchievementId } =
    useAchievementSheet()
  const { executeMutation: markCelebrated } =
    useMarkAchievementCelebratedMutation()

  const route = useRoute()

  // Track which achievements we've already queued/shown in this session
  const celebratedIds = ref(new Set<string>())
  const queue = ref<string[]>([])
  const isProcessing = ref(false)

  function processNext() {
    if (isProcessing.value || queue.value.length === 0) return

    const nextId = queue.value.shift()
    if (!nextId) return

    isProcessing.value = true
    setCelebrating(true)
    openAchievementSheet(nextId)
  }

  // When the drawer closes while celebrating, mark celebrated and process next
  watch(openAchievementId, (newId, oldId) => {
    if (oldId && !newId && isProcessing.value) {
      // Drawer was closed — mark the achievement as celebrated
      markCelebrated({ achievementId: oldId })
      isProcessing.value = false

      // Process next after a short delay
      setTimeout(() => {
        processNext()
      }, 400)
    }
  })

  // Watch achievements for uncelebrated ones
  watch(
    achievements,
    (achs) => {
      if (!achs) return

      const uncelebrated = achs.filter(
        (a) =>
          a.achievedAt && !a.celebratedAt && !celebratedIds.value.has(a.id),
      )

      if (uncelebrated.length === 0) return

      // Skip if achievement is being opened via URL param
      const urlAchievementId = route.query.achievement

      for (const ach of uncelebrated) {
        celebratedIds.value.add(ach.id)
        if (ach.id !== urlAchievementId) {
          queue.value.push(ach.id)
        }
      }

      processNext()
    },
    { immediate: true },
  )
}
