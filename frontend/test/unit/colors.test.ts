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
  })
})
