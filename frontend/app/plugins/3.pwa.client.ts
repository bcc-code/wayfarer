export default defineNuxtPlugin(() => {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker
      .register('/sw.js', { scope: '/' })
      .then((registration) => {
        console.log('[PWA] Service worker registered:', registration.scope)
      })
      .catch((error) => {
        console.error('[PWA] Service worker registration failed:', error)
      })
  }
})
