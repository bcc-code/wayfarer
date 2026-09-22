<script setup lang="ts">
definePageMeta({
  permission: 'consents:view',
  layout: 'admin',
})

gql(`
  query AdminConsentPage($id: ID!) {
    consent(id: $id) {
      id
      key
      version
      title
      shortText
      body {
        markdown
        html
      }
      url
      publishedAt
      managementType
      managedBy
      translationStatus {
        ...TranslationStatus
      }
    }
  }
`)

const route = useRoute('admin-consents-consentId')

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminConsentPageQuery({
  variables: {
    id: route.params.consentId,
  },
  pause: computed(() => !isAuthReady.value),
})

// Supplies the trailing breadcrumb crumb and the navbar title; everything
// above it is derived from the route.
useAdminPage(() => data.value?.consent.title)

const managementTypeLabels: Record<string, string> = {
  LOCAL: 'Lokal',
  REMOTE: 'Ekstern',
}

const { executeMutation: updateConsent } = useUpdateConsentMutation()
const toast = useToast()

// Edit mode state
const isEditing = ref(false)
const editState = reactive({
  title: '',
  shortText: '',
  body: '',
  url: '',
  publishedAt: null as string | null,
  managedBy: '',
})

function startEditing() {
  if (data.value) {
    editState.title = data.value.consent.title
    editState.shortText = data.value.consent.shortText
    editState.body = data.value.consent.body.markdown
    editState.url = data.value.consent.url ?? ''
    editState.publishedAt = data.value.consent.publishedAt ?? null
    editState.managedBy = data.value.consent.managedBy ?? ''
    isEditing.value = true
  }
}

function cancelEditing() {
  isEditing.value = false
}

async function saveChanges() {
  const result = await updateConsent({
    id: route.params.consentId,
    title: editState.title,
    shortText: editState.shortText,
    body: editState.body,
    url: editState.url || null,
    publishedAt: editState.publishedAt || null,
    managedBy: editState.managedBy || null,
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke oppdatere samtykke',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Samtykke oppdatert',
    color: 'success',
  })

  isEditing.value = false
  refetch({ requestPolicy: 'network-only' })
}

async function publishConsent() {
  const result = await updateConsent({
    id: route.params.consentId,
    publishedAt: new Date().toISOString(),
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke publisere samtykke',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Samtykke publisert',
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}
</script>

<template>
  <!-- Capped tighter than the data pages: this one is mostly prose, and the
       body preview at full panel width is unreadable. -->
  <div class="max-w-4xl">
    <AdminQueryState :fetching :error>
      <div v-if="data" class="space-y-8">
        <div class="flex items-start justify-between gap-4">
          <div class="flex flex-wrap items-center gap-3">
            <h1 class="text-3xl font-bold">{{ data.consent.title }}</h1>
            <UBadge variant="soft">v{{ data.consent.version }}</UBadge>
            <UBadge
              v-if="!data.consent.publishedAt"
              variant="soft"
              color="warning"
            >
              Utkast
            </UBadge>
          </div>
          <div class="flex shrink-0 gap-2">
            <UButton
              v-if="!data.consent.publishedAt"
              variant="soft"
              color="success"
              @click="publishConsent"
            >
              Publiser
            </UButton>
            <UButton v-if="!isEditing" variant="soft" @click="startEditing">
              Rediger
            </UButton>
          </div>
        </div>

        <AdminSection v-if="isEditing" title="Rediger samtykke">
          <div class="space-y-4">
            <AdminTranslatableFormField
              label="Tittel"
              :translation-status="data.consent.translationStatus"
              name="title"
            >
              <UInput v-model="editState.title" class="w-full" />
            </AdminTranslatableFormField>
            <AdminTranslatableFormField
              label="Kort tekst"
              :translation-status="data.consent.translationStatus"
              name="shortText"
            >
              <UTextarea
                v-model="editState.shortText"
                class="w-full"
                autoresize
                placeholder="En kort beskrivelse som vises før brukere leser hele samtykket"
              />
            </AdminTranslatableFormField>
            <AdminTranslatableFormField
              label="Innhold (Markdown)"
              :translation-status="data.consent.translationStatus"
              name="body"
            >
              <UTextarea
                v-model="editState.body"
                class="w-full font-mono"
                :rows="10"
                autoresize
              />
            </AdminTranslatableFormField>
            <UFormField label="URL (valgfritt)">
              <UInput
                v-model="editState.url"
                class="w-full"
                type="url"
                placeholder="https://..."
              />
            </UFormField>
            <UFormField label="Administreres av (valgfritt)">
              <UInput
                v-model="editState.managedBy"
                class="w-full"
                placeholder="Ekstern systemidentifikator"
              />
            </UFormField>
          </div>
          <!-- `AdminSection` has no footer slot; the actions sit at the end
               of the section instead. -->
          <div class="flex justify-end gap-3 pt-4">
            <UButton variant="ghost" @click="cancelEditing">Avbryt</UButton>
            <UButton @click="saveChanges">Lagre endringer</UButton>
          </div>
        </AdminSection>

        <!-- Both texts labelled: they are separate fields that often hold
             similar wording, and an unlabelled subtitle above the preview read
             as the same paragraph twice. -->
        <AdminSection v-else title="Tekst">
          <div class="space-y-4">
            <div>
              <p class="text-muted mb-1 text-xs">Kort tekst</p>
              <p>{{ data.consent.shortText }}</p>
            </div>
            <div>
              <p class="text-muted mb-1 text-xs">Fullstendig tekst</p>
              <div
                class="prose prose-sm dark:prose-invert max-w-none"
                v-html="data.consent.body.html"
              />
            </div>
          </div>
        </AdminSection>

        <AdminSection title="Detaljer">
          <dl class="divide-default divide-y text-sm">
            <div class="flex gap-6 py-2">
              <dt class="text-muted w-32 shrink-0">Nøkkel</dt>
              <dd>
                <code class="bg-elevated rounded px-2 py-1">
                  {{ data.consent.key }}
                </code>
              </dd>
            </div>
            <div class="flex gap-6 py-2">
              <dt class="text-muted w-32 shrink-0">Publisert</dt>
              <dd v-if="data.consent.publishedAt">
                {{ formatDateTime(data.consent.publishedAt) }}
              </dd>
              <dd v-else class="text-dimmed">Ikke publisert</dd>
            </div>
            <div class="flex gap-6 py-2">
              <dt class="text-muted w-32 shrink-0">Type</dt>
              <dd>
                {{
                  managementTypeLabels[data.consent.managementType] ??
                  data.consent.managementType
                }}
              </dd>
            </div>
            <div v-if="data.consent.managedBy" class="flex gap-6 py-2">
              <dt class="text-muted w-32 shrink-0">Administrert av</dt>
              <dd>{{ data.consent.managedBy }}</dd>
            </div>
            <div v-if="data.consent.url" class="flex gap-6 py-2">
              <dt class="text-muted w-32 shrink-0">URL</dt>
              <dd>
                <a
                  :href="data.consent.url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-primary hover:underline"
                >
                  {{ data.consent.url }}
                </a>
              </dd>
            </div>
            <!-- Last, matching the user and church pages: a support aid, not
                 the first thing a reader wants. -->
            <div class="flex gap-6 py-2">
              <dt class="text-muted w-32 shrink-0">Samtykke-ID</dt>
              <dd class="font-mono">{{ data.consent.id }}</dd>
            </div>
          </dl>
        </AdminSection>
      </div>
    </AdminQueryState>
  </div>
</template>
