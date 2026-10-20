import { describe, it, expect, afterEach } from 'vitest'
import { effectScope, type EffectScope } from 'vue'
import { Client, cacheExchange, fetchExchange, provideClient } from '@urql/vue'
import { useChallengePageQuery } from '~/api/generated'

// A Firestore ping makes ChallengePage call
// `refresh({ requestPolicy: 'network-only' })` (pages/challenges/[challengeId].vue).
// During a coordinated session start, pings arrive faster than a refetch
// takes. These tests pin the rule the multi-session load test relies on
// (notes/multi-session-loadtest-plan.md §4.3): pings that arrive while a
// refetch is in flight coalesce — at most one request in flight and one
// trailing request — and no request is aborted (an aborted request still
// costs the server a full request).

const LATENCY_MS = 100

interface FetchCall {
  aborted: boolean
  settled: boolean
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

function slowFetch() {
  const calls: FetchCall[] = []
  const fetch = (_url: RequestInfo | URL, init?: RequestInit) => {
    const call: FetchCall = { aborted: false, settled: false }
    calls.push(call)
    return new Promise<Response>((resolve, reject) => {
      const timer = setTimeout(() => {
        call.settled = true
        resolve(
          new Response(JSON.stringify({ data: { challenge: null } }), {
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

describe('Firestore ping refetch on ChallengePage', () => {
  let scope: EffectScope | undefined

  afterEach(() => {
    scope?.stop()
    scope = undefined
  })

  function mountChallengePage(fetch: typeof globalThis.fetch) {
    scope = effectScope()
    return scope.run(() => {
      provideClient(appClient(fetch))
      const { executeQuery: refresh } = useChallengePageQuery({
        variables: { challengeId: 'CL01TESTCHALLENGE0000000000' },
        requestPolicy: 'network-only',
      })
      // The page's useFirestoreRefresh callback.
      return () => refresh({ requestPolicy: 'network-only' })
    })!
  }

  it('sends one request per ping when pings are slower than a refetch', async () => {
    const { fetch, calls } = slowFetch()
    const onPing = mountChallengePage(fetch)
    await sleep(LATENCY_MS * 2) // initial load settles

    for (let i = 0; i < 3; i++) {
      onPing()
      await sleep(LATENCY_MS * 2)
    }

    const pingRequests = calls.slice(1)
    expect(pingRequests).toHaveLength(3)
    expect(pingRequests.filter((c) => c.aborted)).toHaveLength(0)
  })

  // TODO(R4, notes/multi-session-loadtest-plan.md): expected failure — today
  // every ping aborts the in-flight refetch and sends a new request
  // (5 pings → 5 requests, 4 aborted). Once useFirestoreRefresh coalesces
  // refetches, change `it.fails` back to `it` so this guards the fix.
  it.fails(
    'coalesces pings that arrive while a refetch is in flight',
    async () => {
      const { fetch, calls } = slowFetch()
      const onPing = mountChallengePage(fetch)
      await sleep(LATENCY_MS * 2) // initial load settles

      // 5 pings within one refetch latency: ~20 pings/s at the 19:00 peak.
      for (let i = 0; i < 5; i++) {
        onPing()
        await sleep(LATENCY_MS / 5)
      }
      await sleep(LATENCY_MS * 4)

      const pingRequests = calls.slice(1)
      const summary = {
        requests: pingRequests.length,
        aborted: pingRequests.filter((c) => c.aborted).length,
        completed: pingRequests.filter((c) => c.settled).length,
      }
      expect(summary.requests, JSON.stringify(summary)).toBeLessThanOrEqual(2)
      expect(summary.aborted, JSON.stringify(summary)).toBe(0)
    },
  )
})
