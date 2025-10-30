// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: ['@nuxt/ui', '@nuxt/test-utils', '@nuxt/eslint'],
  devtools: { enabled: true },
  css: ['~/assets/styles/main.css'],
  compatibilityDate: '2025-07-15',
})
