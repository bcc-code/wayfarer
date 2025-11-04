import { authExchange, type AuthConfig } from '@urql/exchange-auth'
import urql, { Client, fetchExchange } from '@urql/vue'

export default defineNuxtPlugin((plugin) => {
  const { getAccessTokenSilently } = useAuth()

  plugin.vueApp.use(
    urql,
    new Client({
      url: useRuntimeConfig().public.apiUrl,
      preferGetMethod: false,
      exchanges: [
        authExchange(async (utils) => {
          let token = await getAccessTokenSilently()
          return {
            addAuthToOperation(operation) {
              const headers: Record<string, string> = {}
              if (token) {
                headers.Authorization = `Bearer ${token}`
              }
              return utils.appendHeaders(operation, headers)
            },
            didAuthError() {
              return false
            },
            async refreshAuth() {
              token = await getAccessTokenSilently()
            },
            willAuthError() {
              return true
            },
          } as AuthConfig
        }),
        fetchExchange,
      ],
    }),
  )
})
