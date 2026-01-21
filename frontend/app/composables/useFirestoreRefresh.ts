/**
 * Hook for pages to respond to Firestore updates
 * @param queryNames - Array of query document names to listen for
 * @param callback - Function to call when any of the queries should be refreshed
 */
export function useFirestoreRefresh(
  queryNames: string[],
  callback: () => void,
) {
  const handleUpdate = (event: Event) => {
    const customEvent = event as CustomEvent<{ query: string }>
    if (queryNames.includes(customEvent.detail.query)) {
      callback()
    }
  }

  onMounted(() => {
    window.addEventListener('firestore-update', handleUpdate)
  })

  onUnmounted(() => {
    window.removeEventListener('firestore-update', handleUpdate)
  })
}
