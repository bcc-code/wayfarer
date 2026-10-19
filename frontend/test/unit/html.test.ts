import { describe, it, expect } from 'vitest'
import { isBlankHtml } from '../../layers/admin/app/utils/html'

describe('isBlankHtml', () => {
  it('treats nothing at all as blank', () => {
    expect(isBlankHtml(undefined)).toBe(true)
    expect(isBlankHtml('')).toBe(true)
  })

  // What a Tiptap editor leaves behind when you clear it.
  it('treats an empty paragraph as blank', () => {
    expect(isBlankHtml('<p></p>')).toBe(true)
    expect(isBlankHtml('<p><br></p>')).toBe(true)
    expect(isBlankHtml('<p>&nbsp;</p>')).toBe(true)
    expect(isBlankHtml('<p>   </p>\n<p></p>')).toBe(true)
  })

  it('keeps anything with text', () => {
    expect(isBlankHtml('<p>Les Matteus 5</p>')).toBe(false)
    expect(isBlankHtml('<ul><li>Ett</li></ul>')).toBe(false)
  })

  // Text is not the only kind of content.
  it('keeps media with no text around it', () => {
    expect(isBlankHtml('<p><img src="/kart.png"></p>')).toBe(false)
    expect(isBlankHtml('<hr>')).toBe(false)
  })
})
