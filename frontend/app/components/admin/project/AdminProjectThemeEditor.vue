<script setup lang="ts">
const modelValue = defineModel<Colors>({ required: true })

const colorFields = [
  { key: 'accent', label: 'Accent' },
  { key: 'accentContrast', label: 'Accent Contrast' },
  { key: 'onAccent', label: 'On Accent' },
  { key: 'backgroundDefault', label: 'Background Default' },
  { key: 'backgroundRaised', label: 'Background Raised' },
  { key: 'backgroundIndent', label: 'Background Indent' },
  { key: 'textDefault', label: 'Text Default' },
  { key: 'textMuted', label: 'Text Muted' },
  { key: 'textHint', label: 'Text Hint' },
  { key: 'shadowDefault', label: 'Shadow Default' },
  { key: 'shadowBlank', label: 'Shadow Blank' },
  { key: 'borderDefault', label: 'Border Default' },
] as const

type ColorKey = (typeof colorFields)[number]['key']

const lightStyles = computed(() => {
  return {
    '--color-accent': modelValue.value.light.accent,
    '--color-accent-contrast': modelValue.value.light.accentContrast,
    '--color-on-accent': modelValue.value.light.onAccent,
    '--color-background-default': modelValue.value.light.backgroundDefault,
    '--color-background-raised': modelValue.value.light.backgroundRaised,
    '--color-background-indent': modelValue.value.light.backgroundIndent,
    '--color-text-default': modelValue.value.light.textDefault,
    '--color-text-muted': modelValue.value.light.textMuted,
    '--color-text-hint': modelValue.value.light.textHint,
    '--color-shadow-default': modelValue.value.light.shadowDefault,
    '--color-shadow-blank': modelValue.value.light.shadowBlank,
    '--color-border-default': modelValue.value.light.borderDefault,
  }
})

const darkStyles = computed(() => {
  return {
    '--color-accent': modelValue.value.dark.accent,
    '--color-accent-contrast': modelValue.value.dark.accentContrast,
    '--color-on-accent': modelValue.value.dark.onAccent,
    '--color-background-default': modelValue.value.dark.backgroundDefault,
    '--color-background-raised': modelValue.value.dark.backgroundRaised,
    '--color-background-indent': modelValue.value.dark.backgroundIndent,
    '--color-text-default': modelValue.value.dark.textDefault,
    '--color-text-muted': modelValue.value.dark.textMuted,
    '--color-text-hint': modelValue.value.dark.textHint,
    '--color-shadow-default': modelValue.value.dark.shadowDefault,
    '--color-shadow-blank': modelValue.value.dark.shadowBlank,
    '--color-border-default': modelValue.value.dark.borderDefault,
  }
})

function updateLightColor(key: ColorKey, value: string) {
  modelValue.value = {
    ...modelValue.value,
    light: {
      ...modelValue.value.light,
      [key]: value,
    },
  }
}

function updateDarkColor(key: ColorKey, value: string) {
  modelValue.value = {
    ...modelValue.value,
    dark: {
      ...modelValue.value.dark,
      [key]: value,
    },
  }
}
</script>

<template>
  <UModal title="Theme Editor" fullscreen>
    <UButton variant="soft" block>Open theme editor</UButton>

    <template #body>
      <div class="flex gap-8">
        <div class="grid flex-1 grid-cols-2 gap-6">
          <div>
            <h3 class="mb-4 text-lg font-semibold">Light Mode</h3>
            <div class="space-y-3">
              <UFormField
                v-for="field in colorFields"
                :key="field.key"
                :label="field.label"
              >
                <ColorPickerInput
                  :model-value="modelValue.light[field.key]"
                  @update:model-value="updateLightColor(field.key, $event)"
                />
              </UFormField>
            </div>
          </div>
          <div>
            <h3 class="mb-4 text-lg font-semibold">Dark Mode</h3>
            <div class="space-y-3">
              <UFormField
                v-for="field in colorFields"
                :key="field.key"
                :label="field.label"
              >
                <ColorPickerInput
                  :model-value="modelValue.dark[field.key]"
                  @update:model-value="updateDarkColor(field.key, $event)"
                />
              </UFormField>
            </div>
          </div>
        </div>
        <div class="flex shrink-0 gap-4">
          <div class="text-center">
            <p class="text-muted mb-2 text-sm">Light</p>
            <AdminProjectThemePreview :style="lightStyles" />
          </div>
          <div class="text-center">
            <p class="text-muted mb-2 text-sm">Dark</p>
            <AdminProjectThemePreview :style="darkStyles" />
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>
