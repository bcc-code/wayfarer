<script setup lang="ts">
definePageMeta({
  permission: 'settings:manage',
  layout: 'admin',
})

gql(`
  query AdminSettingsPage {
    settings {
      key
      value
      valueType
      description
      requiresRestart
      envVar
      editable
    }
    currentProject {
      id
      name
    }
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

gql(`
  mutation SetCurrentProject($projectId: ID!) {
    setCurrentProject(projectId: $projectId) {
      id
      name
    }
  }
`)

gql(`
  mutation SetSettings($input: [SettingInput!]!) {
    setSettings(input: $input) {
      key
      value
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching, executeQuery } = useAdminSettingsPageQuery({
  pause: computed(() => !isAuthReady.value),
})

const toast = useToast()
const { confirm } = useConfirm()
const { executeMutation: setCurrentProject, fetching: saving } =
  useSetCurrentProjectMutation()
const { executeMutation: setSettings, fetching: savingSettings } =
  useSetSettingsMutation()

const projects = computed(
  () => data.value?.projects.edges.map((edge) => edge.node) ?? [],
)
const currentProjectId = computed(() => data.value?.currentProject.id)

const { currentProjects, futureProjects, pastProjects } = useGroupedProjects(
  () => projects.value,
)

type SettingsProject = (typeof projects.value)[number]

function groupToItems(group: SettingsProject[]) {
  return group.map((project) => ({ label: project.name, value: project.id }))
}

// Nested arrays rather than inline `type: 'label'` separators: a label entry
// has no `value`, which makes `value-key` unassignable across the item union.
const projectItems = computed(() =>
  [currentProjects.value, futureProjects.value, pastProjects.value]
    .filter((group) => group.length > 0)
    .map(groupToItems),
)

const selectedProjectId = ref<string | undefined>()
watch(currentProjectId, (id) => (selectedProjectId.value = id), {
  immediate: true,
})

const hasChanged = computed(
  () =>
    !!selectedProjectId.value &&
    selectedProjectId.value !== currentProjectId.value,
)

const selectedProjectName = computed(
  () => projects.value.find((p) => p.id === selectedProjectId.value)?.name,
)

// The current project keeps its own confirm-and-save: it changes what every
// end user sees, which does not belong in a batch with log verbosity.
async function save() {
  if (!selectedProjectId.value || !hasChanged.value) return

  const confirmed = await confirm({
    title: `Bytte til "${selectedProjectName.value}"?`,
    description:
      'Dette endrer prosjektet alle brukere ser i appen, med en gang. Poeng, utfordringer, ledertavler og profilsiden byttes samtidig.',
    confirmLabel: 'Bytt prosjekt',
    color: 'primary',
    icon: 'lucide:triangle-alert',
  })
  if (!confirmed) return

  const response = await setCurrentProject({
    projectId: selectedProjectId.value,
  })

  if (response.error) {
    toast.add({
      title: response.error.name,
      description: response.error.message,
      color: 'error',
    })
    // Never leave the control showing a value the server refused.
    selectedProjectId.value = currentProjectId.value
    return
  }

  toast.add({
    title: 'Lagret',
    description: `${response.data?.setCurrentProject.name} er nå gjeldende prosjekt.`,
    color: 'success',
  })

  executeQuery({ requestPolicy: 'network-only' })
}

const configSettings = computed(
  () =>
    data.value?.settings.filter(
      (setting) => setting.key !== CURRENT_PROJECT_KEY,
    ) ?? [],
)

// Drafts are staged locally so several fields can be changed and committed
// together; the mutation applies them in one transaction.
const drafts = ref<Record<string, string>>({})

watch(
  configSettings,
  (settings) => {
    drafts.value = Object.fromEntries(
      settings.map((setting) => [
        setting.key,
        formatSetting(setting.valueType, setting.value),
      ]),
    )
  },
  { immediate: true },
)

const changed = computed(() =>
  configSettings.value.filter(
    (setting) =>
      setting.editable &&
      drafts.value[setting.key] !== undefined &&
      normalizeSetting(setting.valueType, drafts.value[setting.key]!) !==
        normalizeSetting(setting.valueType, setting.value),
  ),
)

const hasInvalid = computed(() =>
  changed.value.some(
    (setting) =>
      !isSettingValueValid(setting.valueType, drafts.value[setting.key]!),
  ),
)

const { markSaved } = useUnsavedChanges(() => drafts.value)

function discard() {
  drafts.value = Object.fromEntries(
    configSettings.value.map((setting) => [
      setting.key,
      formatSetting(setting.valueType, setting.value),
    ]),
  )
}

async function saveSettings() {
  if (!changed.value.length || hasInvalid.value) return

  const response = await setSettings({
    input: changed.value.map((setting) => ({
      key: setting.key,
      value: drafts.value[setting.key]!,
    })),
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
    title: 'Lagret',
    description: `${response.data?.setSettings.length} innstillinger oppdatert.`,
    color: 'success',
  })

  markSaved()
  executeQuery({ requestPolicy: 'network-only' })
}

const CURRENT_PROJECT_KEY = 'current_project_id'
</script>

<template>
  <div class="max-w-3xl space-y-8">
    <h1 class="text-3xl">Systeminnstillinger</h1>

    <AdminQueryState :fetching :error>
      <div v-if="data" class="space-y-8">
        <AdminSection title="Gjeldende prosjekt">
          <div class="space-y-4 py-1">
            <p class="text-muted text-sm">
              Prosjektet alle brukere ser i appen. Endringen gjelder umiddelbart
              for alle.
            </p>

            <UFormField label="Prosjekt">
              <USelectMenu
                v-model="selectedProjectId"
                :items="projectItems"
                value-key="value"
                placeholder="Velg prosjekt"
                searchable
                class="w-full"
              />
            </UFormField>

            <div class="flex items-center gap-3">
              <UButton
                :disabled="!hasChanged"
                :loading="saving"
                icon="lucide:check"
                @click="save"
              >
                Lagre
              </UButton>
              <UButton
                v-if="hasChanged"
                variant="ghost"
                color="neutral"
                @click="selectedProjectId = currentProjectId"
              >
                Avbryt
              </UButton>
              <span v-else class="text-dimmed text-sm">
                {{ data.currentProject.name }} er gjeldende.
              </span>
            </div>
          </div>
        </AdminSection>

        <AdminSection v-if="configSettings.length > 0" title="Konfigurasjon">
          <div class="space-y-1 py-1">
            <p class="text-muted text-sm">
              Verdiene her overstyrer miljøvariablene på serveren. En rad som er
              satt vinner; er den tom, gjelder miljøvariabelen.
            </p>
            <div class="divide-default divide-y">
              <AdminSettingField
                v-for="setting in configSettings"
                :key="setting.key"
                v-model="drafts[setting.key]!"
                :setting
                :disabled="savingSettings"
              />
            </div>

            <div
              v-if="changed.length > 0"
              class="flex flex-wrap items-center gap-3 pt-2"
            >
              <UButton
                icon="lucide:check"
                :loading="savingSettings"
                :disabled="hasInvalid"
                @click="saveSettings"
              >
                Lagre {{ changed.length }}
                {{ changed.length === 1 ? 'endring' : 'endringer' }}
              </UButton>
              <UButton
                variant="ghost"
                color="neutral"
                :disabled="savingSettings"
                @click="discard"
              >
                Forkast
              </UButton>
              <span v-if="hasInvalid" class="text-error text-sm">
                Rett opp de ugyldige feltene før du lagrer.
              </span>
            </div>
          </div>
        </AdminSection>
      </div>
    </AdminQueryState>
  </div>
</template>
