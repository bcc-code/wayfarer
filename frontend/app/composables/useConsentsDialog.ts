export default function useConsentsDialog() {
  const open = useState('consentsDialogOpen', () => false)
  return {
    open,
  }
}
