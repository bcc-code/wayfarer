<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { ExternalContentType, ExternalContentSortBy } from '~/api/generated'
import { fuzzyMatch } from '~/utils/fuzzySearch'

interface ContentItem {
  id: string
  externalContent: {
    id: string
    title?: string | null
    contentType: ExternalContentType
    source: string
    publishedAt?: string | null
  }
}

const props = defineProps<{
  modelValue: ContentItem[]
}>()

const emit = defineEmits<{
  'update:modelValue': [items: ContentItem[]]
}>()

const searchQuery = ref('')
const contentTypeFilter = ref<ExternalContentType | null>(null)
const sourceFilter = ref('')

const contentTypeOptions = [
  { value: null, label: 'Alle typer' },
  { value: ExternalContentType.Media, label: 'Media' },
  { value: ExternalContentType.Song, label: 'Sang' },
  { value: ExternalContentType.BookChapter, label: 'Bokkapittel' },
  { value: ExternalContentType.Article, label: 'Artikkel' },
  { value: ExternalContentType.BibleVerse, label: 'Bibelvers' },
  { value: ExternalContentType.Quiz, label: 'Quiz' },
  { value: ExternalContentType.Text, label: 'Tekst' },
]

// Debounced search
const debouncedSearch = refDebounced(searchQuery, 300)

const queryVariables = computed(() => ({
  filter: {
    ...(contentTypeFilter.value && { contentType: contentTypeFilter.value }),
    ...(sourceFilter.value && { source: sourceFilter.value }),
  },
  sortBy: ExternalContentSortBy.PublishedAtDesc,
  first: 500,
}))

const { data, fetching } = useAdminExternalContentsQuery({
  variables: queryVariables,
})

const searchResults = computed(() => {
  if (!data.value?.externalContents?.edges) return []

  const selectedIds = new Set(
    props.modelValue.map((item) => item.externalContent.id),
  )
  const query = debouncedSearch.value.trim()

  // Filter and score items
  const scoredItems = data.value.externalContents.edges
    .map((edge) => edge.node)
    .filter((node) => !selectedIds.has(node.id))
    .map((node) => {
      if (!query) {
        return { node, score: 0 }
      }

      // Try matching against title and id
      const titleScore = node.title ? fuzzyMatch(query, node.title) : null
      const idScore = fuzzyMatch(query, node.id)

      // Use the best score
      const bestScore = Math.max(titleScore ?? -1, idScore ?? -1)

      return { node, score: bestScore }
    })
    .filter((item) => !query || (item.score !== null && item.score >= 0))

  // Only sort by score when searching, otherwise keep backend order (PublishedAtDesc)
  if (query) {
    scoredItems.sort((a, b) => b.score - a.score)
  }
  return scoredItems.map((item) => item.node)
})

function addItem(externalContent: (typeof searchResults.value)[0]) {
  const newItem: ContentItem = {
    id: `temp-${externalContent.id}`,
    externalContent: {
      id: externalContent.id,
      title: externalContent.title,
      contentType: externalContent.contentType,
      source: externalContent.source,
      publishedAt: externalContent.publishedAt,
    },
  }
  emit('update:modelValue', [...props.modelValue, newItem])
}

function removeItem(index: number) {
  const newItems = [...props.modelValue]
  newItems.splice(index, 1)
  emit('update:modelValue', newItems)
}

function handleReorder() {
  // The v-model binding on VueDraggable automatically updates the array order
  emit('update:modelValue', [...props.modelValue])
}

function formatContentType(type: ExternalContentType): string {
  const labels: Record<ExternalContentType, string> = {
    [ExternalContentType.Media]: 'Media',
    [ExternalContentType.Song]: 'Sang',
    [ExternalContentType.BookChapter]: 'Bokkapittel',
    [ExternalContentType.Article]: 'Artikkel',
    [ExternalContentType.BibleVerse]: 'Bibelvers',
    [ExternalContentType.Quiz]: 'Quiz',
    [ExternalContentType.Text]: 'Tekst',
    [ExternalContentType.ExternalLink]: 'Ekstern lenke',
  }
  return labels[type] || type
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- Search and Filters -->
    <div class="flex gap-3">
      <UInput
        v-model="searchQuery"
        placeholder="Søk etter innhold..."
        icon="lucide:search"
        class="flex-1"
      />
      <USelect
        v-model="contentTypeFilter"
        :items="contentTypeOptions"
        class="w-48"
      />
    </div>

    <!-- Search Results -->
    <div class="border-default max-h-96 overflow-y-auto rounded-lg border">
      <div v-if="fetching" class="text-muted p-4 text-center text-sm">
        Laster...
      </div>
      <div
        v-else-if="searchResults.length === 0"
        class="text-muted p-4 text-center text-sm"
      >
        {{ debouncedSearch ? 'Ingen treff' : 'Ingen innhold tilgjengelig' }}
      </div>
      <div v-else>
        <button
          v-for="item in searchResults"
          :key="item.id"
          type="button"
          class="border-default hover:bg-elevated flex w-full items-center gap-3 border-b px-4 py-2 text-left last:border-b-0"
          @click="addItem(item)"
        >
          <UIcon name="lucide:plus" class="text-primary size-4 shrink-0" />
          <div class="min-w-0 flex-1">
            <div
              class="truncate text-sm font-medium"
              :title="item.title || item.id"
            >
              {{ item.title || item.id }}
            </div>
          </div>
          <UBadge variant="subtle" size="sm">
            {{ formatContentType(item.contentType) }}
          </UBadge>
          <span v-if="item.publishedAt" class="text-muted text-xs">
            {{ formatDate(item.publishedAt) }}
          </span>
        </button>
      </div>
    </div>

    <!-- Selected Items -->
    <div>
      <div class="text-muted mb-2 text-sm font-medium">
        Valgte elementer ({{ modelValue.length }})
        <span v-if="modelValue.length > 1" class="font-normal">
          - dra for å sortere
        </span>
      </div>
      <div class="border-default rounded-lg border">
        <VueDraggable
          :model-value="modelValue"
          handle=".drag-handle"
          ghost-class="opacity-50"
          :animation="200"
          @update:model-value="
            (items: ContentItem[]) => emit('update:modelValue', items)
          "
          @end="handleReorder"
        >
          <div
            v-for="(item, index) in modelValue"
            :key="item.id"
            class="border-default flex items-center gap-3 border-b px-4 py-2 last:border-b-0"
          >
            <div
              class="drag-handle text-muted cursor-grab active:cursor-grabbing"
            >
              <UIcon name="lucide:grip-vertical" class="size-4" />
            </div>
            <span class="text-muted w-6 text-sm">{{ index + 1 }}.</span>
            <div class="min-w-0 flex-1">
              <div class="truncate text-sm font-medium">
                {{ item.externalContent.title || item.externalContent.id }}
              </div>
            </div>
            <UBadge variant="subtle" size="sm">
              {{ formatContentType(item.externalContent.contentType) }}
            </UBadge>
            <span
              v-if="item.externalContent.publishedAt"
              class="text-muted text-xs"
            >
              {{ formatDate(item.externalContent.publishedAt) }}
            </span>
            <UButton
              variant="ghost"
              color="error"
              size="xs"
              icon="lucide:x"
              @click="removeItem(index)"
            />
          </div>
        </VueDraggable>
        <div
          v-if="modelValue.length === 0"
          class="text-muted py-6 text-center text-sm"
        >
          Ingen innhold valgt. Søk og legg til elementer ovenfor.
        </div>
      </div>
    </div>
  </div>
</template>
