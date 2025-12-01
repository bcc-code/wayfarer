<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectEditPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
      description
      startDate
      endDate
      archivedAt
      branding {
        logo
        rounding
        colors {
          light {
            accent
            accentContrast
            onAccent
            backgroundDefault
            backgroundRaised
            backgroundIndent
            textDefault
            textMuted
            textHint
            shadowDefault
            shadowBlank
            borderDefault
          }
          dark {
            accent
            accentContrast
            onAccent
            backgroundDefault
            backgroundRaised
            backgroundIndent
            textDefault
            textMuted
            textHint
            shadowDefault
            shadowBlank
            borderDefault
          }
        }
      }
      rules {
        markdown
        html
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-edit')

const { isAuthReady } = useAuthReady()
const { data } = useAdminProjectEditPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
  pause: computed(() => !isAuthReady.value),
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
  rules: z.string().optional(),
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
        accent: '',
        accentContrast: '',
        onAccent: '',
        backgroundDefault: '',
        backgroundRaised: '',
        backgroundIndent: '',
        textDefault: '',
        textMuted: '',
        textHint: '',
        shadowDefault: '',
        shadowBlank: '',
        borderDefault: '',
      },
      dark: {
        accent: '',
        accentContrast: '',
        onAccent: '',
        backgroundDefault: '',
        backgroundRaised: '',
        backgroundIndent: '',
        textDefault: '',
        textMuted: '',
        textHint: '',
        shadowDefault: '',
        shadowBlank: '',
        borderDefault: '',
      },
    },
  },
  rules: '',
})

watch(
  () => data.value,
  (d) => {
    if (d) {
      state.name = d.project.name
      state.description = d.project.description
      state.startDate = d.project.startDate
      state.endDate = d.project.endDate
      state.branding.logo = d.project.branding.logo ?? undefined
      state.branding.rounding = d.project.branding.rounding
      state.branding.colors = d.project.branding.colors
      state.rules = d.project.rules?.markdown
    }
  },
  { once: true },
)

const { executeMutation } = useUpdateProjectMutation()
const toast = useToast()

async function updateProject(event: FormSubmitEvent<Schema>) {
  if (!event.data) {
    return
  }

  executeMutation({ id: route.params.projectId, input: event.data }).then(
    (response) => {
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
        params: { projectId: response.data.updateProject.id },
      })
    },
  )
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
              label: state.name,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Edit',
              to: {
                name: 'admin-projects-projectId-edit',
                params: { projectId: route.params.projectId },
              },
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <UForm
        :state
        :schema="schema"
        class="flex max-w-md flex-col gap-8"
        @submit.prevent="updateProject"
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
        <UFormField v-if="data?.project" label="Color Theme">
          <AdminProjectThemeEditor
            v-model="state.branding.colors"
            :project-name="data?.project.name"
          />
        </UFormField>
        <UFormField
          name="rules"
          label="Project Rules"
          hint="(optional)"
          help="Explain how users collect points"
        >
          <MarkdownEditor v-model="state.rules" />
        </UFormField>
        <UButton type="submit" size="lg" block>Save changes</UButton>
      </UForm>
    </UContainer>
  </div>
</template>
