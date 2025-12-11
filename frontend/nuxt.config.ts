export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
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

  app: {
    rootAttrs: {
      'data-vaul-drawer-wrapper': '',
    },
    head: {
      title: 'Interact',
      viewport:
        'width=device-width,initial-scale=1,viewport-fit=cover,user-scalable=no',
      charset: 'utf-8',
      meta: [
        {
          name: 'apple-mobile-web-app-status-bar-style',
          content: 'default',
        },
      ],
      link: [
        {
          rel: 'icon',
          href: '/favicon.ico',
          sizes: '32x32',
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
      tokenUrl: 'http://localhost:8080/token',
      loginUrl: 'https://app.bcc.media/r/sigve-test',
      rudderstackWriteKey: '',
      rudderstackDataPlaneUrl: '',
      vapidPublicKey: '',
    },
  },
  experimental: {
    typedPages: true,
  },
  sourcemap: {
    client: 'hidden',
  },
  vite: {
    optimizeDeps: {
      include: [
        '@tiptap/core',
        '@tiptap/vue-3',
        '@tiptap/markdown',
        '@tiptap/starter-kit',
        'workbox-core',
        'workbox-precaching',
        'workbox-routing',
        'workbox-strategies',
        'workbox-cacheable-response',
        'workbox-expiration',
        '@urql/exchange-auth',
        '@urql/vue',
        'graphql-tag',
        'cva',
        'zod',
        '@internationalized/date',
        '@rudderstack/analytics-js',
      ],
    },
  },
  routeRules: {
    // CDN cache rules
    '/manifest.webmanifest': {
      headers: {
        'Content-Type': 'application/manifest+json',
        'Cache-Control': 'public, max-age=0, must-revalidate',
      },
    },
  },

  i18n: {
    defaultLocale: 'nb',
    locales: [
      {
        name: 'Norsk',
        code: 'nb',
        file: 'nb.json',
      },
      {
        name: 'English',
        code: 'en',
        file: 'en_us.json',
      },
      {
        name: 'Deutsch',
        code: 'de',
        file: 'de.json',
      },
      {
        name: 'Français',
        code: 'fr',
        file: 'fr.json',
      },
      {
        name: 'Español',
        code: 'es',
        file: 'es.json',
      },
      {
        name: 'Italiano',
        code: 'it',
        file: 'it.json',
      },
      {
        name: 'Nederlands',
        code: 'nl',
        file: 'nl.json',
      },
      {
        name: 'Eesti',
        code: 'et',
        file: 'et.json',
      },
      {
        name: 'Suomi',
        code: 'fi',
        file: 'fi.json',
      },
      {
        name: 'Magyar',
        code: 'hu',
        file: 'hu.json',
      },
      {
        name: 'Khasi',
        code: 'kha',
        file: 'kha.json',
      },
      {
        name: 'മലയാളം',
        code: 'ml',
        file: 'ml.json',
      },
      {
        name: 'Papiamentu',
        code: 'pap',
        file: 'pap.json',
      },
      {
        name: 'Polski',
        code: 'pl',
        file: 'pl.json',
      },
      {
        name: 'Português',
        code: 'pt',
        file: 'pt.json',
      },
      {
        name: 'Română',
        code: 'ro',
        file: 'ro.json',
      },
      {
        name: 'Русский',
        code: 'ru',
        file: 'ru.json',
      },
      {
        name: 'Slovenščina',
        code: 'sl',
        file: 'sl.json',
      },
      {
        name: 'தமிழ்',
        code: 'ta',
        file: 'ta.json',
      },
      {
        name: 'Türkçe',
        code: 'tr',
        file: 'tr.json',
      },
      {
        name: '中文(简体)',
        code: 'zh-CN',
        file: 'zn_cn.json',
      },
      {
        name: '中文(香港)',
        code: 'zh-HK',
        file: 'zh_hk.json',
      },
    ],
    strategy: 'no_prefix',
  },

  pwa: {
    scope: '/',
    srcDir: '../service-worker',
    filename: 'service-worker.ts',
    strategies: 'injectManifest',
    injectRegister: false,
    injectManifest: {
      globPatterns: [
        '**/*.{js,json,css,html,txt,svg,png,ico,webp,woff,woff2,ttf,eot,otf,wasm}',
      ],
      globIgnores: ['manifest**.webmanifest'],
    },
    devOptions: {
      enabled: true,
      type: 'module',
    },
    manifest: {
      theme_color: '#E8DFA7',
      name: 'Interact',
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
  },

  colorMode: {
    preference: 'dark',
    fallback: 'dark',
  },
})
