<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

gql(`
  query AdminSidebar {
    projects {
      id
      name
      endDate
      startDate
    }
  }
`)

const { data } = useAdminSidebarQuery()
const projectsLinks = computed(() => {
  return data.value?.projects.map((project) => ({
    label: project.name,
    badge: isWithinRange(new Date(), project.startDate, project.endDate)
      ? 'Current'
      : undefined,
    to: `/admin/projects/${project.id}`,
  }))
})

const route = useRoute()

const links = computed<NavigationMenuItem[]>(() => [
  {
    label: 'Home',
    icon: 'lucide:house',
    to: '/admin',
  },
  {
    label: 'Projects',
    icon: 'lucide:layers',
    active: route.fullPath.includes('/projects'),
    to: '/admin/projects',
  },
  {
    label: 'Users',
    icon: 'lucide:users',
    active: route.fullPath.includes('/users'),
    to: '/admin/users',
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
  <div>
    <header class="border-b border-default">
      <UContainer class="flex gap-6 lg:gap-12 items-center">
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
          <UDashboardSearchButton />
          <AdminUserMenu />
        </div>
      </UContainer>
    </header>
    <slot />
    <UDashboardSearch :groups="groups" />
  </div>
</template>
