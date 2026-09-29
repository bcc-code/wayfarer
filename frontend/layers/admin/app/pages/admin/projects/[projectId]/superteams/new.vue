<script setup lang="ts">
definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

// Trailing breadcrumb crumbs; the path above them is derived from the route.
useAdminPage(() => 'Ny')

const route = useRoute('admin-projects-projectId-superteams-new')
const toast = useToast()

const { executeMutation } = useCreateSuperTeamMutation()

const state = reactive({
  name: '',
  description: '',
  imageUrl: null as string | null,
})

const hasColor = ref(false)
const colorValue = ref('#000000')

const { markSaved } = useUnsavedChanges(() => ({
  ...state,
  hasColor: hasColor.value,
  colorValue: colorValue.value,
}))

async function handleSubmit() {
  const response = await executeMutation({
    projectId: route.params.projectId,
    input: {
      name: state.name,
      description: state.description,
      imageUrl: state.imageUrl || undefined,
      color: hasColor.value ? colorValue.value : undefined,
    },
  })

  if (response.error) {
    toast.add({
      title: response.error.name,
      description: response.error.message,
      color: 'error',
    })
    return
  }

  markSaved()
  toast.add({
    title: 'Suksess',
    description: 'Superteam opprettet',
    color: 'success',
  })

  navigateTo({
    name: 'admin-projects-projectId',
    params: { projectId: route.params.projectId },
    query: { tab: 'superteams' },
  })
}
</script>

<template>
  <div>
    <div class="max-w-2xl">
      <h1 class="mb-6 text-2xl font-bold">Opprett superteam</h1>
      <form class="space-y-6" @submit.prevent="handleSubmit">
        <UFormField name="name" label="Navn" required>
          <UInput v-model="state.name" class="w-full" />
        </UFormField>

        <UFormField name="imageUrl" label="Bilde">
          <AdminFileUpload v-model="state.imageUrl" />
        </UFormField>

        <UFormField name="color" label="Farge">
          <div class="space-y-3">
            <UCheckbox
              v-model="hasColor"
              label="Gi superteamet sin egen farge"
              description="Uten dette bruker superteamet fargene til prosjektet."
            />
            <ColorPickerInput v-if="hasColor" v-model="colorValue" />
          </div>
        </UFormField>

        <div class="flex gap-2">
          <UButton type="submit" :disabled="!state.name">
            Opprett superteam
          </UButton>
          <UButton
            variant="ghost"
            :to="{
              name: 'admin-projects-projectId',
              params: { projectId: route.params.projectId },
              query: { tab: 'superteams' },
            }"
          >
            Avbryt
          </UButton>
        </div>
      </form>
    </div>
  </div>
</template>
