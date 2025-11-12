<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import '~/assets/styles/admin.css'

useHead({
  title: 'Wayfarer Admin',
})

gql(`
  query AdminSidebar {
    projects {
      edges {
        node {
          id
          name
          endDate
          startDate
        }
      }
    }
  }
`)

const { data } = useAdminSidebarQuery()
const projectsLinks = computed(() => {
  return data.value?.projects.edges.map(({ node: project }) => ({
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
</script>

<template>
  <div>
    <header class="border-default border-b">
      <UContainer class="flex items-center gap-6 lg:gap-12">
        <NuxtLink to="/admin" class="font-serif text-xl">Wayfarer</NuxtLink>
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
