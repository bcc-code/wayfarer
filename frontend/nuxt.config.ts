export default defineNuxtConfig({
  modules: [
    '@nuxt/ui',
    '@nuxt/test-utils',
    '@nuxt/eslint',
    '@nuxt/image',
    '@vueuse/nuxt',
    '@nuxtjs/i18n',
  ],
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
      link: [
        {
          rel: 'icon',
          type: 'image/svg+xml',
          href: '/favicon.svg',
        },
      ],
    },
  },
  components: {
    dirs: [
      {
        path: '~/components',
        pathPrefix: false,
      },
    ],
  },
  imports: {
    dirs: ['api'],
  },
  runtimeConfig: {
    public: {
      apiUrl: 'http://localhost:8080/graphql',
      callbackUrl: 'http://localhost:8080/callback',
      loginUrl: 'https://app.bcc.media/r/sigve-test',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: [
      {
        name: 'English',
        code: 'en',
        file: 'en.json',
      },
      {
        name: 'Norsk',
        code: 'nb',
        file: 'nb.json',
      },
    ],
    strategy: 'no_prefix',
  },
})
