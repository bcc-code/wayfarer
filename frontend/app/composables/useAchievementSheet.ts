const openAchievementId = ref<string | null>(null)

export function useAchievementSheet() {
  const route = useRoute()
  const router = useRouter()

  function openAchievementSheet(achievementId: string) {
    openAchievementId.value = achievementId ?? null
  }

  function clearOpenAchievementId() {
    openAchievementId.value = null
    const newQuery = { ...route.query }
    delete newQuery['achievement']
    router.replace({ path: route.path, query: newQuery })
  }

  return {
    openAchievementId: readonly(openAchievementId),
    openAchievementSheet,
    clearOpenAchievementId,
  }
}
