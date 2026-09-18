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
} from '~/utils/adminNav'

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
   * a name with that route's own param shape. Every nav target either takes no
   * params or takes exactly `projectId`, which the model guarantees.
   */
  const toLocation = (item: AdminNavItem): RouteLocationRaw =>
    (projectId.value
      ? { name: item.to, params: { projectId: projectId.value } }
      : { name: item.to }) as RouteLocationRaw

  const toMenuItem = (item: AdminNavItem): NavigationMenuItem => ({
    label: item.label,
    icon: item.icon,
    to: toLocation(item),
    active: isNavItemActive(item, route.name as string | undefined),
  })

  const globalNav = computed(() =>
    visibleNavItems(GLOBAL_NAV, permissions, context.value).map(toMenuItem),
  )

  // Only meaningful inside a project; empty elsewhere so the sidebar can drop
  // the whole group.
  const projectNav = computed(() =>
    projectId.value
      ? visibleNavItems(PROJECT_NAV, permissions, context.value).map(toMenuItem)
      : [],
  )

  /** `UNavigationMenu` renders each sub-array as its own divided group. */
  const navItems = computed<NavigationMenuItem[][]>(() =>
    projectNav.value.length
      ? [globalNav.value, projectNav.value]
      : [globalNav.value],
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

  /**
   * Label of the deepest matching nav entry, for the navbar heading. Project
   * entries win over global ones because they are the more specific match.
   */
  const currentTitle = computed(() => {
    const name = route.name as string | undefined
    const match =
      [...PROJECT_NAV].reverse().find((item) => isNavItemActive(item, name)) ??
      [...GLOBAL_NAV].reverse().find((item) => isNavItemActive(item, name))
    return match?.label ?? 'Admin'
  })

  return {
    projectId,
    globalNav,
    projectNav,
    navItems,
    searchGroups,
    currentTitle,
  }
}
