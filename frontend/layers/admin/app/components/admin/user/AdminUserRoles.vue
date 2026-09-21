<script setup lang="ts">
import { RoleType, ScopeType } from '~/api/generated'

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

const scopeTypeOptions = [
  { label: 'Ingen (Global)', value: null },
  { label: 'Menighet', value: ScopeType.Church },
  { label: 'Prosjekt', value: ScopeType.Project },
  { label: 'Lag', value: ScopeType.Team },
]

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
  scopeType: null as ScopeType | null,
  scopeId: '',
})

function resetForm() {
  newRole.role = RoleType.User
  newRole.scopeType = null
  newRole.scopeId = ''
}

function closeAddModal() {
  showAddModal.value = false
  resetForm()
}

async function handleAssign() {
  const result = await assignRole({
    input: {
      userId: props.userId,
      role: newRole.role,
      scopeType: newRole.scopeType,
      scopeId:
        newRole.scopeType && newRole.scopeId ? newRole.scopeId : undefined,
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
        <div class="space-y-4">
          <UFormField label="Rolle">
            <USelect
              v-model="newRole.role"
              :items="roleOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Omfangstype">
            <USelect
              v-model="newRole.scopeType"
              :items="scopeTypeOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>

          <UFormField v-if="newRole.scopeType" label="Omfangs-ID">
            <UInput
              v-model="newRole.scopeId"
              :placeholder="`Skriv inn ${newRole.scopeType.toLowerCase()}-ID`"
              class="w-full"
            />
          </UFormField>
        </div>
      </template>

      <template #footer>
        <div class="flex w-full justify-end gap-3">
          <UButton variant="ghost" color="neutral" @click="closeAddModal">
            Avbryt
          </UButton>
          <UButton @click="handleAssign">Tildel rolle</UButton>
        </div>
      </template>
    </UModal>
  </UCard>
</template>
