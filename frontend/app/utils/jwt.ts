// Safety margin so a token that expires within the next 30s is treated as
// expired — avoids firing requests that will 401 mid-flight.
const CLOCK_SKEW_MS = 30_000

/**
 * Checks whether a JWT is expired (or will expire within the clock-skew
 * margin). Malformed tokens and tokens without a numeric `exp` claim are
 * treated as expired so callers always fall back to re-authentication.
 */
export function isTokenExpired(token: string | null | undefined): boolean {
  if (!token) return true

  const payloadPart = token.split('.')[1]
  if (!payloadPart) return true

  try {
    const payload = JSON.parse(
      atob(payloadPart.replace(/-/g, '+').replace(/_/g, '/')),
    )
    if (typeof payload.exp !== 'number') return true
    return payload.exp * 1000 <= Date.now() + CLOCK_SKEW_MS
  } catch {
    return true
  }
}
