export default defineNuxtPlugin(() => {
  const { me } = useAuth()
  const { identify, page } = useAnalytics()
  const router = useRouter()

  let identifiedUserId: string

  watch(
    () => me.value,
    (currentMe) => {
      if (currentMe && currentMe.id !== identifiedUserId) {
        identifiedUserId = currentMe.id
        hashUserId(currentMe.id).then((hashedId) => {
          identify(hashedId, {
            age_group: getAgeGroup(currentMe.age),
            gender: currentMe.gender,
            church_id: currentMe.church.id,
            church_name: currentMe.church.name,
            church_country: currentMe.church.country,
          })
        })
      }
    },
    { immediate: true },
  )

  // Track page views on navigation
  router.afterEach(() => {
    page()
  })
})
