import { execSync } from 'node:child_process'

function getAppVersion(): string {
  if (process.env.APP_VERSION) {
    return process.env.APP_VERSION
  }

  try {
    return execSync('git rev-parse --short HEAD').toString().trim()
  } catch {
    throw new Error(
      'APP_VERSION environment variable not set and git is not available. Set APP_VERSION in your build environment.',
    )
  }
}

const gitCommitHash = getAppVersion()

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',

  modules: [
    '@nuxt/ui',
    '@nuxt/test-utils',
    '@nuxt/eslint',
    '@vueuse/nuxt',
    '@nuxtjs/i18n',
    '@vite-pwa/nuxt',
    '@posthog/nuxt',
    '@sentry/nuxt/module',
  ],

  devtools: { enabled: false },
  ssr: false,
  css: ['~/assets/styles/main.css'],

  typescript: {
    tsConfig: {
      compilerOptions: {
        allowSyntheticDefaultImports: true,
      },
    },
  },

  app: {
    rootAttrs: {
      'data-vaul-drawer-wrapper': '',
      class: 'bg-background-default',
    },
    head: {
      title: 'Interact',
      viewport:
        'width=device-width, initial-scale=1, viewport-fit=cover, user-scalable=no',
      charset: 'utf-8',
      meta: [
        {
          name: 'apple-mobile-web-app-capable',
          content: 'yes',
        },
        {
          name: 'apple-mobile-web-app-status-bar-style',
          content: 'black-translucent',
        },
      ],
      link: [
        {
          rel: 'icon',
          href: '/favicon.ico',
          sizes: '48x48',
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
      auth0Domain: 'login.bcc.no',
      auth0ClientId: '',
      auth0Audience: '',
      rudderstackWriteKey: '',
      rudderstackDataPlaneUrl: '',
      vapidPublicKey: '',
      appVersion: gitCommitHash,
      isStaging: false,
      firebaseDatabase: '',
      firebaseApiKey: '',
      firebaseAuthDomain: '',
      firebaseProjectId: '',
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
        '@auth0/auth0-vue',
        '@internationalized/date',
        '@neoconfetti/vue',
        '@rudderstack/analytics-js',
        '@sentry/nuxt',
        '@tiptap/core',
        '@tiptap/markdown',
        '@tiptap/starter-kit',
        '@tiptap/vue-3',
        '@unovis/vue',
        '@urql/exchange-auth',
        '@urql/vue',
        '@zag-js/slider',
        '@zag-js/vue',
        'blurhash',
        'cva',
        'dayjs', // CJS
        'firebase/app',
        'firebase/auth',
        'firebase/firestore',
        'graphql-tag',
        'gsap',
        'uqr',
        'vue-draggable-plus',
        'workbox-cacheable-response',
        'workbox-core',
        'workbox-expiration',
        'workbox-precaching',
        'workbox-routing',
        'workbox-strategies',
        'zod',
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
      // The commented ones are not actively translated.
      // Verified by Milenko
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
      // {
      //   name: 'മലയാളം',
      //   code: 'ml',
      //   file: 'ml.json',
      // },
      // {
      //   name: 'Papiamentu',
      //   code: 'pap',
      //   file: 'pap.json',
      // },
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
      // {
      //   name: 'Slovenščina',
      //   code: 'sl',
      //   file: 'sl.json',
      // },
      // {
      //   name: 'தமிழ்',
      //   code: 'ta',
      //   file: 'ta.json',
      // },
      {
        name: 'Türkçe',
        code: 'tr',
        file: 'tr.json',
      },
      {
        name: '中文(简体)',
        code: 'zh-CN',
        file: 'zh_cn.json',
      },
      // {
      //   name: '中文(香港)',
      //   code: 'zh-HK',
      //   file: 'zh_hk.json',
      // },
    ],
    strategy: 'no_prefix',
  },

  pwa: {
    scope: '/',
    srcDir: '../service-worker',
    filename: 'service-worker.ts',
    strategies: 'injectManifest',
    registerType: 'prompt',
    injectRegister: 'auto',
    injectManifest: {
      globPatterns: [
        '**/*.{js,json,css,html,txt,svg,png,ico,webp,woff,woff2,ttf,eot,otf,wasm}',
      ],
      globIgnores: ['manifest**.webmanifest', '**/admin/**'],
    },
    devOptions: {
      enabled: true,
      type: 'module',
    },
    manifest: {
      theme_color: '#E8DFA7',
      name: 'Interact',
      short_name: 'Interact',
      display: 'standalone',
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
      opt_out_capturing_by_default: import.meta.dev,
    },
  },

  colorMode: {
    preference: 'dark',
    fallback: 'dark',
  },

  sentry: {
    org: 'bcc-media-sti',
    project: 'interact',
  },
})
