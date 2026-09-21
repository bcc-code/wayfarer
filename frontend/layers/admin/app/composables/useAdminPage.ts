import type { BreadcrumbItem } from '@nuxt/ui'
import type { RouteLocationRaw } from 'vue-router'
import { GLOBAL_NAV, PROJECT_NAV, isNavItemActive } from '../utils/adminNav'

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
 * What a page contributes to the trail. A bare string is the common case — the
 * name of the thing being shown. Routes nested below a detail page need two,
 * and the first of those links back: a quiz belongs to a challenge.
 */
export type AdminPageCrumb = string | { label: string; to?: RouteLocationRaw }

/**
 * Module-level so the layout can read what the page set. The layout renders the
 * chrome and is an *ancestor* of the page, so provide/inject cannot carry this
 * upward.
 */
// shallowRef, not ref: `BreadcrumbItem['to']` is the full typed-routes union,
// and deep-unwrapping it blows TypeScript's instantiation depth.
const pageCrumbs = shallowRef<BreadcrumbItem[]>([])

function normalise(
  value: AdminPageCrumb | AdminPageCrumb[] | undefined,
): BreadcrumbItem[] {
  if (!value) return []
  const list = Array.isArray(value) ? value : [value]
  return list
    .filter((crumb) => (typeof crumb === 'string' ? crumb : crumb.label))
    .map((crumb) =>
      typeof crumb === 'string' ? { label: crumb } : { ...crumb },
    )
}

export function useAdminPage(
  label?: MaybeRefOrGetter<AdminPageCrumb | AdminPageCrumb[] | undefined>,
) {
  // Only a caller that passes something is a setter; the layout calls this bare
  // to read.
  if (label !== undefined) {
    const source = toRef(label)
    /**
     * `flush: 'post'` so the getter is never called during the caller's setup.
     *
     * A plain `watchEffect` runs its effect **synchronously on creation**, which
     * evaluates the page's getter mid-setup. Pages pass
     * `() => data.value?.user.name`, and `data` is very often declared *below*
     * this call — the query it comes from needs an id the lines above compute.
     * Reading it then hits the temporal dead zone and throws
     * `can't access lexical declaration 'data' before initialization`, which
     * Nuxt renders as a **500 page**, not a blank breadcrumb.
     *
     * Two pages did exactly that (`users/[userId]/index.vue` and
     * `users/[userId]/achievements.vue`), so every user detail view was a 500.
     * Deferring the first run to after render makes call order irrelevant, so
     * the next page to be written this way cannot reintroduce it. The cost is
     * that the crumb appears one tick late, which is invisible — the data it
     * names has not loaded yet either.
     */
    watchEffect(
      () => {
        pageCrumbs.value = normalise(source.value)
      },
      { flush: 'post' },
    )
    // Without this the previous page's name lingers on the next route until its
    // own query resolves.
    onScopeDispose(() => {
      pageCrumbs.value = []
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

    items.push(...pageCrumbs.value)

    // The last crumb is where you already are.
    const last = items.at(-1)
    if (last) delete last.to

    return items
  })

  /** Navbar heading: the page's own label when it has one, else its section. */
  const title = computed(() => {
    const own = pageCrumbs.value.at(-1)?.label
    if (own) return own
    const routeName = String(route.name ?? '')
    const match =
      [...PROJECT_NAV].reverse().find((i) => isNavItemActive(i, routeName)) ??
      [...GLOBAL_NAV].reverse().find((i) => isNavItemActive(i, routeName))
    return match?.label ?? 'Admin'
  })

  return { breadcrumb, title }
}
