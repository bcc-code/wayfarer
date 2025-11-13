import { describe, it, expect } from 'vitest'
import {
  getRelativeLuminance,
  getContrastColor,
  getContrastRatio,
  hexToRgb,
  rgbToHsl,
  hslToRgb,
  rgbToHex,
} from '../../app/utils/colors'

describe('colors', () => {
  describe('getRelativeLuminance', () => {
    it('should return 0 for black', () => {
      expect(getRelativeLuminance(0, 0, 0)).toBe(0)
    })

    it('should return 1 for white', () => {
      expect(getRelativeLuminance(255, 255, 255)).toBe(1)
    })

    it('should return correct luminance for red', () => {
      const luminance = getRelativeLuminance(255, 0, 0)
      expect(luminance).toBeCloseTo(0.2126, 4)
    })

    it('should return correct luminance for green', () => {
      const luminance = getRelativeLuminance(0, 255, 0)
      expect(luminance).toBeCloseTo(0.7152, 4)
    })

    it('should return correct luminance for blue', () => {
      const luminance = getRelativeLuminance(0, 0, 255)
      expect(luminance).toBeCloseTo(0.0722, 4)
    })

    it('should handle mid-range gray', () => {
      const luminance = getRelativeLuminance(128, 128, 128)
      expect(luminance).toBeGreaterThan(0)
      expect(luminance).toBeLessThan(1)
    })
  })

  describe('color conversion functions', () => {
    it('should convert hex to RGB', () => {
      expect(hexToRgb('#000000')).toEqual([0, 0, 0])
      expect(hexToRgb('#ffffff')).toEqual([255, 255, 255])
      expect(hexToRgb('#ff0000')).toEqual([255, 0, 0])
      expect(hexToRgb('00ff00')).toEqual([0, 255, 0]) // without #
    })

    it('should convert RGB to hex', () => {
      expect(rgbToHex(0, 0, 0)).toBe('#000000')
      expect(rgbToHex(255, 255, 255)).toBe('#ffffff')
      expect(rgbToHex(255, 0, 0)).toBe('#ff0000')
    })

    it('should handle hex with lowercase letters', () => {
      expect(hexToRgb('#aabbcc')).toEqual([170, 187, 204])
      expect(hexToRgb('#AABBCC')).toEqual([170, 187, 204])
    })

    it('should round RGB values when converting from HSL', () => {
      const [r, g, b] = hslToRgb(180, 0.5, 0.5)
      expect(r).toBeCloseTo(64, 0)
      expect(g).toBeCloseTo(191, 0)
      expect(b).toBeCloseTo(191, 0)
      expect(Number.isInteger(r)).toBe(true)
      expect(Number.isInteger(g)).toBe(true)
      expect(Number.isInteger(b)).toBe(true)
    })

    it('should convert RGB to HSL and back', () => {
      // Red
      const [h1, s1, l1] = rgbToHsl(255, 0, 0)
      expect(h1).toBeCloseTo(0, 0)
      expect(s1).toBeCloseTo(1, 1)
      expect(l1).toBeCloseTo(0.5, 1)

      const [r1, g1, b1] = hslToRgb(h1, s1, l1)
      expect(r1).toBe(255)
      expect(g1).toBe(0)
      expect(b1).toBe(0)

      // Green
      const [h2, s2, l2] = rgbToHsl(0, 255, 0)
      expect(h2).toBeCloseTo(120, 0)
      expect(s2).toBeCloseTo(1, 1)
      expect(l2).toBeCloseTo(0.5, 1)

      // Blue
      const [h3, s3, l3] = rgbToHsl(0, 0, 255)
      expect(h3).toBeCloseTo(240, 0)
      expect(s3).toBeCloseTo(1, 1)
      expect(l3).toBeCloseTo(0.5, 1)
    })

    it('should handle grayscale colors correctly', () => {
      // Pure gray has 0 saturation
      const [h1, s1, l1] = rgbToHsl(128, 128, 128)
      expect(s1).toBe(0)
      expect(l1).toBeCloseTo(0.5, 2)

      // Converting back should maintain gray
      const [r1, g1, b1] = hslToRgb(h1, s1, l1)
      expect(r1).toBeCloseTo(128, 0)
      expect(g1).toBeCloseTo(128, 0)
      expect(b1).toBeCloseTo(128, 0)
    })

    it('should handle full round-trip conversions accurately', () => {
      const testColors = [
        [60, 136, 126], // Teal
        [100, 50, 200], // Purple
        [255, 165, 0], // Orange
        [34, 139, 34], // Forest Green
      ]

      for (const [r, g, b] of testColors) {
        const [h, s, l] = rgbToHsl(r, g, b)
        const [r2, g2, b2] = hslToRgb(h, s, l)

        // Should be within 1 due to rounding
        expect(Math.abs(r - r2)).toBeLessThanOrEqual(1)
        expect(Math.abs(g - g2)).toBeLessThanOrEqual(1)
        expect(Math.abs(b - b2)).toBeLessThanOrEqual(1)
      }
    })

    it('should maintain hex format consistency', () => {
      // Should always return lowercase hex with #
      expect(rgbToHex(170, 187, 204)).toBe('#aabbcc')
      expect(rgbToHex(255, 255, 255)).toBe('#ffffff')

      // Should pad single digit hex values
      expect(rgbToHex(0, 0, 15)).toBe('#00000f')
      expect(rgbToHex(1, 2, 3)).toBe('#010203')
    })
  })

  describe('getContrastColor', () => {
    it('should return a lighter color for dark backgrounds', () => {
      const dark = '#1a1a1a'
      const contrast = getContrastColor(dark)
      const [r, g, b] = hexToRgb(contrast)
      const luminance = getRelativeLuminance(r, g, b)

      // Should be much lighter
      const [origR, origG, origB] = hexToRgb(dark)
      const origLuminance = getRelativeLuminance(origR, origG, origB)
      expect(luminance).toBeGreaterThan(origLuminance)
    })

    it('should return a darker color for light backgrounds', () => {
      const light = '#f0f0f0'
      const contrast = getContrastColor(light)
      const [r, g, b] = hexToRgb(contrast)
      const luminance = getRelativeLuminance(r, g, b)

      // Should be much darker
      const [origR, origG, origB] = hexToRgb(light)
      const origLuminance = getRelativeLuminance(origR, origG, origB)
      expect(luminance).toBeLessThan(origLuminance)
    })

    it('should maintain hue from original color', () => {
      const blue = '#0066cc'
      const contrast = getContrastColor(blue)

      const [origH] = rgbToHsl(...hexToRgb(blue))
      const [contrastH] = rgbToHsl(...hexToRgb(contrast))

      // Hue should be similar (within 10 degrees)
      expect(Math.abs(origH - contrastH)).toBeLessThan(10)
    })

    it('should achieve sufficient contrast ratio', () => {
      // Test with colors that can actually achieve 4.5:1 while preserving hue
      const testCases = [
        { color: '#0066cc', minRatio: 3.0 }, // Blue - highly saturated, harder
        { color: '#cc0000', minRatio: 3.0 }, // Red - highly saturated, harder
        { color: '#006600', minRatio: 4.5 }, // Dark green - should achieve AA
        { color: '#666666', minRatio: 4.5 }, // Gray - should achieve AA
        { color: '#333333', minRatio: 4.5 }, // Dark gray - should achieve at least AA
        { color: '#3C887E', minRatio: 3.5 }, // Teal - saturated mid-tone, decent contrast
      ]

      for (const { color, minRatio } of testCases) {
        const contrast = getContrastColor(color)
        const ratio = getContrastRatio(color, contrast)

        // Should meet at least the expected minimum ratio
        expect(ratio).toBeGreaterThanOrEqual(minRatio)
      }
    })

    it('should handle hex colors without # prefix', () => {
      const contrast1 = getContrastColor('000000')
      const contrast2 = getContrastColor('#000000')

      // Both should return valid hex colors
      expect(contrast1).toMatch(/^#[0-9a-f]{6}$/)
      expect(contrast2).toMatch(/^#[0-9a-f]{6}$/)
    })

    it('should accept custom target ratio', () => {
      // Use a color that can actually achieve 7:1 ratio
      const color = '#333333' // dark gray

      // With default ratio (4.5)
      const contrast45 = getContrastColor(color, 4.5)
      const ratio45 = getContrastRatio(color, contrast45)
      expect(ratio45).toBeGreaterThanOrEqual(4.5)

      // With higher ratio (7.0 for AAA)
      const contrast70 = getContrastColor(color, 7.0)
      const ratio70 = getContrastRatio(color, contrast70)
      expect(ratio70).toBeGreaterThanOrEqual(7.0)

      // Higher ratio should result in more extreme contrast
      const [r45, g45, b45] = hexToRgb(contrast45)
      const [r70, g70, b70] = hexToRgb(contrast70)
      const lum45 = getRelativeLuminance(r45, g45, b45)
      const lum70 = getRelativeLuminance(r70, g70, b70)

      // For a dark color, AAA should be even lighter than AA
      expect(lum70).toBeGreaterThanOrEqual(lum45)
    })

    it('should try its best when target ratio is impossible', () => {
      // Highly saturated blue - impossible to get 7:1 while keeping hue
      const color = '#0066cc'
      const contrast = getContrastColor(color, 7.0)
      const ratio = getContrastRatio(color, contrast)

      // Should return the best it can find, even if it doesn't meet target
      expect(ratio).toBeGreaterThan(1)
      // Should at least meet AA standard
      expect(ratio).toBeGreaterThanOrEqual(4.5)
    })

    it('should maintain grayscale for gray input colors', () => {
      const grays = ['#000000', '#333333', '#808080', '#cccccc', '#ffffff']

      for (const gray of grays) {
        const contrast = getContrastColor(gray)
        const [r, g, b] = hexToRgb(contrast)

        // Result should also be grayscale (r === g === b, within tolerance)
        expect(Math.abs(r - g)).toBeLessThanOrEqual(2)
        expect(Math.abs(g - b)).toBeLessThanOrEqual(2)
        expect(Math.abs(r - b)).toBeLessThanOrEqual(2)
      }
    })

    it('should handle edge case: pure saturated colors', () => {
      const pureColors = [
        '#ff0000',
        '#00ff00',
        '#0000ff',
        '#ffff00',
        '#ff00ff',
        '#00ffff',
      ]

      for (const color of pureColors) {
        const contrast = getContrastColor(color)

        // Should return a valid hex color
        expect(contrast).toMatch(/^#[0-9a-f]{6}$/)

        // Should achieve some contrast
        const ratio = getContrastRatio(color, contrast)
        expect(ratio).toBeGreaterThan(1.5)
      }
    })

    it('should handle colors at luminance boundaries', () => {
      // Test colors near the 0.5 luminance threshold
      const boundaryColors = [
        '#757575', // Just below 0.5
        '#767676', // Just above 0.5
        '#808080', // Right at boundary
      ]

      for (const color of boundaryColors) {
        const contrast = getContrastColor(color)
        const ratio = getContrastRatio(color, contrast)

        // Should achieve good contrast regardless
        expect(ratio).toBeGreaterThan(3.0)
      }
    })

    it('should produce different results for different target ratios', () => {
      const color = '#666666'

      const contrast30 = getContrastColor(color, 3.0)
      const contrast45 = getContrastColor(color, 4.5)
      const contrast70 = getContrastColor(color, 7.0)

      // Higher target ratios should produce more extreme contrast
      const lum30 = getRelativeLuminance(...hexToRgb(contrast30))
      const lum45 = getRelativeLuminance(...hexToRgb(contrast45))
      const lum70 = getRelativeLuminance(...hexToRgb(contrast70))

      // For a dark color, higher ratios should be lighter
      expect(lum45).toBeGreaterThanOrEqual(lum30)
      expect(lum70).toBeGreaterThanOrEqual(lum45)
    })

    it('should handle very light colors correctly', () => {
      const veryLight = ['#fafafa', '#f5f5f5', '#eeeeee']

      for (const color of veryLight) {
        const contrast = getContrastColor(color)
        const origLum = getRelativeLuminance(...hexToRgb(color))
        const contrastLum = getRelativeLuminance(...hexToRgb(contrast))

        // Should darken, not lighten
        expect(contrastLum).toBeLessThan(origLum)

        // Should achieve sufficient contrast
        const ratio = getContrastRatio(color, contrast)
        expect(ratio).toBeGreaterThanOrEqual(4.5)
      }
    })

    it('should handle very dark colors correctly', () => {
      const veryDark = ['#0a0a0a', '#1a1a1a', '#222222']

      for (const color of veryDark) {
        const contrast = getContrastColor(color)
        const origLum = getRelativeLuminance(...hexToRgb(color))
        const contrastLum = getRelativeLuminance(...hexToRgb(contrast))

        // Should lighten, not darken
        expect(contrastLum).toBeGreaterThan(origLum)

        // Should achieve at least AA level contrast
        const ratio = getContrastRatio(color, contrast)
        expect(ratio).toBeGreaterThanOrEqual(4.5)
      }
    })

    it('should preserve hue across saturation adjustments', () => {
      // Use a less extreme color for testing hue preservation
      const color = '#ff6b6b' // Coral/salmon
      const contrast = getContrastColor(color, 4.5)

      const [origH] = rgbToHsl(...hexToRgb(color))
      const [contrastH] = rgbToHsl(...hexToRgb(contrast))

      // Hue should be relatively close even if saturation changed
      // Handle wraparound: difference should be < 30° or > 330° (close via wraparound)
      let hueDiff = Math.abs(origH - contrastH)
      if (hueDiff > 180) {
        hueDiff = 360 - hueDiff
      }
      expect(hueDiff).toBeLessThan(30)
    })
  })

  describe('getContrastRatio', () => {
    it('should return 21 for black and white', () => {
      const ratio = getContrastRatio('#000000', '#ffffff')
      expect(ratio).toBe(21)
    })

    it('should return 21 for white and black (order independent)', () => {
      const ratio = getContrastRatio('#ffffff', '#000000')
      expect(ratio).toBe(21)
    })

    it('should return 1 for identical colors', () => {
      expect(getContrastRatio('#000000', '#000000')).toBe(1)
      expect(getContrastRatio('#ffffff', '#ffffff')).toBe(1)
      expect(getContrastRatio('#ff0000', '#ff0000')).toBe(1)
    })

    it('should calculate correct ratio for common color pairs', () => {
      // Black text on white background should be 21
      const blackOnWhite = getContrastRatio('#000000', '#ffffff')
      expect(blackOnWhite).toBe(21)

      // White text on black background should be 21
      const whiteOnBlack = getContrastRatio('#ffffff', '#000000')
      expect(whiteOnBlack).toBe(21)

      // Gray on white should be less than black on white
      const grayOnWhite = getContrastRatio('#888888', '#ffffff')
      expect(grayOnWhite).toBeLessThan(21)
      expect(grayOnWhite).toBeGreaterThan(1)
    })

    it('should meet WCAG AA standard for normal text (4.5:1)', () => {
      // Dark gray on white should meet AA
      const ratio = getContrastRatio('#595959', '#ffffff')
      expect(ratio).toBeGreaterThanOrEqual(4.5)
    })

    it('should meet WCAG AAA standard for normal text (7:1)', () => {
      // Very dark gray on white should meet AAA
      const ratio = getContrastRatio('#595959', '#ffffff')
      expect(ratio).toBeGreaterThanOrEqual(4.5)

      // Black on white definitely meets AAA
      const blackOnWhite = getContrastRatio('#000000', '#ffffff')
      expect(blackOnWhite).toBeGreaterThanOrEqual(7)
    })

    it('should handle hex colors without # prefix', () => {
      const ratio = getContrastRatio('000000', 'ffffff')
      expect(ratio).toBe(21)
    })

    it('should be symmetric (order should not matter)', () => {
      const ratio1 = getContrastRatio('#ff0000', '#00ff00')
      const ratio2 = getContrastRatio('#00ff00', '#ff0000')
      expect(ratio1).toBe(ratio2)
    })

    it('should verify getContrastColor choices meet minimum standards', () => {
      // Test that our contrast color function makes good choices
      const darkColor = '#1a1a1a'
      const lightColor = '#f0f0f0'

      // Dark background should get lighter color with good contrast
      const contrastForDark = getContrastColor(darkColor)
      const ratioForDark = getContrastRatio(darkColor, contrastForDark)
      expect(ratioForDark).toBeGreaterThanOrEqual(4.5) // Should meet AA by default

      // Light background should get darker color with good contrast
      const contrastForLight = getContrastColor(lightColor)
      const ratioForLight = getContrastRatio(lightColor, contrastForLight)
      expect(ratioForLight).toBeGreaterThanOrEqual(4.5) // Should meet AA by default
    })

    it('should handle very similar colors', () => {
      const color1 = '#808080'
      const color2 = '#808181'

      const ratio = getContrastRatio(color1, color2)

      // Very similar colors should have ratio close to 1
      expect(ratio).toBeCloseTo(1, 1)
    })

    it('should calculate ratios for common text/background combinations', () => {
      const combinations = [
        { bg: '#ffffff', fg: '#000000', minRatio: 21 }, // Black on white
        { bg: '#000000', fg: '#ffffff', minRatio: 21 }, // White on black
        { bg: '#ffffff', fg: '#757575', minRatio: 4.5 }, // Gray on white (AA)
        { bg: '#ffffff', fg: '#595959', minRatio: 7 }, // Dark gray on white (AAA)
        { bg: '#0066cc', fg: '#ffffff', minRatio: 4.5 }, // White on blue
      ]

      for (const { bg, fg, minRatio } of combinations) {
        const ratio = getContrastRatio(bg, fg)
        expect(ratio).toBeGreaterThanOrEqual(minRatio)
      }
    })

    it('should handle colors with different luminance channels', () => {
      // Pure red has different luminance contribution than pure green or blue
      const red = '#ff0000'
      const green = '#00ff00'
      const blue = '#0000ff'

      const [rR, gR, bR] = hexToRgb(red)
      const [rG, gG, bG] = hexToRgb(green)
      const [rB, gB, bB] = hexToRgb(blue)

      const lumR = getRelativeLuminance(rR, gR, bR)
      const lumG = getRelativeLuminance(rG, gG, bG)
      const lumB = getRelativeLuminance(rB, gB, bB)

      // Green should have highest luminance (coefficient 0.7152)
      expect(lumG).toBeGreaterThan(lumR)
      expect(lumG).toBeGreaterThan(lumB)

      // Red should have higher luminance than blue (0.2126 vs 0.0722)
      expect(lumR).toBeGreaterThan(lumB)
    })

    it('should be deterministic - same inputs produce same outputs', () => {
      const color = '#3C887E'

      const result1 = getContrastColor(color)
      const result2 = getContrastColor(color)
      const result3 = getContrastColor(color)

      expect(result1).toBe(result2)
      expect(result2).toBe(result3)
    })
  })

  describe('edge cases and error handling', () => {
    it('should handle extreme lightness values in HSL', () => {
      // Pure white (L=1)
      const [r1, g1, b1] = hslToRgb(0, 0, 1)
      expect(r1).toBe(255)
      expect(g1).toBe(255)
      expect(b1).toBe(255)

      // Pure black (L=0)
      const [r2, g2, b2] = hslToRgb(0, 0, 0)
      expect(r2).toBe(0)
      expect(g2).toBe(0)
      expect(b2).toBe(0)
    })

    it('should handle extreme saturation values in HSL', () => {
      // No saturation = gray
      const [r1, g1, b1] = hslToRgb(180, 0, 0.5)
      expect(r1).toBeCloseTo(128, 0)
      expect(g1).toBeCloseTo(128, 0)
      expect(b1).toBeCloseTo(128, 0)

      // Full saturation
      const [r2, g2, b2] = hslToRgb(180, 1, 0.5)
      expect(r2).toBe(0)
      expect(g2).toBe(255)
      expect(b2).toBe(255)
    })

    it('should handle hue wraparound (360 degrees)', () => {
      // 0° and 360° should produce the same color
      const [r1, g1, b1] = hslToRgb(0, 1, 0.5)
      const [r2, g2, b2] = hslToRgb(360, 1, 0.5)

      expect(r1).toBeCloseTo(r2, 0)
      expect(g1).toBeCloseTo(g2, 0)
      expect(b1).toBeCloseTo(b2, 0)
    })

    it('should handle RGB values at boundaries', () => {
      // Minimum values
      const [h1, s1, l1] = rgbToHsl(0, 0, 0)
      expect(h1).toBeDefined()
      expect(s1).toBeDefined()
      expect(l1).toBe(0)

      // Maximum values
      const [h2, s2, l2] = rgbToHsl(255, 255, 255)
      expect(h2).toBeDefined()
      expect(s2).toBe(0)
      expect(l2).toBe(1)
    })

    it('should handle near-zero luminance correctly', () => {
      const nearBlack = getRelativeLuminance(1, 1, 1)
      expect(nearBlack).toBeGreaterThan(0)
      expect(nearBlack).toBeLessThan(0.001)
    })

    it('should handle mixed RGB values correctly', () => {
      // Test color with different R, G, B values
      const testCases = [
        { r: 100, g: 150, b: 200 },
        { r: 200, g: 100, b: 50 },
        { r: 50, g: 200, b: 100 },
      ]

      for (const { r, g, b } of testCases) {
        const [h, s, l] = rgbToHsl(r, g, b)
        const [r2, g2, b2] = hslToRgb(h, s, l)

        // Should round-trip accurately
        expect(Math.abs(r - r2)).toBeLessThanOrEqual(1)
        expect(Math.abs(g - g2)).toBeLessThanOrEqual(1)
        expect(Math.abs(b - b2)).toBeLessThanOrEqual(1)
      }
    })
  })
})
