export default defineNuxtConfig({
  modules: [
    '@nuxt/ui',
    '@nuxt/test-utils',
    '@nuxt/eslint',
    '@nuxt/image',
    '@vueuse/nuxt',
    '@nuxtjs/i18n',
    '@vite-pwa/nuxt',
    '@posthog/nuxt',
  ],
  devtools: { enabled: false },
  ssr: false,
  css: ['~/assets/styles/main.css'],
  compatibilityDate: '2025-07-15',
  experimental: {
    typedPages: true,
  },
  app: {
    rootAttrs: {
      'data-vaul-drawer-wrapper': '',
    },
    head: {
      title: 'Interact',
      meta: [
        {
          name: 'viewport',
          content: 'width=device-width, initial-scale=1, user-scalable=no',
        },
      ],
      link: [
        {
          rel: 'icon',
          href: '/favicon.ico',
          sizes: '32x32',
        },
        {
          rel: 'icon',
          type: 'image/svg+xml',
          href: '/favicon.svg',
          sizes: 'any',
        },
        {
          rel: 'apple-touch-icon',
          href: '/apple-touch-icon-180x180.png',
          sizes: '180x180',
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
      {
        path: '~/components/global',
        global: true,
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
      posthogKey: '',
      posthogHost: 'https://eu.i.posthog.com',
    },
  },
  sourcemap: {
    client: 'hidden',
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

  pwa: {
    registerType: 'autoUpdate',
    devOptions: {
      enabled: true,
    },
    manifest: {
      theme_color: '#ffaedf',
      name: 'Interact by BCC Media',
      short_name: 'Interact',
      orientation: 'portrait',
      icons: [
        {
          src: 'pwa-64x64.png',
          sizes: '64x64',
          type: 'image/png',
        },
        {
          src: 'pwa-192x192.png',
          sizes: '192x192',
          type: 'image/png',
        },
        {
          src: 'pwa-512x512.png',
          sizes: '512x512',
          type: 'image/png',
          purpose: 'any',
        },
        {
          src: 'maskable-icon-512x512.png',
          sizes: '512x512',
          type: 'image/png',
          purpose: 'maskable',
        },
      ],
    },
  },

  posthogConfig: {
    publicKey: 'phc_l88yVnYQJShvE2rFd1f7Cask76jMuK7qLVVyPlA9FLl',
    host: 'https://eu.i.posthog.com',
    clientConfig: {
      capture_exceptions: true,
    },
    sourcemaps: {
      enabled: true,
      personalApiKey: import.meta.env.NUXT_POSTHOG_API_KEY,
      envId: import.meta.env.NUXT_POSTHOG_PROJECT_ID,
    },
  },
})
