<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    /**
     * Lets the preview be used. Off by default: a preview renders the
     * participant app's real components, links and buttons included, and
     * clicking one would navigate the admin into the participant app.
     * `AdminAchievementPreview` is still a replica with a state switcher
     * inside it, and needs this until it is rebuilt on the real badge.
     */
    interactive?: boolean
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
  }>(),
  { interactive: false },
)

const colorMode = useColorMode()

const themeStyles = computed(() => {
  if (!props.colors) return {}

  const colors =
    colorMode.value === 'dark' ? props.colors.dark : props.colors.light

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
  <!-- A width and the project's palette, nothing else. 390px is the viewport
       the user layer's breakpoints are drawn for, so the content is judged at
       a width a participant will actually see. No surface of its own: a card
       already carries one, and a frame around a frame is just two borders. A
       preview that needs a background paints it itself. -->
  <div :inert="!interactive" :style="themeStyles" class="w-[390px] max-w-full">
    <slot />
  </div>
</template>
