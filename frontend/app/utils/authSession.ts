/**
 * Wayfarer session tokens.
 *
 * After the one-time BCC (Auth0) login, the backend issues a pair:
 * - an access token (JWT, 7 days) sent on every API request
 * - a refresh token (opaque, ~6 months, sliding) used only to get a new pair
 *
 * Both live in localStorage so every tab shares them. Refreshing is
 * serialised across tabs with the Web Locks API, because the backend rotates
 * the refresh token on every use.
 */

export const ACCESS_TOKEN_KEY = 'token'
export const REFRESH_TOKEN_KEY = 'refresh_token'

/** Refresh once the access token is older than this, so the session slides with normal use. */
const REFRESH_AFTER_MS = 24 * 60 * 60 * 1000
/** Also refresh when the access token has less than this left. */
const REFRESH_BEFORE_EXPIRY_MS = 24 * 60 * 60 * 1000
/** After a failed refresh (e.g. offline), wait this long before proactively retrying. */
const RETRY_COOLDOWN_MS = 60 * 1000

const LOCK_NAME = 'wayfarer-token-refresh'

export interface SessionTokenPair {
  access_token: string
  access_expires_at: string
  refresh_token: string
  refresh_expires_at: string
}

/** Base URL of the backend auth endpoints, e.g. `https://api.example/auth`. */
export function authBaseUrl(config: { authUrl?: string; apiUrl: string }) {
  if (config.authUrl) return config.authUrl.replace(/\/$/, '')
  return config.apiUrl.replace(/\/graphql\/?$/, '') + '/auth'
}

function readStorage(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

/**
 * Writes to localStorage and notifies same-tab `useLocalStorage` refs, which
 * otherwise only react to writes from other tabs.
 */
function writeStorage(key: string, value: string | null) {
  try {
    const oldValue = localStorage.getItem(key)
    if (value === null) localStorage.removeItem(key)
    else localStorage.setItem(key, value)
    if (typeof window !== 'undefined' && typeof StorageEvent !== 'undefined') {
      window.dispatchEvent(
        new StorageEvent('storage', {
          key,
          oldValue,
          newValue: value,
          storageArea: localStorage,
        }),
      )
    }
  } catch {
    // Storage unavailable (private mode); the session just won't persist.
  }
}

export function getStoredAccessToken() {
  return readStorage(ACCESS_TOKEN_KEY)
}

export function getStoredRefreshToken() {
  return readStorage(REFRESH_TOKEN_KEY)
}

export function hasRefreshToken() {
  return !!getStoredRefreshToken()
}

export function storeTokenPair(pair: SessionTokenPair) {
  writeStorage(REFRESH_TOKEN_KEY, pair.refresh_token)
  writeStorage(ACCESS_TOKEN_KEY, pair.access_token)
}

export function clearSessionTokens() {
  writeStorage(ACCESS_TOKEN_KEY, null)
  writeStorage(REFRESH_TOKEN_KEY, null)
}

function decodeClaims(token: string): { iat?: number; exp?: number } | null {
  const payloadPart = token.split('.')[1]
  if (!payloadPart) return null
  try {
    return JSON.parse(atob(payloadPart.replace(/-/g, '+').replace(/_/g, '/')))
  } catch {
    return null
  }
}

/**
 * Whether the access token is due for a proactive refresh: older than a day,
 * or less than a day from expiring. Only meaningful with a refresh token.
 */
export function shouldRefreshAccessToken(
  token: string | null | undefined,
): boolean {
  if (!token) return true
  const claims = decodeClaims(token)
  if (!claims || typeof claims.exp !== 'number') return true
  const now = Date.now()
  if (claims.exp * 1000 - now < REFRESH_BEFORE_EXPIRY_MS) return true
  return (
    typeof claims.iat === 'number' && now - claims.iat * 1000 > REFRESH_AFTER_MS
  )
}

let lastFailedAt = 0

/** Whether a proactive refresh should be attempted now. */
export function canRefreshProactively() {
  return (
    hasRefreshToken() &&
    shouldRefreshAccessToken(getStoredAccessToken()) &&
    Date.now() - lastFailedAt > RETRY_COOLDOWN_MS
  )
}

/**
 * Exchanges a BCC (Auth0) access token for a Wayfarer session.
 * Returns false when the backend rejects the token.
 */
export async function exchangeExternalToken(
  baseUrl: string,
  externalToken: string,
): Promise<boolean> {
  const res = await fetch(`${baseUrl}/exchange`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token: externalToken }),
  })
  if (!res.ok) return false
  storeTokenPair((await res.json()) as SessionTokenPair)
  lastFailedAt = 0
  return true
}

let inflight: Promise<boolean> | null = null

/**
 * Refreshes the session. Concurrent calls in this tab share one request, and
 * the Web Lock makes other tabs wait and then reuse the result.
 *
 * Resolves true when a usable access token is stored afterwards. On a
 * definitive rejection the tokens are cleared; on a network error they are
 * kept so the user isn't logged out while offline.
 */
export function refreshSession(baseUrl: string): Promise<boolean> {
  // The token that triggered the refresh. If another tab replaces it while
  // we wait for the lock, that tab already did the work.
  const staleAccessToken = getStoredAccessToken()
  inflight ??= withLock(() => doRefresh(baseUrl, staleAccessToken)).finally(
    () => {
      inflight = null
    },
  )
  return inflight
}

async function withLock<T>(fn: () => Promise<T>): Promise<T> {
  const locks = typeof navigator !== 'undefined' ? navigator.locks : undefined
  if (!locks) return fn()
  return locks.request(LOCK_NAME, fn)
}

async function doRefresh(
  baseUrl: string,
  staleAccessToken: string | null,
): Promise<boolean> {
  const current = getStoredAccessToken()
  if (current && current !== staleAccessToken) return true

  const refreshToken = getStoredRefreshToken()
  if (!refreshToken) return false

  let res: Response
  try {
    res = await fetch(`${baseUrl}/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  } catch {
    lastFailedAt = Date.now()
    return false
  }

  if (res.ok) {
    storeTokenPair((await res.json()) as SessionTokenPair)
    lastFailedAt = 0
    return true
  }

  if (res.status === 409) {
    // Rotated moments ago by a request we couldn't coordinate with (e.g. a
    // browser without Web Locks). Use what it stored, if anything.
    return getStoredRefreshToken() !== refreshToken && !!getStoredAccessToken()
  }

  if (res.status === 400 || res.status === 401) {
    clearSessionTokens()
    return false
  }

  // Server error: keep the session and try again later.
  lastFailedAt = Date.now()
  return false
}

/** Revokes the session on the server (best effort) and clears local tokens. */
export async function logoutSession(baseUrl: string) {
  const refreshToken = getStoredRefreshToken()
  clearSessionTokens()
  if (!refreshToken) return
  try {
    await fetch(`${baseUrl}/logout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
      keepalive: true,
    })
  } catch {
    // The session expires on its own; nothing else to do.
  }
}
