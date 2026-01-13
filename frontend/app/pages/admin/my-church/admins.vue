<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query ChurchAdminsPage($churchId: ID!) {
    usersWithRole(role: CHURCH_ADMIN, scopeType: CHURCH, scopeId: $churchId) {
      id
      name
      email
    }
    users(filter: { churchId: $churchId }, first: 500) {
      edges {
        node {
          id
          name
          email
        }
      }
    }
  }
`)

const { me } = useAuth()
const { isAuthReady } = useAuthReady()
const toast = useToast()
const { t } = useI18n()

const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useChurchAdminsPageQuery({
  variables: computed(() => ({
    churchId: me.value?.church.id ?? '',
  })),
  pause: computed(() => !isAuthReady.value || !me.value?.church.id),
})

// Track initial load
const hasLoadedOnce = ref(false)
watch(data, (newData) => {
  if (!newData) return
  hasLoadedOnce.value = true
})

const { executeMutation: assignRole } = useAssignRoleMutation()
const { executeMutation: revokeRole } = useRevokeRoleMutation()

// Get current admins
const admins = computed(() => data.value?.usersWithRole ?? [])

// Get all users from church for autocomplete
const allUsers = computed(
  () => data.value?.users.edges.map((edge) => edge.node) ?? [],
)

// Filter out users who are already admins for the autocomplete
const nonAdminUsers = computed(() => {
  const adminIds = new Set(admins.value.map((a) => a.id))
  return allUsers.value.filter((u) => !adminIds.has(u.id))
})

// Search state
const searchQuery = ref('')

// Filtered admins based on search
const filteredAdmins = computed(() => {
  if (!searchQuery.value) return admins.value
  const search = searchQuery.value.toLowerCase()
  return admins.value.filter(
    (admin) =>
      admin.name.toLowerCase().includes(search) ||
      admin.email.toLowerCase().includes(search),
  )
})

// Loading states
const addingAdminId = ref<string | null>(null)
const removingAdminId = ref<string | null>(null)

// Remove confirmation modal state
const removeConfirmOpen = ref(false)
const pendingRemove = ref<{
  userId: string
  userName: string
} | null>(null)

// Add admin handler
async function handleAddAdmin(userId: string) {
  if (!me.value?.church.id) return

  addingAdminId.value = userId
  const result = await assignRole({
    input: {
      userId,
      role: RoleType.ChurchAdmin,
      scopeType: ScopeType.Church,
      scopeId: me.value.church.id,
    },
  })
  addingAdminId.value = null

  if (result.error) {
    toast.add({
      title: t('admin.admins.errors.addFailed'),
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: t('admin.admins.success.added'),
    color: 'success',
  })

  await refetch({ requestPolicy: 'network-only' })
}

// Remove admin handlers
function handleRemoveAdmin(userId: string, userName: string) {
  pendingRemove.value = { userId, userName }
  removeConfirmOpen.value = true
}

async function confirmRemove() {
  if (!pendingRemove.value || !me.value?.church.id) return

  removingAdminId.value = pendingRemove.value.userId
  const result = await revokeRole({
    input: {
      userId: pendingRemove.value.userId,
      role: RoleType.ChurchAdmin,
      scopeType: ScopeType.Church,
      scopeId: me.value.church.id,
    },
  })
  removingAdminId.value = null

  if (result.error) {
    toast.add({
      title: t('admin.admins.errors.removeFailed'),
      description: result.error.message,
      color: 'error',
    })
    removeConfirmOpen.value = false
    pendingRemove.value = null
    return
  }

  toast.add({
    title: t('admin.admins.success.removed'),
    color: 'success',
  })

  removeConfirmOpen.value = false
  pendingRemove.value = null
  await refetch({ requestPolicy: 'network-only' })
}

function cancelRemove() {
  removeConfirmOpen.value = false
  pendingRemove.value = null
}

// User item type for autocomplete
type UserItem = {
  id: string
  label: string
  description: string
  user: { id: string }
}

// Autocomplete items for adding users
const userItems = computed<UserItem[]>(() =>
  nonAdminUsers.value.map((user) => ({
    id: user.id,
    label: user.name,
    description: user.email,
    user: { id: user.id },
  })),
)

// Selected user for autocomplete
const selectedUser = ref<UserItem | undefined>(undefined)

// Handle user selection from autocomplete
function handleUserSelect(item: UserItem | undefined) {
  if (item) {
    handleAddAdmin(item.user.id)
    nextTick(() => {
      selectedUser.value = undefined
    })
  }
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: $t('admin.breadcrumb.home'),
              to: { name: 'admin-my-church' },
            },
            {
              label: $t('admin.churchHome.administrators'),
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-6">
      <UButton
        color="neutral"
        variant="soft"
        size="lg"
        :to="{ name: 'admin-my-church' }"
      >
        <Icon name="lucide:arrow-left" />
        {{ $t('admin.common.back') }}
      </UButton>

      <LoadingState v-if="fetching && !hasLoadedOnce" />
      <ErrorState v-else-if="error" :error />
      <div v-else-if="data" class="mt-12 max-w-2xl">
        <h2 class="text-2xl font-semibold mb-6">
          {{ $t('admin.churchHome.administrators') }}
        </h2>

        <!-- Search -->
        <UInput
          v-model="searchQuery"
          :placeholder="$t('admin.admins.searchPlaceholder')"
          icon="lucide:search"
          class="mb-4 mr-2"
        />

        <!-- Add admin autocomplete -->
        <UInputMenu
          v-model="selectedUser"
          :items="userItems"
          :placeholder="$t('admin.admins.addPlaceholder')"
          icon="lucide:user-plus"
          class="mb-6"
          :loading="addingAdminId !== null"
          @update:model-value="handleUserSelect"
        />

        <!-- Admins list -->
        <div class="space-y-2">
          <div
            v-for="admin in filteredAdmins"
            :key="admin.id"
            class="flex items-center justify-between p-4 rounded-xl border border-default bg-elevated/50"
          >
            <div class="flex items-center gap-3">
              <div>
                <div class="font-medium flex gap-2 items-center">
                  {{ admin.name }}
                  <UBadge v-if="admin.id === me?.id" size="sm" variant="soft">
                    {{ $t('admin.admins.you') }}
                  </UBadge>
                </div>
                <div class="text-sm text-dimmed">{{ admin.email }}</div>
              </div>
            </div>
            <UButton
              v-if="admin.id !== me?.id"
              color="error"
              variant="soft"
              square
              :loading="removingAdminId === admin.id"
              @click="handleRemoveAdmin(admin.id, admin.name)"
            >
              <Icon name="lucide:trash-2" />
            </UButton>
          </div>

          <p
            v-if="filteredAdmins.length === 0"
            class="text-dimmed text-sm text-center py-4"
          >
            {{ $t('admin.admins.noAdminsFound') }}
          </p>
        </div>
      </div>
    </UContainer>

    <!-- Remove confirmation modal -->
    <UModal v-model:open="removeConfirmOpen">
      <template #content>
        <div class="p-6">
          <Icon name="lucide:alert-triangle" class="size-6 text-error" />
          <h3 class="my-2 text-lg font-semibold">
            {{ $t('admin.admins.confirmRemove.title') }}
          </h3>
          <p class="text-dimmed mb-6">
            {{
              $t('admin.admins.confirmRemove.message', {
                name: pendingRemove?.userName,
              })
            }}
          </p>
          <div class="flex justify-end gap-3">
            <UButton variant="ghost" color="neutral" @click="cancelRemove">
              {{ $t('admin.common.cancel') }}
            </UButton>
            <UButton
              color="error"
              :loading="removingAdminId !== null"
              @click="confirmRemove"
            >
              {{ $t('admin.admins.confirmRemove.confirm') }}
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
