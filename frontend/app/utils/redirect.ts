/**
 * The path to return to after signing in, or `undefined` when there is nothing
 * safe to return to.
 *
 * Only same-site paths survive: the value travels through a query parameter,
 * so an absolute URL — or the `//host` form the browser also treats as one —
 * would turn the login flow into an open redirect.
 */
export function safeRedirectPath(
  value: unknown,
  fallback?: string,
): string | undefined {
  const path = Array.isArray(value) ? value[0] : value
  if (typeof path !== 'string' || path === '') return fallback
  if (!path.startsWith('/') || path.startsWith('//')) return fallback
  if (path.startsWith('/\\')) return fallback
  return path
}
