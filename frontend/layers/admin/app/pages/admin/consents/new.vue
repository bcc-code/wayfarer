<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import { z } from 'zod'

definePageMeta({
  permission: 'consents:view',
  layout: 'admin',
})

// Trailing breadcrumb crumbs; the path above them is derived from the route.
useAdminPage(() => 'Nytt samtykke')

gql(`
  query AdminNewConsentProjects {
    projects(first: 100, filter: { archived: false }) {
      edges {
        node {
          id
          name
          startDate
          endDate
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data: projectsData } = useAdminNewConsentProjectsQuery({
  pause: computed(() => !isAuthReady.value),
})

const projects = computed(
  () => projectsData.value?.projects.edges.map((edge) => edge.node) ?? [],
)

const { currentProjects, futureProjects, pastProjects } = useGroupedProjects(
  () => projects.value,
)

type ConsentProject = (typeof projects.value)[number]

function groupToItems(group: ConsentProject[]) {
  return group.map((project) => ({ label: project.name, value: project.id }))
}

// A sentinel rather than '': reka-ui's Combobox reserves the empty string for
// "cleared", and an item carrying it throws. Sent to the server as no
// projectId at all, which is what makes a consent apply to every project.
const GLOBAL_PROJECT = '__all__'

const projectItems = computed(() => [
  [{ label: 'Alle prosjekter', value: GLOBAL_PROJECT }],
  ...[currentProjects.value, futureProjects.value, pastProjects.value]
    .filter((group) => group.length > 0)
    .map(groupToItems),
])

const schema = z.object({
  key: z
    .string()
    .min(1, 'Nøkkel er påkrevd')
    .regex(
      /^[a-z0-9_-]+$/,
      'Nøkkel må være små bokstaver og tall med understrek eller bindestrek',
    ),
  title: z.string().min(1, 'Tittel er påkrevd'),
  shortText: z.string().optional(),
  body: z.string().min(1, 'Innhold er påkrevd'),
  url: z.url('Må være en gyldig URL').optional().or(z.literal('')),
  publishNow: z.boolean().default(false),
  isRemote: z.boolean().default(false),
  isRequired: z.boolean().default(false),
  managedBy: z.string().optional(),
  projectId: z.string().default(GLOBAL_PROJECT),
})

type Schema = z.infer<typeof schema>

const state = reactive<Schema>({
  key: '',
  title: '',
  shortText: '',
  body: '',
  url: '',
  publishNow: false,
  isRemote: false,
  isRequired: false,
  managedBy: '',
  projectId: GLOBAL_PROJECT,
})

const { markSaved } = useUnsavedChanges(() => ({ ...state }))

const { executeMutation: createConsent } = useCreateConsentMutation()
const toast = useToast()

async function handleSubmit(event: FormSubmitEvent<Schema>) {
  markSaved()
  if (!event.data) return

  const result = await createConsent({
    key: event.data.key,
    title: event.data.title,
    shortText: event.data.shortText || undefined,
    body: event.data.body,
    url: event.data.url || undefined,
    publishedAt: event.data.publishNow ? new Date().toISOString() : undefined,
    isRemote: event.data.isRemote || undefined,
    isRequired: event.data.isRequired || undefined,
    managedBy: event.data.managedBy || undefined,
    projectId:
      event.data.projectId === GLOBAL_PROJECT
        ? undefined
        : event.data.projectId,
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke opprette samtykke',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Samtykke opprettet',
    color: 'success',
  })

  const consentId = result.data?.createConsent.id
  if (consentId) {
    navigateTo({
      name: 'admin-consents-consentId',
      params: { consentId },
    })
  } else {
    navigateTo({ name: 'admin-consents' })
  }
}
</script>

<template>
  <div>
    <div>
      <div class="max-w-2xl">
        <h1 class="mb-6 text-3xl font-bold">Opprett nytt samtykke</h1>

        <UForm :state :schema @submit="handleSubmit">
          <div class="space-y-6">
            <UFormField name="key" label="Nøkkel" required>
              <UInput
                v-model="state.key"
                class="w-full"
                placeholder="personvern"
                description="Unik identifikator for dette samtykket. Bruk små bokstaver med understrek."
              />
            </UFormField>

            <UFormField name="projectId" label="Prosjekt">
              <USelectMenu
                v-model="state.projectId"
                :items="projectItems"
                value-key="value"
                class="w-full"
              />
              <template #description>
                Hvilket prosjekt samtykket gjelder for. Kan ikke endres senere.
              </template>
            </UFormField>

            <AdminTranslatableFormField name="title" label="Tittel">
              <UInput
                v-model="state.title"
                class="w-full"
                placeholder="Personvernerklæring"
              />
            </AdminTranslatableFormField>

            <AdminTranslatableFormField name="shortText" label="Kort tekst">
              <UTextarea
                v-model="state.shortText"
                class="w-full"
                autoresize
                placeholder="En kort beskrivelse som vises til brukere før de leser hele samtykket"
              />
            </AdminTranslatableFormField>

            <AdminTranslatableFormField name="body" label="Innhold (Markdown)">
              <UTextarea
                v-model="state.body"
                class="w-full font-mono"
                :rows="10"
                autoresize
                placeholder="# Personvernerklæring&#10;&#10;Din markdown-innhold her..."
              />
            </AdminTranslatableFormField>

            <UFormField name="url" label="Ekstern URL">
              <UInput
                v-model="state.url"
                class="w-full"
                type="url"
                placeholder="https://example.com/personvern"
              />
            </UFormField>

            <USeparator />

            <div class="space-y-4">
              <UFormField name="isRemote">
                <UCheckbox
                  v-model="state.isRemote"
                  label="Ekstern administrasjon"
                />
                <template #description>
                  Aktiver hvis dette samtykket administreres av et eksternt
                  system
                </template>
              </UFormField>

              <UFormField
                v-if="state.isRemote"
                name="managedBy"
                label="Administreres av"
              >
                <UInput
                  v-model="state.managedBy"
                  class="w-full"
                  placeholder="Eksternt systemnavn"
                />
              </UFormField>

              <UFormField name="isRequired">
                <UCheckbox v-model="state.isRequired" label="Påkrevd" />
                <template #description>
                  Må godtas for å delta i prosjektet, og kan ikke avslås
                </template>
              </UFormField>

              <UFormField name="publishNow">
                <UCheckbox
                  v-model="state.publishNow"
                  label="Publiser umiddelbart"
                />
                <template #description>
                  Hvis ikke valgt, lagres samtykket som utkast
                </template>
              </UFormField>
            </div>

            <div class="flex justify-end gap-3 pt-4">
              <UButton variant="ghost" :to="{ name: 'admin-consents' }">
                Avbryt
              </UButton>
              <UButton type="submit"> Opprett samtykke </UButton>
            </div>
          </div>
        </UForm>
      </div>
    </div>
  </div>
</template>
