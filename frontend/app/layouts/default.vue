<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const links = [
  {
    label: 'Standings',
    icon: 'lucide:list',
    to: { name: 'index' },
  },
  {
    label: 'Challenges',
    icon: 'lucide:medal',
    to: { name: 'challenges' },
  },
  {
    label: 'Unit',
    icon: 'lucide:users',
    to: { name: 'unit' },
  },
  {
    label: 'Profile',
    icon: 'lucide:user',
    to: { name: 'profile' },
  },
] satisfies NavigationMenuItem[]

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

// Force light theme for now
onBeforeMount(() => {
  document.documentElement.classList.remove('dark')
  document.documentElement.classList.add('light')
})

const { data } = useCurrentProjectQuery()

watch(data, (newData) => {
  if (!newData) return

  const styleElement = document.createElement('style')
  styleElement.innerHTML = `
    /* Current project theme */
    :root {
      --ui-primary: ${newData.myCurrentProject.branding.colors.primary};
      --ui-radius: ${newData.myCurrentProject.branding.rounding}px;
    }

    body {
      background-color: var(--color-background-default);
    }
  `
  document.body.appendChild(styleElement)
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

const { left, height, width } = useElementBounding(activeMenuItem)
</script>

<template>
  <div class="text-default relative h-full">
    <div class="h-full">
      <slot />
    </div>
    <div
      class="p-navigation-outside from-shadow-blank/0 to-shadow-default fixed inset-x-0 bottom-0 bg-linear-to-b"
    >
      <ul
        class="bg-background-raised shadow-large rounded-navigation p-navigation-inset relative grid grid-cols-4"
      >
        <li v-for="link in links" :key="link.label" class="grow">
          <NuxtLink
            :to="link.to"
            class="px-default rounded-navigation-inset text-tiny flex h-14 flex-col items-center justify-center gap-0.5 transition-all duration-150 ease-out"
            active-class="text-accent-contrast"
          >
            <UIcon :name="link.icon" class="shrink-0" />
            <span class="text-xs">{{ link.label }}</span>
          </NuxtLink>
        </li>
        <div
          class="m-navigation-inset rounded-navigation-inset bg-background-indent absolute aspect-square transition-all duration-150 ease-out"
          :style="{
            left: `calc(${left}px - var(--spacing-default))`,
            height: height + 'px',
            width: width + 'px',
          }"
        />
      </ul>
    </div>
  </div>
</template>
