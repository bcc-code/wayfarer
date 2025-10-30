export default defineEventHandler(async (event) => {
  const challengeId = getRouterParam(event, 'challengeId')
  const challenges = await $fetch('/api/challenges')

  return challenges.find((challenge) => challenge.id === challengeId)
})
