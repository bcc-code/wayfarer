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

// Access control lives in `middleware/admin-permission.global.ts`; the layout
// only needs to know whether to render the main nav. Church admins who hold no
// other admin role are confined to /admin/my-church, which uses its own layout.
const { me } = useAuth()
const showNav = computed(() => !isChurchAdminOnly(me.value?.roles))

const { globalNav, projectNav, searchGroups, currentTitle } = useAdminNav()
// `ssr: false`, so this is safe to gate rendering on.
const isDesktop = useMediaQuery('(min-width: 1024px)')
</script>

<template>
  <!--
    The page sits one step behind the surfaces on it, in both modes and in a
    single palette. Inheriting the root's `bg-background-default` instead would
    mix palettes: it is the user app's grey (#efefef / #222222), and #222222 is
    actually *lighter* than the zinc-900 that UCard and the tables sit on, so
    panels read as sunk into the page rather than raised off it.
  -->
  <UDashboardGroup class="bg-neutral-100 dark:bg-neutral-950">
    <!--
      `m-2` floats the sidebar off the chrome, which needs the theme's
      `min-h-svh` undone or it overflows the viewport by that margin, and its
      `border-e` dropped since the floating surface is framed by its own ring.

      `default-size` is a percentage of the viewport (the dashboard context uses
      `unit: '%'`), so 16% passes 300px on anything wider than ~1875px. The cap
      is a max-width rather than a smaller percentage, which would make the
      sidebar too narrow on a laptop.
    -->
    <UDashboardSidebar
      id="admin-sidebar"
      :default-size="16"
      :ui="{
        root: 'm-2 min-h-0 max-w-[300px] rounded-lg overflow-hidden border-e-0',
      }"
    >
      <template #header>
        <NuxtLink to="/admin" class="flex items-center">
          <UColorModeImage
            light="/images/logo/logo.svg"
            dark="/images/logo/logo-light.svg"
            class="h-6"
          />
        </NuxtLink>
      </template>

      <template #default>
        <template v-if="showNav">
          <!-- Renders the meta+K hint itself and opens UDashboardSearch, which
               was previously reachable only by the shortcut. -->
          <UDashboardSearchButton label="Søk" variant="soft" />
          <UNavigationMenu
            :items="globalNav"
            orientation="vertical"
            highlight
          />

          <!-- Below `lg` the project section has no column of its own, so it
               rides along in this sidebar's slideover. -->
          <template v-if="projectNav.length">
            <USeparator class="lg:hidden" />
            <div class="lg:hidden">
              <AdminProjectSwitcher />
            </div>
            <UNavigationMenu
              :items="projectNav"
              orientation="vertical"
              highlight
              class="lg:hidden"
            />
          </template>
        </template>
      </template>

      <template #footer>
        <AdminUserMenu />
      </template>
    </UDashboardSidebar>

    <!--
      Project context gets its own column rather than a second block in the
      primary sidebar, so the global nav stays a fixed landmark while this one
      comes and goes with the route. The switcher heads this column rather than
      the primary: it is the project's identity, and having it in both places
      printed the same truncated name twice, side by side.

      `shadow-none` is the hierarchy signal — the primary rail keeps its shadow
      and reads as the top layer, this one sits flatter behind it. Only the inline margin is set: the primary
      already keeps an 8px gutter on its end, so `me-2` alone leaves an even
      rhythm between the two panels and the content.

      The header is not decoration — `min-h-(--ui-header-height)` is what lines
      its first nav item up with the primary sidebar's, and with the navbar.
    -->
    <UDashboardSidebar
      v-if="isDesktop && showNav && projectNav.length"
      id="admin-project-sidebar"
      :default-size="13"
      :ui="{
        root: 'my-2 me-2 min-h-0 max-w-[300px] rounded-lg overflow-hidden border-e-0 shadow-none',
      }"
    >
      <template #header>
        <AdminProjectSwitcher />
      </template>

      <template #default>
        <UNavigationMenu :items="projectNav" orientation="vertical" highlight />
      </template>
    </UDashboardSidebar>

    <UDashboardPanel id="admin-main">
      <template #header>
        <UDashboardNavbar :title="currentTitle">
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
