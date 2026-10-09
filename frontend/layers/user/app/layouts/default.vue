<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import { gsap } from 'gsap'
import {
  pageTransitionDirection,
  pageTransitionMeta,
} from '../utils/pageTransition'
import { syncAppBadge } from '../utils/appBadge'
import '~/assets/styles/user.css'

const { t } = useI18n()

// Theme caching to prevent flash of default theme
const cachedTheme = useLocalStorage<BrandingColorsFieldsFragment | null>(
  'projectTheme',
  null,
  {
    serializer: {
      read: (v) => (v ? JSON.parse(v) : null),
      write: (v) => JSON.stringify(v),
    },
  },
)

function isValidTheme(colors: unknown): colors is BrandingColorsFieldsFragment {
  return (
    typeof colors === 'object' &&
    colors !== null &&
    'light' in colors &&
    'dark' in colors &&
    typeof (colors as BrandingColorsFieldsFragment).light?.accent === 'string'
  )
}

function applyTheme(colors: BrandingColorsFieldsFragment) {
  document.getElementById('theme')?.remove()

  const style = `
  <style id="theme">
  :root {
    --color-accent: ${colors.light.accent};
    --color-accent-contrast: ${colors.light.accentContrast};
    --color-on-accent: ${colors.light.onAccent};
    --color-background-default: ${colors.light.backgroundDefault};
    --color-background-raised: ${colors.light.backgroundRaised};
    --color-background-indent: ${colors.light.backgroundIndent};
    --color-text-default: ${colors.light.textDefault};
    --color-text-muted: ${colors.light.textMuted};
    --color-text-hint: ${colors.light.textHint};
    --color-shadow-default: ${colors.light.shadowDefault};
    --color-shadow-blank: ${colors.light.shadowBlank};
    --color-border-default: ${colors.light.borderDefault};
  }

  .dark {
    --color-accent: ${colors.dark.accent};
    --color-accent-contrast: ${colors.dark.accentContrast};
    --color-on-accent: ${colors.dark.onAccent};
    --color-background-default: ${colors.dark.backgroundDefault};
    --color-background-raised: ${colors.dark.backgroundRaised};
    --color-background-indent: ${colors.dark.backgroundIndent};
    --color-text-default: ${colors.dark.textDefault};
    --color-text-muted: ${colors.dark.textMuted};
    --color-text-hint: ${colors.dark.textHint};
    --color-shadow-default: ${colors.dark.shadowDefault};
    --color-shadow-blank: ${colors.dark.shadowBlank};
    --color-border-default: ${colors.dark.borderDefault};
  }
  </style>
  `

  document.head.insertAdjacentHTML('beforeend', style)
}

// Initialize Firestore sync for realtime updates
const {
  initialize: initFirestoreSync,
  subscribeProject,
  isAuthenticated: firestoreAuthenticated,
} = useFirestoreSync()
onMounted(() => {
  initFirestoreSync()

  // Apply cached theme immediately to prevent flash
  if (isValidTheme(cachedTheme.value)) {
    applyTheme(cachedTheme.value)
  }
})

const availableChallengesBadge = computed(
  () => data.value?.myCurrentProject.activeChallengesCount,
)

// Hide the standings tab when the project has no leaderboards configured —
// the page is config-driven, so with no configs there is nothing for it to
// render. This used to key off whether the persons leaderboard had any rows,
// which asked the wrong question: a configured board that is still empty is a
// page worth opening, and rows without a config are not reachable at all.
const hasLeaderboard = computed(
  () => (data.value?.myCurrentProject.leaderboards.length ?? 0) > 0,
)

const links = computed<NavigationMenuItem[]>(() =>
  (
    [
      {
        label: t('navigation.profile'),
        icon: 'IconProfile',
        to: { name: 'index' },
      },
      hasLeaderboard.value
        ? {
            label: t('navigation.standings'),
            icon: 'IconStandings',
            to: { name: 'standings' },
          }
        : null,
      {
        label: t('navigation.challenges'),
        icon: 'IconChallenges',
        to: { name: 'challenges' },
        badge: availableChallengesBadge.value,
      },
    ] as (NavigationMenuItem | null)[]
  ).filter((link): link is NavigationMenuItem => link !== null),
)

