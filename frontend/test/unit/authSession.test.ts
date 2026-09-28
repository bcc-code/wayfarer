import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

/**
 * utils/authSession runs in the browser; in the node test environment
 * localStorage, fetch and navigator.locks are stubbed. The module keeps
 * per-tab state (in-flight refresh, failure cooldown), so each test loads a
 * fresh copy.
 */

type AuthSession = typeof import('../../app/utils/authSession')

function makeToken(payload: Record<string, unknown>): string {
  const encode = (obj: Record<string, unknown>) =>
    Buffer.from(JSON.stringify(obj))
      .toString('base64')
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '')
  return `${encode({ alg: 'HS256', typ: 'JWT' })}.${encode(payload)}.signature`
}

const nowSeconds = () => Math.floor(Date.now() / 1000)
const DAY = 24 * 60 * 60

function tokenIssuedAgo(seconds: number, ttl = 7 * DAY) {
  const iat = nowSeconds() - seconds
  return makeToken({ iat, exp: iat + ttl })
}

let store: Map<string, string>
const fetchMock = vi.fn()

function jsonResponse(status: number, body: unknown = {}) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

function pair(access: string, refresh: string) {
  return {
    access_token: access,
    access_expires_at: '2024-06-22T12:00:00Z',
    refresh_token: refresh,
    refresh_expires_at: '2024-12-15T12:00:00Z',
  }
}

async function load(): Promise<AuthSession> {
  vi.resetModules()
  return import('../../app/utils/authSession')
}

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2024-06-15T12:00:00'))
  store = new Map()
  fetchMock.mockReset()
  vi.stubGlobal('localStorage', {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => store.set(k, v),
    removeItem: (k: string) => store.delete(k),
  })
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('authBaseUrl', () => {
  it('derives the auth base from the GraphQL URL', async () => {
    const { authBaseUrl } = await load()
    expect(authBaseUrl({ apiUrl: 'https://api.test/graphql' })).toBe(
      'https://api.test/auth',
    )
    expect(authBaseUrl({ apiUrl: 'https://api.test/graphql/' })).toBe(
      'https://api.test/auth',
    )
  })

  it('prefers an explicit authUrl', async () => {
    const { authBaseUrl } = await load()
    expect(
      authBaseUrl({
        apiUrl: 'https://api.test/graphql',
        authUrl: 'https://auth.test/auth/',
      }),
    ).toBe('https://auth.test/auth')
  })
})

describe('shouldRefreshAccessToken', () => {
  it('is false for a fresh token', async () => {
    const { shouldRefreshAccessToken } = await load()
    expect(shouldRefreshAccessToken(tokenIssuedAgo(60))).toBe(false)
  })

  it('is true once the token is older than a day', async () => {
    const { shouldRefreshAccessToken } = await load()
    expect(shouldRefreshAccessToken(tokenIssuedAgo(DAY + 60))).toBe(true)
  })

  it('is true when less than a day is left', async () => {
    const { shouldRefreshAccessToken } = await load()
    // A legacy 24h token issued an hour ago
    expect(shouldRefreshAccessToken(tokenIssuedAgo(3600, DAY))).toBe(true)
  })

  it('is true for missing or malformed tokens', async () => {
    const { shouldRefreshAccessToken } = await load()
    expect(shouldRefreshAccessToken(null)).toBe(true)
    expect(shouldRefreshAccessToken('garbage')).toBe(true)
  })
})

describe('exchangeExternalToken', () => {
  it('posts the token in the body and stores the pair', async () => {
    const {
      exchangeExternalToken,
      getStoredAccessToken,
      getStoredRefreshToken,
    } = await load()
    fetchMock.mockResolvedValue(jsonResponse(200, pair('access-1', 'wfr_1')))

    await expect(
      exchangeExternalToken('https://api.test/auth', 'auth0-token'),
    ).resolves.toBe(true)

    expect(fetchMock).toHaveBeenCalledWith(
      'https://api.test/auth/exchange',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ token: 'auth0-token' }),
      }),
    )
    expect(getStoredAccessToken()).toBe('access-1')
    expect(getStoredRefreshToken()).toBe('wfr_1')
  })

  it('returns false when the backend rejects the token', async () => {
    const { exchangeExternalToken, getStoredAccessToken } = await load()
    fetchMock.mockResolvedValue(jsonResponse(403))

    await expect(
      exchangeExternalToken('https://api.test/auth', 'auth0-token'),
    ).resolves.toBe(false)
    expect(getStoredAccessToken()).toBeNull()
  })
})

