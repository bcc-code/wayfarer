// Service worker registration is handled by @vite-pwa/nuxt module
// This plugin only adds custom hooks for debugging/logging
export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.hook('app:mounted', () => {
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.ready.then((registration) => {
        console.log('[PWA] Service worker ready:', registration.scope)
      })
    }
  })
})
