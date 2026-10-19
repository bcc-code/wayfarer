/**
 * Whether a rich text value holds nothing worth storing.
 *
 * A Tiptap editor never returns an empty string: clearing it leaves an empty
 * paragraph behind, which would be saved as a description and render as a gap
 * in the app.
 */
export function isBlankHtml(html: string | undefined): boolean {
  if (!html) return true

  // Content that carries no text of its own still counts as content.
  if (/<(img|hr|iframe|video|audio)\b/i.test(html)) return false

  return (
    html
      .replace(/<[^>]*>/g, '')
      .replace(/&nbsp;/gi, ' ')
      .trim() === ''
  )
}
