<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      id
    }
  }
`)

const schema = z.object({
  name: z.string().nonempty({ error: 'Name is required' }),
  description: z.string().optional(),
  startDate: z.string().nonempty({ error: 'Start date is required' }),
  endDate: z.string().nonempty({ error: 'End date is required' }),
  branding: z.object({
    logo: z.string().optional(),
    colors: z.object({
      primary: z.string(),
      secondary: z.string(),
      tertiary: z.string(),
    }),
    rounding: z.number(),
  }),
})
type Schema = z.infer<typeof schema>
const state = reactive<Schema>({
  name: '',
  description: undefined,
  startDate: '',
  endDate: '',
  branding: {
    logo: undefined,
    rounding: 0,
    colors: {
      primary: '#000000',
      secondary: '#000000',
      tertiary: '#000000',
    },
  },
})

const { executeMutation } = useCreateProjectMutation()
const toast = useToast()

async function createProject(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({ input: event.data }).then((response) => {
    console.log(response)
    if (response.error) {
      toast.add({
        title: response.error.name,
        description: response.error.message,
        color: 'error',
      })
      return
    }
    if (!response.data) {
      return
    }
    navigateTo({
      name: 'admin-projects-projectId',
      params: { projectId: response.data.createProject.id },
    })
  })
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Projects', to: { name: 'admin-projects' } },
            {
              label: state.name || 'New project',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <UForm
        :state
        :schema="schema"
        class="flex max-w-md flex-col gap-6"
        @submit.prevent="createProject"
      >
        <UFormField name="name" label="Name">
          <UInput v-model="state.name" size="xl" required class="w-full" />
        </UFormField>
        <UFormField
          name="description"
          label="Description"
          hint="(optional)"
          help="This is only for admins to have better context"
        >
          <UTextarea v-model="state.description" class="w-full" autoresize />
        </UFormField>
        <DateRangeField
          v-model:start="state.startDate"
          v-model:end="state.endDate"
        />
        <UFormField label="Accent color">
          <ColorPickerInput v-model="state.branding.colors.primary" />
        </UFormField>
        <UButton type="submit" size="lg" block>Create project</UButton>
      </UForm>
    </UContainer>
  </div>
</template>
