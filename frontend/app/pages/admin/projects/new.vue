<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

const schema = z.object({
  name: z.string().nonempty({ error: 'Name is required' }),
  description: z.string().optional(),
  startDate: z.string().nonempty({ error: 'Start date is required' }),
  endDate: z.string().nonempty({ error: 'End date is required' }),
  branding: z.object({
    logo: z.string().optional(),
    colors: z.object({
      light: z.object({
        accent: z.string(),
        accentContrast: z.string(),
        onAccent: z.string(),
        backgroundDefault: z.string(),
        backgroundRaised: z.string(),
        backgroundIndent: z.string(),
        textDefault: z.string(),
        textMuted: z.string(),
        textHint: z.string(),
        shadowDefault: z.string(),
        shadowBlank: z.string(),
        borderDefault: z.string(),
      }),
      dark: z.object({
        accent: z.string(),
        accentContrast: z.string(),
        onAccent: z.string(),
        backgroundDefault: z.string(),
        backgroundRaised: z.string(),
        backgroundIndent: z.string(),
        textDefault: z.string(),
        textMuted: z.string(),
        textHint: z.string(),
        shadowDefault: z.string(),
        shadowBlank: z.string(),
        borderDefault: z.string(),
      }),
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
      light: {
        accent: '#ffaedf',
        accentContrast: '#5d0045',
        onAccent: '#5d0045',
        backgroundDefault: '#efefef',
        backgroundRaised: '#ffffff',
        backgroundIndent: 'rgb(0 0 0 / 0.05)',
        textDefault: '#202020',
        textMuted: 'rgb(32 32 32 / 0.7)',
        textHint: 'rgb(32 32 32 / 0.4)',
        shadowDefault: 'rgb(0 0 0 / 0.1)',
        shadowBlank: 'rgb(0 0 0 / 0)',
        borderDefault: 'rgb(0 0 0 / 0.15)',
      },
      dark: {
        accent: '#ffaedf',
        accentContrast: '#ffaedf',
        onAccent: '#5d0045',
        backgroundDefault: '#222222',
        backgroundRaised: '#343434',
        backgroundIndent: 'rgb(0 0 0 / 0.25)',
        textDefault: '#f5f5f5',
        textMuted: 'rgb(255 255 255 / 0.7)',
        textHint: 'rgb(255 255 255 / 0.4)',
        shadowDefault: 'rgb(0 0 0 / 0.3)',
        shadowBlank: 'rgb(0 0 0 / 0)',
        borderDefault: 'rgb(255 255 255 / 0.09)',
      },
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
        <UFormField label="Color Theme">
          <AdminProjectThemeEditor
            v-model="state.branding.colors"
            :project-name="state.name || 'New Project'"
          />
        </UFormField>
        <UButton type="submit" size="lg" block>Create project</UButton>
      </UForm>
    </UContainer>
  </div>
</template>
