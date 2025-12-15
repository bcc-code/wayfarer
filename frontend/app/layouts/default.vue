<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import '~/assets/styles/user.css'

const { t } = useI18n()

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
        logo
        colors{
          dark{
            accent
            accentContrast
            onAccent
            backgroundDefault
            backgroundRaised
            backgroundIndent
            textDefault
            textMuted
            textHint
            shadowDefault
            shadowBlank
            borderDefault
          }
          light{
            accent
            accentContrast
            onAccent
            backgroundDefault
            backgroundRaised
            backgroundIndent
            textDefault
            textMuted
            textHint
            shadowDefault
            shadowBlank
            borderDefault
          }
        }
        rounding
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
const activeMenuItem = ref()

watch(
  () => route.path,
  () => {
    nextTick(() => {
      activeMenuItem.value = document.querySelector(`[aria-current="page"]`)
    })
  },
  { immediate: true },
)

const { left, height, width, top } = useElementBounding(activeMenuItem)

const showNavigation = computed(() => {
  const path = route.path
  if (path.startsWith('/settings')) return false
  if (path === '/challenges') return true
  if (path.startsWith('/challenges/')) return false
  return true
})
</script>

<template>
  <div class="text-default relative mx-auto h-full w-full max-w-xl">
    <div class="h-full">
      <slot />
    </div>
    <div v-if="showNavigation" class="fixed inset-x-0 bottom-0">
      <ProgressiveBlur
        class="p-navigation-outside pb-[calc(var(--spacing-navigation-outside)+env(safe-area-inset-bottom))] from-shadow-blank/0 to-shadow-default bg-linear-to-b"
      >
        <ul
          class="bg-background-raised shadow-large rounded-navigation p-navigation-inset relative mx-auto grid w-full max-w-xl grid-cols-3"
        >
          <li v-for="link in links" :key="link.label" class="grow z-10">
            <NuxtLink
              :to="link.to"
              class="px-default rounded-navigation-inset text-tiny flex h-14 flex-col items-center justify-center gap-0.5 transition-all duration-150 ease-out"
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
            class="rounded-navigation-inset bg-background-indent fixed aspect-square transition-all duration-150 ease-out"
            :style="{
              left: left + 'px',
              top: top + 'px',
              height: height + 'px',
              width: width + 'px',
            }"
          />
        </ul>
      </ProgressiveBlur>
    </div>
  </div>
</template>
