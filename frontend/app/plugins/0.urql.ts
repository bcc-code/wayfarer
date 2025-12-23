import { authExchange, type AuthConfig } from '@urql/exchange-auth'
import { useAuth0 } from '@auth0/auth0-vue'
import urql, { Client, fetchExchange } from '@urql/vue'

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(
    urql,
    new Client({
      url: useRuntimeConfig().public.apiUrl,
      preferGetMethod: false,
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
              // Prevent multiple redirects
              if (isRedirecting) return
              isRedirecting = true

              // Clear Wayfarer token
              wayfarerToken.value = null

              // Redirect to Auth0 login via the middleware
              // The middleware will handle the Auth0 login flow
              const auth0 = useAuth0()
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
