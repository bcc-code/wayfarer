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
const { currentProjects, futureProjects } = useGroupedProjects(
  () => data.value?.admin.projects,
)
const projectsLinks = computed(() => {
  return [...currentProjects.value, ...futureProjects.value].map((project) => ({
    label: project.name,
    badge: isWithinRange(new Date(), project.startDate, project.endDate)
      ? 'Current'
      : undefined,
    to: `/admin/projects/${project.id}`,
  }))
})

const links = computed<NavigationMenuItem[]>(() => [
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
])

const groups = computed(() => [
  {
    id: 'links',
    label: 'Go to',
    items: links.value.flat(),
  },
  {
    id: 'projects',
    label: 'Projects',
    items: projectsLinks.value,
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
    <UDashboardPanel>
      <header class="border-b border-default">
        <UContainer class="flex gap-8 items-center">
          <NuxtLink to="/admin" class="font-serif text-xl text-primary">
            Wayfarer
          </NuxtLink>
          <UNavigationMenu
            :items="links"
            highlight
            variant="link"
            orientation="horizontal"
          />
          <div class="ml-auto flex gap-2">
            <UDashboardSearchButton class="bg-transparent ring-default" />
            <AdminUserMenu />
          </div>
        </UContainer>
      </header>
      <slot />
      <UDashboardSearch :groups="groups" />
    </UDashboardPanel>
  </UDashboardGroup>
</template>
