export default defineNuxtPlugin(() => {
  const { me } = useAuth()
  const { identify } = useAnalytics()

  watch(
    () => me.value,
    (currentMe) => {
      if (currentMe) {
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
})
