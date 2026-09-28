<script setup lang="ts">
import { RoleType, ScopeType } from '~/api/generated'

// Scope pickers, loaded only while the dialog is open.
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

/** The page owns the query and refetches. */
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

// The role determines its scope, so the dialog does not ask. Nothing validates
// the pairing server-side, so a mismatch would be stored if offered.
const ROLE_SCOPE: Partial<Record<RoleType, ScopeType>> = {
  [RoleType.ChurchAdmin]: ScopeType.Church,
  [RoleType.ProjectAdmin]: ScopeType.Project,
  [RoleType.TeamLead]: ScopeType.Team,
}

// `RoleScope` resolves these server-side. The id is the fallback only: a scope
// pointing at something deleted still has to render.
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

// A half-filled scope must not travel to a role that does not use it.
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

// Teams via their project: `TeamFilter` has no free-text field, and one
// project can hold over a thousand teams.
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

/** The id for whichever scope this role needs. */
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
  <AdminSection title="Roller og tillatelser" :count="roles.length">
    <template #actions>
      <UButton
        v-if="canManage"
        icon="i-lucide-plus"
        size="sm"
        @click="showAddModal = true"
      >
        Legg til rolle
      </UButton>
    </template>

    <div v-if="roles.length" class="space-y-1">
      <div
        v-for="role in roles"
        :key="role.id"
        class="flex items-center justify-between gap-4 py-3"
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
    <p v-else class="text-dimmed text-sm">Ingen roller tildelt</p>

    <UModal v-model:open="showAddModal">
      <template #header>
        <h3 class="text-lg font-semibold">Legg til rolle</h3>
      </template>

      <template #body>
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
          <!-- Without its scope, a scoped role would be assigned globally. -->
          <UButton :disabled="!canSubmit" @click="handleAssign">
            Tildel rolle
          </UButton>
        </div>
      </template>
    </UModal>
  </AdminSection>
</template>
