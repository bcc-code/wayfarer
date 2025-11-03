export function useAuth() {
  const isLoading = useState('isLoading', () => true)
  const user = useState('user', () => ({
    id: 'e52257a2-0240-4e4d-bcf3-9953162e491d',
    personId: 99999,
    name: 'Ola Nordmann',
    image: 'https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=Felix',
  }))

  // TODO: implement
  function getAccessTokenSilently() {
    return 'token'
  }

  // TODO: implement
  function loginWithRedirect() {
    return Promise.resolve()
  }

  return { getAccessTokenSilently, loginWithRedirect, isLoading, user }
}
