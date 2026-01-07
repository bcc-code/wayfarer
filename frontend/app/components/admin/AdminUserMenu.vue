<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

defineProps<{
  collapsed?: boolean
}>()

const colorMode = useColorMode()
const themes: Record<string, { label: string; icon: string }> = {
  system: {
    label: 'System',
    icon: 'lucide:monitor',
  },
  dark: {
    label: 'Mørk',
    icon: 'lucide:moon',
  },
  light: {
    label: 'Lys',
    icon: 'lucide:sun',
  },
}

const { me } = useAuth()

const items = computed<DropdownMenuItem[][]>(() => [
  [
    {
      label: themes[colorMode.preference]?.label,
      icon: themes[colorMode.preference]?.icon,
      children: Object.entries(themes).map(([key, theme]) => ({
        label: theme.label,
        icon: theme.icon,
        type: 'checkbox',
        checked: colorMode.preference === key,
        onUpdateChecked(checked: boolean) {
          if (checked) {
            colorMode.preference = key
          }
        },
        onSelect(e: Event) {
          e.preventDefault()
        },
      })),
    },
  ],
  [
    {
      label: 'Logg ut',
      icon: 'lucide:log-out',
    },
  ],
])
</script>

<template>
  <UDropdownMenu
    :items="items"
    :content="{ align: 'center', collisionPadding: 12 }"
    :ui="{
      content: collapsed ? 'w-48' : 'w-(--reka-dropdown-menu-trigger-width)',
    }"
  >
    <UButton
      v-bind="{
        label: collapsed ? undefined : me?.name,
        trailingIcon: collapsed ? undefined : 'lucide:chevrons-up-down',
      }"
      color="neutral"
      variant="ghost"
      block
      :square="collapsed"
      class="data-[state=open]:bg-elevated"
      :ui="{
        trailingIcon: 'text-dimmed',
      }"
    />
  </UDropdownMenu>
</template>
