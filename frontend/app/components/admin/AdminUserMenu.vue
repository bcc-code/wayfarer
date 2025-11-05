<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

defineProps<{
  collapsed?: boolean
}>()

const colorMode = useColorMode()

const { me } = useAuth()

const items = computed<DropdownMenuItem[][]>(() => [
  [
    {
      type: 'label',
      label: me.value?.name,
      avatar: { src: me.value?.image ?? '', alt: me.value?.name },
    },
  ],
  [
    {
      label: colorMode.value === 'dark' ? 'Dark' : 'Light',
      icon: colorMode.value === 'dark' ? 'lucide:moon' : 'lucide:sun',
      children: [
        {
          label: 'Light',
          icon: 'lucide:sun',
          type: 'checkbox',
          checked: colorMode.value === 'light',
          onSelect(e: Event) {
            e.preventDefault()

            colorMode.preference = 'light'
          },
        },
        {
          label: 'Dark',
          icon: 'lucide:moon',
          type: 'checkbox',
          checked: colorMode.value === 'dark',
          onUpdateChecked(checked: boolean) {
            if (checked) {
              colorMode.preference = 'dark'
            }
          },
          onSelect(e: Event) {
            e.preventDefault()
          },
        },
      ],
    },
  ],
  [
    {
      label: 'Log out',
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
        avatar: { src: me?.image ?? '', alt: me?.name },
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

    <template #chip-leading="{ item }">
      <div class="inline-flex items-center justify-center shrink-0 size-5">
        <span
          class="rounded-full ring ring-bg bg-(--chip-light) dark:bg-(--chip-dark) size-2"
          :style="{
            '--chip-light': `var(--color-${(item as any).chip}-500)`,
            '--chip-dark': `var(--color-${(item as any).chip}-400)`,
          }"
        />
      </div>
    </template>
  </UDropdownMenu>
</template>
