<script setup lang="ts">
import { Editor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import { Markdown } from '@tiptap/markdown'

const modelValue = defineModel<string>({ default: '' })

const editor = ref<Editor>()

onMounted(() => {
  editor.value = new Editor({
    extensions: [StarterKit, Markdown],
    content: modelValue.value,
    onUpdate: ({ editor }) => {
      modelValue.value = editor.getMarkdown()
    },
  })
})

watch(modelValue, (value) => {
  const currentMarkdown = editor.value?.getMarkdown()
  if (value === currentMarkdown) return
  editor.value?.commands.setContent(value, { contentType: 'markdown' })
})

onBeforeUnmount(() => {
  editor.value?.destroy()
})

const toolbarItems = [
  {
    icon: 'i-lucide-bold',
    action: () => editor.value?.chain().focus().toggleBold().run(),
    isActive: () => editor.value?.isActive('bold'),
    title: 'Fet',
  },
  {
    icon: 'i-lucide-italic',
    action: () => editor.value?.chain().focus().toggleItalic().run(),
    isActive: () => editor.value?.isActive('italic'),
    title: 'Kursiv',
  },
  {
    icon: 'i-lucide-strikethrough',
    action: () => editor.value?.chain().focus().toggleStrike().run(),
    isActive: () => editor.value?.isActive('strike'),
    title: 'Gjennomstreking',
  },
  { type: 'divider' },
  {
    icon: 'i-lucide-heading-2',
    action: () =>
      editor.value?.chain().focus().toggleHeading({ level: 2 }).run(),
    isActive: () => editor.value?.isActive('heading', { level: 2 }),
    title: 'Overskrift 2',
  },
  {
    icon: 'i-lucide-heading-3',
    action: () =>
      editor.value?.chain().focus().toggleHeading({ level: 3 }).run(),
    isActive: () => editor.value?.isActive('heading', { level: 3 }),
    title: 'Overskrift 3',
  },
  { type: 'divider' },
  {
    icon: 'i-lucide-list',
    action: () => editor.value?.chain().focus().toggleBulletList().run(),
    isActive: () => editor.value?.isActive('bulletList'),
    title: 'Punktliste',
  },
  {
    icon: 'i-lucide-list-ordered',
    action: () => editor.value?.chain().focus().toggleOrderedList().run(),
    isActive: () => editor.value?.isActive('orderedList'),
    title: 'Nummerert liste',
  },
]
</script>

<template>
  <div class="border-accented overflow-hidden rounded-md border">
    <div
      v-if="editor"
      class="bg-elevated border-accented flex flex-wrap gap-1 border-b p-1"
    >
      <template v-for="(item, index) in toolbarItems" :key="index">
        <div
          v-if="item.type === 'divider'"
          class="bg-accented mx-1 w-px self-stretch"
        />
        <UButton
          v-else
          :icon="item.icon"
          size="xs"
          square
          :variant="item.isActive?.() ? 'soft' : 'ghost'"
          :title="item.title"
          @click="
            () => {
              item.action?.()
            }
          "
        />
      </template>
    </div>
    <EditorContent
      :editor
      class="prose prose-sm dark:prose-invert max-w-none p-3 focus:outline-none [&_.ProseMirror]:min-h-[150px] [&_.ProseMirror]:outline-none"
    />
  </div>
</template>