// Current project config
gql(`
  query CurrentProject {
    myCurrentProject {
      id
      branding {
        ...BrandingFields
      }
      activeChallengesCount
      leaderboards {
        id
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, executeQuery: refresh } = useCurrentProjectQuery({
  pause: computed(() => !isAuthReady.value),
})

// Mirrors the nav badge onto the home-screen icon.
watch(availableChallengesBadge, (count) => syncAppBadge(navigator, count), {
  immediate: true,
})

// Listen for Firestore realtime updates
useFirestoreRefresh(['CurrentProjectDocument'], () => {
  refresh({ requestPolicy: 'network-only' })
})

watch(data, (newData) => {
  if (!newData) return

  const colors = newData.myCurrentProject.branding.colors
  applyTheme(colors)
  cachedTheme.value = JSON.parse(JSON.stringify(colors))
})

// Subscribe to project-level quiz session notifications
const projectSubscriptionCleanup = ref<(() => void) | null>(null)
watch(
  [() => data.value?.myCurrentProject.id, firestoreAuthenticated],
  ([projectId, isAuth]) => {
    // Cleanup previous subscription if any
    if (projectSubscriptionCleanup.value) {
      projectSubscriptionCleanup.value()
      projectSubscriptionCleanup.value = null
    }

    if (projectId && isAuth) {
      const quizCleanup = subscribeProject(projectId, 'quiz_sessions')
      const challengesCleanup = subscribeProject(projectId, 'challenges')
      projectSubscriptionCleanup.value = () => {
        quizCleanup()
        challengesCleanup()
      }
    }
  },
  { immediate: true },
)

onUnmounted(() => {
  if (projectSubscriptionCleanup.value) {
    projectSubscriptionCleanup.value()
  }
})

// The browser UI colour follows the project's branding and the active colour
// mode. Until branding resolves the manifest's own theme_color stands in.
const colorMode = useColorMode()
const activeColors = computed<BrandingColorsFieldsFragment | null>(() => {
  const live = data.value?.myCurrentProject.branding.colors
  if (live) return live
  return isValidTheme(cachedTheme.value) ? cachedTheme.value : null
})

useSeoMeta({
  themeColor: () => {
    const colors = activeColors.value
    if (!colors) return null
    return colorMode.value === 'dark'
      ? colors.dark.backgroundDefault
      : colors.light.backgroundDefault
  },
})

const { pressListeners } = useButtonPress()

const route = useRoute()
const navRef = ref<HTMLElement | null>(null)
const indicatorRef = ref<HTMLElement | null>(null)

/** Until measured, the indicator is zero-sized — keep it hidden rather than
 * letting it animate open out of nothing. */
const isIndicatorPositioned = ref(false)

// A nav that unmounted on a route that hides it comes back unmeasured.
watch(navRef, () => {
  isIndicatorPositioned.value = false
})

function updateIndicator() {
  if (!navRef.value || !indicatorRef.value) return

  const activeLink = navRef.value.querySelector<HTMLElement>(
    '[aria-current="page"]',
  )
  if (!activeLink) return

  const targetLeft = activeLink.offsetLeft
  const targetTop = activeLink.offsetTop
  const targetWidth = activeLink.offsetWidth
  const targetHeight = activeLink.offsetHeight

  // The link measures zero before the nav is laid out.
  if (!targetWidth || !targetHeight) return

  const prefersReducedMotion = window.matchMedia(
    '(prefers-reduced-motion: reduce)',
  ).matches

  // Asked of the element, not a flag: a flag outlives the DOM it describes, and
  // a remounted nav would then animate open from zero.
  const isPlaced = indicatorRef.value.offsetWidth > 0

  if (!isPlaced || prefersReducedMotion) {
    gsap.set(indicatorRef.value, {
      left: targetLeft,
      top: targetTop,
      width: targetWidth,
      height: targetHeight,
    })
    isIndicatorPositioned.value = true
  } else {
    gsap.to(indicatorRef.value, {
      left: targetLeft,
      top: targetTop,
      width: targetWidth,
      height: targetHeight,
      duration: 0.3,
      ease: 'power2.out',
    })
  }
}

watch([() => route.path, () => links.value.length], () => {
  nextTick(() => {
    // Small delay to ensure aria-current is set
    setTimeout(updateIndicator, 10)
  })
})

// Measured when the nav is actually laid out, rather than on a guessed delay.
useResizeObserver(navRef, updateIndicator)

const showNavigation = computed(() => {
  const path = route.path
  if (path.startsWith('/settings')) return false
  if (path === '/challenges') return true
  if (path.startsWith('/challenges/')) return false
  return true
})

// A router hook rather than a global middleware file: a layer's `*.global.ts`
// would run ahead of the root `01.auth.global.ts`. The name is constant and
// only the direction variable changes, because under `mode: 'out-in'` the
// leaving page reads the route being left.
const router = useRouter()
const stopPageTransitionHook = router.beforeEach((to, from) => {
  const direction = pageTransitionDirection(from.path, to.path)
  if (direction === null) return

  const meta = pageTransitionMeta(direction)
  to.meta.pageTransition = meta
  if (!meta) return

  document.documentElement.style.setProperty(
    '--page-direction',
    String(direction),
  )
})
onUnmounted(stopPageTransitionHook)

const { $pwa } = useNuxtApp()
</script>

<template>
  <div class="text-default relative mx-auto h-dvh w-full max-w-xl">
    <div class="h-full">
      <slot />
    </div>
    <Transition
      enter-active-class="transition duration-200 ease-out motion-reduce:transition-none"
      enter-from-class="opacity-0 scale-95 translate-y-4"
      leave-active-class="transition duration-150 ease-in motion-reduce:transition-none"
      leave-to-class="opacity-0 scale-95 translate-y-4"
    >
      <div
        v-if="showNavigation"
        class="origin-bottom fixed inset-x-0 bottom-0 flex flex-col"
      >
        <Transition
          enter-active-class="transition duration-300 ease-out"
          enter-from-class="opacity-0 translate-y-4"
        >
          <PwaUpdateBanner v-if="$pwa?.needRefresh" />
        </Transition>
        <ProgressiveBlur
          class="p-navigation-outside pb-[max(var(--spacing-navigation-outside),env(safe-area-inset-bottom))] from-shadow-blank/0 to-shadow-default bg-linear-to-b"
        >
          <ul
            ref="navRef"
            class="bg-background-raised shadow-large rounded-navigation p-navigation-inset relative mx-auto grid w-full max-w-xl gradient-border"
            :style="{
              gridTemplateColumns: `repeat(${links.length}, minmax(0, 1fr))`,
            }"
          >
            <li v-for="link in links" :key="link.label" class="grow z-10">
              <NuxtLink
                :to="link.to"
                class="px-default text-center rounded-navigation-inset text-tiny flex h-14 flex-col items-center justify-center gap-0.5"
                active-class="text-accent-contrast"
                v-on="pressListeners"
              >
                <span class="relative">
                  <UIcon
                    v-if="link.icon"
                    :name="link.icon"
                    class="size-7 shrink-0"
                  />

                  <span
                    v-if="link.badge"
                    class="rounded-full h-4.5 min-w-4.5 px-1.25 bg-accent-negative flex items-center justify-center absolute -top-0.5 left-5 text-caption text-text-default text-start"
                  >
                    {{ link.badge }}
                  </span>
                </span>
                <span class="text-xs">{{ link.label }}</span>
              </NuxtLink>
            </li>
            <div
              ref="indicatorRef"
              :class="[
                'rounded-navigation-inset bg-background-indent absolute top-0 left-0',
                isIndicatorPositioned ? 'opacity-100' : 'opacity-0',
              ]"
            />
          </ul>
        </ProgressiveBlur>
      </div>
    </Transition>
  </div>
</template>
