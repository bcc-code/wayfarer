import { describe, it, expect } from 'vitest'
import { computed } from 'vue'
import {
  GLOBAL_NAV,
  PROJECT_NAV,
  isNavItemActive,
  visibleNavItems,
  type AdminNavItem,
} from '../../layers/admin/app/utils/adminNav'

/**
 * The nav model is plain data plus two pure functions, so the rules that used
 * to be buried in `layouts/admin.vue` (permission gating and active state) are
 * testable without mounting a layout.
 */

type Perms = Parameters<typeof visibleNavItems>[1]

/** Minimal stand-in for usePermissions(), defaulting everything to denied. */
function permissions(granted: Partial<Record<string, boolean>> = {}): Perms {
  const flag = (key: string) => computed(() => granted[key] ?? false)
  return {
    canAccessProjects: flag('canAccessProjects'),
    canAccessTeams: flag('canAccessTeams'),
    canAccessUsers: flag('canAccessUsers'),
    canAccessScores: flag('canAccessScores'),
    canAccessConsents: flag('canAccessConsents'),
    canAccessFeedback: flag('canAccessFeedback'),
    canAccessMaintenance: flag('canAccessMaintenance'),
    canEditProject: (projectId: string) =>
      !!granted.canEditProject && !!projectId,
  } as unknown as Perms
}

const labels = (items: AdminNavItem[]) => items.map((i) => i.label)

describe('admin nav model', () => {
  describe('visibility', () => {
    it('shows only ungated entries when nothing is permitted', () => {
      expect(labels(visibleNavItems(GLOBAL_NAV, permissions()))).toEqual([
        'Hjem',
      ])
    })

    it('reveals an entry when its permission is granted', () => {
      const items = visibleNavItems(
        GLOBAL_NAV,
        permissions({ canAccessProjects: true }),
      )
      expect(labels(items)).toEqual(['Hjem', 'Prosjekter'])
    })

    it('shows everything to a fully permitted user, in declaration order', () => {
      const items = visibleNavItems(
        GLOBAL_NAV,
        permissions({
          canAccessProjects: true,
          canAccessUsers: true,
          canAccessConsents: true,
          canAccessFeedback: true,
          canAccessMaintenance: true,
        }),
      )
      // Teams and scores are not here: they are project-scoped now and live in
      // PROJECT_NAV.
      expect(labels(items)).toEqual([
        'Hjem',
        'Prosjekter',
        'Brukere',
        'Samtykker',
        'Tilbakemeldinger',
        'Vedlikehold',
      ])
    })

    it('passes the project context to project-scoped predicates', () => {
      const perms = permissions({ canEditProject: true })

      expect(labels(visibleNavItems(PROJECT_NAV, perms, {}))).not.toContain(
        'Innstillinger',
      )
      expect(
        labels(visibleNavItems(PROJECT_NAV, perms, { projectId: 'PR01' })),
      ).toContain('Innstillinger')
    })

    it('hides project settings from someone who cannot edit the project', () => {
      const items = visibleNavItems(PROJECT_NAV, permissions(), {
        projectId: 'PR01',
      })
      expect(labels(items)).not.toContain('Innstillinger')
    })
  })

  describe('active state', () => {
    const find = (label: string) =>
      [...GLOBAL_NAV, ...PROJECT_NAV].find((i) => i.label === label)!

    it('matches a prefix entry on itself and on nested routes', () => {
      const projects = find('Prosjekter')
      expect(isNavItemActive(projects, 'admin-projects')).toBe(true)
      expect(isNavItemActive(projects, 'admin-projects-projectId')).toBe(true)
      expect(
        isNavItemActive(projects, 'admin-projects-projectId-challenges-new'),
      ).toBe(true)
    })

    // The old check was `route.fullPath.includes('/teams')`, which would light
    // the project "Lag" entry on the legacy global /admin/teams route as well.
    // Name-prefix matching keeps the two branches apart, which is what made the
    // move safe.
    it('does not match a different branch that merely shares a word', () => {
      const teams = find('Lag')
      expect(isNavItemActive(teams, 'admin-projects-projectId-teams')).toBe(
        true,
      )
      expect(
        isNavItemActive(teams, 'admin-projects-projectId-teams-teamId'),
      ).toBe(true)
      // The legacy global route, which now only redirects.
      expect(isNavItemActive(teams, 'admin-teams')).toBe(false)
    })

    it('does not treat a longer sibling name as nested', () => {
      const teams = find('Lag')
      expect(isNavItemActive(teams, 'admin-teamsomething')).toBe(false)
    })

    it('requires an exact match for entries without a prefix', () => {
      const home = find('Hjem')
      expect(isNavItemActive(home, 'admin')).toBe(true)
      expect(isNavItemActive(home, 'admin-projects')).toBe(false)
    })

    it('is inactive when the route has no name', () => {
      expect(isNavItemActive(find('Hjem'), undefined)).toBe(false)
    })
  })

  // The route-manifest test asserts these resolve; this one guards the model's
  // own shape so a typo cannot silently produce a nav entry that links nowhere.
  it('every entry targets a route name, never a path', () => {
    for (const item of [...GLOBAL_NAV, ...PROJECT_NAV]) {
      expect(item.to).not.toMatch(/^\//)
      expect(item.to).toMatch(/^[a-z][a-zA-Z0-9-]*$/)
    }
  })
})
