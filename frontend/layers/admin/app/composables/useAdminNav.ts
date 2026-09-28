import type {
  CommandPaletteGroup,
  CommandPaletteItem,
  NavigationMenuItem,
} from '@nuxt/ui'
import type { RouteLocationRaw } from 'vue-router'
import {
  GLOBAL_NAV,
  PROJECT_NAV,
  isNavItemActive,
  visibleNavItems,
  type AdminNavItem,
} from '../utils/adminNav'

/**
 * Turns the declarative nav model into the shapes `UNavigationMenu` and
 * `UDashboardSearch` expect. Both read the same source, so the sidebar and the
 * command palette cannot drift apart.
 */
export function useAdminNav() {
  const permissions = usePermissions()
  const route = useRoute()

  // `route.params` is a union across every route under typedPages, so the key
  // has to be probed rather than read directly.
  const projectId = computed(() =>
    'projectId' in route.params ? route.params.projectId : undefined,
  )

  const context = computed(() => ({ projectId: projectId.value }))

  /**
   * Route names are only known at runtime here, so TypeScript cannot correlate
   * a name with that route's own param shape. `scoped` says which of the two
   * shapes the target has: PROJECT_NAV entries take exactly `projectId`,
   * GLOBAL_NAV entries take none. Handing `projectId` to a route that does not
   * declare it makes vue-router discard it with a "Discarded invalid param(s)"
   * warning on every render while inside a project.
   */
  const toLocation = (item: AdminNavItem, scoped: boolean): RouteLocationRaw =>
    (scoped && projectId.value
      ? { name: item.to, params: { projectId: projectId.value } }
      : { name: item.to }) as RouteLocationRaw

  const toMenuItem = (
    item: AdminNavItem,
    scoped: boolean,
  ): NavigationMenuItem => ({
    label: item.label,
    icon: item.icon,
    to: toLocation(item, scoped),
    active: isNavItemActive(item, route.name as string | undefined),
  })

  const globalNav = computed(() =>
    visibleNavItems(GLOBAL_NAV, permissions, context.value).map((item) =>
      toMenuItem(item, false),
    ),
  )

  // Only meaningful inside a project; empty elsewhere so the sidebar can drop
  // the whole group.
  const projectNav = computed(() =>
    projectId.value
      ? visibleNavItems(PROJECT_NAV, permissions, context.value).map((item) =>
          toMenuItem(item, true),
        )
      : [],
  )

  const toPaletteItems = (items: NavigationMenuItem[]): CommandPaletteItem[] =>
    items.map((item) => ({
      label: item.label,
      icon: item.icon,
      to: item.to,
    }))

  const searchGroups = computed<CommandPaletteGroup<CommandPaletteItem>[]>(
    () => [
      {
        id: 'global',
        label: 'Gå til',
        items: toPaletteItems(globalNav.value),
      },
      ...(projectNav.value.length
        ? [
            {
              id: 'project',
              label: 'I dette prosjektet',
              items: toPaletteItems(projectNav.value),
            },
          ]
        : []),
    ],
  )

  return {
    projectId,
    globalNav,
    projectNav,
    searchGroups,
  }
}
