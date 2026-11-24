export interface FirestoreRefreshOptions {
  /** Minimum time between the starts of two refetches, in ms */
  minIntervalMs?: number
  /**
   * Random delay (0..jitterMs) before a refetch, so the clients that get the
   * same ping don't all hit the server in the same millisecond
   */
  jitterMs?: number
  /** Checked when a refetch is about to run; false drops it */
  enabled?: () => boolean
}

/**
 * Coalesces Firestore pings into refetches: at most one refetch in flight and
 * at most one more queued, never closer together than minIntervalMs.
 *
 * A Firestore ping only says "this data changed", so pings can be merged.
 * Calling the urql refetch on every ping instead aborted the refetch still in
 * flight and sent a new one: during a burst of session pings every client
 * sent one request per ping, most of them cancelled (R4,
 * firestoreRefetchCoalescing.test.ts).
 *
 * refetch should return the urql executeQuery result: awaiting it waits for
 * the request (urql never rejects; errors end up in the query's state). A
 * refetch of the same query started elsewhere (another caller, or reactive
 * variables changing) replaces ours, and ours then resolves early. The queued
 * refetch still waits minIntervalMs after ours started, so it can only cancel
 * such a request if that one takes longer than that.
 */
export function createCoalescedRefresh(
  refetch: () => unknown,
  {
    minIntervalMs = 2000,
    jitterMs = 500,
    enabled = () => true,
  }: FirestoreRefreshOptions = {},
) {
  let inFlight = false
  let queued = false // a ping arrived while a refetch was in flight
  let timer: ReturnType<typeof setTimeout> | undefined
  let lastStart = -Infinity
  let disposed = false

  function schedule() {
    if (disposed || timer !== undefined) return
    const wait = Math.max(
      Math.random() * jitterMs,
      lastStart + minIntervalMs - Date.now(),
    )
    timer = setTimeout(run, wait)
  }

  async function run() {
    timer = undefined
    if (disposed || !enabled()) return
    inFlight = true
    lastStart = Date.now()
    try {
      await refetch()
    } catch (error) {
      console.error('Firestore refetch failed:', error)
    } finally {
      inFlight = false
    }
    if (queued) {
      queued = false
      schedule()
    }
  }

  /** Drop any scheduled or queued refetch (e.g. on logout) */
  function cancel() {
    clearTimeout(timer)
    timer = undefined
    queued = false
  }

  return {
    notify() {
      if (disposed) return
      if (inFlight) {
        queued = true
        return
      }
      schedule()
    },
    cancel,
    dispose() {
      disposed = true
      cancel()
    },
  }
}

/**
 * Hook for pages to respond to Firestore updates
 * @param queryNames - Array of query document names to listen for
 * @param callback - Refetches the data; return the urql executeQuery result so
 *   that pings arriving during the refetch are merged into one more refetch
 * @param options - See FirestoreRefreshOptions
 */
export function useFirestoreRefresh(
  queryNames: string[],
  callback: () => unknown,
  options: FirestoreRefreshOptions = {},
) {
  // Refetches only run while authenticated: on logout the token is cleared
  // and anything scheduled is dropped, also where the component (e.g. the
  // layout) stays mounted
  const { isAuthReady } = useAuthReady()
  const refresh = createCoalescedRefresh(callback, {
    ...options,
    enabled: () => isAuthReady.value && (options.enabled?.() ?? true),
  })
  watch(isAuthReady, (ready) => {
    if (!ready) refresh.cancel()
  })

  const handleUpdate = (event: Event) => {
    const customEvent = event as CustomEvent<{ query: string }>
    if (queryNames.includes(customEvent.detail.query)) {
      refresh.notify()
    }
  }

  onMounted(() => {
    window.addEventListener('firestore-update', handleUpdate)
  })

  onUnmounted(() => {
    window.removeEventListener('firestore-update', handleUpdate)
    refresh.dispose()
  })
}
