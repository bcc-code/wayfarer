<script setup lang="ts">
import { RoleType, ScopeType } from '~/api/generated'

/**
 * Pickers for the scope a role needs. Loaded only while the dialog is open —
 * there is no reason to fetch every church on a page view.
 */
gql(`
  query AdminUserRoleScopeOptions {
    churches(first: 500) {
      edges {
        node {
          id
          name
        }
      }
    }
    projects(first: 200, filter: { archived: false }) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`)

gql(`
  query AdminUserRoleTeamOptions($projectId: ID!) {
    teams(first: 500, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`)

interface RoleScope {
  id: string
  type: ScopeType
  church?: { id: string; name: string } | null
  project?: { id: string; name: string } | null
  team?: { id: string; name: string } | null
}

interface UserRole {
  id: string
  role: RoleType
  scope?: RoleScope | null
}

const props = defineProps<{
  userId: string
  roles: UserRole[]
  canManage?: boolean
}>()

/** The page owns the query, so it refetches when the role set changes. */
const emit = defineEmits<{ changed: [] }>()

const roleLabels: Record<RoleType, string> = {
  [RoleType.User]: 'Bruker',
  [RoleType.Admin]: 'Admin',
  [RoleType.Superadmin]: 'Superadmin',
  [RoleType.ChurchAdmin]: 'Menighetsadmin',
  [RoleType.ProjectAdmin]: 'Prosjektadmin',
  [RoleType.TeamLead]: 'Lagleder',
  [RoleType.M2M]: 'M2M',
}

const roleOptions = Object.entries(roleLabels).map(([value, label]) => ({
  label,
  value: value as RoleType,
}))

/**
 * The scope a role takes, derived from the role itself.
 *
 * It was a second dropdown the admin had to answer, with exactly one correct
 * answer per role — and nothing stopped them picking a wrong one, so
 * "Menighetsadmin scoped to a Prosjekt" was expressible and would have been
 * stored. The pairing comes from the service: its own tests assign
 * `ChurchAdmin` with a church id, `ProjectAdmin` with a project id, `TeamLead`
 * with a team id and `Admin` with none (`internal/services/roles_test.go`).
 */
const ROLE_SCOPE: Partial<Record<RoleType, ScopeType>> = {
  [RoleType.ChurchAdmin]: ScopeType.Church,
  [RoleType.ProjectAdmin]: ScopeType.Project,
  [RoleType.TeamLead]: ScopeType.Team,
}

/**
 * What a role is scoped to, by name. `RoleScope` resolves `church`/`project`/
 * `team` server-side, so this says "Østfold" where the card used to print
 * `CH01K9VZ865699692N7FVTXYR4AQ`.
 *
 * The id is the *fallback*, not the default: a scope pointing at something
 * deleted still has to render, and then the raw id is the only honest thing
 * left to show.
 */
function scopeLabel(scope: RoleScope): string {
  return (
    scope.church?.name ?? scope.project?.name ?? scope.team?.name ?? scope.id
  )
}

const { executeMutation: assignRole } = useAssignRoleMutation()
const { executeMutation: revokeRole } = useRevokeRoleMutation()
const toast = useToast()

const showAddModal = ref(false)
const newRole = reactive({
  role: RoleType.User as RoleType,
  churchId: '',
  projectId: '',
  teamId: '',
})

const requiredScope = computed(() => ROLE_SCOPE[newRole.role])

function resetForm() {
  newRole.role = RoleType.User
  newRole.churchId = ''
  newRole.projectId = ''
  newRole.teamId = ''
}

// Changing the role changes which scope applies, so a half-filled one from the
// previous role must not travel with it.
watch(
  () => newRole.role,
  () => {
    newRole.churchId = ''
    newRole.projectId = ''
    newRole.teamId = ''
  },
)

const { isAuthReady } = useAuthReady()

const { data: scopeOptions } = useAdminUserRoleScopeOptionsQuery({
  pause: computed(() => !isAuthReady.value || !showAddModal.value),
  requestPolicy: 'cache-first',
})

const churchItems = computed(() =>
  (scopeOptions.value?.churches.edges ?? []).map((edge) => ({
    label: edge.node.name,
    value: edge.node.id,
  })),
)

const projectItems = computed(() =>
  (scopeOptions.value?.projects.edges ?? []).map((edge) => ({
    label: edge.node.name,
    value: edge.node.id,
  })),
)

/**
 * Teams are picked through their project rather than from one global list:
 * `TeamFilter` has no free-text field, so a flat list of every team could not
 * be searched server-side, and a single project can hold over a thousand.
 */
const { data: teamOptions } = useAdminUserRoleTeamOptionsQuery({
  variables: computed(() => ({ projectId: newRole.projectId })),
  pause: computed(
    () =>
      !isAuthReady.value ||
      !showAddModal.value ||
      requiredScope.value !== ScopeType.Team ||
      !newRole.projectId,
  ),
})

const teamItems = computed(() =>
  (teamOptions.value?.teams.edges ?? []).map((edge) => ({
    label: edge.node.name,
    value: edge.node.id,
  })),
)

