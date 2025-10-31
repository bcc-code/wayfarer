<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import { useCurrentProjectQuery } from '~/api/generated'

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
  user {
    currentProject {
      branding {
        logo
        colors {
          primary
          secondary
          tertiary
        }
      }
    }
  }
}
`)

const { data } = useCurrentProjectQuery()

watch(data, (newData) => {
  if (!newData) return

  const styleElement = document.createElement('style')
  styleElement.innerHTML = `
    /* Current project theme */
    :root {
      --ui-primary: ${newData.user.currentProject.branding.colors.primary};
      --ui-secondary: ${newData.user.currentProject.branding.colors.secondary};
      --ui-tertiary: ${newData.user.currentProject.branding.colors.tertiary};
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
    <ul
      class="fixed bottom-0 right-0 left-0 bg-default border-t border-default grid grid-cols-4"
    >
      <li v-for="link in links" :key="link.label" class="grow">
        <NuxtLink
          :to="link.to"
          class="flex flex-col items-center justify-center h-full gap-2 p-4"
          active-class="bg-primary/5 text-primary"
        >
          <UIcon :name="link.icon" class="shrink-0" />
          <span class="text-xs">{{ link.label }}</span>
        </NuxtLink>
      </li>
    </ul>
  </UContainer>
</template>
