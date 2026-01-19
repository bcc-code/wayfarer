<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import z from 'zod'
import { toLocalDatetimeLocal, toISOString } from '~/utils/dates'

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
        ...BrandingFields
      }
      rules {
        markdown
        html
      }
      infoMessage {
        markdown
        html
      }
      infoMessageStart
      infoMessageEnd
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
    logo: z.string().nullish(),
    banner: z.string().nullish(),
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
  infoMessage: z.string().optional(),
  infoMessageStart: z.string().nullish(),
  infoMessageEnd: z.string().nullish(),
})
type Schema = z.infer<typeof schema>
const state = reactive<Schema>({
  name: '',
  description: undefined,
  startDate: '',
  endDate: '',
  branding: {
    logo: undefined,
    banner: undefined,
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
  infoMessage: '',
  infoMessageStart: undefined,
  infoMessageEnd: undefined,
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
      state.branding.banner = d.project.branding.banner ?? undefined
      state.branding.rounding = d.project.branding.rounding
      state.branding.colors = d.project.branding.colors
      state.rules = d.project.rules?.markdown
      state.infoMessage = d.project.infoMessage?.markdown
      state.infoMessageStart = toLocalDatetimeLocal(d.project.infoMessageStart)
      state.infoMessageEnd = toLocalDatetimeLocal(d.project.infoMessageEnd)
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

  // Convert nullish logo/banner to empty string so backend can clear them
  // Convert datetime-local values to ISO strings
  const input = {
    ...event.data,
    branding: {
      ...event.data.branding,
      logo: event.data.branding.logo ?? '',
      banner: event.data.branding.banner ?? '',
    },
    infoMessageStart: toISOString(event.data.infoMessageStart ?? undefined),
    infoMessageEnd: toISOString(event.data.infoMessageEnd ?? undefined),
  }

  executeMutation({ id: route.params.projectId, input }).then(
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
            { label: 'Prosjekter', to: { name: 'admin-projects' } },
            {
              label: state.name,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Rediger',
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
        <UFormField name="branding.logo" label="Logo" hint="(valgfritt)">
          <AdminFileUpload v-model="state.branding.logo" />
        </UFormField>
        <UFormField name="branding.banner" label="Banner" hint="(valgfritt)">
          <AdminFileUpload v-model="state.branding.banner" />
        </UFormField>
        <UFormField name="name" label="Navn">
          <UInput v-model="state.name" size="xl" required class="w-full" />
        </UFormField>
        <UFormField
          name="description"
          label="Beskrivelse"
          hint="(valgfritt)"
          help="Dette er kun for at admins skal ha bedre kontekst"
        >
          <UTextarea v-model="state.description" class="w-full" autoresize />
        </UFormField>
        <UFormField label="Prosjektvarighet">
          <DateRangeField
            v-model:start="state.startDate"
            v-model:end="state.endDate"
          />
        </UFormField>
        <UFormField v-if="data?.project" label="Fargetema">
          <AdminProjectThemeEditor
            v-model="state.branding.colors"
            :project-name="data?.project.name"
          />
        </UFormField>
        <UFormField
          name="rules"
          label="Prosjektregler"
          hint="(valgfritt)"
          help="Forklar hvordan brukere samler poeng"
        >
          <MarkdownEditor v-model="state.rules" />
        </UFormField>
        <UFormField
          name="infoMessage"
          label="Info-melding"
          hint="(valgfritt)"
          help="Vises som banner på forsiden. Brukere kan lukke den."
        >
          <MarkdownEditor v-model="state.infoMessage" />
        </UFormField>
        <UFormField
          name="infoMessageStart"
          label="Info-melding synlig fra"
          hint="(valgfritt)"
          help="Når info-meldingen skal begynne å vises"
        >
          <UInput
            v-model="state.infoMessageStart"
            type="datetime-local"
            size="xl"
            class="w-full"
          />
        </UFormField>
        <UFormField
          name="infoMessageEnd"
          label="Info-melding synlig til"
          hint="(valgfritt)"
          help="Når info-meldingen skal slutte å vises"
        >
          <UInput
            v-model="state.infoMessageEnd"
            type="datetime-local"
            size="xl"
            class="w-full"
          />
        </UFormField>
        <UButton type="submit" size="lg" block>Lagre endringer</UButton>
      </UForm>
    </UContainer>
  </div>
</template>
