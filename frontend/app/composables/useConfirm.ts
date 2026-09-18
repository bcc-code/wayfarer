import { reactive } from 'vue'

export interface ConfirmOptions {
  /** Dialog heading. */
  title: string
  /** Optional body text explaining the consequence. */
  description?: string
  /** Label for the confirming action. Defaults to `admin.common.delete`. */
  confirmLabel?: string
  /** Label for the dismissing action. Defaults to `admin.common.cancel`. */
  cancelLabel?: string
  /** Colour of the confirming button. */
  color?: 'error' | 'primary' | 'neutral'
  /** Icon shown next to the title. */
  icon?: string
}

interface ConfirmState extends ConfirmOptions {
  open: boolean
  resolve: ((value: boolean) => void) | null
}

// Module-level singleton so any page can open the one dialog rendered by the
// admin layout (`layouts/admin.vue` mounts `AdminConfirmDialog`).
const state = reactive<ConfirmState>({
  open: false,
  title: '',
  description: undefined,
  confirmLabel: undefined,
  cancelLabel: undefined,
  color: 'error',
  icon: undefined,
  resolve: null,
})

function settle(value: boolean) {
  state.resolve?.(value)
  state.resolve = null
  state.open = false
}

/**
 * Promise-based confirmation dialog, replacing native `window.confirm()`.
 *
 * ```ts
 * const { confirm } = useConfirm()
 * if (!(await confirm({ title: 'Slette laget?' }))) return
 * ```
 */
export function useConfirm() {
  function confirm(options: ConfirmOptions): Promise<boolean> {
    // A second call while one is pending dismisses the first rather than
    // leaving its promise dangling forever.
    settle(false)

    // Assign every optional field explicitly: a bare Object.assign would leave
    // the previous dialog's description/label/icon in place when the next one
    // omits them.
    Object.assign(state, {
      title: options.title,
      description: options.description,
      confirmLabel: options.confirmLabel,
      cancelLabel: options.cancelLabel,
      icon: options.icon,
      color: options.color ?? 'error',
      open: true,
    })

    return new Promise<boolean>((resolve) => {
      state.resolve = resolve
    })
  }

  return { confirm, state, settle }
}
