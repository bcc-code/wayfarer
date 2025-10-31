export function useAuth() {
  const isLoading = useState('isLoading', () => true)

  // TODO: implement
  function getAccessTokenSilently() {
    return 'token'
  }

  // TODO: implement
  function loginWithRedirect() {
    return Promise.resolve()
  }

  return { getAccessTokenSilently, loginWithRedirect, isLoading }
}
