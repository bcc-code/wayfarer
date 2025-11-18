<script setup lang="ts">
import { onKeyDown } from '@vueuse/core'

const themes = [
  { color: '#EA8C55', rounding: 1 },
  { color: '#5448C8', rounding: 0.5 },
  { color: '#BEEDAA', rounding: 0 },
]
function setTheme(theme: number) {
  const themeData = themes[theme]
  if (!themeData) return

  // Set dynamic project theme color
  document.documentElement.style.setProperty(
    '--color-accent-base',
    themeData.color,
  )

  // Calculate and set the on-accent contrast color
  const contrastColor = getContrastColor(themeData.color)
  document.documentElement.style.setProperty('--color-on-accent', contrastColor)

  // Set rounding
  document.documentElement.style.setProperty(
    '--radius-multiplier',
    themeData.rounding.toString(),
  )
}

onKeyDown('1', () => setTheme(0))
onKeyDown('2', () => setTheme(1))
onKeyDown('3', () => setTheme(2))
</script>
