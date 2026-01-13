<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import '~/assets/styles/admin.css'

// Force Norwegian locale in admin
const { setLocale } = useI18n()
setLocale('nb')

useHead({
  title: 'Interact Admin',
})

const { me, isLoading } = useAuth()
const {
  canAccessAdmin,
  canAccessProjects,
  canAccessUsers,
  canAccessTeams,
  canAccessConsents,
  canAccessScores,
  canAccessFeedback,
} = usePermissions()

const route = useRoute()

// Check if user is church-admin-only
const isChurchAdminOnly = computed(() => {
  if (!me.value) return false
  const hasFullAdminRole = me.value.roles.some((role: { role: RoleType }) =>
    [RoleType.Admin, RoleType.Superadmin].includes(role.role),
  )
  return (
    !hasFullAdminRole &&
    me.value.roles.some(
      (role: { role: RoleType }) => role.role === RoleType.ChurchAdmin,
    )
  )
})

// Redirect unauthorized users after auth loads
watch(
  [isLoading, me, () => route.path],
  ([loading, user, path]) => {
    if (loading) return
    if (!user || !canAccessAdmin.value) {
      navigateTo('/')
      return
    }

    // Restrict church-admin-only users to /admin/my-church
    if (isChurchAdminOnly.value && !path.startsWith('/admin/my-church')) {
      navigateTo('/admin/my-church')
    }
  },
  { immediate: true },
)

const links = computed<NavigationMenuItem[]>(() => {
  // Church-admin-only users use a different layout, don't show main admin nav
  if (isChurchAdminOnly.value) {
    return []
  }

  const items: NavigationMenuItem[] = [
    {
      label: 'Hjem',
      icon: 'lucide:house',
      to: '/admin',
    },
  ]

  if (canAccessProjects.value) {
    items.push({
      label: 'Prosjekter',
      icon: 'lucide:layers',
      active: route.fullPath.includes('/projects'),
      to: '/admin/projects',
    })
  }

  if (canAccessTeams.value) {
    items.push({
      label: 'Lag',
      icon: 'lucide:users-round',
      active: route.fullPath.includes('/teams'),
      to: '/admin/teams',
    })
  }

  if (canAccessUsers.value) {
    items.push({
      label: 'Brukere',
      icon: 'lucide:user',
      active: route.fullPath.includes('/users'),
      to: '/admin/users',
    })
  }

  if (canAccessScores.value) {
    items.push({
      label: 'Poeng',
      icon: 'lucide:trophy',
      active: route.fullPath.includes('/scores'),
      to: '/admin/scores',
    })
  }

  if (canAccessConsents.value) {
    items.push({
      label: 'Samtykker',
      icon: 'lucide:file-check',
      active: route.fullPath.includes('/consents'),
      to: '/admin/consents',
    })
  }

  if (canAccessFeedback.value) {
    items.push({
      label: 'Tilbakemeldinger',
      icon: 'lucide:message-square',
      active: route.fullPath.includes('/feedback'),
      to: '/admin/feedback',
    })
  }

  return items
})

const groups = computed(() => [
  {
    id: 'links',
    label: 'Gå til',
    items: links.value.flat(),
  },
])
</script>

<template>
  <div class="bg-default h-full">
    <header class="border-default border-b">
      <UContainer class="flex items-center gap-6 lg:gap-12">
        <NuxtLink to="/admin" class="font-serif text-xl">
          <UColorModeImage
            light="/images/logo/logo.svg"
            dark="/images/logo/logo-light.svg"
            class="h-6"
          />
        </NuxtLink>
        <UNavigationMenu
          :items="links"
          highlight
          variant="link"
          orientation="horizontal"
        />
        <div class="ml-auto flex gap-2">
          <AdminUserMenu />
        </div>
      </UContainer>
    </header>
    <slot />
    <UDashboardSearch :groups="groups" />
    <QuickAccess />
  </div>
</template>
