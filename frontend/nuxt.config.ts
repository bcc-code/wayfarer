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
      vapidPublicKey: '', // Set via NUXT_PUBLIC_VAPID_PUBLIC_KEY env var
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

  colorMode: {
    preference: 'dark',
    fallback: 'dark',
  },
})
