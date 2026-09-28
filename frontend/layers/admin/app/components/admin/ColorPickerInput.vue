<script setup lang="ts">
const color = defineModel<string>({ required: true })

type ColorFormat = 'hex' | 'rgba'

const format = ref<ColorFormat>(color.value.startsWith('rgb') ? 'rgba' : 'hex')

// Parse color to extract RGB and alpha
function parseColor(colorStr: string): {
  r: number
  g: number
  b: number
  a: number
} {
  // Modern format: rgb(r g b / a) or rgba(r g b / a)
  const modernMatch = colorStr.match(
    /rgba?\((\d+)\s+(\d+)\s+(\d+)(?:\s*\/\s*([\d.]+))?\)/,
  )
  if (modernMatch) {
    return {
      r: parseInt(modernMatch[1]!),
      g: parseInt(modernMatch[2]!),
      b: parseInt(modernMatch[3]!),
      a: modernMatch[4] ? parseFloat(modernMatch[4]) : 1,
    }
  }

  // Legacy format: rgba(r, g, b, a) or rgb(r, g, b)
  const legacyMatch = colorStr.match(
    /rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)/,
  )
  if (legacyMatch) {
    return {
      r: parseInt(legacyMatch[1]!),
      g: parseInt(legacyMatch[2]!),
      b: parseInt(legacyMatch[3]!),
      a: legacyMatch[4] ? parseFloat(legacyMatch[4]) : 1,
    }
  }

  let hex = colorStr.replace('#', '')
  if (hex.length === 3) {
    hex = hex
      .split('')
      .map((c) => c + c)
      .join('')
  }
  if (hex.length === 6) {
    hex += 'ff'
  }

  return {
    r: parseInt(hex.slice(0, 2), 16) || 0,
    g: parseInt(hex.slice(2, 4), 16) || 0,
    b: parseInt(hex.slice(4, 6), 16) || 0,
    a: (parseInt(hex.slice(6, 8), 16) || 255) / 255,
  }
}

function toHex(r: number, g: number, b: number): string {
  return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`
}

function toRgb(r: number, g: number, b: number, a: number): string {
  return `rgb(${r} ${g} ${b} / ${a.toFixed(2)})`
}

const hexColor = computed({
  get() {
    const { r, g, b } = parseColor(color.value)
    return toHex(r, g, b)
  },
  set(newHex: string) {
    const { a } = parseColor(color.value)
    const hex = newHex.replace('#', '')
    const r = parseInt(hex.slice(0, 2), 16) || 0
    const g = parseInt(hex.slice(2, 4), 16) || 0
    const b = parseInt(hex.slice(4, 6), 16) || 0
    color.value = format.value === 'rgba' ? toRgb(r, g, b, a) : toHex(r, g, b)
  },
})

const opacity = computed({
  get() {
    const { a } = parseColor(color.value)
    return Math.round(a * 100)
  },
  set(newOpacity: number) {
    const { r, g, b } = parseColor(color.value)
    color.value = toRgb(r, g, b, newOpacity / 100)
  },
})

function switchFormat(newFormat: ColorFormat) {
  format.value = newFormat
  const { r, g, b, a } = parseColor(color.value)
  if (newFormat === 'hex') {
    color.value = toHex(r, g, b)
  } else {
    color.value = toRgb(r, g, b, a)
  }
}

const chip = computed(() => ({ backgroundColor: color.value }))
</script>

<template>
  <UInput v-model="color">
    <template #leading>
      <UPopover :ui="{ content: 'p-1' }">
        <button :style="chip" class="border-accented size-4 rounded border" />
        <template #content>
          <div class="flex flex-col gap-2 p-2">
            <UColorPicker v-model="hexColor" />
            <div class="flex items-center gap-2">
              <UTabs
                :items="[
                  { label: 'HEX', value: 'hex' },
                  { label: 'RGBA', value: 'rgba' },
                ]"
                :model-value="format"
                size="xs"
                class="flex-1"
                @update:model-value="switchFormat($event as ColorFormat)"
              />
            </div>
            <div v-if="format === 'rgba'" class="flex items-center gap-2">
              <USlider
                v-model="opacity"
                :min="0"
                :max="100"
                :step="1"
                class="flex-1"
              />
              <span class="text-muted w-8 text-right text-xs"
                >{{ opacity }}%</span
              >
            </div>
          </div>
        </template>
      </UPopover>
    </template>
  </UInput>
</template>
