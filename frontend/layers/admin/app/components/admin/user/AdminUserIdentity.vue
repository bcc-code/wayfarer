<script setup lang="ts">
import { dbLanguageToLocale } from '../../../utils/languageMapping'

// Identity, plus the two actions that apply to the user as a whole.
gql(`
  mutation SyncUser($userId: ID!) {
    syncUser(userId: $userId) {
      user {
        id
        name
        personUuid
        churchLockedUntil
        church {
          id
          name
        }
      }
      contentEventsProcessed
      churchUpdated
      churchLockSkipped
      personUuidUpdated
    }
  }
`)

gql(`
  mutation LockUserChurch($userId: ID!) {
    lockUserChurch(userId: $userId) {
      id
      churchLockedUntil
      church {
        id
        name
      }
    }
  }
`)

gql(`
  mutation UnlockUserChurch($userId: ID!) {
    unlockUserChurch(userId: $userId) {
      id
      churchLockedUntil
      church {
        id
        name
      }
    }
  }
`)

const props = defineProps<{
  user: {
    id: string
    name: string
    email?: string | null
    image?: string | null
    age?: number | null
    language?: string | null
    membersId?: string | number | null
    personUuid?: string | null
    createdAt: string
    churchLockedUntil?: string | null
    church: { id: string; name: string }
  }
  /** Gates the church lock. */
  canManage?: boolean
  canCheckAchievements?: boolean
}>()

/** The page owns the query and refetches. */
const emit = defineEmits<{ changed: [] }>()

const { executeMutation: syncUser } = useSyncUserMutation()
const { executeMutation: lockChurch } = useLockUserChurchMutation()
const { executeMutation: unlockChurch } = useUnlockUserChurchMutation()
const toast = useToast()

const syncing = ref(false)
const locking = ref(false)
const unlocking = ref(false)

const isChurchLocked = computed(() => {
  const until = props.user.churchLockedUntil
  return !!until && new Date(until) > new Date()
})

// The DB stores `no` where the app uses `nb`.
const LANGUAGE_NAMES = new Intl.DisplayNames(['nb'], { type: 'language' })
const languageLabel = computed(() => {
  const code = props.user.language
  if (!code) return undefined
  const locale = dbLanguageToLocale(code)
  try {
    return LANGUAGE_NAMES.of(locale) ?? locale
  } catch {
    // Intl throws on a malformed tag; the raw code is better than nothing.
    return locale
  }
})

