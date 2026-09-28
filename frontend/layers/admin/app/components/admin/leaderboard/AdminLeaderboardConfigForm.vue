<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import type {
  LeaderboardConfigFieldsFragment,
  LeaderboardFilter,
} from '~/api/generated'
import { ChurchCategory, Gender, LeaderboardEntityType } from '~/api/generated'
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
  maxEntries: number | null
  sortOrder: number
  isActive: boolean
  filter: LeaderboardFilter | null
}

gql(`
  query AdminLeaderboardConfigFormOptions($projectId: ID!) {
    project(id: $projectId) {
      id
      events {
        id
        name
      }
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

const eventItems = computed(() =>
  (options.value?.project.events ?? []).map((event) => ({
    label: event.name,
    value: event.id,
  })),
)
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
const churchCategoryItems = CHURCH_CATEGORY_ITEMS
const genderItems = GENDER_ITEMS

/**
 * An optional field is unset when it is falsy. Which falsy value depends on the
 * control: an emptied `UInput` gives `''`, and a cleared `USelectMenu` gives
 * `null`. `buildFilter` tests truthiness rather than a specific sentinel so
 * both read the same.
 *
 * An explicit "Alle" *item* is not an option — reka-ui rejects a `SelectItem`
 * with an empty-string value, because it reserves that value for "cleared".
 * Hence `clear` on every optional picker.
 */
const optionalNumber = z.union([z.number().int(), z.literal('')]).optional()
const optionalText = z.string().optional()
const optionalId = z.string().nullish()

/**
 * `ageRange` is flattened to `ageMin`/`ageMax` because a form control cannot
 * bind to a nested object that is itself optional.
 *
 * Every `LeaderboardFilter` field is represented, including
 * ones an admin rarely sets. The update mutation is full-replace: a field this
 * form does not carry would be silently dropped from an existing config the
 * first time someone opened it and saved.
 */
const schema = z
  .object({
    name: z.string().min(1, 'Navn er påkrevd'),
    entityType: z.nativeEnum(LeaderboardEntityType),
    eventId: optionalId,
    maxEntries: z
      .union([z.number().int().min(1).max(2147483647), z.literal('')])
      .optional(),
    sortOrder: z.number().int('Rekkefølge må være et heltall'),
    isActive: z.boolean(),
    filter: z.object({
      myChurch: z.boolean().optional(),
      myTeam: z.boolean().optional(),
      mySuperTeam: z.boolean().optional(),
      relativeAge: z.boolean().optional(),
      yearsYounger: z.number().int().min(0).max(150).optional(),
      yearsOlder: z.number().int().min(0).max(150).optional(),
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
    const { minScore, maxScore, ageMin, ageMax } = value.filter
    const persons = value.entityType === LeaderboardEntityType.Persons
    for (const key of ['myTeam', 'mySuperTeam', 'relativeAge'] as const) {
      if (value.filter[key] && !persons) {
        ctx.addIssue({
          code: 'custom',
          path: ['filter', key],
          message: 'Krever en persontavle',
        })
      }
    }
    if (
      value.filter.myChurch &&
      !persons &&
      value.entityType !== LeaderboardEntityType.Teams
    ) {
      ctx.addIssue({
        code: 'custom',
        path: ['filter', 'myChurch'],
        message: 'Krever en person- eller lagtavle',
      })
    }
    if (value.filter.relativeAge) {
      for (const key of ['yearsYounger', 'yearsOlder'] as const) {
        if (value.filter[key] == null)
          ctx.addIssue({
            code: 'custom',
            path: ['filter', key],
            message: 'Antall år er påkrevd',
          })
      }
    }
    const hasMin = typeof ageMin === 'number'
    const hasMax = typeof ageMax === 'number'

    // `AgeRangeInput` has `min: Int!` / `max: Int!` — half an age range cannot
    // be sent, so it has to be rejected here rather than quietly dropped.
    if (!value.filter.relativeAge && hasMin !== hasMax) {
      ctx.addIssue({
        code: 'custom',
        path: ['filter', hasMin ? 'ageMax' : 'ageMin'],
        message: 'Aldersgrense krever både fra og til',
      })
    }
    if (
      !value.filter.relativeAge &&
      hasMin &&
      hasMax &&
      (ageMin as number) > (ageMax as number)
    ) {
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
    relativeAge: false,
    yearsYounger: 3,
    yearsOlder: 3,
    minScore: '',
    maxScore: '',
    churchId: null,
    country: '',
    churchCategory: null,
    gender: null,
    ageMin: '',
    ageMax: '',
    teamId: null,
    superTeamId: null,
  }
}

const state = reactive<Schema>({
  name: '',
  entityType: LeaderboardEntityType.Persons,
  eventId: null,
  maxEntries: '',
  sortOrder: 0,
  isActive: true,
  filter: emptyFilter(),
})

watch(
  () => props.initialData,
  (config) => {
    if (!config) return
    state.name = config.name
    state.entityType = config.entityType
    state.eventId = config.event?.id ?? null
    state.maxEntries = config.maxEntries ?? ''
    state.sortOrder = config.sortOrder
    state.isActive = config.isActive
    // `LeaderboardFilterView` on the way out, `LeaderboardFilter` on the way
    // in — same fields, so the mapping is field-by-field in both directions.
    state.filter = {
      myChurch: config.filter?.myChurch ?? false,
      myTeam: config.filter?.myTeam ?? false,
      mySuperTeam: config.filter?.mySuperTeam ?? false,
      relativeAge: !!config.filter?.relativeAgeRange,
      yearsYounger: config.filter?.relativeAgeRange?.yearsYounger ?? 3,
      yearsOlder: config.filter?.relativeAgeRange?.yearsOlder ?? 3,
      minScore: config.filter?.minScore ?? '',
      maxScore: config.filter?.maxScore ?? '',
      churchId: config.filter?.churchId ?? null,
      country: config.filter?.country ?? '',
      churchCategory: config.filter?.churchCategory ?? null,
      gender: config.filter?.gender ?? null,
      ageMin: config.filter?.ageRange?.min ?? '',
      ageMax: config.filter?.ageRange?.max ?? '',
      teamId: config.filter?.teamId ?? null,
      superTeamId: config.filter?.superTeamId ?? null,
    }
  },
  { immediate: true },
)

/** `null` rather than an all-empty object, so "no filter" round-trips as null. */
function buildFilter(filter: Schema['filter']): LeaderboardFilter | null {
  const built: LeaderboardFilter = {}

  if (typeof filter.minScore === 'number') built.minScore = filter.minScore
  if (typeof filter.maxScore === 'number') built.maxScore = filter.maxScore
  if (filter.myChurch) built.myChurch = true
  else if (filter.churchId) built.churchId = filter.churchId
  if (filter.country) built.country = filter.country
  if (filter.churchCategory) built.churchCategory = filter.churchCategory
  if (filter.gender) built.gender = filter.gender
  if (filter.relativeAge) {
    built.relativeAgeRange = {
      yearsYounger: filter.yearsYounger!,
      yearsOlder: filter.yearsOlder!,
    }
  } else if (
    typeof filter.ageMin === 'number' &&
    typeof filter.ageMax === 'number'
  ) {
    built.ageRange = { min: filter.ageMin, max: filter.ageMax }
  }
  if (filter.myTeam) built.myTeam = true
  else if (filter.teamId) built.teamId = filter.teamId
  if (filter.mySuperTeam) built.mySuperTeam = true
  else if (filter.superTeamId) built.superTeamId = filter.superTeamId

  return Object.keys(built).length ? built : null
}

function onSubmit(event: FormSubmitEvent<Schema>) {
  emit('submit', {
    name: event.data.name,
    entityType: event.data.entityType,
    eventId: event.data.eventId || null,
    maxEntries:
      typeof event.data.maxEntries === 'number' ? event.data.maxEntries : null,
    sortOrder: event.data.sortOrder,
    isActive: event.data.isActive,
    filter: buildFilter(event.data.filter),
  })
}

function clearFilter() {
  state.filter = emptyFilter()
}
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
        <UFormField name="name" label="Navn">
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
          v-if="!isEditMode"
          name="eventId"
          label="Arrangement"
          help="La stå tom for en tavle på prosjektnivå. Kan ikke endres senere."
        >
          <USelectMenu
            v-model="state.eventId"
            :items="eventItems"
            value-key="value"
            placeholder="Hele prosjektet"
            clear
            class="w-full"
          />
        </UFormField>

        <UFormField
          name="maxEntries"
          label="Maks antall oppføringer"
          help="La stå tom for ingen grense. Egen plassering og nærmeste rivaler vises i tillegg."
        >
          <UInput
            v-model.number="state.maxEntries"
            type="number"
            :min="1"
            :max="2147483647"
            :step="1"
            placeholder="Ingen grense"
            class="w-full"
          />
        </UFormField>

        <UFormField
          name="sortOrder"
          label="Rekkefølge"
          help="Lav verdi vises først."
        >
          <UInput
            v-model.number="state.sortOrder"
            type="number"
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
        <UButton variant="ghost" size="sm" @click="clearFilter">
          Nullstill filter
        </UButton>
      </template>

      <div class="flex flex-col gap-6">
        <p class="text-sm text-muted">
          Filtrene kombineres. «Min» og «mitt» følger personen som ser tavlen.
        </p>
        <div class="flex flex-col gap-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <h3 class="text-sm font-medium">Alder</h3>
            <UFormField name="filter.relativeAge">
              <USwitch
                v-model="state.filter.relativeAge"
                label="Min aldersgruppe"
              />
            </UFormField>
          </div>
          <div
            v-if="state.filter.relativeAge"
            class="grid grid-cols-1 gap-4 sm:grid-cols-2"
          >
            <UFormField name="filter.yearsYounger" label="År yngre">
              <UInput
                v-model.number="state.filter.yearsYounger"
                type="number"
                class="w-full"
                :min="0"
                :max="150"
              />
            </UFormField>
            <UFormField name="filter.yearsOlder" label="År eldre">
              <UInput
                v-model.number="state.filter.yearsOlder"
                type="number"
                class="w-full"
                :min="0"
                :max="150"
              />
            </UFormField>
          </div>
          <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <UFormField name="filter.ageMin" label="Alder fra">
              <UInput
                v-model.number="state.filter.ageMin"
                type="number"
                placeholder="Ingen grense"
                class="w-full"
              />
            </UFormField>
            <UFormField name="filter.ageMax" label="Alder til">
              <UInput
                v-model.number="state.filter.ageMax"
                type="number"
                placeholder="Ingen grense"
                class="w-full"
              />
            </UFormField>
          </div>
        </div>
        <div
          class="grid grid-cols-1 gap-5 border-t border-default pt-6 sm:grid-cols-2"
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
          <div class="flex min-w-0 flex-col gap-3">
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
          <div class="flex min-w-0 flex-col gap-3">
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
          <div class="flex min-w-0 flex-col justify-end">
            <UFormField name="filter.gender" label="Kjønn">
              <USelectMenu
                v-model="state.filter.gender"
                :items="genderItems"
                value-key="value"
                placeholder="Alle"
                clear
                class="w-full"
              />
            </UFormField>
          </div>
        </div>
        <div
          class="grid grid-cols-1 gap-4 border-t border-default pt-6 sm:grid-cols-2"
        >
          <UFormField name="filter.churchCategory" label="Menighetsstørrelse">
            <USelectMenu
              v-model="state.filter.churchCategory"
              :items="churchCategoryItems"
              value-key="value"
              placeholder="Alle"
              clear
              class="w-full"
            />
          </UFormField>
          <UFormField name="filter.country" label="Land">
            <UInput
              v-model="state.filter.country"
              placeholder="Alle"
              class="w-full"
            />
          </UFormField>
          <UFormField name="filter.minScore" label="Min. poeng">
            <UInput
              v-model.number="state.filter.minScore"
              type="number"
              placeholder="Ingen grense"
              class="w-full"
            />
          </UFormField>
          <UFormField name="filter.maxScore" label="Maks poeng">
            <UInput
              v-model.number="state.filter.maxScore"
              type="number"
              placeholder="Ingen grense"
              class="w-full"
            />
          </UFormField>
        </div>
      </div>
    </AdminSection>

    <UButton type="submit" size="lg" block>{{ submitLabel }}</UButton>
  </UForm>
</template>
