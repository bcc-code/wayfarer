export function useAuth() {
  const token = useLocalStorage<string | null>('token', () => null)
  const isLoading = useState('isLoading', () => true)
  const user = useState('user', () => ({
    id: 'e52257a2-0240-4e4d-bcf3-9953162e491d',
    personId: 99999,
    name: 'Ola Nordmann',
    image: 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=Felix',
  }))

  const getAccessToken = async () => {
    try {
      while (isLoading.value) {
        await new Promise((resolve) => setTimeout(resolve, 10))
      }
      return token.value
    } catch {
      await loginWithRedirect()
    }
  }

  const getAccessTokenSilently = async () => {
    return token.value
  }

  const setAccessToken = (value: string) => {
    token.value = value
    isLoading.value = false
  }

  const config = useRuntimeConfig()
  const route = useRoute()

  function loginWithRedirect() {
    return navigateTo(`${config.public.loginUrl}?redirect=${route.path}`, {
      external: true,
    })
  }

  return {
    getAccessTokenSilently,
    getAccessToken,
    setAccessToken,
    loginWithRedirect,
    isLoading,
    user,
    token,
  }
}