async function handleSync() {
  syncing.value = true
  const result = await syncUser({ userId: props.user.id })
  syncing.value = false

  if (result.error) {
    toast.add({
      title: 'Synkronisering feilet',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  const syncResult = result.data?.syncUser
  const details: string[] = []
  if (syncResult) {
    if (syncResult.contentEventsProcessed > 0)
      details.push(`${syncResult.contentEventsProcessed} innholdseventer`)
    if (syncResult.churchUpdated) details.push('menighet oppdatert')
    if (syncResult.churchLockSkipped)
      details.push('menighet hoppet over (låst)')
    if (syncResult.personUuidUpdated) details.push('person-UUID oppdatert')
  }

  toast.add({
    title: 'Synkronisering fullført',
    description: details.length ? details.join(', ') : 'Ingen endringer',
    color: 'success',
  })

  emit('changed')
}

async function handleLock() {
  locking.value = true
  const result = await lockChurch({ userId: props.user.id })
  locking.value = false

  if (result.error) {
    toast.add({
      title: 'Kunne ikke låse menighet',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Menighet låst',
    description: 'Menigheten er låst i 6 måneder',
    color: 'success',
  })

  emit('changed')
}

async function handleUnlock() {
  unlocking.value = true
  const result = await unlockChurch({ userId: props.user.id })
  unlocking.value = false

  if (result.error) {
    toast.add({
      title: 'Kunne ikke låse opp menighet',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Menighet låst opp',
    description: 'Menighetslåsing er fjernet',
    color: 'success',
  })

  emit('changed')
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div class="flex items-start gap-4">
        <UAvatar :src="user.image ?? undefined" :alt="user.name" size="3xl" />
        <div class="min-w-0">
          <h1 class="text-3xl font-bold">{{ user.name }}</h1>
          <p v-if="user.email" class="text-muted text-sm">{{ user.email }}</p>
          <div
            class="text-muted mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-sm"
          >
            <NuxtLink
              :to="{
                name: 'admin-churches-churchId',
                params: { churchId: user.church.id },
              }"
              class="hover:underline"
            >
              {{ user.church.name }}
            </NuxtLink>
            <template v-if="user.age">
              <span aria-hidden="true">·</span>
              <span>{{ user.age }} år</span>
            </template>
            <template v-if="user.language">
              <span aria-hidden="true">·</span>
              <span>{{ languageLabel }}</span>
            </template>
          </div>
        </div>
      </div>

      <div class="flex flex-wrap gap-2">
        <UButton
          v-if="canCheckAchievements"
          icon="i-lucide-trophy"
          variant="soft"
          :to="{
            name: 'admin-users-userId-achievements',
            params: { userId: user.id },
          }"
        >
          Sjekk prestasjoner
        </UButton>
        <UButton
          icon="i-lucide-refresh-cw"
          variant="soft"
          :loading="syncing"
          @click="handleSync"
        >
          Synkroniser
        </UButton>
      </div>
    </div>

    <!-- Sync-lock and the ids share one quiet row: both are rarely needed. -->
    <div
      class="border-default flex flex-wrap items-center gap-x-6 gap-y-2 border-t py-3 text-sm"
    >
      <UCollapsible>
        <UButton
          variant="link"
          color="neutral"
          size="sm"
          class="px-0"
          trailing-icon="i-lucide-chevron-down"
          label="Tekniske detaljer"
        />
        <template #content>
          <dl
            class="divide-default mt-2 grid grid-cols-[auto_1fr] gap-x-6 divide-y text-sm"
          >
            <div class="col-span-full grid grid-cols-subgrid py-2">
              <dt class="text-muted w-36 shrink-0">Bruker-ID</dt>
              <dd class="font-mono">{{ user.id }}</dd>
            </div>
            <div class="col-span-full grid grid-cols-subgrid py-2">
              <dt class="text-muted w-36 shrink-0">Members-ID</dt>
              <dd class="font-mono">{{ user.membersId }}</dd>
            </div>
            <div class="col-span-full grid grid-cols-subgrid py-2">
              <dt class="text-muted w-36 shrink-0">Members-UUID</dt>
              <dd class="font-mono">{{ user.personUuid }}</dd>
            </div>
            <div class="col-span-full grid grid-cols-subgrid py-2">
              <dt class="text-muted w-36 shrink-0">Menighets-ID</dt>
              <dd class="font-mono">{{ user.church.id }}</dd>
            </div>
            <div class="col-span-full grid grid-cols-subgrid py-2">
              <dt class="text-muted w-36 shrink-0">Bruker opprettet</dt>
              <dd>{{ formatDateTime(user.createdAt) }}</dd>
            </div>
          </dl>
        </template>
      </UCollapsible>

      <div v-if="canManage" class="flex items-center gap-2">
        <span class="text-muted">Menighetslås:</span>
        <template v-if="isChurchLocked">
          <UBadge color="warning" variant="soft">
            Låst til {{ formatDateTime(user.churchLockedUntil!) }}
          </UBadge>
          <UButton
            size="xs"
            variant="soft"
            color="neutral"
            :loading="unlocking"
            @click="handleUnlock"
          >
            Lås opp
          </UButton>
        </template>
        <template v-else>
          <span class="text-dimmed">Ikke låst</span>
          <UButton
            size="xs"
            variant="soft"
            color="neutral"
            :loading="locking"
            @click="handleLock"
          >
            Lås i 6 måneder
          </UButton>
        </template>
      </div>
    </div>
  </div>
</template>
