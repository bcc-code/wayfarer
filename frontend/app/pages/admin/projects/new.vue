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
  name: z.string().nonempty(),
  description: z.string().optional(),
  startDate: z.string().nonempty(),
  endDate: z.string().nonempty(),
  branding: z
    .object({
      logo: z.string(),
      colors: z.object({
        primary: z.string(),
        secondary: z.string(),
        tertiary: z.string(),
      }),
      rounding: z.number(),
    })
    .optional(),
})
type Schema = z.infer<typeof schema>
const state = reactive<Schema>({
  name: '',
  description: '',
  startDate: '',
  endDate: '',
  branding: {
    logo: '',
    colors: {
      primary: '',
      secondary: '',
      tertiary: '',
    },
    rounding: 0,
  },
})

const { executeMutation } = useCreateProjectMutation()

async function createProject(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({ input: event.data }).then((response) => {
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
    <div class="py-2 border-b border-default">
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
    <UContainer>
      <UForm
        :state
        :schema="schema"
        class="py-12 space-y-6"
        @submit.prevent="createProject"
      >
        <UFormField name="name" label="Name" required>
          <UInput
            v-model="state.name"
            placeholder="New project"
            size="xl"
            type="text"
            autofocus
            required
          />
        </UFormField>
        <UFormField name="description" label="Description">
          <UTextarea v-model="state.description" class="w-sm" autoresize />
        </UFormField>
        <div class="gap-4 flex">
          <UFormField name="startDate" label="Starts at" required>
            <UInput v-model="state.startDate" type="datetime-local" required />
          </UFormField>
          <UFormField name="endDate" label="Ends at" required>
            <UInput v-model="state.endDate" type="datetime-local" required />
          </UFormField>
        </div>
        <UButton type="submit">Create project</UButton>
      </UForm>
    </UContainer>
  </div>
</template>
