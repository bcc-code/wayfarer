const openAchievementId = ref<string | null>(null)
const celebrating = ref(false)

export function useAchievementSheet() {
  const route = useRoute()
  const router = useRouter()

  function openAchievementSheet(achievementId: string) {
    openAchievementId.value = achievementId ?? null
  }

  function clearOpenAchievementId() {
    openAchievementId.value = null
    celebrating.value = false
    const newQuery = { ...route.query }
    delete newQuery['achievement']
    router.replace({ path: route.path, query: newQuery })
  }

  function setCelebrating(value: boolean) {
    celebrating.value = value
  }

  return {
    openAchievementId: readonly(openAchievementId),
    celebrating: readonly(celebrating),
    openAchievementSheet,
    clearOpenAchievementId,
    setCelebrating,
  }
}
