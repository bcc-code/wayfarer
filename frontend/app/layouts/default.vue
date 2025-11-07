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
        class="bg-background-raised shadow-large rounded-navigation p-navigation-inset grid grid-cols-4"
      >
        <li v-for="link in links" :key="link.label" class="grow">
          <NuxtLink
            :to="link.to"
            class="px-default rounded-navigation-inset text-tiny flex h-14 flex-col items-center justify-center gap-0.5"
            active-class="bg-background-indent text-accent-contrast"
          >
            <UIcon :name="link.icon" class="shrink-0" />
            <span class="text-xs">{{ link.label }}</span>
          </NuxtLink>
        </li>
      </ul>
    </div>
  </div>
</template>
