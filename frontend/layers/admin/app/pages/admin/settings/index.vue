<script setup lang="ts">
import { SettingValueType } from '~/api/generated'

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
      editable
      updatedAt
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

const { isAuthReady } = useAuthReady()
const { data, error, fetching, executeQuery } = useAdminSettingsPageQuery({
  pause: computed(() => !isAuthReady.value),
})

const toast = useToast()
const { confirm } = useConfirm()
const { executeMutation: setCurrentProject, fetching: saving } =
  useSetCurrentProjectMutation()

const projects = computed(
  () => data.value?.projects.edges.map((edge) => edge.node) ?? [],
)
const currentProjectId = computed(() => data.value?.currentProject.id)

// Grouped the same way the project switcher groups them, so "which one is
// running right now" is as obvious here as it is in the sidebar.
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

// Mirrors the backend's selection: everything else in the table duplicates an
// environment variable the server reads instead, and `editable` says so.
const otherSettings = computed(
  () => data.value?.settings.filter((setting) => !setting.editable) ?? [],
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
    // Back to what the server actually holds, so the control never shows a
    // value that was refused.
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

const VALUE_TYPE_LABELS: Record<SettingValueType, string> = {
  [SettingValueType.Text]: 'tekst',
  [SettingValueType.Int]: 'heltall',
  [SettingValueType.Bool]: 'av/på',
  [SettingValueType.Float]: 'desimaltall',
  [SettingValueType.Json]: 'JSON',
}
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

        <AdminSection
          v-if="otherSettings.length > 0"
          title="Øvrige innstillinger"
        >
          <div class="space-y-3 py-1">
            <p class="text-muted text-sm">
              Disse leses fra miljøvariabler på serveren, ikke herfra. Verdiene
              under er radene i databasen og har ingen effekt — de vises bare
              for innsyn.
            </p>
            <dl class="divide-default divide-y">
              <div
                v-for="setting in otherSettings"
                :key="setting.key"
                class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 py-2"
              >
                <div>
                  <dt class="font-medium">{{ setting.key }}</dt>
                  <dd v-if="setting.description" class="text-muted text-sm">
                    {{ setting.description }}
                  </dd>
                </div>
                <div class="text-right">
                  <code class="text-dimmed text-sm">{{ setting.value }}</code>
                  <p class="text-dimmed text-xs">
                    {{ VALUE_TYPE_LABELS[setting.valueType] }} · endret
                    {{ formatDateTime(setting.updatedAt) }}
                  </p>
                </div>
              </div>
            </dl>
          </div>
        </AdminSection>
      </div>
    </AdminQueryState>
  </div>
</template>
