import { authExchange, type AuthConfig } from '@urql/exchange-auth'
import { useAuth0 } from '@auth0/auth0-vue'
import urql, { cacheExchange, Client, fetchExchange } from '@urql/vue'

export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig()
  nuxtApp.vueApp.use(
    urql,
    new Client({
      url: config.public.apiUrl,
      preferGetMethod: false,
      requestPolicy: 'cache-and-network',
      fetchOptions() {
        return {
          headers: {
            'Accept-Language':
              (nuxtApp.$i18n as { locale?: { value?: string } })?.locale
                ?.value || 'en',
          },
        }
      },
      exchanges: [
        cacheExchange,
        authExchange(async (utils) => {
          // Use Wayfarer token from localStorage for API calls
          const wayfarerToken = useLocalStorage<string>(
            ACCESS_TOKEN_KEY,
            () => null,
          )
          const authUrl = authBaseUrl(config.public)
          let isRedirecting = false

          return {
            addAuthToOperation(operation) {
              const headers: Record<string, string> = {}
              if (wayfarerToken.value) {
                headers.Authorization = `Bearer ${wayfarerToken.value}`
              }
              return utils.appendHeaders(operation, headers)
            },
            didAuthError(error) {
              // Don't check for auth errors if we're already redirecting
              if (isRedirecting) return false

              // Check for authentication errors
              return (
                error.response?.status === 401 ||
                error.graphQLErrors?.some(
                  (e) =>
                    e.extensions?.code === 'UNAUTHENTICATED' ||
                    e.extensions?.code === 'UNAUTHORIZED',
                ) ||
                false
              )
            },
            async refreshAuth() {
              // Prevent multiple concurrent refresh attempts
              if (isRedirecting) return

              // 1. Wayfarer session: rotate the refresh token. No Auth0 needed.
              if (hasRefreshToken()) {
                if (await refreshSession(authUrl)) return
                // A network or server hiccup keeps the tokens. If the access
                // token still works, carry on and try again later.
                const current = getStoredAccessToken()
                if (hasRefreshToken() && current && !isTokenExpired(current)) {
                  return
                }
              }

              // This runs from a urql exchange callback, outside any Vue setup
              // or app context — useAuth0() relies on inject(), so it must be
              // resolved inside runWithContext or it returns undefined.
              const auth0 = () =>
                nuxtApp.vueApp.runWithContext(() => useAuth0())

              // 2. No session (legacy token or session ended): exchange a
              // fresh Auth0 token if the Auth0 session is still alive.
              try {
                const auth0Token = await auth0().getAccessTokenSilently()
                if (
                  auth0Token &&
                  (await exchangeExternalToken(authUrl, auth0Token))
                ) {
                  return // Success - urql will retry the operation
                }
              } catch {
                // Silent refresh failed, fall through to login redirect
              }

              // Silent refresh failed - clear the token and do a full login
              // redirect. The token is only cleared here, after the refresh
              // attempt, so a recoverable 401 never pauses in-flight queries.
              isRedirecting = true
              clearSessionTokens()
              try {
                await auth0().loginWithRedirect({
                  appState: {
                    targetUrl:
                      window.location.pathname + window.location.search,
                  },
                })
              } catch {
                // Auth0 client unavailable — still get the user to login
                // instead of leaving the app stuck on a blank page.
                window.location.assign('/login')
              }
            },
            willAuthError() {
              // Refresh ahead of expiry (and daily, so the session keeps
              // sliding) instead of waiting for a 401.
              return !isRedirecting && canRefreshProactively()
            },
          } as AuthConfig
        }),
        fetchExchange,
      ],
    }),
  )
})