describe('refreshSession', () => {
  beforeEach(() => {
    store.set('token', 'old-access')
    store.set('refresh_token', 'wfr_old')
  })

  it('rotates and stores the new pair', async () => {
    const { refreshSession, getStoredAccessToken, getStoredRefreshToken } =
      await load()
    fetchMock.mockResolvedValue(
      jsonResponse(200, pair('new-access', 'wfr_new')),
    )

    await expect(refreshSession('https://api.test/auth')).resolves.toBe(true)

    expect(fetchMock).toHaveBeenCalledWith(
      'https://api.test/auth/refresh',
      expect.objectContaining({
        body: JSON.stringify({ refresh_token: 'wfr_old' }),
      }),
    )
    expect(getStoredAccessToken()).toBe('new-access')
    expect(getStoredRefreshToken()).toBe('wfr_new')
  })

  it('shares one request between concurrent callers', async () => {
    const { refreshSession } = await load()
    fetchMock.mockResolvedValue(
      jsonResponse(200, pair('new-access', 'wfr_new')),
    )

    const results = await Promise.all([
      refreshSession('https://api.test/auth'),
      refreshSession('https://api.test/auth'),
      refreshSession('https://api.test/auth'),
    ])

    expect(results).toEqual([true, true, true])
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('reuses tokens another tab stored while waiting for the lock', async () => {
    const { refreshSession } = await load()
    vi.stubGlobal('navigator', {
      locks: {
        request: async (_name: string, fn: () => Promise<boolean>) => {
          // Another tab held the lock and refreshed first.
          store.set('token', 'other-tab-access')
          store.set('refresh_token', 'wfr_other')
          return fn()
        },
      },
    })

    await expect(refreshSession('https://api.test/auth')).resolves.toBe(true)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('clears the session when the refresh token is rejected', async () => {
    const { refreshSession, getStoredAccessToken, getStoredRefreshToken } =
      await load()
    fetchMock.mockResolvedValue(
      jsonResponse(401, { code: 'invalid_refresh_token' }),
    )

    await expect(refreshSession('https://api.test/auth')).resolves.toBe(false)
    expect(getStoredAccessToken()).toBeNull()
    expect(getStoredRefreshToken()).toBeNull()
  })

  it('keeps the session on a network error and backs off', async () => {
    const { refreshSession, getStoredRefreshToken, canRefreshProactively } =
      await load()
    store.set('token', tokenIssuedAgo(2 * DAY))
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(refreshSession('https://api.test/auth')).resolves.toBe(false)
    expect(getStoredRefreshToken()).toBe('wfr_old')

    expect(canRefreshProactively()).toBe(false)
    vi.advanceTimersByTime(61_000)
    expect(canRefreshProactively()).toBe(true)
  })

  it('keeps the session on a server error', async () => {
    const { refreshSession, getStoredRefreshToken } = await load()
    fetchMock.mockResolvedValue(jsonResponse(500))

    await expect(refreshSession('https://api.test/auth')).resolves.toBe(false)
    expect(getStoredRefreshToken()).toBe('wfr_old')
  })

  it('treats 409 as success when another request already stored new tokens', async () => {
    const { refreshSession } = await load()
    fetchMock.mockImplementation(async () => {
      store.set('refresh_token', 'wfr_from_other_request')
      store.set('token', 'access-from-other-request')
      return jsonResponse(409, { code: 'token_rotated' })
    })

    await expect(refreshSession('https://api.test/auth')).resolves.toBe(true)
  })

  it('returns false without a refresh token', async () => {
    const { refreshSession } = await load()
    store.delete('refresh_token')

    await expect(refreshSession('https://api.test/auth')).resolves.toBe(false)
    expect(fetchMock).not.toHaveBeenCalled()
  })
})

describe('canRefreshProactively', () => {
  it('requires a refresh token', async () => {
    const { canRefreshProactively } = await load()
    store.set('token', tokenIssuedAgo(2 * DAY))
    expect(canRefreshProactively()).toBe(false)

    store.set('refresh_token', 'wfr_1')
    expect(canRefreshProactively()).toBe(true)
  })

  it('skips fresh tokens', async () => {
    const { canRefreshProactively } = await load()
    store.set('token', tokenIssuedAgo(60))
    store.set('refresh_token', 'wfr_1')
    expect(canRefreshProactively()).toBe(false)
  })
})

describe('logoutSession', () => {
  it('clears tokens and revokes the session', async () => {
    const { logoutSession, getStoredAccessToken, getStoredRefreshToken } =
      await load()
    store.set('token', 'access')
    store.set('refresh_token', 'wfr_1')
    fetchMock.mockResolvedValue(jsonResponse(204))

    await logoutSession('https://api.test/auth')

    expect(getStoredAccessToken()).toBeNull()
    expect(getStoredRefreshToken()).toBeNull()
    expect(fetchMock).toHaveBeenCalledWith(
      'https://api.test/auth/logout',
      expect.objectContaining({
        body: JSON.stringify({ refresh_token: 'wfr_1' }),
      }),
    )
  })

  it('does not call the backend without a refresh token', async () => {
    const { logoutSession } = await load()
    store.set('token', 'legacy-access')

    await logoutSession('https://api.test/auth')

    expect(store.has('token')).toBe(false)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('still clears tokens when the backend is unreachable', async () => {
    const { logoutSession } = await load()
    store.set('refresh_token', 'wfr_1')
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(
      logoutSession('https://api.test/auth'),
    ).resolves.toBeUndefined()
    expect(store.has('refresh_token')).toBe(false)
  })
})
