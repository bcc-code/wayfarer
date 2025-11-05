import { authExchange, type AuthConfig } from '@urql/exchange-auth'
import urql, { Client, fetchExchange } from '@urql/vue'

export default defineNuxtPlugin((plugin) => {
  plugin.vueApp.use(
    urql,
    new Client({
      url: useRuntimeConfig().public.apiUrl,
      preferGetMethod: false,
      exchanges: [
        authExchange(async (utils) => {
          // Defer getting the token until the auth exchange is actually used
          const token = useLocalStorage<string | null>('token', () => null)

          return {
            addAuthToOperation(operation) {
              const headers: Record<string, string> = {}
              if (token.value) {
                headers.Authorization = `Bearer ${token.value}`
              }
              return utils.appendHeaders(operation, headers)
            },
            didAuthError() {
              return false
            },
            async refreshAuth() {
              // Token is reactive, so it will automatically update
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
