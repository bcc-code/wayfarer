<script setup lang="ts">
import { refDebounced } from '@vueuse/core'

const props = defineProps<{
  /** The selected user's id, or an empty string. */
  modelValue: string
  /** Scopes the search to a project's participants. */
  projectId?: string
  placeholder?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [string] }>()

// `first: 20` is a search result list, not a page of data: the picker is for
// finding one known person, so more rows would not help.
gql(`
  query AdminUserPickerSearch($query: String, $projectId: ID) {
    users(first: 20, filter: { query: $query, projectId: $projectId }) {
      edges {
        node {
          id
          name
          image
          church {
            id
            name
          }
        }
      }
    }
  }
`)

/** Labels a value that arrived from outside, so an edit form can prefill it. */
gql(`
  query AdminUserPickerSelected($ids: [ID!]) {
    users(first: 1, filter: { ids: $ids }) {
      edges {
        node {
          id
          name
          image
          church {
            id
            name
          }
        }
      }
    }
  }
`)

type PickerUser = {
  id: string
  name: string
  image?: string | null
  church: { name: string }
}

const searchTerm = ref('')
const debouncedSearch = refDebounced(searchTerm, 300)

const { isAuthReady } = useAuthReady()
const { data, fetching } = useAdminUserPickerSearchQuery({
  variables: computed(() => ({
    query: debouncedSearch.value || undefined,
    projectId: props.projectId,
  })),
  pause: computed(() => !isAuthReady.value),
})

const results = computed<PickerUser[]>(
  () => data.value?.users.edges.map((edge) => edge.node) ?? [],
)

const selected = ref<PickerUser | undefined>()

const { data: selectedData } = useAdminUserPickerSelectedQuery({
  variables: computed(() => ({ ids: [props.modelValue] })),
  pause: computed(
    () =>
      !isAuthReady.value ||
      !props.modelValue ||
      selected.value?.id === props.modelValue,
  ),
})

watch(
  () => [props.modelValue, selectedData.value] as const,
  ([id, resolved]) => {
    if (!id) {
      selected.value = undefined
      return
    }
    if (selected.value?.id === id) return
    const node = resolved?.users.edges[0]?.node
    if (node?.id === id) selected.value = node
  },
  { immediate: true },
)

function onSelect(user: PickerUser | undefined) {
  selected.value = user
  emit('update:modelValue', user?.id ?? '')
}
</script>

<template>
  <!-- `ignore-filter`: the search runs server-side, so the menu must not also
       filter the 20 rows it was given. -->
  <USelectMenu
    v-model="selected"
    v-model:search-term="searchTerm"
    :items="results"
    :loading="fetching"
    ignore-filter
    label-key="name"
    :placeholder="placeholder ?? 'Søk etter navn...'"
    icon="lucide:user-search"
    class="w-full"
    @update:model-value="onSelect"
  >
    <template #item="{ item }">
      <div class="flex min-w-0 items-center gap-2">
        <UAvatar :src="item.image ?? undefined" :alt="item.name" size="2xs" />
        <span class="truncate">{{ item.name }}</span>
        <span class="text-dimmed truncate text-xs">{{ item.church.name }}</span>
      </div>
    </template>
    <template #empty>
      <span v-if="searchTerm">Ingen treff på «{{ searchTerm }}»</span>
      <span v-else>Skriv for å søke</span>
    </template>
  </USelectMenu>
</template>
