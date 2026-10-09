import type { BrandingColorsFieldsFragment } from '~/api/generated'

/**
 * The theme is cached so a cold start paints branded colours before the
 * project query resolves. It carries the project it belongs to: without that
 * it is just "some colours", and an admin switching the current project leaves
 * every client applying the previous project's theme on each mount with no way
 * to tell it is stale.
 */
export interface CachedTheme {
  projectId: string
  colors: BrandingColorsFieldsFragment
}

export function isValidTheme(value: unknown): value is CachedTheme {
  const cached = value as CachedTheme | null
  return (
    typeof cached === 'object' &&
    cached !== null &&
    typeof cached.projectId === 'string' &&
    typeof cached.colors?.light?.accent === 'string' &&
    typeof cached.colors?.dark?.accent === 'string'
  )
}

const TOKENS = [
  ['accent', 'accent'],
  ['accent-contrast', 'accentContrast'],
  ['on-accent', 'onAccent'],
  ['background-default', 'backgroundDefault'],
  ['background-raised', 'backgroundRaised'],
  ['background-indent', 'backgroundIndent'],
  ['text-default', 'textDefault'],
  ['text-muted', 'textMuted'],
  ['text-hint', 'textHint'],
  ['shadow-default', 'shadowDefault'],
  ['shadow-blank', 'shadowBlank'],
  ['border-default', 'borderDefault'],
] as const

function block(
  selector: string,
  set: BrandingColorsFieldsFragment['light'],
): string {
  const declarations = TOKENS.map(
    ([token, field]) => `--color-${token}: ${set[field]};`,
  ).join('\n    ')
  return `${selector} {\n    ${declarations}\n  }`
}

// Unlayered, so it beats Tailwind's `@theme` defaults without extra specificity.
export function buildThemeCss(colors: BrandingColorsFieldsFragment): string {
  return `${block(':root', colors.light)}\n  ${block('.dark', colors.dark)}`
}
