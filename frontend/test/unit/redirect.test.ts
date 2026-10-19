import { describe, it, expect } from 'vitest'
import { safeRedirectPath } from '../../app/utils/redirect'

describe('safeRedirectPath', () => {
  it('keeps a path within the app', () => {
    expect(safeRedirectPath('/admin')).toBe('/admin')
    expect(safeRedirectPath('/admin/projects/PR1?tab=teams')).toBe(
      '/admin/projects/PR1?tab=teams',
    )
  })

  it('falls back when there is nothing to return to', () => {
    expect(safeRedirectPath(undefined)).toBeUndefined()
    expect(safeRedirectPath('')).toBeUndefined()
    expect(safeRedirectPath(undefined, '/')).toBe('/')
  })

  // The value rides in a query parameter, so it is attacker-supplied.
  it('refuses anything that could leave the site', () => {
    expect(safeRedirectPath('https://evil.example/admin', '/')).toBe('/')
    expect(safeRedirectPath('//evil.example', '/')).toBe('/')
    expect(safeRedirectPath('/\\evil.example', '/')).toBe('/')
    expect(safeRedirectPath('javascript:alert(1)', '/')).toBe('/')
  })

  // Vue Router hands repeated query parameters over as an array.
  it('takes the first of a repeated parameter', () => {
    expect(safeRedirectPath(['/admin', '/elsewhere'])).toBe('/admin')
  })
})
