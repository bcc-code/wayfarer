<script lang="ts" setup>
import { onKeyDown } from '@vueuse/core'

definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query AdminGameNightPage {
    frontendConfig
    myCurrentProject {
      myChurchTeams {
        members {
          id
          isTeamLead
        }
      }
    }
  }
`)

const { data: pageData } = useAdminGameNightPageQuery()

const route = useRoute('admin-my-church-gamenight-gamenight')
const gamenight = computed(() => route.params.gamenight)

const quizId = computed(() => {
  const frontendConfig = pageData.value?.frontendConfig
  if (!frontendConfig) return
  const frontendConfigJson = JSON.parse(frontendConfig)
  return frontendConfigJson[`gamenight_${gamenight.value}_betting_quiz_id`] as
    | string
    | undefined
})

const unitLeaders = computed(() =>
  pageData.value?.myCurrentProject.myChurchTeams.flatMap((team) =>
    team.members.filter((member) => member.isTeamLead),
  ),
)

const unitLeaderChallengeId = computed(() => {
  const frontendConfig = pageData.value?.frontendConfig
  if (!frontendConfig) return
  const frontendConfigJson = JSON.parse(frontendConfig)
  return frontendConfigJson[
    `gamenight_${gamenight.value}_unit_leader_challenge_id`
  ] as string | undefined
})

// We need to keep track of quite a bit of state
// that should be persisted between refreshes
const state = useLocalStorage(`state:gamenight-${gamenight.value}`, {
  filmsReady: false,
  bigScreenReady: false,
  step: 1,
  quizSessionId: undefined as string | undefined,
  quizSessionState: undefined as QuizSessionState | undefined,
})

onKeyDown('ArrowUp', () => {
  state.value.step--
})
onKeyDown('ArrowDown', () => {
  state.value.step++
})

const config = useRuntimeConfig()
const { getAccessToken, me } = useAuth()
const { data: cryptexUrl, status: cryptexUrlStatus } = useFetch<{
  url: string
}>(
  `${config.public.apiUrl.replace('/graphql', '')}/plugins/ladder-to-heaven/cryptex-admin-url`,
  {
    method: 'get',
    headers: {
      Authorization: `Bearer ${await getAccessToken()}`,
    },
  },
)

const bigScreenUrl = computed(() => {
  if (!cryptexUrl.value) return
  return `${cryptexUrl.value.url}&redirect=/so`
})

const filmsUrl = 'https://pc26.bcc.media/gamenight-filmer-1237'

const toast = useToast()

// Betting action loading state to prevent spam clicking
const bettingActionLoading = ref(false)

// Betting
const { executeMutation: createQuizSession } = useCreateQuizSessionMutation()
const { executeMutation: openQuizSession } = useOpenQuizSessionMutation()
const { executeMutation: grantQuizSessionAccess } =
  useGrantQuizSessionAccessMutation()
const { executeMutation: lockQuizSession } = useLockQuizSessionMutation()
const { executeMutation: reopenQuizSession } = useReopenQuizSessionMutation()
const { executeMutation: finishQuizSession } = useFinishQuizSessionMutation()
const { executeMutation: bulkEnrollUsersInChallenge } =
  useBulkEnrollUsersInChallengeMutation()
const { data: quizDetails, executeQuery: getQuizDetails } = useQuizDetailsQuery(
  {
    variables: computed(() => ({
      id: quizId.value ?? '',
    })),
  },
)

async function createBetting() {
  if (!quizId.value) {
    throw new Error('Quiz ID not configured for this gamenight')
  }

  const { data } = await createQuizSession({
    input: {
      quizId: quizId.value,
      name: `Gamenight ${gamenight.value} Betting for ${me.value?.church.name}`,
    },
  })
  state.value.quizSessionId = data?.createQuizSession.id

  if (data?.createQuizSession.id) {
    const { data: openData } = await openQuizSession({
      id: data.createQuizSession.id,
    })
    state.value.quizSessionState = openData?.openQuizSession.state
  }
}

async function startBetting() {
  if (bettingActionLoading.value) return
  bettingActionLoading.value = true
  try {
    await createBetting()
    await giveUsersBettingAccess()
    await enrollUsersInChallenge()
  } catch (err) {
    console.error(err)
    toast.add({
      color: 'error',
      icon: 'lucide:alert-triangle',
      title: 'Kunne ikke starte tipping',
    })
  } finally {
    bettingActionLoading.value = false
  }
}

async function lockBetting() {
  if (!state.value.quizSessionId || bettingActionLoading.value) return
  bettingActionLoading.value = true

  try {
    const { data } = await lockQuizSession({ id: state.value.quizSessionId })
    if (data?.lockQuizSession.state) {
      state.value.quizSessionState = data.lockQuizSession.state
    }
  } catch (err) {
    console.error(err)
    toast.add({
      color: 'error',
      icon: 'lucide:alert-triangle',
      title: 'Kunne ikke låse tipping',
    })
  } finally {
    bettingActionLoading.value = false
  }
}

async function reopenBetting() {
  if (!state.value.quizSessionId || bettingActionLoading.value) return
  bettingActionLoading.value = true

  try {
    const { data } = await reopenQuizSession({ id: state.value.quizSessionId })
    if (data?.reopenQuizSession.state) {
      state.value.quizSessionState = data.reopenQuizSession.state
    }
  } catch (err) {
    console.error(err)
    toast.add({
      color: 'error',
      icon: 'lucide:alert-triangle',
      title: 'Kunne ikke gjenåpne tipping',
    })
  } finally {
    bettingActionLoading.value = false
  }
}

async function finishBetting() {
  if (!state.value.quizSessionId) return

  try {
    const { data } = await finishQuizSession({ id: state.value.quizSessionId })
    if (data?.finishQuizSession.state) {
      state.value.quizSessionState = data.finishQuizSession.state
      state.value.step++
    }
  } catch (err) {
    console.error(err)
    toast.add({
      color: 'error',
      icon: 'lucide:alert-triangle',
      title: 'Kunne ikke avslutte tipping',
    })
  }
}

async function giveUsersBettingAccess() {
  if (!state.value.quizSessionId || !me.value?.church.id) return

  await grantQuizSessionAccess({
    input: {
      sessionId: state.value.quizSessionId,
      churchIds: [me.value.church.id],
    },
  })
}

async function enrollUsersInChallenge() {
  if (!quizId.value || !me.value?.church.id) return

  await getQuizDetails()
  if (!quizDetails.value?.quiz.challenge.id) return

  await bulkEnrollUsersInChallenge({
    challengeId: quizDetails.value.quiz.challenge.id,
    target: {
      churchInProject: {
        churchId: me.value.church.id,
        projectId: quizDetails.value?.quiz.project.id,
      },
    },
  })
}

async function enrollUnitLeadersInChallenge() {
  if (!unitLeaderChallengeId.value || !unitLeaders.value?.length) return

  await bulkEnrollUsersInChallenge({
    challengeId: unitLeaderChallengeId.value,
    target: {
      userIds: unitLeaders.value.map((leader) => leader.id),
    },
  })

  state.value.step++
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
              label: $t('admin.churchHome.gameNights'),
              to: { name: 'admin-my-church-gamenights' },
            },
            {
              label: $t('admin.churchHome.gameNight', { number: gamenight }),
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-6 relative">
      <UButton
        color="neutral"
        variant="soft"
        size="lg"
        :to="{ name: 'admin-my-church-gamenights' }"
      >
        <Icon name="lucide:arrow-left" />
        {{ $t('admin.common.back') }}
      </UButton>

      <h1 class="text-3xl font-semibold my-12">
        {{ $t('admin.churchHome.gameNight', { number: gamenight }) }}
      </h1>

      <section
        class="space-y-8 my-6 rounded-2xl p-8 pr-12 border border-muted w-max"
      >
        <h2 class="text-2xl font-semibold">Før du starter</h2>
        <div>
          <UCheckbox
            v-model="state.filmsReady"
            label="Pass på at filmer er klare for avspilling"
          />
          <NuxtLink
            :to="filmsUrl"
            external
            target="_blank"
            class="pl-7 underline text-primary flex items-center gap-1"
          >
            Du finner filmene her
            <Icon name="lucide:external-link" />
          </NuxtLink>
        </div>
        <div>
          <UCheckbox
            v-model="state.bigScreenReady"
            label="Ha storskjermvisningen for spillet klart"
          />
          <NuxtLink
            :to="bigScreenUrl"
            external
            target="_blank"
            :data-loading="!bigScreenUrl"
            class="pl-7 underline text-primary data-[loading=true]:pointer-events-none data-[loading=true]:opacity-50 flex items-center gap-1"
          >
            Du finner storskjermvisningen her
            <Icon name="lucide:external-link" />
          </NuxtLink>
        </div>
      </section>

      <div class="divide-y divide-muted mb-24">
        <AdminStep
          :step="1"
          title="Vis Ladder to Heaven logo på storskjermen"
          description="Dette kan være på storskjermen når ungdommene kommer inn i salen"
          :active="state.step === 1"
        >
          <UButton
            size="xl"
            trailing-icon="lucide:external-link"
            :to="bigScreenUrl"
            target="_blank"
            external
          >
            Åpne storskjermvisningen i ny fane
          </UButton>
          <UButton
            class="flex mt-2"
            variant="ghost"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            Logo vises på storskjermen
          </UButton>
        </AdminStep>
        <AdminStep
          :step="2"
          title="Spill av film 1"
          description="Dette gjøres når ungdommene satt seg og kvelden er i gang"
          :active="state.step === 2"
        >
          <UButton
            size="xl"
            :to="`${filmsUrl}?film=1`"
            trailing-icon="lucide:external-link"
          >
            Gå til film 1
          </UButton>
          <UButton
            class="flex mt-2"
            variant="ghost"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            Filmen er avspilt
          </UButton>
        </AdminStep>
        <AdminStep
          :step="3"
          title="Start tippingen"
          description="Alle må gå inn på 'Utfordringer' siden i Interact appen"
          :active="state.step === 3"
        >
          <div class="flex gap-2 items-center">
            <UButton
              size="xl"
              :loading="bettingActionLoading && !state.quizSessionState"
              :disabled="state.quizSessionState !== undefined"
              :variant="state.quizSessionState !== undefined ? 'soft' : 'solid'"
              @click="startBetting"
            >
              Start tipping
            </UButton>
            <UButton
              v-if="state.quizSessionState === QuizSessionState.Open"
              size="xl"
              :loading="bettingActionLoading"
              @click="lockBetting"
            >
              Lås tipping
            </UButton>
            <UButton
              v-if="state.quizSessionState === QuizSessionState.Locked"
              size="xl"
              :loading="bettingActionLoading"
              @click="reopenBetting"
            >
              Gjenåpne tipping
            </UButton>
          </div>
        </AdminStep>
        <AdminStep
          :step="4"
          title="Spill av film 2"
          description="Dette gjøres etter at tipping er gjennomført"
          :active="state.step === 4"
        >
          <UButton
            size="xl"
            :to="`${filmsUrl}?film=2`"
            trailing-icon="lucide:external-link"
          >
            Gå til film 2
          </UButton>
          <UButton
            class="flex mt-2"
            variant="ghost"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            Filmen er avspilt
          </UButton>
        </AdminStep>
        <AdminStep
          :step="5"
          title="Send ut login-kode til pc26.bcc.media"
          :active="state.step === 5"
        >
          <UButton size="xl" @click="enrollUnitLeadersInChallenge">
            Gi unitledere tilgang til utfordringen
          </UButton>
        </AdminStep>
        <AdminStep
          :step="6"
          title="Gå til adminsiden for å starte spillet"
          :active="state.step === 6"
        >
          <UButton
            :loading="cryptexUrlStatus !== 'success'"
            size="xl"
            :to="cryptexUrl?.url"
            trailing-icon="lucide:arrow-right"
          >
            Gå til adminsiden for spillet
          </UButton>
        </AdminStep>
        <div
          class="py-12 pl-22.75 text-2xl font-semibold data-[active=false]:pointer-events-none data-[active=false]:opacity-50"
          :data-active="state.step === 6"
        >
          <p>Fullfør steg 7 til 11 på den andre siden</p>
        </div>
        <AdminStep
          :step="12"
          title="Spill av film 3"
          :active="state.step === 7"
        >
          <UButton
            size="xl"
            :to="`${filmsUrl}?film=3`"
            trailing-icon="lucide:external-link"
          >
            Gå til film 3
          </UButton>
          <UButton
            class="flex mt-2"
            variant="ghost"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            Filmen er avspilt
          </UButton>
        </AdminStep>
        <AdminStep
          :step="13"
          title="Gi deltakerne poeng for tippingen"
          :active="state.step === 8"
        >
          <UButton
            size="xl"
            :disabled="state.quizSessionState !== QuizSessionState.Locked"
            @click="finishBetting"
          >
            Frigi poeng nå
          </UButton>
        </AdminStep>
        <div
          class="py-24 text-center text-2xl font-semibold data-[active=false]:pointer-events-none data-[active=false]:opacity-50"
          :data-active="state.step === 9"
        >
          <p>Du er nå ferdig! 🎉</p>
          <p>Game Night {{ gamenight }} kan trygt avsluttes.</p>
        </div>
      </div>
    </UContainer>
  </div>
</template>
