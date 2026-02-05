<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
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
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Samtykker', to: { name: 'admin-consents' } },
            {
              label: data?.consent.title ?? route.params.consentId,
              to: {
                name: 'admin-consents-consentId',
                params: { consentId: route.params.consentId },
              },
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <div v-else-if="data" class="space-y-6">
        <!-- Consent Header -->
        <div class="flex items-start justify-between">
          <div>
            <div class="mb-2 flex items-center gap-3">
              <h1 class="text-3xl font-bold">{{ data.consent.title }}</h1>
              <UBadge variant="soft">
                v{{ data.consent.version }}
              </UBadge>
              <UBadge
                v-if="!data.consent.publishedAt"
                variant="soft"
                color="warning"
              >
                Utkast
              </UBadge>
            </div>
            <p class="text-dimmed">{{ data.consent.shortText }}</p>
          </div>
          <div class="flex gap-2">
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

        <!-- Edit Form -->
        <UCard v-if="isEditing">
          <template #header>
            <h2 class="text-xl font-semibold">Rediger samtykke</h2>
          </template>
          <div class="space-y-4">
            <UFormField label="Tittel">
              <UInput v-model="editState.title" class="w-full" />
            </UFormField>
            <UFormField label="Kort tekst">
              <UTextarea
                v-model="editState.shortText"
                class="w-full"
                autoresize
                placeholder="En kort beskrivelse som vises før brukere leser hele samtykket"
              />
            </UFormField>
            <UFormField label="Innhold (Markdown)">
              <UTextarea
                v-model="editState.body"
                class="w-full font-mono"
                :rows="10"
                autoresize
              />
            </UFormField>
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
          <template #footer>
            <div class="flex justify-end gap-3">
              <UButton variant="ghost" @click="cancelEditing">Avbryt</UButton>
              <UButton @click="saveChanges">Lagre endringer</UButton>
            </div>
          </template>
        </UCard>

        <!-- Consent Info -->
        <dl class="text-sm">
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Samtykke-ID</dt>
            <dd class="font-mono">{{ data.consent.id }}</dd>
          </div>
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Nøkkel</dt>
            <dd>
              <code class="bg-background-indent rounded px-2 py-1">
                {{ data.consent.key }}
              </code>
            </dd>
          </div>
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Versjon</dt>
            <dd class="font-medium">{{ data.consent.version }}</dd>
          </div>
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Publisert</dt>
            <dd v-if="data.consent.publishedAt" class="font-medium">
              {{ formatDateTime(data.consent.publishedAt) }}
            </dd>
            <dd v-else class="text-muted">Ikke publisert</dd>
          </div>
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Type</dt>
            <dd>
              <UBadge
                :color="data.consent.managementType === 'LOCAL' ? 'primary' : 'neutral'"
                variant="soft"
              >
                {{ data.consent.managementType }}
              </UBadge>
            </dd>
          </div>
          <div v-if="data.consent.managedBy" class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Administrert av</dt>
            <dd class="font-medium">{{ data.consent.managedBy }}</dd>
          </div>
          <div v-if="data.consent.url" class="flex gap-6 py-2">
            <dt class="text-muted w-24 shrink-0">URL</dt>
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
        </dl>

        <!-- Body Preview -->
        <UCard>
          <template #header>
            <h2 class="text-xl font-semibold">Forhåndsvisning av innhold</h2>
          </template>
          <div
            class="prose prose-sm dark:prose-invert max-w-none"
            v-html="data.consent.body.html"
          />
        </UCard>

      </div>
    </UContainer>
  </div>
</template>
