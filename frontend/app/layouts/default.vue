<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import '~/assets/styles/user.css'
import { getContrastColor } from '~/utils/colors'

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
        colors {
          primary
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

// watch(data, (newData) => {
//   if (!newData) return

//   const primaryColor = newData.myCurrentProject.branding.colors.primary

//   // Set dynamic project theme color
//   document.documentElement.style.setProperty(
//     '--color-accent-base',
//     primaryColor,
//   )

//   // Calculate and set the on-accent contrast color
//   const contrastColor = getContrastColor(primaryColor)
//   document.documentElement.style.setProperty('--color-on-accent', contrastColor)
// })

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
</script>

<template>
  <div class="text-default relative mx-auto h-full w-full max-w-xl">
    <div class="h-full">
      <slot />
    </div>
    <div v-if="route.path != '/settings'" class="fixed inset-x-0 bottom-0">
      <ProgressiveBlur
        class="p-navigation-outside from-shadow-blank/0 to-shadow-default bg-linear-to-b"
      >
        <ul
          class="bg-background-raised shadow-large rounded-navigation p-navigation-inset relative mx-auto grid w-full max-w-xl grid-cols-3"
        >
          <li v-for="link in links" :key="link.label" class="grow">
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
