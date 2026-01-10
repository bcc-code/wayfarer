<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import { gsap } from 'gsap'
import '~/assets/styles/user.css'

const { t } = useI18n()

// Initialize Firestore sync for realtime updates
const { initialize: initFirestoreSync } = useFirestoreSync()
onMounted(() => {
  initFirestoreSync()
})

const links = computed<NavigationMenuItem[]>(() => [
  {
    label: t('navigation.profile'),
    icon: 'IconProfile',
    to: { name: 'index' },
  },
  {
    label: t('navigation.standings'),
    icon: 'IconStandings',
    to: { name: 'standings' },
  },
  {
    label: t('navigation.challenges'),
    icon: 'IconChallenges',
    to: { name: 'challenges' },
  },
])

// Current project theme
gql(`
  query CurrentProject {
    myCurrentProject {
      branding {
        ...BrandingFields
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data } = useCurrentProjectQuery({
  pause: computed(() => !isAuthReady.value),
})

watch(data, (newData) => {
  if (!newData) return

  const style = `
  <style id="theme">
  :root {
    --color-accent: ${newData.myCurrentProject.branding.colors.light.accent};
    --color-accent-contrast: ${newData.myCurrentProject.branding.colors.light.accentContrast};
    --color-on-accent: ${newData.myCurrentProject.branding.colors.light.onAccent};
    --color-background-default: ${newData.myCurrentProject.branding.colors.light.backgroundDefault};
    --color-background-raised: ${newData.myCurrentProject.branding.colors.light.backgroundRaised};
    --color-background-indent: ${newData.myCurrentProject.branding.colors.light.backgroundIndent};
    --color-text-default: ${newData.myCurrentProject.branding.colors.light.textDefault};
    --color-text-muted: ${newData.myCurrentProject.branding.colors.light.textMuted};
    --color-text-hint: ${newData.myCurrentProject.branding.colors.light.textHint};
    --color-shadow-default: ${newData.myCurrentProject.branding.colors.light.shadowDefault};
    --color-shadow-blank: ${newData.myCurrentProject.branding.colors.light.shadowBlank};
    --color-border-default: ${newData.myCurrentProject.branding.colors.light.borderDefault};
  }

  .dark {
    --color-accent: ${newData.myCurrentProject.branding.colors.dark.accent};
    --color-accent-contrast: ${newData.myCurrentProject.branding.colors.dark.accentContrast};
    --color-on-accent: ${newData.myCurrentProject.branding.colors.dark.onAccent};
    --color-background-default: ${newData.myCurrentProject.branding.colors.dark.backgroundDefault};
    --color-background-raised: ${newData.myCurrentProject.branding.colors.dark.backgroundRaised};
    --color-background-indent: ${newData.myCurrentProject.branding.colors.dark.backgroundIndent};
    --color-text-default: ${newData.myCurrentProject.branding.colors.dark.textDefault};
    --color-text-muted: ${newData.myCurrentProject.branding.colors.dark.textMuted};
    --color-text-hint: ${newData.myCurrentProject.branding.colors.dark.textHint};
    --color-shadow-default: ${newData.myCurrentProject.branding.colors.dark.shadowDefault};
    --color-shadow-blank: ${newData.myCurrentProject.branding.colors.dark.shadowBlank};
    --color-border-default: ${newData.myCurrentProject.branding.colors.dark.borderDefault};
  }
  </style>
  `

  document.head.insertAdjacentHTML('beforeend', style)
})

const route = useRoute()
const navRef = ref<HTMLElement | null>(null)
const indicatorRef = ref<HTMLElement | null>(null)
const isFirstRender = ref(true)

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

  const prefersReducedMotion = window.matchMedia(
    '(prefers-reduced-motion: reduce)',
  ).matches

  if (isFirstRender.value || prefersReducedMotion) {
    gsap.set(indicatorRef.value, {
      left: targetLeft,
      top: targetTop,
      width: targetWidth,
      height: targetHeight,
    })
    isFirstRender.value = false
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

watch(
  () => route.path,
  () => {
    nextTick(() => {
      // Small delay to ensure aria-current is set
      setTimeout(updateIndicator, 10)
    })
  },
)

onMounted(() => {
  setTimeout(updateIndicator, 50)
})

const showNavigation = computed(() => {
  const path = route.path
  if (path.startsWith('/settings')) return false
  if (path === '/challenges') return true
  if (path.startsWith('/challenges/')) return false
  return true
})

const { $pwa } = useNuxtApp()
</script>

<template>
  <div class="text-default relative mx-auto h-full w-full max-w-xl">
    <div class="h-full">
      <slot />
    </div>
    <div v-if="showNavigation" class="fixed inset-x-0 bottom-0 flex flex-col">
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
          class="bg-background-raised shadow-large rounded-navigation p-navigation-inset relative mx-auto grid w-full max-w-xl grid-cols-3 gradient-border"
        >
          <li v-for="link in links" :key="link.label" class="grow z-10">
            <NuxtLink
              :to="link.to"
              class="px-default text-center rounded-navigation-inset text-tiny flex h-14 flex-col items-center justify-center gap-0.5"
              active-class="text-accent-contrast"
            >
              <UIcon
                v-if="link.icon"
                :name="link.icon"
                class="size-7 shrink-0"
              />
              <span class="text-xs">{{ link.label }}</span>
            </NuxtLink>
          </li>
          <div
            ref="indicatorRef"
            class="rounded-navigation-inset bg-background-indent absolute top-0 left-0"
          />
        </ul>
      </ProgressiveBlur>
    </div>
  </div>
</template>
