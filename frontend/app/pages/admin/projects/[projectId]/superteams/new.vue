<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminSuperTeamNewPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
    }
  }
`)

const route = useRoute('admin-projects-projectId-superteams-new')
const toast = useToast()

const { isAuthReady } = useAuthReady()
const { data } = useAdminSuperTeamNewPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation } = useCreateSuperTeamMutation()

const state = reactive({
  name: '',
  description: '',
  imageUrl: null as string | null,
})

const hasColor = ref(false)
const colorValue = ref('#000000')

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
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: 'Prosjekter',
              to: { name: 'admin-projects' },
            },
            {
              label: data?.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Superteams',
            },
            {
              label: 'Ny',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="max-w-2xl py-12">
      <h1 class="mb-6 text-2xl font-bold">Opprett superteam</h1>
      <form class="space-y-6" @submit.prevent="handleSubmit">
        <UFormField name="name" label="Navn" required>
          <UInput v-model="state.name" class="w-full" />
        </UFormField>

        <UFormField name="description" label="Beskrivelse" required>
          <UTextarea v-model="state.description" class="w-full" />
        </UFormField>

        <UFormField name="imageUrl" label="Bilde">
          <AdminFileUpload v-model="state.imageUrl" />
        </UFormField>

        <UFormField name="color" label="Farge">
          <div class="flex items-center gap-3">
            <UCheckbox v-model="hasColor" />
            <template v-if="hasColor">
              <ColorPickerInput v-model="colorValue" />
            </template>
            <span v-else class="text-muted text-sm">Ingen farge valgt</span>
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
    </UContainer>
  </div>
</template>
