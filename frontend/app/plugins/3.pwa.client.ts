export default defineNuxtPlugin(async () => {
  // Skip service worker registration in dev mode
  if (import.meta.env.DEV) {
    console.log('[PWA] Skipping service worker registration in dev mode')
    return
  }

  if (!('serviceWorker' in navigator)) {
    console.log('[PWA] Service workers not supported')
    return
  }

  try {
    const registration = await navigator.serviceWorker.register(
      import.meta.env.DEV ? '/dev-sw.js?dev-sw' : '/service-worker.js',
      {
        scope: '/',
      },
    )
    console.log('[PWA] Service worker registered:', registration.scope)
  } catch (error) {
    console.error('[PWA] Service worker registration failed:', error)
  }
})
