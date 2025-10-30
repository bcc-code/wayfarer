<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const open = ref(false)

const links = [
  [
    {
      label: 'Home',
      icon: 'lucide:house',
      to: '/admin',
    },
    {
      label: 'Projects',
      icon: 'lucide:layers',
      to: '/admin/projects',
    },
  ],
] satisfies NavigationMenuItem[][]

const groups = computed(() => [
  {
    id: 'links',
    label: 'Go to',
    items: links.flat(),
  },
])

// Black and white theme
onMounted(() => {
  const styleElement = document.createElement('style')
  styleElement.innerHTML = `
    /* Admin theme */
    :root {
      --ui-primary: black;
    }
    .dark {
      --ui-primary: white;
    }
  `
  document.body.appendChild(styleElement)
})
</script>

<template>
  <UDashboardGroup unit="rem">
    <UDashboardSidebar
      id="default"
      v-model:open="open"
      collapsible
      resizable
      class="bg-elevated/25"
      :ui="{ footer: 'lg:border-t lg:border-default' }"
    >
      <template #header>
        <p class="font-serif text-xl ml-2 text-primary">Wayfarer</p>
      </template>

      <template #default="{ collapsed }">
        <UDashboardSearchButton
          :collapsed="collapsed"
          class="bg-transparent ring-default"
        />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[0]"
          orientation="vertical"
          tooltip
          popover
        />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[1]"
          orientation="vertical"
          tooltip
          class="mt-auto"
        />
      </template>

      <template #footer="{ collapsed }">
        <AdminUserMenu :collapsed="collapsed" />
      </template>
    </UDashboardSidebar>

    <UDashboardSearch :groups="groups" />

    <slot />
  </UDashboardGroup>
</template>
