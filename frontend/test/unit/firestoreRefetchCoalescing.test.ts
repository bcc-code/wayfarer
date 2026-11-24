import { describe, it, expect, afterEach, vi } from 'vitest'
import { computed, effectScope, ref, type EffectScope, type Ref } from 'vue'
import { Client, cacheExchange, fetchExchange, provideClient } from '@urql/vue'
import { useChallengePageQuery } from '~/api/generated'
import {
  createCoalescedRefresh,
  type FirestoreRefreshOptions,
} from '~/composables/useFirestoreRefresh'

// A Firestore ping makes ChallengePage refetch with
// `refresh({ requestPolicy: 'network-only' })` (pages/challenges/[challengeId].vue),
// through useFirestoreRefresh. During a coordinated session start, pings
// arrive faster than a refetch takes. Calling the urql refetch on every ping
// aborted the one in flight and sent a new request. These tests run the
// app's urql setup with the pings going through createCoalescedRefresh: at
// most one request in flight and one trailing request, and no request
// aborted (an aborted request still costs the server a full request).

const LATENCY_MS = 100
// The interval and jitter themselves are covered in coalescedRefresh.test.ts;
// here they are off unless a test needs them, to keep the timing readable
const IMMEDIATE: FirestoreRefreshOptions = { minIntervalMs: 0, jitterMs: 0 }

type Outcome = 'ok' | 'graphqlError' | 'networkError'

