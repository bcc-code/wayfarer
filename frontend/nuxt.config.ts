export default defineNuxtConfig({
  modules: ['@nuxt/ui', '@nuxt/test-utils', '@nuxt/eslint', '@nuxt/image'],
  devtools: { enabled: false },
  ssr: false,
  css: ['~/assets/styles/main.css'],
  compatibilityDate: '2025-07-15',
  experimental: {
    typedPages: true,
  },
  app: {
    head: {
      title: 'Wayfarer',
      meta: [
        {
          name: 'viewport',
          content: 'width=device-width, initial-scale=1, user-scalable=no',
        },
      ],
    },
  },
})
