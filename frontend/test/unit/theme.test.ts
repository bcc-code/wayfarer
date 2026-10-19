import { describe, it, expect } from 'vitest'
import {
  buildThemeCss,
  isValidTheme,
  type CachedTheme,
} from '../../layers/user/app/utils/theme'

const colorSet = (accent: string) => ({
  accent,
  accentContrast: '#000',
  onAccent: '#fff',
  backgroundDefault: '#eee',
  backgroundRaised: '#fff',
  backgroundIndent: '#ddd',
  textDefault: '#111',
  textMuted: '#555',
  textHint: '#999',
  shadowDefault: '#0001',
  shadowBlank: '#0000',
  borderDefault: '#ccc',
})

const colors = {
  light: colorSet('#9ed63c'),
  dark: colorSet('#7ab32a'),
} as CachedTheme['colors']

describe('buildThemeCss', () => {
  it('emits every token for both schemes', () => {
    const css = buildThemeCss(colors)

    expect(css).toContain('--color-accent: #9ed63c;')
    expect(css).toContain('--color-background-indent: #ddd;')
    expect(css).toMatch(/\.dark \{[\s\S]*--color-accent: #7ab32a;/)
  })

  // Tailwind declares these tokens inside `@layer theme`, and unlayered CSS
  // beats any layer — so the generated block must stay unlayered.
  it('stays unlayered', () => {
    expect(buildThemeCss(colors)).not.toContain('@layer')
  })
})

describe('isValidTheme', () => {
  it('accepts a cache that names its project', () => {
    expect(isValidTheme({ projectId: 'PR01', colors })).toBe(true)
  })

  // The previous shape was the bare colour sets, with no way to tell which
  // project they came from. Those entries must be ignored rather than applied.
  it('rejects the old project-less shape', () => {
    expect(isValidTheme(colors)).toBe(false)
  })

  it.each([
    ['null', null],
    ['a string', 'nope'],
    ['missing colors', { projectId: 'PR01' }],
    [
      'missing dark set',
      { projectId: 'PR01', colors: { light: colorSet('#f') } },
    ],
  ])('rejects %s', (_label, value) => {
    expect(isValidTheme(value)).toBe(false)
  })
})
