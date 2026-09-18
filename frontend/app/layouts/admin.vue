<script setup lang="ts">
import '~/assets/styles/admin.css'

// Force Norwegian locale in admin.
// Deliberately kept in this layout rather than a shared composable: the
// my-church subtree uses the church-admin layout and is the one fully
// translated admin area, so it must not inherit this.
const { setLocale } = useI18n()
setLocale('nb')

// Initialize Firestore sync for realtime updates
const { initialize: initFirestoreSync } = useFirestoreSync()
onMounted(() => {
  initFirestoreSync()
})

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
  canAccessMaintenance,
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

// Redirect unauthorized users after auth loads.
// TODO: replaced by a `permission` route-meta key plus one global middleware —
// see notes/frontend-admin-restructure.md. Left as-is here so the shell rewrite
// and the permission change land in separate, separately revertible commits.
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
      return
    }

    // Redirect users away from routes they don't have permission for
    const routePermissions: [string, boolean][] = [
      ['/admin/users', !!canAccessUsers.value],
      ['/admin/consents', !!canAccessConsents.value],
      ['/admin/maintenance', !!canAccessMaintenance.value],
      ['/admin/teams', !!canAccessTeams.value],
      ['/admin/projects', !!canAccessProjects.value],
      ['/admin/scores', !!canAccessScores.value],
      ['/admin/feedback', !!canAccessFeedback.value],
    ]
    for (const [route, allowed] of routePermissions) {
      if (path.startsWith(route) && !allowed) {
        navigateTo('/admin')
        return
      }
    }
  },
  { immediate: true },
)

const { navItems, searchGroups } = useAdminNav()

// Church-admin-only users are redirected to the church-admin layout by the
// watcher above; suppress the nav in the frame or two before that lands.
const showNav = computed(() => !isChurchAdminOnly.value)
</script>

<template>
  <UDashboardGroup storage="local" storage-key="wayfarer-admin">
    <UDashboardSidebar
      id="admin-sidebar"
      collapsible
      resizable
      :default-size="16"
    >
      <template #header="{ collapsed }">
        <NuxtLink to="/admin" class="flex items-center">
          <UColorModeImage
            v-if="!collapsed"
            light="/images/logo/logo.svg"
            dark="/images/logo/logo-light.svg"
            class="h-6"
          />
        </NuxtLink>
        <UDashboardSidebarCollapse class="ms-auto" />
      </template>

      <template #default="{ collapsed }">
        <template v-if="showNav">
          <AdminProjectSwitcher :collapsed />
          <UNavigationMenu
            :items="navItems"
            :collapsed
            orientation="vertical"
            highlight
          />
        </template>
      </template>

      <template #footer="{ collapsed }">
        <AdminUserMenu :collapsed />
      </template>
    </UDashboardSidebar>

    <!--
      Pages still bring their own `UContainer` and vertical padding, so the
      panel body's default padding is removed here to avoid doubling it. The
      responsive variants have to be zeroed explicitly: the body default is
      `p-4 sm:p-6 gap-4 sm:gap-6`, and tailwind-merge is variant-aware, so a
      bare `p-0` would only override the base and leave `sm:p-6` in place.
      As pages move onto a shared page header this override goes away.
    -->
    <UDashboardPanel
      id="admin-main"
      :ui="{ body: 'p-0 sm:p-0 gap-0 sm:gap-0' }"
    >
      <template #header>
        <UDashboardNavbar>
          <template #right>
            <AdminUserFeedback />
          </template>
        </UDashboardNavbar>
      </template>

      <template #body>
        <slot />
      </template>
    </UDashboardPanel>

    <UDashboardSearch :groups="searchGroups" />
    <AdminConfirmDialog />
    <QuickAccess />
  </UDashboardGroup>
</template>
