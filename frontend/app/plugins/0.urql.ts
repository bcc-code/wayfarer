import { authExchange, type AuthConfig } from '@urql/exchange-auth'
import { useAuth0 } from '@auth0/auth0-vue'
import urql, { cacheExchange, Client, fetchExchange } from '@urql/vue'

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(
    urql,
    new Client({
      url: useRuntimeConfig().public.apiUrl,
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
          const wayfarerToken = useLocalStorage<string>('token', () => null)
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

              // Clear expired Wayfarer token
              wayfarerToken.value = null

              const auth0 = useAuth0()

              // Try silent refresh first - get fresh Auth0 token and exchange it
              try {
                const auth0Token = await auth0.getAccessTokenSilently()
                if (auth0Token) {
                  const config = useRuntimeConfig()
                  const response = await $fetch<{ token: string }>(
                    `${config.public.tokenUrl}?token=${auth0Token}`,
                    { method: 'GET' },
                  )
                  if (response?.token) {
                    wayfarerToken.value = response.token
                    return // Success - urql will retry the operation
                  }
                }
              } catch {
                // Silent refresh failed, fall through to login redirect
              }

              // Silent refresh failed - do full login redirect
              isRedirecting = true
              await auth0.loginWithRedirect({
                appState: {
                  targetUrl: window.location.pathname,
                },
              })
            },
            willAuthError() {
              return false
            },
          } as AuthConfig
        }),
        fetchExchange,
      ],
    }),
  )
})
