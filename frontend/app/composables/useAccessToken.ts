export function useAccessToken() {
  const { getAccessTokenSilently, loginWithRedirect, isLoading } = useAuth()

  const token = useState<string | null>('token', () => null)

  const getToken = async () => {
    try {
      while (isLoading.value) {
        await new Promise((resolve) => setTimeout(resolve, 10))
      }
      return token.value ?? (await getAccessTokenSilently())
    } catch {
      await loginWithRedirect()
    }
  }

  return {
    getToken,
    setToken: (value: string) => {
      token.value = value
    },
  }
}
