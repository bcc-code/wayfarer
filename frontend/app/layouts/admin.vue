<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import { useAdminSidebarQuery } from '~/api/generated'

gql(`
  query AdminSidebar {
    admin {
      projects {
        id
        name
        endDate
        startDate
      }
    }
  }
`)

const { data } = useAdminSidebarQuery()
const projects = computed(() => {
  return (
    data.value?.admin.projects
      .filter((project) => new Date(project.startDate) < new Date())
      .sort((a, b) =>
        isWithinRange(new Date(), a.startDate, a.endDate)
          ? -1
          : isWithinRange(new Date(), b.startDate, b.endDate)
            ? 1
            : 0,
      ) ?? []
  )
})

const open = ref(false)

const links = computed<NavigationMenuItem[][]>(() => [
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
      children:
        projects.value.map((project) => ({
          label: project.name,
          badge: isWithinRange(new Date(), project.startDate, project.endDate)
            ? 'Current'
            : undefined,
          to: `/admin/projects/${project.id}`,
        })) ?? [],
    },
  ],
])

const groups = computed(() => [
  {
    id: 'links',
    label: 'Go to',
    items: links.value.flat(),
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
          :items="links"
          orientation="vertical"
          tooltip
          popover
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
