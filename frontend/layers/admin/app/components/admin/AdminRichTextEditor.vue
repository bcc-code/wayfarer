<script setup lang="ts">
import type { EditorToolbarItem } from '@nuxt/ui'

withDefaults(
  defineProps<{
    /** What the field stores: a challenge description is HTML, a project's
     *  rules and info message are markdown. */
    contentType?: 'markdown' | 'html'
  }>(),
  { contentType: 'markdown' },
)

const modelValue = defineModel<string>({ default: '' })

const toolbarItems: EditorToolbarItem[][] = [
  [
    {
      kind: 'mark',
      mark: 'bold',
      icon: 'i-lucide-bold',
      tooltip: { text: 'Fet' },
    },
    {
      kind: 'mark',
      mark: 'italic',
      icon: 'i-lucide-italic',
      tooltip: { text: 'Kursiv' },
    },
    {
      kind: 'mark',
      mark: 'strike',
      icon: 'i-lucide-strikethrough',
      tooltip: { text: 'Gjennomstreking' },
    },
  ],
  [
    {
      kind: 'heading',
      level: 2,
      icon: 'i-lucide-heading-2',
      tooltip: { text: 'Overskrift 2' },
    },
    {
      kind: 'heading',
      level: 3,
      icon: 'i-lucide-heading-3',
      tooltip: { text: 'Overskrift 3' },
    },
  ],
  [
    {
      kind: 'bulletList',
      icon: 'i-lucide-list',
      tooltip: { text: 'Punktliste' },
    },
    {
      kind: 'orderedList',
      icon: 'i-lucide-list-ordered',
      tooltip: { text: 'Nummerert liste' },
    },
  ],
]
</script>

<template>
  <UEditor
    v-model="modelValue"
    :content-type="contentType"
    :image="false"
    :mention="false"
    class="border-accented overflow-hidden rounded-md border"
    :ui="{ base: 'min-h-[150px] p-3 sm:px-3 *:my-2' }"
  >
    <template #default="{ editor }">
      <UEditorToolbar
        :editor="editor"
        :items="toolbarItems"
        size="xs"
        class="bg-elevated border-accented border-b p-1"
      />
    </template>
  </UEditor>
</template>
