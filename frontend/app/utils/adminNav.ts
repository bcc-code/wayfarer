import type { RouteNamedMap } from 'vue-router/auto-routes'
import type { usePermissions } from '~/composables/usePermissions'

/** Any route name Nuxt actually generated — typed, so a typo will not compile. */
export type AdminRouteName = keyof RouteNamedMap

type Permissions = ReturnType<typeof usePermissions>

export interface AdminNavContext {
  /** Present while the current route is inside a project. */
  projectId?: string
}

export interface AdminNavItem {
  label: string
  icon: string
  /** Typed route name — never a path, so `experimental.typedPages` checks it. */
  to: AdminRouteName
  /**
   * Route-name prefix marking this entry active. Omit for an exact match:
   * 'admin' must not light up for every `admin-*` route.
   */
  match?: AdminRouteName
  /** Visibility predicate. Omitted means always visible. */
  can?: (permissions: Permissions, context: AdminNavContext) => boolean
}

/**
 * Admin navigation, as data.
 *
 * Replaces the hand-pushed `links` array that used to live in
 * `layouts/admin.vue`, and is consumed by both the sidebar and the command
 * palette so the two cannot drift apart. Being plain data makes the filtering
 * and active-state rules unit-testable without mounting anything, and turns the
 * eventual i18n pass into an edit of this file rather than a template rewrite.
 *
 * Active state keys off route *names* rather than `route.fullPath.includes(...)`.
 * The old substring check is actively wrong once teams and scores move under a
 * project: `/admin/projects/x/teams` would light up the global "Lag" entry.
 */
export const GLOBAL_NAV: AdminNavItem[] = [
  {
    label: 'Hjem',
    icon: 'lucide:house',
    to: 'admin',
  },
  {
    label: 'Prosjekter',
    icon: 'lucide:layers',
    to: 'admin-projects',
    match: 'admin-projects',
    can: (p) => !!p.canAccessProjects.value,
  },
  {
    label: 'Lag',
    icon: 'lucide:users-round',
    to: 'admin-teams',
    match: 'admin-teams',
    can: (p) => !!p.canAccessTeams.value,
  },
  {
    label: 'Brukere',
    icon: 'lucide:user',
    to: 'admin-users',
    match: 'admin-users',
    can: (p) => !!p.canAccessUsers.value,
  },
  {
    label: 'Poeng',
    icon: 'lucide:trophy',
    to: 'admin-scores',
    match: 'admin-scores',
    can: (p) => !!p.canAccessScores.value,
  },
  {
    label: 'Samtykker',
    icon: 'lucide:file-check',
    to: 'admin-consents',
    match: 'admin-consents',
    can: (p) => !!p.canAccessConsents.value,
  },
  {
    label: 'Tilbakemeldinger',
    icon: 'lucide:message-square',
    to: 'admin-feedback',
    match: 'admin-feedback',
    can: (p) => !!p.canAccessFeedback.value,
  },
  {
    label: 'Vedlikehold',
    icon: 'lucide:wrench',
    to: 'admin-maintenance',
    match: 'admin-maintenance',
    can: (p) => !!p.canAccessMaintenance.value,
  },
]

/**
 * Shown only while inside `/admin/projects/:projectId`.
 *
 * Limited to routes that exist today. The sections still living as tabs on
 * `projects/[projectId]/index.vue` — challenges, achievements, events — plus the
 * project-scoped teams and scores join this list once those routes exist; at
 * that point this is an append, not a redesign.
 */
export const PROJECT_NAV: AdminNavItem[] = [
  {
    label: 'Oversikt',
    icon: 'lucide:layout-dashboard',
    to: 'admin-projects-projectId',
  },
  {
    label: 'Superlag',
    icon: 'lucide:users',
    to: 'admin-projects-projectId-superteams',
    match: 'admin-projects-projectId-superteams',
  },
  {
    label: 'Innstillinger',
    icon: 'lucide:settings',
    to: 'admin-projects-projectId-edit',
    can: (p, ctx) => !!ctx.projectId && p.canEditProject(ctx.projectId),
  },
]

/** Entries the current user may see, in declaration order. */
export function visibleNavItems(
  items: AdminNavItem[],
  permissions: Permissions,
  context: AdminNavContext = {},
): AdminNavItem[] {
  return items.filter((item) => !item.can || item.can(permissions, context))
}

/**
 * Whether `item` should render as active for the given route name.
 * Prefix entries match themselves and anything nested beneath them; entries
 * without a `match` require an exact hit.
 */
export function isNavItemActive(
  item: AdminNavItem,
  routeName: string | undefined,
): boolean {
  if (!routeName) return false
  if (!item.match) return routeName === item.to
  return routeName === item.match || routeName.startsWith(`${item.match}-`)
}
