// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { useFirestoreRefresh } from '~/composables/useFirestoreRefresh'

// The wiring of useFirestoreRefresh: the 'firestore-update' events, auth and
// unmounting. The coalescing itself is covered in
// test/unit/coalescedRefresh.test.ts.

const authReady = ref(true)
mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: authReady }))

const JITTER_MS = 50
const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

function ping(query = 'ChallengePageDocument') {
  window.dispatchEvent(
    new CustomEvent('firestore-update', { detail: { query } }),
  )
}

// Stands in for a page or the layout: it stays mounted while auth changes
function mountListener(latencyMs = 0) {
  const refetch = vi.fn(() => sleep(latencyMs))
  const Listener = defineComponent({
    setup() {
      useFirestoreRefresh(['ChallengePageDocument'], refetch, {
        minIntervalMs: 0,
        jitterMs: JITTER_MS,
      })
      return () => h('div')
    },
  })
  return { refetch, mounted: mountSuspended(Listener) }
}

describe('useFirestoreRefresh', () => {
  beforeEach(() => {
    authReady.value = true
  })

  it('refetches on a ping for its query and ignores other queries', async () => {
    const { refetch, mounted } = mountListener()
    const wrapper = await mounted

    ping('ProfilePageDocument')
    await sleep(JITTER_MS * 2)
    expect(refetch).not.toHaveBeenCalled()

    ping()
    await sleep(JITTER_MS * 2)
    expect(refetch).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  it('drops a scheduled refetch on logout while the component stays mounted', async () => {
    const { refetch, mounted } = mountListener()
    const wrapper = await mounted

    ping()
    authReady.value = false // logout before the jitter elapsed
    await sleep(JITTER_MS * 2)
    expect(refetch).not.toHaveBeenCalled()

    ping() // a ping that still arrives while logged out
    await sleep(JITTER_MS * 2)
    expect(refetch).not.toHaveBeenCalled()

    authReady.value = true // logged in again
    ping()
    await sleep(JITTER_MS * 2)
    expect(refetch).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  it('drops the queued trailing refetch on logout during a refetch', async () => {
    const { refetch, mounted } = mountListener(JITTER_MS * 2)
    const wrapper = await mounted

    ping()
    await sleep(JITTER_MS * 1.5) // the refetch is in flight
    ping() // queued behind it
    authReady.value = false
    await sleep(JITTER_MS * 6)

    expect(refetch).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  it('stops listening and drops a scheduled refetch on unmount', async () => {
    const { refetch, mounted } = mountListener()
    const wrapper = await mounted

    ping()
    wrapper.unmount()
    await sleep(JITTER_MS * 2)
    ping()
    await sleep(JITTER_MS * 2)

    expect(refetch).not.toHaveBeenCalled()
  })
})
