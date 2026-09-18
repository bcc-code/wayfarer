<script setup lang="ts">
const props = defineProps<{
  colors?: {
    light: {
      accent: string
      accentContrast: string
      onAccent: string
      backgroundDefault: string
      backgroundRaised: string
      backgroundIndent: string
      textDefault: string
      textMuted: string
      textHint: string
      shadowDefault: string
      shadowBlank: string
      borderDefault: string
    }
    dark: {
      accent: string
      accentContrast: string
      onAccent: string
      backgroundDefault: string
      backgroundRaised: string
      backgroundIndent: string
      textDefault: string
      textMuted: string
      textHint: string
      shadowDefault: string
      shadowBlank: string
      borderDefault: string
    }
  }
}>()

const colorMode = useColorMode()

const themeStyles = computed(() => {
  if (!props.colors) return {}

  const colors = colorMode.value === 'dark' ? props.colors.dark : props.colors.light

  return {
    '--color-accent': colors.accent,
    '--color-accent-contrast': colors.accentContrast,
    '--color-on-accent': colors.onAccent,
    '--color-background-default': colors.backgroundDefault,
    '--color-background-raised': colors.backgroundRaised,
    '--color-background-indent': colors.backgroundIndent,
    '--color-text-default': colors.textDefault,
    '--color-text-muted': colors.textMuted,
    '--color-text-hint': colors.textHint,
    '--color-shadow-default': colors.shadowDefault,
    '--color-shadow-blank': colors.shadowBlank,
    '--color-border-default': colors.borderDefault,
  }
})
</script>

<template>
  <div :style="themeStyles">
    <slot />
  </div>
</template>
