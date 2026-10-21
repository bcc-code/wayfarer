import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createCoalescedRefresh } from '~/composables/useFirestoreRefresh'

// A refetch that takes `latency` ms and records when each one started and
// how many ran at once
function fakeRefetch(
  latency: number,
  outcome: () => 'ok' | 'reject' = () => 'ok',
) {
  const starts: number[] = []
  let running = 0
  let maxRunning = 0
  const refetch = vi.fn(() => {
    starts.push(Date.now())
    running++
    maxRunning = Math.max(maxRunning, running)
    return new Promise<void>((resolve, reject) =>
      setTimeout(() => {
        running--
        if (outcome() === 'reject') reject(new Error('network down'))
        else resolve()
      }, latency),
    )
  })
  return { refetch, starts, maxRunning: () => maxRunning }
}

describe('createCoalescedRefresh', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-10-21T12:00:00'))
    vi.spyOn(Math, 'random').mockReturnValue(0.5)
  })
  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('refetches once for a single ping, after the jitter', async () => {
    const { refetch } = fakeRefetch(100)
    const r = createCoalescedRefresh(refetch, {
      jitterMs: 500,
      minIntervalMs: 2000,
    })

    r.notify()
    await vi.advanceTimersByTimeAsync(249)
    expect(refetch).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1)
    expect(refetch).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(5000)
    expect(refetch).toHaveBeenCalledTimes(1)
  })

  it('keeps the jitter within jitterMs', async () => {
    vi.mocked(Math.random).mockReturnValue(0.999)
    const { refetch } = fakeRefetch(100)
    const r = createCoalescedRefresh(refetch, { jitterMs: 500 })

    r.notify()
    await vi.advanceTimersByTimeAsync(500)
    expect(refetch).toHaveBeenCalledTimes(1)
  })

  it('merges pings before the refetch starts into that refetch', async () => {
    const { refetch } = fakeRefetch(100)
    const r = createCoalescedRefresh(refetch, { jitterMs: 500 })

    r.notify()
    await vi.advanceTimersByTimeAsync(100)
    r.notify()
    r.notify()
    await vi.advanceTimersByTimeAsync(5000)
    expect(refetch).toHaveBeenCalledTimes(1)
  })

  it('merges pings during a refetch into exactly one trailing refetch, started after them', async () => {
    const { refetch, starts } = fakeRefetch(1000)
    const r = createCoalescedRefresh(refetch, {
      jitterMs: 0,
      minIntervalMs: 2000,
    })

    r.notify()
    await vi.advanceTimersByTimeAsync(0) // first refetch starts
    let lastPing = 0
    for (let i = 0; i < 5; i++) {
      await vi.advanceTimersByTimeAsync(150)
      r.notify()
      lastPing = Date.now()
    }
    await vi.advanceTimersByTimeAsync(10_000)

    expect(starts).toHaveLength(2)
    expect(starts[1]! - starts[0]!).toBeGreaterThanOrEqual(2000)
    expect(starts[1]!).toBeGreaterThan(lastPing) // the data reflects the last ping
  })

  it('never overlaps refetches and stays bounded by minIntervalMs under a sustained burst', async () => {
    const { refetch, starts, maxRunning } = fakeRefetch(300)
    const r = createCoalescedRefresh(refetch) // defaults: 2000 ms / 500 ms

    // 20 pings per second for 10 seconds, like the 19:00 session storm
    for (let t = 0; t < 10_000; t += 50) {
      r.notify()
      await vi.advanceTimersByTimeAsync(50)
    }
    await vi.advanceTimersByTimeAsync(5000)

    expect(maxRunning()).toBe(1)
    expect(starts.length).toBeGreaterThanOrEqual(5)
    expect(starts.length).toBeLessThanOrEqual(10_000 / 2000 + 1)
    for (let i = 1; i < starts.length; i++) {
      expect(starts[i]! - starts[i - 1]!).toBeGreaterThanOrEqual(2000)
    }
  })

  it('keeps working after a refetch rejects (network error)', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    let fail = true
    const { refetch } = fakeRefetch(100, () => (fail ? 'reject' : 'ok'))
    const r = createCoalescedRefresh(refetch, { jitterMs: 0, minIntervalMs: 0 })

    r.notify()
    await vi.advanceTimersByTimeAsync(200)
    fail = false
    r.notify()
    await vi.advanceTimersByTimeAsync(200)

    expect(refetch).toHaveBeenCalledTimes(2)
  })

  it('keeps working after a refetch throws synchronously', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const refetch = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new Error('boom')
      })
      .mockImplementation(() => Promise.resolve())
    const r = createCoalescedRefresh(refetch, { jitterMs: 0, minIntervalMs: 0 })

    r.notify()
    await vi.advanceTimersByTimeAsync(10)
    r.notify()
    await vi.advanceTimersByTimeAsync(10)

    expect(refetch).toHaveBeenCalledTimes(2)
  })

  it('treats a refetch that returns nothing as done', async () => {
    const refetch = vi.fn(() => undefined) // e.g. the Completed tab is not open
    const r = createCoalescedRefresh(refetch, { jitterMs: 0, minIntervalMs: 0 })

    r.notify()
    await vi.advanceTimersByTimeAsync(10)
    r.notify()
    await vi.advanceTimersByTimeAsync(10)

    expect(refetch).toHaveBeenCalledTimes(2)
  })

  it('dispose drops a scheduled refetch', async () => {
    const { refetch } = fakeRefetch(100)
    const r = createCoalescedRefresh(refetch, { jitterMs: 500 })

    r.notify()
    r.dispose()
    await vi.advanceTimersByTimeAsync(5000)
    r.notify()
    await vi.advanceTimersByTimeAsync(5000)

    expect(refetch).not.toHaveBeenCalled()
  })

  it('dispose during a refetch drops the queued trailing refetch', async () => {
    const { refetch } = fakeRefetch(1000)
    const r = createCoalescedRefresh(refetch, { jitterMs: 0, minIntervalMs: 0 })

    r.notify()
    await vi.advanceTimersByTimeAsync(0)
    r.notify() // queued behind the one in flight
    r.dispose()
    await vi.advanceTimersByTimeAsync(5000)

    expect(refetch).toHaveBeenCalledTimes(1)
  })

  it('cancel (logout) drops scheduled and queued refetches, later pings work again', async () => {
    const { refetch } = fakeRefetch(1000)
    const r = createCoalescedRefresh(refetch, { jitterMs: 0, minIntervalMs: 0 })

    r.notify()
    await vi.advanceTimersByTimeAsync(0)
    r.notify() // queued
    r.cancel()
    await vi.advanceTimersByTimeAsync(5000)
    expect(refetch).toHaveBeenCalledTimes(1)

    r.notify()
    await vi.advanceTimersByTimeAsync(10)
    expect(refetch).toHaveBeenCalledTimes(2)
  })

  it('drops a refetch when enabled() is false at the time it would run', async () => {
    let authed = true
    const { refetch } = fakeRefetch(100)
    const r = createCoalescedRefresh(refetch, {
      jitterMs: 500,
      enabled: () => authed,
    })

    r.notify()
    authed = false // logged out while the refetch waited for its jitter
    await vi.advanceTimersByTimeAsync(5000)

    expect(refetch).not.toHaveBeenCalled()
  })
})
