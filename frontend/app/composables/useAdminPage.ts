import type { BreadcrumbItem } from '@nuxt/ui'
import type { RouteLocationRaw } from 'vue-router'
import { GLOBAL_NAV, PROJECT_NAV, isNavItemActive } from '~/utils/adminNav'

/**
 * The page's place in the admin chrome: its breadcrumb and navbar title.
 *
 * Everything structural is derived from the route and the nav model, so a page
 * only supplies what cannot be known from the URL — the name of the thing it is
 * showing. `Prosjekter → <project> → Utfordringer` comes for free;
 * `→ Bibelquiz` is the page's to give.
 *
 * Replaces ~25 hand-written `<UBreadcrumb :items="[...]">` blocks, each wrapped
 * in an identical bordered `div`, which drifted from one another and rendered
 * inside the page body rather than in the shell.
 */

/**
 * Module-level so the layout can read what the page set. The layout renders the
 * chrome and is an *ancestor* of the page, so provide/inject cannot carry this
 * upward.
 */
const pageLabel = ref<string | null>(null)

export function useAdminPage(label?: MaybeRefOrGetter<string | undefined>) {
  // Only a caller that passes something is a setter; the layout calls this bare
  // to read.
  if (label !== undefined) {
    const source = toRef(label)
    watchEffect(() => {
      pageLabel.value = source.value ?? null
    })
    // Without this the previous page's name lingers on the next route until its
    // own query resolves.
    onScopeDispose(() => {
      pageLabel.value = null
    })
  }

  const route = useRoute()
  const { projectId, project } = useCurrentProject()

  const projectLocation = (name: string): RouteLocationRaw =>
    ({ name, params: { projectId: projectId.value } }) as RouteLocationRaw

  const breadcrumb = computed<BreadcrumbItem[]>(() => {
    const routeName = String(route.name ?? '')
    const items: BreadcrumbItem[] = []

    // Only prefix-matching entries can be ancestors; an exact-match entry like
    // "Hjem" is a destination, not a parent.
    const global = GLOBAL_NAV.find(
      (item) => item.match && isNavItemActive(item, routeName),
    )
    if (global) {
      // Same reason as `projectLocation`: the name is data here, so TypeScript
      // cannot correlate it with that route's own params.
      items.push({
        label: global.label,
        to: { name: global.to } as RouteLocationRaw,
      })
    }

    if (projectId.value) {
      items.push({
        label: project.value?.name ?? projectId.value,
        to: projectLocation('admin-projects-projectId'),
      })

      // The overview is the project crumb itself, so it never appears twice.
      const section = PROJECT_NAV.find(
        (item) =>
          item.to !== 'admin-projects-projectId' &&
          isNavItemActive(item, routeName),
      )
      if (section) {
        items.push({ label: section.label, to: projectLocation(section.to) })
      }
    }

    if (pageLabel.value) items.push({ label: pageLabel.value })

    // The last crumb is where you already are.
    const last = items.at(-1)
    if (last) delete last.to

    return items
  })

  /** Navbar heading: the page's own label when it has one, else its section. */
  const title = computed(() => {
    if (pageLabel.value) return pageLabel.value
    const routeName = String(route.name ?? '')
    const match =
      [...PROJECT_NAV].reverse().find((i) => isNavItemActive(i, routeName)) ??
      [...GLOBAL_NAV].reverse().find((i) => isNavItemActive(i, routeName))
    return match?.label ?? 'Admin'
  })

  return { breadcrumb, title }
}
