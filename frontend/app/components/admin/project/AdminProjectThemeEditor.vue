<script setup lang="ts">
const props = defineProps<{
  projectName: string
}>()

const modelValue = defineModel<Colors>({ required: true })

const toast = useToast()

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

// Convert camelCase to kebab-case
function toKebabCase(str: string): string {
  return str.replace(/([a-z])([A-Z])/g, '$1-$2').toLowerCase()
}

// Convert kebab-case to camelCase
function toCamelCase(str: string): string {
  return str.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase())
}

function toFileName(str: string): string {
  return str.split(' ').join('-').toLowerCase()
}

// Export theme as JSON file
function exportTheme() {
  const themeJson = {
    light: Object.fromEntries(
      Object.entries(modelValue.value.light)
        .filter(([key]) => key !== '__typename')
        .map(([key, value]) => [toKebabCase(key), value]),
    ),
    dark: Object.fromEntries(
      Object.entries(modelValue.value.dark)
        .filter(([key]) => key !== '__typename')
        .map(([key, value]) => [toKebabCase(key), value]),
    ),
  }
  console.log(themeJson)

  const blob = new Blob([JSON.stringify(themeJson, null, 2)], {
    type: 'application/json',
  })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${toFileName(props.projectName)}.theme.json`
  a.click()
  URL.revokeObjectURL(url)
}

// Import theme from JSON file
function importTheme() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = async (e) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return

    try {
      const text = await file.text()
      const json = JSON.parse(text)

      if (!json.light || !json.dark) {
        throw new Error('Invalid theme format: missing light or dark keys')
      }

      const convertColorSet = (
        colorSet: Record<string, string>,
      ): Record<string, string> => {
        return Object.fromEntries(
          Object.entries(colorSet).map(([key, value]) => [
            toCamelCase(key),
            value,
          ]),
        )
      }

      modelValue.value = {
        ...modelValue.value,
        light: {
          ...modelValue.value.light,
          ...convertColorSet(json.light),
        },
        dark: {
          ...modelValue.value.dark,
          ...convertColorSet(json.dark),
        },
      }

      toast.add({
        title: 'Theme imported',
        color: 'success',
      })
    } catch (err) {
      toast.add({
        title: 'Failed to import theme',
        description: err instanceof Error ? err.message : 'Invalid JSON file',
        color: 'error',
      })
    }
  }
  input.click()
}

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
      <div class="flex flex-col gap-6">
        <div class="mx-auto flex gap-8">
          <div>
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
          <div>
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
        <div class="flex justify-center gap-2">
          <UButton variant="soft" icon="i-lucide-upload" @click="importTheme">
            Import
          </UButton>
          <UButton variant="soft" icon="i-lucide-download" @click="exportTheme">
            Export
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
