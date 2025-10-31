import urql, { Client, fetchExchange } from '@urql/vue'
import { authExchange, type AuthConfig } from '@urql/exchange-auth'

export default defineNuxtPlugin((plugin) => {
  const { getToken } = useAccessToken()

  plugin.vueApp.use(
    urql,
    new Client({
      url: useRuntimeConfig().public.apiUrl,
      exchanges: [
        authExchange(async (utils) => {
          let token = await getToken()

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
              token = await getToken()
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
