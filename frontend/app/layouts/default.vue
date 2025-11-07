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
  `
  document.body.appendChild(styleElement)
})
</script>

<template>
  <UContainer class="relative h-full">
    <div class="py-4 h-full">
      <slot />
    </div>
    <div class="fixed inset-x-0 bottom-0 p-navigation-outside">
      <ul
        class="bg-background-raised shadow-large h-16 grid grid-cols-4 text-default rounded-navigation p-navigation-inset"
      >
        <li v-for="link in links" :key="link.label" class="grow">
          <NuxtLink
            :to="link.to"
            class="flex flex-col items-center justify-center h-full gap-0.5 px-default py-small rounded-navigation-inset"
            active-class="bg-background-indent text-accent-contrast"
          >
            <UIcon :name="link.icon" class="shrink-0" />
            <span class="text-xs">{{ link.label }}</span>
          </NuxtLink>
        </li>
      </ul>
    </div>
  </UContainer>
</template>