interface FetchCall {
  aborted: boolean
  settled: boolean
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

function slowFetch(outcome: () => Outcome = () => 'ok') {
  const calls: FetchCall[] = []
  const fetch = (_url: RequestInfo | URL, init?: RequestInit) => {
    const call: FetchCall = { aborted: false, settled: false }
    calls.push(call)
    return new Promise<Response>((resolve, reject) => {
      const timer = setTimeout(() => {
        call.settled = true
        const o = outcome()
        if (o === 'networkError') {
          reject(new TypeError('Failed to fetch'))
          return
        }
        const body =
          o === 'graphqlError'
            ? { data: null, errors: [{ message: 'boom' }] }
            : { data: { challenge: null } }
        resolve(
          new Response(JSON.stringify(body), {
            headers: { 'content-type': 'application/json' },
          }),
        )
      }, LATENCY_MS)
      // urql also aborts the controller of a request that already settled
      // when it tears the operation down; only an abort before the response
      // cancels a request.
      init?.signal?.addEventListener('abort', () => {
        if (call.settled) return
        call.aborted = true
        clearTimeout(timer)
        reject(new DOMException('The operation was aborted.', 'AbortError'))
      })
    })
  }
  return { fetch, calls }
}

// Mirrors app/plugins/0.urql.ts minus the auth exchange, which only adds a
// header here.
function appClient(fetch: typeof globalThis.fetch) {
  return new Client({
    url: 'http://test.invalid/graphql',
    preferGetMethod: false,
    requestPolicy: 'cache-and-network',
    exchanges: [cacheExchange, fetchExchange],
    fetch,
  })
}

function summary(calls: FetchCall[]) {
  return {
    requests: calls.length,
    aborted: calls.filter((c) => c.aborted).length,
    completed: calls.filter((c) => c.settled).length,
  }
}

describe('Firestore ping refetch on ChallengePage', () => {
  let scope: EffectScope | undefined

  afterEach(() => {
    scope?.stop()
    scope = undefined
    vi.restoreAllMocks()
  })

  function mountChallengePage(
    fetch: typeof globalThis.fetch,
    options: FirestoreRefreshOptions = IMMEDIATE,
    challengeId: Ref<string> = ref('CL01TESTCHALLENGE0000000000'),
  ) {
    scope = effectScope()
    return scope.run(() => {
      provideClient(appClient(fetch))
      const { executeQuery: refresh } = useChallengePageQuery({
        // computed, like the pages with reactive variables (pages/index.vue)
        variables: computed(() => ({ challengeId: challengeId.value })),
        requestPolicy: 'network-only',
      })
      // What useFirestoreRefresh wires to the 'firestore-update' event
      const pings = createCoalescedRefresh(
        () => refresh({ requestPolicy: 'network-only' }),
        options,
      )
      return {
        onPing: () => pings.notify(),
        externalRefresh: () => refresh({ requestPolicy: 'network-only' }),
        challengeId,
      }
    })!
  }

  it('sends one request per ping when pings are slower than a refetch', async () => {
    const { fetch, calls } = slowFetch()
    const { onPing } = mountChallengePage(fetch)
    await sleep(LATENCY_MS * 2) // initial load settles

    for (let i = 0; i < 3; i++) {
      onPing()
      await sleep(LATENCY_MS * 2)
    }

    const pingRequests = calls.slice(1)
    expect(pingRequests).toHaveLength(3)
    expect(pingRequests.filter((c) => c.aborted)).toHaveLength(0)
  })

  it('coalesces pings that arrive while a refetch is in flight', async () => {
    const { fetch, calls } = slowFetch()
    const { onPing } = mountChallengePage(fetch)
    await sleep(LATENCY_MS * 2) // initial load settles

    // 5 pings within one refetch latency: ~20 pings/s at the 19:00 peak.
    for (let i = 0; i < 5; i++) {
      onPing()
      await sleep(LATENCY_MS / 5)
    }
    await sleep(LATENCY_MS * 4)

    const s = summary(calls.slice(1))
    expect(s.requests, JSON.stringify(s)).toBeLessThanOrEqual(2)
    expect(s.aborted, JSON.stringify(s)).toBe(0)
  })

  it.each(['graphqlError', 'networkError'] as const)(
    'refetches again on the next ping after a %s',
    async (failure) => {
      vi.spyOn(console, 'error').mockImplementation(() => {})
      let outcome: Outcome = 'ok'
      const { fetch, calls } = slowFetch(() => outcome)
      const { onPing } = mountChallengePage(fetch)
      await sleep(LATENCY_MS * 2)

      outcome = failure
      onPing()
      await sleep(LATENCY_MS * 2)
      outcome = 'ok'
      onPing()
      await sleep(LATENCY_MS * 2)

      const s = summary(calls.slice(1))
      expect(s, JSON.stringify(s)).toEqual({
        requests: 2,
        aborted: 0,
        completed: 2,
      })
    },
  )

  // Limitation (urql semantics): a refetch of the same query started
  // elsewhere replaces the one in flight, so it aborts ours. That can't be
  // prevented without a global coordinator. What the coalescing guarantees:
  // our queued refetch waits minIntervalMs after ours started, so it doesn't
  // abort the external request in turn (no cancellation loop), as long as
  // that request finishes within minIntervalMs.
  it('an external refresh during a refetch causes no cancellation loop', async () => {
    const { fetch, calls } = slowFetch()
    const { onPing, externalRefresh } = mountChallengePage(fetch, {
      minIntervalMs: LATENCY_MS * 3,
      jitterMs: 0,
    })
    await sleep(LATENCY_MS * 2)

    onPing() // our refetch starts
    await sleep(LATENCY_MS / 4)
    externalRefresh() // e.g. the enroll flow: replaces (aborts) ours
    onPing() // arrives while ours counted as in flight: queued
    await sleep(LATENCY_MS * 8)

    const pingRequests = calls.slice(1)
    const s = summary(pingRequests)
    expect(s, JSON.stringify(s)).toEqual({
      requests: 3, // ours, the external one, our one trailing refetch
      aborted: 1, // only ours, by the external refresh
      completed: 2,
    })
    expect(pingRequests[1]!.settled, 'the external refresh completed').toBe(
      true,
    )
  })

  it('a reactive variable change during a refetch adds no trailing refetch', async () => {
    const { fetch, calls } = slowFetch()
    const { onPing, challengeId } = mountChallengePage(fetch)
    await sleep(LATENCY_MS * 2)

    onPing()
    await sleep(LATENCY_MS / 4)
    challengeId.value = 'CL01OTHERCHALLENGE000000000' // urql re-runs the query itself
    await sleep(LATENCY_MS * 4)

    const s = summary(calls.slice(1))
    // ours (replaced by urql for the new variables) and the new query; no
    // further request without another ping
    expect(s, JSON.stringify(s)).toEqual({
      requests: 2,
      aborted: 1,
      completed: 1,
    })
  })
})
