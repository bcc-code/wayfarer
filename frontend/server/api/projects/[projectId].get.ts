export default defineEventHandler(async (event) => {
  const projectId = getRouterParam(event, 'projectId')
  const projects = await $fetch('/api/projects')

  return projects.find((project) => project.id === projectId)
})
