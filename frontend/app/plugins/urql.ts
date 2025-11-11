import { authExchange, type AuthConfig } from '@urql/exchange-auth'
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
            'Accept-Language': nuxtApp.$i18n?.locale?.value || 'en',
          },
        }
      },
      exchanges: [
        authExchange(async (utils) => {
          // Defer getting the token until the auth exchange is actually used
          const token = useCookie('token')
          let isRedirecting = false

          return {
            addAuthToOperation(operation) {
              const headers: Record<string, string> = {}
              if (token.value) {
                headers.Authorization = `Bearer ${token.value}`
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

              // Clear token and redirect to login
              token.value = null
              const loginUrl =
                useRuntimeConfig().public.loginUrl +
                '?redirect=' +
                encodeURIComponent(window.location.pathname)
              window.location.href = loginUrl
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
