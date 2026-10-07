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
  mutation SetSetting($key: String!, $value: String!) {
    setSetting(key: $key, value: $value) {
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
const { executeMutation: setSetting } = useSetSettingMutation()

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

const configSettings = computed(
  () =>
    data.value?.settings.filter(
      (setting) => setting.key !== CURRENT_PROJECT_KEY,
    ) ?? [],
)

const savingKey = ref<string | undefined>()

async function saveSetting(key: string, value: string) {
  savingKey.value = key
  const response = await setSetting({ key, value })
  savingKey.value = undefined

  if (response.error) {
    toast.add({
      title: response.error.name,
      description: response.error.message,
      color: 'error',
    })
    return
  }

  toast.add({ title: 'Lagret', description: key, color: 'success' })
  executeQuery({ requestPolicy: 'network-only' })
}

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
                :setting
                :saving="savingKey === setting.key"
                @save="(value) => saveSetting(setting.key, value)"
              />
            </div>
          </div>
        </AdminSection>
      </div>
    </AdminQueryState>
  </div>
</template>
