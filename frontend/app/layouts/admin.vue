<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import '~/assets/styles/admin.css'

// Force Norwegian locale in admin
const { setLocale } = useI18n()
setLocale('nb')

// PWA update notification
const { $pwa } = useNuxtApp()
const toast = useToast()

watch(
  () => $pwa?.needRefresh,
  (needRefresh) => {
    if (needRefresh) {
      toast.add({
        id: 'pwa-update',
        title: 'Oppdatering tilgjengelig',
        description: 'En ny versjon av appen er klar.',
        icon: 'lucide:download',
        close: false,
        duration: 0,
        color: 'neutral',
        actions: [
          {
            label: 'Oppdater nå',
            color: 'neutral',
            onClick: () => $pwa?.updateServiceWorker(true),
          },
        ],
      })
    }
  },
  { immediate: true },
)

useHead({
  title: 'Interact Admin',
})

const { me, isLoading, isAuth0Loading, token } = useAuth()
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
  [isLoading, isAuth0Loading, me, token, () => route.path],
  ([loading, auth0Loading, user, hasToken, path]) => {
    // Wait for both Wayfarer auth and Auth0 to finish loading
    if (loading || auth0Loading) return
    // If we have a token but no user data yet, wait for the query to complete
    if (hasToken && !user) return
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
