/**
 * Color utility functions for calculating contrast and accessibility
 */

/**
 * Convert hex color to RGB
 * @param hexColor - Hex color string (with or without #)
 * @returns RGB values as [r, g, b] array (0-255)
 */
export function hexToRgb(hexColor: string): [number, number, number] {
  const hex = hexColor.replace('#', '')
  const r = Number.parseInt(hex.substring(0, 2), 16)
  const g = Number.parseInt(hex.substring(2, 4), 16)
  const b = Number.parseInt(hex.substring(4, 6), 16)
  return [r, g, b]
}

/**
 * Convert RGB to HSL
 * @param r - Red channel (0-255)
 * @param g - Green channel (0-255)
 * @param b - Blue channel (0-255)
 * @returns HSL values as [h, s, l] where h is 0-360, s and l are 0-1
 */
export function rgbToHsl(
  r: number,
  g: number,
  b: number,
): [number, number, number] {
  r /= 255
  g /= 255
  b /= 255

  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  const delta = max - min

  let h = 0
  let s = 0
  const l = (max + min) / 2

  if (delta !== 0) {
    s = l > 0.5 ? delta / (2 - max - min) : delta / (max + min)

    switch (max) {
      case r:
        h = ((g - b) / delta + (g < b ? 6 : 0)) / 6
        break
      case g:
        h = ((b - r) / delta + 2) / 6
        break
      case b:
        h = ((r - g) / delta + 4) / 6
        break
    }
  }

  return [h * 360, s, l]
}

/**
 * Convert HSL to RGB
 * @param h - Hue (0-360)
 * @param s - Saturation (0-1)
 * @param l - Lightness (0-1)
 * @returns RGB values as [r, g, b] array (0-255)
 */
export function hslToRgb(
  h: number,
  s: number,
  l: number,
): [number, number, number] {
  h /= 360

  const hueToRgb = (p: number, q: number, t: number) => {
    if (t < 0) t += 1
    if (t > 1) t -= 1
    if (t < 1 / 6) return p + (q - p) * 6 * t
    if (t < 1 / 2) return q
    if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6
    return p
  }

  let r: number, g: number, b: number

  if (s === 0) {
    r = g = b = l
  } else {
    const q = l < 0.5 ? l * (1 + s) : l + s - l * s
    const p = 2 * l - q
    r = hueToRgb(p, q, h + 1 / 3)
    g = hueToRgb(p, q, h)
    b = hueToRgb(p, q, h - 1 / 3)
  }

  return [Math.round(r * 255), Math.round(g * 255), Math.round(b * 255)]
}

/**
 * Convert RGB to hex
 * @param r - Red channel (0-255)
 * @param g - Green channel (0-255)
 * @param b - Blue channel (0-255)
 * @returns Hex color string with # prefix
 */
export function rgbToHex(r: number, g: number, b: number): string {
  const toHex = (n: number) => {
    const hex = Math.round(n).toString(16)
    return hex.length === 1 ? '0' + hex : hex
  }
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}

/**
 * Calculate relative luminance for a color according to WCAG 2.1
 * @see https://www.w3.org/TR/WCAG21/#dfn-relative-luminance
 *
 * @param r - Red channel (0-255)
 * @param g - Green channel (0-255)
 * @param b - Blue channel (0-255)
 * @returns Relative luminance value (0-1)
 */
export function getRelativeLuminance(r: number, g: number, b: number): number {
  const [rs, gs, bs] = [r, g, b].map((c) => {
    const sRGB = c / 255
    return sRGB <= 0.03928 ? sRGB / 12.92 : ((sRGB + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs
}

/**
 * Calculate the best contrasting color based on the original color
 * For dark colors, returns a much lighter version
 * For light colors, returns a much darker version
 * Maintains the hue and saturation for brand consistency
 * If target ratio cannot be achieved, reduces saturation to get closer
 *
 * @param hexColor - Hex color string (with or without #)
 * @param targetRatio - Minimum contrast ratio to achieve (default 4.5 for WCAG AA)
 * @returns Lightened or darkened version of the original color
 */
export function getContrastColor(
  hexColor: string,
  targetRatio: number = 4.5,
): string {
  const [r, g, b] = hexToRgb(hexColor)
  const [h, s, l] = rgbToHsl(r, g, b)

  // Determine if we need to lighten or darken
  const luminance = getRelativeLuminance(r, g, b)
  const shouldLighten = luminance < 0.5

  let bestColor = hexColor
  let bestRatio = 1

  // Try different saturation levels (starting with original, then reducing)
  for (let saturation = s; saturation >= 0; saturation -= 0.1) {
    if (shouldLighten) {
      // For dark colors, try lighter values
      // Go from significantly lighter than current up to very light
      const minL = Math.min(l + 0.2, 0.5) // Start well above current but not too high
      for (let lightness = minL; lightness <= 0.98; lightness += 0.02) {
        const [testR, testG, testB] = hslToRgb(h, saturation, lightness)
        const testHex = rgbToHex(testR, testG, testB)
        const ratio = getContrastRatio(hexColor, testHex)

        if (ratio > bestRatio) {
          bestRatio = ratio
          bestColor = testHex
        }

        if (ratio >= targetRatio) {
          return testHex
        }
      }
    } else {
      // For light colors, try darker values
      // Go from very dark up to significantly darker than current
      const maxL = Math.max(l - 0.2, 0.5) // End well below current but not too low
      for (let lightness = 0.02; lightness <= maxL; lightness += 0.02) {
        const [testR, testG, testB] = hslToRgb(h, saturation, lightness)
        const testHex = rgbToHex(testR, testG, testB)
        const ratio = getContrastRatio(hexColor, testHex)

        if (ratio > bestRatio) {
          bestRatio = ratio
          bestColor = testHex
        }

        if (ratio >= targetRatio) {
          return testHex
        }
      }
    }
  }

  // Return the best color we found, even if it doesn't meet the target
  return bestColor
}

/**
 * Calculate the contrast ratio between two colors according to WCAG 2.1
 * @see https://www.w3.org/TR/WCAG21/#dfn-contrast-ratio
 *
 * @param hexColor1 - First hex color string (with or without #)
 * @param hexColor2 - Second hex color string (with or without #)
 * @returns Contrast ratio (1-21)
 */
export function getContrastRatio(hexColor1: string, hexColor2: string): number {
  const hex1 = hexColor1.replace('#', '')
  const hex2 = hexColor2.replace('#', '')

  const r1 = Number.parseInt(hex1.substring(0, 2), 16)
  const g1 = Number.parseInt(hex1.substring(2, 4), 16)
  const b1 = Number.parseInt(hex1.substring(4, 6), 16)

  const r2 = Number.parseInt(hex2.substring(0, 2), 16)
  const g2 = Number.parseInt(hex2.substring(2, 4), 16)
  const b2 = Number.parseInt(hex2.substring(4, 6), 16)

  const lum1 = getRelativeLuminance(r1, g1, b1)
  const lum2 = getRelativeLuminance(r2, g2, b2)

  const lighter = Math.max(lum1, lum2)
  const darker = Math.min(lum1, lum2)

  return (lighter + 0.05) / (darker + 0.05)
}
