<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import type {
  LeaderboardConfigFieldsFragment,
  LeaderboardFilter,
} from '~/api/generated'
import {
  ChurchCategory,
  Gender,
  LeaderboardEntityType,
  LeaderboardLimitMode,
} from '~/api/generated'
import z from 'zod'

const props = defineProps<{
  projectId: string
  /** Absent in create mode. */
  initialData?: LeaderboardConfigFieldsFragment
  /**
   * Event scope is fixed at creation: `UpdateLeaderboardConfigInput` has no
   * `eventId`, so the picker is only offered while creating.
   */
  isEditMode?: boolean
  submitLabel: string
}>()

const emit = defineEmits<{
  submit: [data: LeaderboardConfigFormData]
}>()

export interface LeaderboardConfigFormData {
  name: string
  entityType: LeaderboardEntityType
  eventId: string | null
  limitMode: LeaderboardLimitMode
  maxEntries: number | null
  sortOrder: number
  isActive: boolean
  filter: LeaderboardFilter | null
}

gql(`
  query AdminLeaderboardConfigFormOptions($projectId: ID!) {
    project(id: $projectId) {
      id
    }
    teams(filter: { projectId: $projectId }, first: 500) {
      edges {
        node {
          id
          name
        }
      }
    }
    superteams(filter: { projectId: $projectId }, first: 500) {
      edges {
        node {
          id
          name
        }
      }
    }
    churches(first: 500) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data: options } = useAdminLeaderboardConfigFormOptionsQuery({
  variables: computed(() => ({ projectId: props.projectId })),
  pause: computed(() => !isAuthReady.value),
})

const teamItems = computed(() =>
  (options.value?.teams.edges ?? []).map((edge) => ({
    label: edge.node.name,
    value: edge.node.id,
  })),
)
const superTeamItems = computed(() =>
  (options.value?.superteams.edges ?? []).map((edge) => ({
    label: edge.node.name,
    value: edge.node.id,
  })),
)
const churchItems = computed(() =>
  (options.value?.churches.edges ?? []).map((edge) => ({
    label: edge.node.name,
    value: edge.node.id,
  })),
)

// Shared with the list page, so a label cannot drift between the two.
const entityTypeItems = LEADERBOARD_ENTITY_TYPE_ITEMS
const ageGroups = LEADERBOARD_AGE_GROUPS

/**
 * An optional field is unset when it is falsy. Which falsy value depends on the
 * control: an empty `UInputNumber` gives `undefined`, an emptied `UInput` `''`,
 * and a cleared `USelectMenu` `null`. `buildFilter` tests truthiness rather
 * than a specific sentinel so all three read the same.
 *
 * An explicit "Alle" *item* is not an option — reka-ui rejects a `SelectItem`
 * with an empty-string value, because it reserves that value for "cleared".
 * Hence `clear` on every optional picker.
 */
const optionalNumber = z.number().int().optional()
const optionalText = z.string().optional()
const optionalId = z.string().nullish()

/**
 * `ageRange` is flattened to `ageMin`/`ageMax` because a form control cannot
 * bind to a nested object that is itself optional.
 *
 * Every `LeaderboardFilter` field is represented, including `gender`, `country`
 * and `churchCategory`, which have no control (see the Filter section). The
 * update mutation is full-replace: a field this form does not carry would be
 * silently dropped from a config the first time someone opened it and saved.
 */
const schema = z
  .object({
    name: z.string().min(1, 'Tittel er påkrevd'),
    entityType: z.nativeEnum(LeaderboardEntityType),
    eventId: optionalId,
    limitMode: z.nativeEnum(LeaderboardLimitMode),
    maxEntries: z.number().int().min(1).max(2147483647).optional(),
    sortOrder: z.number().int('Rekkefølge må være et heltall'),
    isActive: z.boolean(),
    filter: z.object({
      myChurch: z.boolean().optional(),
      myTeam: z.boolean().optional(),
      mySuperTeam: z.boolean().optional(),
      minScore: optionalNumber,
      maxScore: optionalNumber,
      churchId: optionalId,
      country: optionalText,
      churchCategory: z.nativeEnum(ChurchCategory).nullish(),
      gender: z.nativeEnum(Gender).nullish(),
      ageMin: optionalNumber,
      ageMax: optionalNumber,
      teamId: optionalId,
      superTeamId: optionalId,
    }),
  })
  .superRefine((value, ctx) => {
    /*
     * Only the conditions an admin can still see and fix are checked here.
     * A control the entity type hides cannot carry an error message, so the
     * entity-type rules live in `buildFilter` (which drops what the backend
     * would reject) and in the `limitMode` watcher instead.
     */
    if (
      value.limitMode === LeaderboardLimitMode.ChurchSize &&
      !value.filter.myChurch &&
      !value.filter.churchId
    ) {
      ctx.addIssue({
        code: 'custom',
        path: ['filter', 'churchId'],
        message: 'Velg en menighet for automatisk grense',
      })
    }
    const { minScore, maxScore, ageMin, ageMax } = value.filter
    const hasMin = typeof ageMin === 'number'
    const hasMax = typeof ageMax === 'number'

    // `AgeRangeInput` has `min: Int!` / `max: Int!` — half an age range cannot
    // be sent, so it has to be rejected here rather than quietly dropped.
    if (hasMin !== hasMax) {
      ctx.addIssue({
        code: 'custom',
        path: ['filter', hasMin ? 'ageMax' : 'ageMin'],
        message: 'Aldersgrense krever både fra og til',
      })
    }
    if (hasMin && hasMax && (ageMin as number) > (ageMax as number)) {
      ctx.addIssue({
        code: 'custom',
        path: ['filter', 'ageMax'],
        message: 'Til-alder må være høyere enn fra-alder',
      })
    }
    if (
      typeof minScore === 'number' &&
      typeof maxScore === 'number' &&
      minScore > maxScore
    ) {
      ctx.addIssue({
        code: 'custom',
        path: ['filter', 'maxScore'],
        message: 'Maks poeng må være høyere enn min poeng',
      })
    }
  })

type Schema = z.infer<typeof schema>

function emptyFilter(): Schema['filter'] {
  return {
    myChurch: false,
    myTeam: false,
    mySuperTeam: false,
    minScore: undefined,
    maxScore: undefined,
    churchId: null,
    country: '',
    churchCategory: null,
    gender: null,
    ageMin: undefined,
    ageMax: undefined,
    teamId: null,
    superTeamId: null,
  }
}

const state = reactive<Schema>({
  name: '',
  entityType: LeaderboardEntityType.Persons,
  eventId: null,
  limitMode: LeaderboardLimitMode.Manual,
  maxEntries: undefined,
  sortOrder: 0,
  isActive: true,
  filter: emptyFilter(),
})

/*
 * Which filters the chosen entity type can actually apply. The live board
 * queries are the `GetFull*` ones in `backend/.../queries/leaderboards.sql`,
 * and their parameter builders pass only these through — a control for
 * anything else would promise a narrowing that never happens.
 */
const isPersons = computed(
  () => state.entityType === LeaderboardEntityType.Persons,
)
const supportsAge = isPersons
const supportsTeam = isPersons
const supportsChurch = computed(
  () => isPersons.value || state.entityType === LeaderboardEntityType.Teams,
)

/**
 * `leaderboardLimitModeToDB` rejects CHURCH_SIZE unless the config carries a
 * persons entity type and a concrete `churchId`. `myChurch` resolves per
 * viewer at request time, so it does not satisfy that — offering the mode in
 * either case would only produce a server error on save.
 */
const supportsChurchSize = computed(
  () => isPersons.value && !state.filter.myChurch,
)

const limitModeItems = computed(() => [
  ...(supportsChurchSize.value
    ? [
        {
          label: 'Automatisk etter menighetsstørrelse',
          value: LeaderboardLimitMode.ChurchSize,
        },
      ]
    : []),
  { label: 'Manuell grense', value: LeaderboardLimitMode.Manual },
])

/*
 * The tiers are the whole substance of the automatic mode, so they belong on
 * the picker rather than in a paragraph under it. "Applies to persons boards
 * with a church" is left unsaid — the picker only offers the mode when that
 * already holds.
 */
const limitModeHelp = computed(() =>
  state.limitMode === LeaderboardLimitMode.ChurchSize
    ? 'Antall deltakere i den filtrerte tavlen bestemmer grensen: under 20: topp 3; 20–49: topp 10; 50–99: topp 20; 100–199: topp 50; 200 eller flere: topp 100. Egen plassering og nærmeste rivaler vises i tillegg.'
    : undefined,
)

// Also fires while hydrating an existing config, which is wanted: a board the
// backend would no longer accept on CHURCH_SIZE falls back to the manual cap.
watch(supportsChurchSize, (supported) => {
  if (!supported) state.limitMode = LeaderboardLimitMode.Manual
})

const { markSaved } = useUnsavedChanges(() => ({ ...state }))

watch(
  () => props.initialData,
  (config) => {
    if (!config) return
    state.name = config.name
    state.entityType = config.entityType
    state.eventId = config.event?.id ?? null
    state.maxEntries = config.maxEntries ?? undefined
    state.limitMode = config.limitMode ?? LeaderboardLimitMode.Manual
    state.sortOrder = config.sortOrder
    state.isActive = config.isActive
    // `LeaderboardFilterView` on the way out, `LeaderboardFilter` on the way
    // in — same fields, so the mapping is field-by-field in both directions.
    state.filter = {
      myChurch: config.filter?.myChurch ?? false,
      myTeam: config.filter?.myTeam ?? false,
      mySuperTeam: config.filter?.mySuperTeam ?? false,
      minScore: config.filter?.minScore ?? undefined,
      maxScore: config.filter?.maxScore ?? undefined,
      churchId: config.filter?.churchId ?? null,
      country: config.filter?.country ?? '',
      churchCategory: config.filter?.churchCategory ?? null,
      gender: config.filter?.gender ?? null,
      ageMin: config.filter?.ageRange?.min ?? undefined,
      ageMax: config.filter?.ageRange?.max ?? undefined,
      teamId: config.filter?.teamId ?? null,
      superTeamId: config.filter?.superTeamId ?? null,
    }
    // What the server holds, not an edit.
    nextTick(markSaved)
  },
  { immediate: true },
)

/**
 * `null` rather than an all-empty object, so "no filter" round-trips as null.
 *
 * Fields the chosen entity type cannot apply are kept rather than dropped —
 * the board ignores them, and keeping them means switching a board's type back
 * and forth does not quietly destroy what was configured. The exception is the
 * three viewer-relative flags: `ValidateLeaderboardRelativeFilter` *rejects*
 * those outside their entity types, so sending one the admin can no longer see
 * would fail the save with an error pointing at a hidden control.
 */
function buildFilter(
  filter: Schema['filter'],
  entityType: LeaderboardEntityType,
): LeaderboardFilter | null {
  const built: LeaderboardFilter = {}
  const persons = entityType === LeaderboardEntityType.Persons
  const teams = entityType === LeaderboardEntityType.Teams

  if (typeof filter.minScore === 'number') built.minScore = filter.minScore
  if (typeof filter.maxScore === 'number') built.maxScore = filter.maxScore
  if (filter.myChurch && (persons || teams)) built.myChurch = true
  else if (filter.churchId) built.churchId = filter.churchId
  if (filter.country) built.country = filter.country
  if (filter.churchCategory) built.churchCategory = filter.churchCategory
  if (filter.gender) built.gender = filter.gender
  if (typeof filter.ageMin === 'number' && typeof filter.ageMax === 'number') {
    built.ageRange = { min: filter.ageMin, max: filter.ageMax }
  }
  if (filter.myTeam && persons) built.myTeam = true
  else if (filter.teamId) built.teamId = filter.teamId
  if (filter.mySuperTeam && persons) built.mySuperTeam = true
  else if (filter.superTeamId) built.superTeamId = filter.superTeamId

  return Object.keys(built).length ? built : null
}

function onSubmit(event: FormSubmitEvent<Schema>) {
  markSaved()
  emit('submit', {
    name: event.data.name,
    entityType: event.data.entityType,
    eventId: event.data.eventId || null,
    limitMode: event.data.limitMode,
    maxEntries:
      typeof event.data.maxEntries === 'number' ? event.data.maxEntries : null,
    sortOrder: event.data.sortOrder,
    isActive: event.data.isActive,
    filter: buildFilter(event.data.filter, event.data.entityType),
  })
}

function clearFilter() {
  state.filter = emptyFilter()
}

/** No age bounds. A preset rather than an empty state, so the control always
 * reads as a choice and "no age limit" is one click rather than two cleared
 * fields. */
const AGE_GROUP_ALL = 'Alle'

const ageGroupItems = [
  { label: AGE_GROUP_ALL, value: AGE_GROUP_ALL },
  ...ageGroups.map((group) => ({ label: group.label, value: group.label })),
]

/**
 * `''` — no segment at all — for a custom range typed into the two fields
 * below, which matches no preset. Not `undefined`: that leaves `UTabs`
 * uncontrolled, and it would fall back to its first item and read as "Alle".
 */
const activeAgeGroup = computed(() => {
  const { ageMin, ageMax } = state.filter
  if (typeof ageMin !== 'number' && typeof ageMax !== 'number')
    return AGE_GROUP_ALL
  if (typeof ageMin !== 'number' || typeof ageMax !== 'number') return ''
  return ageGroupLabel({ min: ageMin, max: ageMax }) ?? ''
})

function selectAgeGroup(label: string | number) {
  const group = ageGroups.find((candidate) => candidate.label === label)
  state.filter.ageMin = group?.min
  state.filter.ageMax = group?.max
}

/**
 * `trigger` undoes the theme's `w-full`, which is right where the bar spans a
 * column and truncates these labels to "U…" in a heading row.
 *
 * `indicator` covers reka-ui's `updateIndicatorStyle`, which returns early when
 * no tab is active and so leaves the pill parked on the last selection instead
 * of clearing it. A custom range typed into the fields below is the only way to
 * reach that state now that "Alle" covers the empty one.
 */
const ageGroupUi = computed(() => ({
  trigger: 'w-auto',
  indicator: activeAgeGroup.value ? undefined : 'hidden',
}))
</script>

<template>
  <UForm
    :state
    :schema
    loading-auto
    class="flex max-w-2xl flex-col gap-8"
    @submit.prevent="onSubmit"
  >
    <AdminSection title="Ledertavle">
      <div class="flex flex-col gap-6">
        <UFormField name="name" label="Tittel">
          <UInput v-model="state.name" size="xl" required class="w-full" />
        </UFormField>

        <UFormField
          name="entityType"
          label="Type"
          help="Hva tavlen rangerer — personer, lag, superlag eller menigheter."
        >
          <USelect
            v-model="state.entityType"
            :items="entityTypeItems"
            value-key="value"
            class="w-full"
          />
        </UFormField>

        <UFormField
          v-if="limitModeItems.length > 1"
          name="limitMode"
          label="Antall plasseringer"
          :help="limitModeHelp"
        >
          <USelect
            v-model="state.limitMode"
            :items="limitModeItems"
            value-key="value"
            class="w-full"
          />
        </UFormField>

        <UFormField
          name="maxEntries"
          label="Hvor mange vises på tavlen"
          :help="
            state.limitMode === LeaderboardLimitMode.ChurchSize
              ? 'Valgfri øvre grense. Den laveste av denne og menighetsgrensen brukes.'
              : 'Skriv 10 for en topp 10-liste. La stå tom for å vise alle. Deltakeren ser sin egen plassering og de nærmeste rivalene uansett.'
          "
        >
          <UInputNumber
            v-model="state.maxEntries"
            :min="1"
            :max="2147483647"
            :step="1"
            placeholder="Ingen grense"
            class="w-full"
          />
        </UFormField>

        <UFormField
          name="isActive"
          label="Aktiv"
          help="Inaktive tavler er skjult for vanlige brukere."
        >
          <USwitch v-model="state.isActive" />
        </UFormField>
      </div>
    </AdminSection>

    <AdminSection title="Filter">
      <template #actions>
        <UButton
          icon="lucide:rotate-ccw"
          variant="ghost"
          size="sm"
          @click="clearFilter"
        >
          Nullstill filter
        </UButton>
      </template>

      <div class="flex flex-col gap-6">
        <p class="text-sm text-muted">
          Filtrene kombineres. Hvilke filtre som vises avhenger av hva tavlen
          rangerer.<template v-if="supportsChurch">
            «Min» og «mitt» følger personen som ser tavlen.</template
          >
        </p>
        <div v-if="supportsAge" class="flex flex-col gap-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <h3 class="text-sm font-medium">Alder</h3>
            <UTabs
              :model-value="activeAgeGroup"
              :items="ageGroupItems"
              :content="false"
              variant="pill"
              color="neutral"
              size="sm"
              aria-label="Aldersgruppe"
              class="shrink-0"
              :ui="ageGroupUi"
              @update:model-value="selectAgeGroup"
            />
          </div>
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <UFormField
              name="filter.ageMin"
              label="Yngste alder"
              help="Alder regnes etter fødselsår, ikke bursdag."
            >
              <UInputNumber
                v-model="state.filter.ageMin"
                placeholder="Ingen grense"
                class="w-full"
              />
            </UFormField>
            <UFormField
              name="filter.ageMax"
              label="Eldste alder"
              help="Denne alderen er med."
            >
              <UInputNumber
                v-model="state.filter.ageMax"
                placeholder="Ingen grense"
                class="w-full"
              />
            </UFormField>
          </div>
        </div>
        <div
          v-if="supportsChurch"
          class="grid grid-cols-1 gap-5 sm:grid-cols-2"
          :class="supportsAge && 'border-t border-default pt-6'"
        >
          <div class="flex min-w-0 flex-col gap-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h3 class="text-sm font-medium">Menighet</h3>
              <UFormField name="filter.myChurch"
                ><USwitch v-model="state.filter.myChurch" label="Min menighet"
              /></UFormField>
            </div>
            <UFormField v-if="!state.filter.myChurch" name="filter.churchId">
              <USelectMenu
                v-model="state.filter.churchId"
                :items="churchItems"
                aria-label="Menighet"
                value-key="value"
                placeholder="Alle"
                clear
                class="w-full"
              />
            </UFormField>
            <p v-else class="flex min-h-8 items-center text-sm text-muted">
              Følger personens menighet.
            </p>
          </div>
          <div v-if="supportsTeam" class="flex min-w-0 flex-col gap-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h3 class="text-sm font-medium">Lag</h3>
              <UFormField name="filter.myTeam"
                ><USwitch v-model="state.filter.myTeam" label="Mitt lag"
              /></UFormField>
            </div>
            <UFormField v-if="!state.filter.myTeam" name="filter.teamId">
              <USelectMenu
                v-model="state.filter.teamId"
                :items="teamItems"
                aria-label="Lag"
                value-key="value"
                placeholder="Alle"
                clear
                class="w-full"
              />
            </UFormField>
            <p v-else class="flex min-h-8 items-center text-sm text-muted">
              Følger personens lag.
            </p>
          </div>
          <div v-if="supportsTeam" class="flex min-w-0 flex-col gap-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h3 class="text-sm font-medium">Superlag</h3>
              <UFormField name="filter.mySuperTeam"
                ><USwitch
                  v-model="state.filter.mySuperTeam"
                  label="Mitt superlag"
              /></UFormField>
            </div>
            <UFormField
              v-if="!state.filter.mySuperTeam"
              name="filter.superTeamId"
            >
              <USelectMenu
                v-model="state.filter.superTeamId"
                :items="superTeamItems"
                aria-label="Superlag"
                value-key="value"
                placeholder="Alle"
                clear
                class="w-full"
              />
            </UFormField>
            <p v-else class="flex min-h-8 items-center text-sm text-muted">
              Følger personens superlag.
            </p>
          </div>
        </div>
        <div
          class="grid grid-cols-1 gap-4 sm:grid-cols-2"
          :class="
            (supportsAge || supportsChurch) && 'border-t border-default pt-6'
          "
        >
          <UFormField name="filter.minScore" label="Min. poeng">
            <UInputNumber
              v-model="state.filter.minScore"
              placeholder="Ingen grense"
              class="w-full"
            />
          </UFormField>
          <UFormField name="filter.maxScore" label="Maks poeng">
            <UInputNumber
              v-model="state.filter.maxScore"
              placeholder="Ingen grense"
              class="w-full"
            />
          </UFormField>
        </div>
      </div>
    </AdminSection>

    <UButton
      :icon="isEditMode ? 'lucide:check' : 'lucide:plus'"
      type="submit"
      size="lg"
      block
      >{{ submitLabel }}</UButton
    >
  </UForm>
</template>
