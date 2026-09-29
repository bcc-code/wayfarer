/**
 * Warns before leaving a form that has unsaved work.
 *
 * Dirtiness is a comparison against a baseline rather than a flag set by the
 * first change: an admin form fills itself in when its query resolves, which
 * would set a flag and make every page warn on the way out. Take the snapshot
 * again with `markSaved()` whenever the form legitimately matches what is
 * stored — after initial data arrives, and after a save lands.
 */
export function useUnsavedChanges(snapshot: () => unknown) {
  const { confirm } = useConfirm()

  const serialise = () => JSON.stringify(snapshot())
  const baseline = ref(serialise())

  const isDirty = computed(() => serialise() !== baseline.value)

  function markSaved() {
    baseline.value = serialise()
  }

  onBeforeRouteLeave(async () => {
    if (!isDirty.value) return true
    return confirm({
      title: 'Forlate siden med ulagrede endringer?',
      description: 'Endringene du har gjort blir ikke lagret.',
      confirmLabel: 'Forlat siden',
    })
  })

  // A refresh or a closed tab never reaches the router guard. The browser
  // shows its own wording here; preventDefault is all it takes to ask.
  useEventListener('beforeunload', (event: BeforeUnloadEvent) => {
    if (isDirty.value) event.preventDefault()
  })

  return { isDirty, markSaved }
}
