/**
 * The project the current route is inside.
 *
 * Owned by a composable rather than provided by `pages/admin/projects/[projectId].vue`
 * because the consumers that need it most — the sidebar, the project switcher,
 * the navbar title — are rendered by the *layout*, which is an ancestor of the
 * page. Anything the page provided would be invisible to them.
 *
 * urql's document cache is keyed by operation + variables, so every caller
 * shares one result regardless of where it sits in the tree.
 */
gql(`
  query AdminProjectShell($projectId: ID!) {
    project(id: $projectId) {
      id
      name
      description
      startDate
      endDate
      branding {
        logoImage {
          ...ImageFields
        }
        rounding
        colors {
          light {
            accent
          }
          dark {
            accent
          }
        }
      }
    }
  }
`)

export function useCurrentProject() {
  const route = useRoute()
  const { isAuthReady } = useAuthReady()

  // `route.params` is a union across every route under typedPages.
  const projectId = computed(() =>
    'projectId' in route.params ? route.params.projectId : undefined,
  )

  const { data, fetching, error, executeQuery } = useAdminProjectShellQuery({
    // Must be a computed. The parent route does not remount when the projectId
    // changes, so a plain object would pin the variables to whichever project
    // happened to be open first and silently keep querying it.
    variables: computed(() => ({ projectId: projectId.value as string })),
    pause: computed(() => !isAuthReady.value || !projectId.value),
    // The client default is cache-and-network. With the page, the sidebar, the
    // switcher and the breadcrumb all subscribed, that fires several identical
    // requests on first paint; mutations refresh explicitly instead.
    requestPolicy: 'cache-first',
  })

  return {
    projectId,
    project: computed(() => data.value?.project ?? null),
    fetching,
    error,
    /** Call after a mutation that changes the project itself. */
    refresh: () => executeQuery({ requestPolicy: 'network-only' }),
  }
}
