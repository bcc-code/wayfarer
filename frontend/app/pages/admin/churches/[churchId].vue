<script setup lang="ts">
import { ChurchCategory } from '~/api/generated'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminChurchPage($id: ID!) {
    church(id: $id) {
      id
      name
      country
      category
    }
  }
`)

gql(`
  mutation UpdateChurch($id: ID!, $input: UpdateChurchInput!) {
    updateChurch(id: $id, input: $input) {
      id
      name
      country
      category
    }
  }
`)

const route = useRoute('admin-churches-churchId')

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminChurchPageQuery({
  variables: {
    id: route.params.churchId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation: updateChurch } = useUpdateChurchMutation()
const toast = useToast()

// Edit mode state
const isEditing = ref(false)
const editState = reactive({
  name: '',
  country: '',
  category: ChurchCategory.S as ChurchCategory,
})

const categoryOptions = [
  { label: 'Small (S)', value: ChurchCategory.S },
  { label: 'Large (L)', value: ChurchCategory.L },
  { label: 'Extra Large (XL)', value: ChurchCategory.Xl },
]

function startEditing() {
  if (data.value) {
    editState.name = data.value.church.name
    editState.country = data.value.church.country
    editState.category = data.value.church.category
    isEditing.value = true
  }
}

function cancelEditing() {
  isEditing.value = false
}

async function saveChanges() {
  const result = await updateChurch({
    id: route.params.churchId,
    input: {
      name: editState.name,
      country: editState.country,
      category: editState.category,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Failed to update church',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Church updated',
    color: 'success',
  })

  isEditing.value = false
  refetch({ requestPolicy: 'network-only' })
}
</script>

<template>
  <UContainer class="space-y-6 my-12">
    <div class="flex items-center justify-between">
      <h1>Church Details</h1>
      <UButton
        v-if="!isEditing && data"
        variant="soft"
        icon="i-heroicons-pencil"
        @click="startEditing"
      >
        Edit
      </UButton>
    </div>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <!-- View Mode -->
      <UCard v-if="!isEditing">
        <dl class="space-y-4">
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">ID</dt>
            <dd class="font-mono text-sm">{{ data.church.id }}</dd>
          </div>
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">Name</dt>
            <dd class="font-medium">{{ data.church.name }}</dd>
          </div>
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">Country</dt>
            <dd>{{ data.church.country }}</dd>
          </div>
          <div class="flex gap-6 py-2">
            <dt class="text-muted w-24 shrink-0">Category</dt>
            <dd>{{ data.church.category }}</dd>
          </div>
        </dl>
      </UCard>

      <!-- Edit Mode -->
      <UCard v-else>
        <div class="space-y-4">
          <UFormField label="Name">
            <UInput v-model="editState.name" class="w-full" />
          </UFormField>

          <UFormField label="Country">
            <UInput v-model="editState.country" class="w-full" />
          </UFormField>

          <UFormField label="Category">
            <USelect
              v-model="editState.category"
              :items="categoryOptions"
              class="w-full"
            />
          </UFormField>

          <div class="flex justify-end gap-2 pt-4">
            <UButton variant="ghost" @click="cancelEditing">Cancel</UButton>
            <UButton @click="saveChanges">Save Changes</UButton>
          </div>
        </div>
      </UCard>
    </template>
  </UContainer>
</template>
