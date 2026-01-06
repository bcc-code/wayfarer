export default defineNuxtRouteMiddleware(async (to) => {
  if (!to.path.startsWith('/church-admin')) {
    return
  }
})