/** The id the mutation should carry, for whichever scope the role needs. */
const scopeId = computed(() => {
  switch (requiredScope.value) {
    case ScopeType.Church:
      return newRole.churchId
    case ScopeType.Project:
      return newRole.projectId
    case ScopeType.Team:
      return newRole.teamId
    default:
      return ''
  }
})

/** A scoped role without its scope would be assigned globally by mistake. */
const canSubmit = computed(() => !requiredScope.value || !!scopeId.value)

function closeAddModal() {
  showAddModal.value = false
  resetForm()
}

async function handleAssign() {
  if (!canSubmit.value) return

  const result = await assignRole({
    input: {
      userId: props.userId,
      role: newRole.role,
      scopeType: requiredScope.value ?? null,
      scopeId: scopeId.value || undefined,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke tildele rolle',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Rolle tildelt',
    description: `Tildelte rollen ${roleLabels[newRole.role]}`,
    color: 'success',
  })

  closeAddModal()
  emit('changed')
}

async function handleRevoke(role: UserRole) {
  const result = await revokeRole({
    input: {
      userId: props.userId,
      role: role.role,
      scopeType: role.scope?.type ?? undefined,
      scopeId: role.scope?.id ?? undefined,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke fjerne rolle',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Rolle fjernet',
    description: `Fjernet rollen ${roleLabels[role.role]}`,
    color: 'success',
  })

  emit('changed')
}
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <h2 class="text-xl font-semibold">Roller og tillatelser</h2>
        <UButton
          v-if="canManage"
          icon="i-lucide-plus"
          size="sm"
          @click="showAddModal = true"
        >
          Legg til rolle
        </UButton>
      </div>
    </template>

    <div v-if="roles.length" class="space-y-3">
      <div
        v-for="role in roles"
        :key="role.id"
        class="border-default flex items-center justify-between rounded-md border p-3"
      >
        <div class="flex items-center gap-3">
          <UBadge variant="soft" size="lg">
            {{ roleLabels[role.role] ?? role.role }}
          </UBadge>
          <div v-if="role.scope" class="text-sm">
            <span class="text-dimmed">
              {{ capitalizeFirst(role.scope.type) }}
            </span>
            <span class="ml-2 font-medium">{{ scopeLabel(role.scope) }}</span>
          </div>
        </div>
        <UButton
          v-if="canManage"
          icon="i-lucide-trash-2"
          color="error"
          variant="ghost"
          size="sm"
          :aria-label="`Fjern rollen ${roleLabels[role.role] ?? role.role}`"
          @click="handleRevoke(role)"
        />
      </div>
    </div>
    <div v-else class="text-dimmed">Ingen roller tildelt</div>

    <UModal v-model:open="showAddModal">
      <template #header>
        <h3 class="text-lg font-semibold">Legg til rolle</h3>
      </template>

      <template #body>
        <!--
          One question, then the scope it implies. The dialog used to ask for a
          scope *type* the role already determines, and then for the scope's
          **ULID as free text** — "Skriv inn church-ID" — which meant leaving
          the page to find an id and pasting it back.
        -->
        <div class="space-y-4">
          <UFormField label="Rolle">
            <USelect
              v-model="newRole.role"
              :items="roleOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>

          <UFormField
            v-if="requiredScope === ScopeType.Church"
            label="Menighet"
          >
            <USelectMenu
              v-model="newRole.churchId"
              :items="churchItems"
              value-key="value"
              placeholder="Velg menighet"
              searchable
              class="w-full"
            />
          </UFormField>

          <UFormField
            v-if="
              requiredScope === ScopeType.Project ||
              requiredScope === ScopeType.Team
            "
            label="Prosjekt"
          >
            <USelectMenu
              v-model="newRole.projectId"
              :items="projectItems"
              value-key="value"
              placeholder="Velg prosjekt"
              searchable
              class="w-full"
            />
          </UFormField>

          <!--
            Teams are reached through their project: `TeamFilter` has no
            free-text field, so one global list could not be searched
            server-side, and a single project can hold over a thousand teams.
          -->
          <UFormField v-if="requiredScope === ScopeType.Team" label="Lag">
            <USelectMenu
              v-model="newRole.teamId"
              :items="teamItems"
              value-key="value"
              :disabled="!newRole.projectId"
              :placeholder="
                newRole.projectId ? 'Velg lag' : 'Velg prosjekt først'
              "
              searchable
              class="w-full"
            />
          </UFormField>

          <p v-if="!requiredScope" class="text-muted text-sm">
            Denne rollen gjelder globalt og trenger ikke noe omfang.
          </p>
        </div>
      </template>

      <template #footer>
        <div class="flex w-full justify-end gap-3">
          <UButton variant="ghost" color="neutral" @click="closeAddModal">
            Avbryt
          </UButton>
          <!--
            Disabled until the scope is chosen: a scoped role submitted without
            one would be assigned globally, which is a much larger grant than
            the admin asked for.
          -->
          <UButton :disabled="!canSubmit" @click="handleAssign">
            Tildel rolle
          </UButton>
        </div>
      </template>
    </UModal>
  </UCard>
</template>
