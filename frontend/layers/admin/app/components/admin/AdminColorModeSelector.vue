<script setup lang="ts">
const colorMode = useColorMode()

const { t } = useI18n()
const items = computed(() => {
  return ['system', 'dark', 'light'].map((mode) => {
    return {
      label: t(`settings.colorModes.${mode}`),
      value: mode,
      onClick() {
        colorMode.preference = mode
      },
    }
  })
})

const getIcon = (mode: string) => {
  switch (mode) {
    case 'system':
      return 'lucide:monitor'
    case 'dark':
      return 'lucide:moon'
    case 'light':
      return 'lucide:sun'
  }
}
</script>

<template>
  <UDropdownMenu v-model="colorMode.preference" :items>
    <UButton
      :icon="getIcon(colorMode.preference)"
      square
      color="neutral"
      variant="soft"
    />
    <template #item="{ item }">
      <div class="flex items-center gap-2 justify-between w-full">
        <span>
          {{ item.label }}
        </span>
        <Icon v-if="item.value === colorMode.preference" name="lucide:check" />
      </div>
    </template>
  </UDropdownMenu>
</template>
