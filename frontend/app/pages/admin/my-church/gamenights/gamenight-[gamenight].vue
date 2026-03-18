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

const route = useRoute('admin-my-church-gamenights-gamenight-gamenight')
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
  externalAdminVisited: false,
})

function goToPreviousStep() {
  state.value.step = Math.max(state.value.step - 1, 1)
}
function goToNextStep() {
  state.value.step = Math.min(state.value.step + 1, 10)
}

onKeyDown('ArrowUp', goToPreviousStep)
onKeyDown('ArrowDown', goToNextStep)

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
const { t } = useI18n()

// Betting action loading state to prevent spam clicking
const bettingActionLoading = ref(false)

// Countdown before betting starts
const {
  remaining: bettingCountdown,
  start: startBettingCountdown,
  stop: cancelBettingCountdown,
  isActive: isBettingCountdownActive,
} = useCountdown(10, {
  onComplete: () => startBetting(),
})

// Betting
const { executeMutation: createQuizSession } = useCreateQuizSessionMutation()
const { executeMutation: openQuizSession } = useOpenQuizSessionMutation()
const { executeMutation: grantQuizSessionAccess } =
  useGrantQuizSessionAccessAsyncMutation()
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
      title: t('admin.gamenight.errors.couldNotStartBetting'),
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
      title: t('admin.gamenight.errors.couldNotLockBetting'),
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
      title: t('admin.gamenight.errors.couldNotReopenBetting'),
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
      title: t('admin.gamenight.errors.couldNotFinishBetting'),
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

async function copyLink(url: string) {
  await navigator.clipboard.writeText(url)
  toast.add({
    title: t('admin.common.linkCopied'),
    color: 'success',
  })
}

const unitLeaderAccessLoading = ref(false)
const unitLeaderAccessSent = ref(false)

async function enrollUnitLeadersInChallenge() {
  if (!unitLeaderChallengeId.value || !unitLeaders.value?.length) return
  if (unitLeaderAccessLoading.value) return
  unitLeaderAccessLoading.value = true

  try {
    await bulkEnrollUsersInChallenge({
      challengeId: unitLeaderChallengeId.value,
      target: {
        userIds: unitLeaders.value.map((leader) => leader.id),
      },
    })

    unitLeaderAccessSent.value = true
    state.value.step++
  } catch (err) {
    console.error(err)
    toast.add({
      color: 'error',
      icon: 'lucide:alert-triangle',
      title: t('admin.gamenight.errors.couldNotSendUnitLeaderAccess'),
    })
  } finally {
    unitLeaderAccessLoading.value = false
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
        <div>
          <h2 class="text-2xl font-semibold mb-2">
            {{ $t('admin.gamenight.importantResources') }}
          </h2>
          <p>{{ $t('admin.gamenight.importantResourcesDescription') }}</p>
        </div>
        <div>
          <UCheckbox
            v-model="state.filmsReady"
            :label="$t('admin.gamenight.linkToAllFilms')"
          />
          <UButton
            class="ml-7 mt-2"
            variant="outline"
            trailing-icon="lucide:copy"
            @click="copyLink(filmsUrl)"
          >
            {{ $t('admin.gamenight.findFilmsHere') }}
          </UButton>
        </div>
        <div>
          <UCheckbox
            v-model="state.bigScreenReady"
            :label="$t('admin.gamenight.bigScreen')"
          />
          <UButton
            class="ml-7 mt-2"
            variant="outline"
            trailing-icon="lucide:copy"
            :disabled="!bigScreenUrl"
            @click="copyLink(bigScreenUrl!)"
          >
            {{ $t('admin.gamenight.findBigScreenHere') }}
          </UButton>
        </div>
      </section>

      <div class="divide-y divide-muted mb-24">
        <AdminStep
          :step="1"
          :title="$t('admin.gamenight.step2Title')"
          :description="$t('admin.gamenight.step2Description')"
          :active="state.step === 1"
        >
          <UButton
            class="flex mt-2"
            variant="outline"
            trailing-icon="lucide:check"
            size="xl"
            @click="goToNextStep"
          >
            {{ $t('admin.gamenight.filmFinishedNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="2"
          :title="$t('admin.gamenight.step3Title')"
          :active="state.step === 2"
        >
          <div class="flex flex-col gap-4 items-start">
            <p>1. {{ $t('admin.gamenight.step3Description') }}</p>
            <div class="flex gap-2 items-center">
              <UButton
                v-if="!state.quizSessionState"
                size="xl"
                :loading="bettingActionLoading"
                :disabled="isBettingCountdownActive"
                color="error"
                trailing-icon="lucide:triangle-alert"
                @click="() => startBettingCountdown()"
              >
                {{ $t('admin.gamenight.startBetting') }}
              </UButton>
              <div
                v-if="isBettingCountdownActive"
                class="flex gap-4 items-center tabular-nums"
              >
                <p>
                  {{
                    $t('admin.gamenight.bettingStartingIn', {
                      seconds: bettingCountdown,
                    })
                  }}
                </p>
                <UButton
                  variant="soft"
                  color="neutral"
                  trailing-icon="lucide:x"
                  @click="cancelBettingCountdown"
                >
                  {{ $t('admin.gamenight.cancelBettingCountdown') }}
                </UButton>
              </div>
              <UButton
                v-if="state.quizSessionState !== undefined"
                size="xl"
                :loading="bettingActionLoading && !state.quizSessionState"
                disabled
                variant="soft"
                color="neutral"
                trailing-icon="lucide:triangle-alert"
              >
                {{ $t('admin.gamenight.startBetting') }}
              </UButton>
            </div>
            <p class="mt-4">
              2. {{ $t('admin.gamenight.lockBettingDescription') }}
            </p>
            <div
              v-if="state.quizSessionState === QuizSessionState.Open"
              class="flex gap-4 items-center px-4 py-3 border border-muted rounded-2xl"
            >
              <UButton
                size="xl"
                :loading="bettingActionLoading"
                @click="lockBetting"
              >
                {{ $t('admin.gamenight.lockBetting') }}
              </UButton>
              <p>
                {{ $t('admin.gamenight.bettingLockedNotice') }} <br >
                {{ $t('admin.gamenight.reopenBettingHint') }}
              </p>
            </div>
            <div
              v-if="state.quizSessionState === QuizSessionState.Locked"
              class="flex gap-4 items-center p-4 border border-muted rounded-2xl"
            >
              <UButton
                size="xl"
                :loading="bettingActionLoading"
                :disabled="
                  state.quizSessionState === QuizSessionState.Locked &&
                  state.step !== 2
                "
                :variant="state.step === 2 ? 'outline' : 'soft'"
                :color="state.step === 2 ? 'primary' : 'neutral'"
                @click="reopenBetting"
              >
                {{ $t('admin.gamenight.reopenBetting') }}
              </UButton>
              <p>{{ $t('admin.gamenight.reopenBettingHint') }}</p>
            </div>
          </div>
          <p class="mt-8">
            3. {{ $t('admin.gamenight.makeSureEverybodyBet') }}
          </p>
          <div
            v-if="state.quizSessionState === QuizSessionState.Locked"
            class="flex gap-4 items-center mt-4"
          >
            <UButton
              trailing-icon="lucide:triangle-alert"
              size="xl"
              :variant="state.step === 2 ? 'solid' : 'soft'"
              :color="state.step === 2 ? 'error' : 'neutral'"
              :disabled="state.step !== 2"
              @click="goToNextStep"
            >
              {{ $t('admin.gamenight.allBetNextStep') }}
            </UButton>
            <p>{{ $t('admin.gamenight.cannotReopenAfterContinue') }}</p>
          </div>
        </AdminStep>
        <AdminStep
          :step="3"
          :title="$t('admin.gamenight.step4Title')"
          :description="$t('admin.gamenight.step4Description')"
          :active="state.step === 3"
        >
          <UButton
            class="flex mt-2"
            variant="outline"
            trailing-icon="lucide:check"
            size="xl"
            @click="goToNextStep"
          >
            {{ $t('admin.gamenight.filmFinishedNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="4"
          :title="$t('admin.gamenight.step5TitleBigScreen')"
          :active="state.step === 4"
        >
          <UButton
            size="xl"
            variant="outline"
            trailing-icon="lucide:check"
            @click="goToNextStep"
          >
            {{ $t('admin.gamenight.goToNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="5"
          :title="$t('admin.gamenight.step5Title')"
          :description="$t('admin.gamenight.step5Description')"
          :active="state.step === 5"
        >
          <UButton
            size="xl"
            :loading="unitLeaderAccessLoading"
            :disabled="unitLeaderAccessSent"
            :trailing-icon="unitLeaderAccessSent ? 'lucide:check' : undefined"
            :color="unitLeaderAccessSent ? 'success' : undefined"
            @click="enrollUnitLeadersInChallenge"
          >
            {{
              unitLeaderAccessSent
                ? $t('admin.gamenight.unitLeaderAccessSent')
                : $t('admin.gamenight.giveUnitLeadersAccess')
            }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="6"
          :title="$t('admin.gamenight.step6Title')"
          :description="$t('admin.gamenight.step6Description')"
          :active="state.step === 6"
        >
          <UButton
            :loading="cryptexUrlStatus !== 'success'"
            size="xl"
            :to="cryptexUrl?.url"
            trailing-icon="lucide:arrow-right"
            @click="state.externalAdminVisited = true"
          >
            {{ $t('admin.gamenight.goToUnitTaskAdmin') }}
          </UButton>
        </AdminStep>
        <div
          class="py-12 pl-22.75 data-[active=false]:pointer-events-none data-[active=false]:opacity-50"
          :data-active="state.step === 6"
        >
          <p class="text-2xl font-semibold mb-4">
            {{
              $t('admin.gamenight.completeStepsOnOtherPage', {
                from: 7,
                to: 10,
              })
            }}
          </p>
          <UButton
            v-if="state.externalAdminVisited"
            variant="outline"
            trailing-icon="lucide:check"
            size="xl"
            @click="goToNextStep"
          >
            {{ $t('admin.gamenight.goToNextStep') }}
          </UButton>
        </div>
        <AdminStep
          :step="11"
          :title="$t('admin.gamenight.step12Title')"
          :active="state.step === 7"
        >
          <UButton
            class="flex mt-2"
            variant="outline"
            trailing-icon="lucide:check"
            size="xl"
            @click="goToNextStep"
          >
            {{ $t('admin.gamenight.filmFinishedNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="12"
          :title="$t('admin.gamenight.step13Title')"
          :active="state.step === 8"
        >
          <UAlert
            :title="$t('admin.gamenight.tip')"
            :description="$t('admin.gamenight.step13Description')"
            color="info"
            variant="subtle"
            icon="lucide:lightbulb"
            :ui="{ description: 'text-default' }"
            class="mb-6"
          />
          <UButton
            size="xl"
            :disabled="state.quizSessionState !== QuizSessionState.Locked"
            @click="finishBetting"
          >
            {{ $t('admin.gamenight.releasePointsNow') }}
          </UButton>
        </AdminStep>
        <div
          class="py-24 text-center text-2xl font-semibold data-[active=false]:pointer-events-none data-[active=false]:opacity-50"
          :data-active="state.step === 9"
        >
          <p>{{ $t('admin.gamenight.youAreDone') }}</p>
          <p>
            {{
              $t('admin.gamenight.canClosePageSafely', { number: gamenight })
            }}
          </p>
        </div>
      </div>
    </UContainer>
  </div>
</template>
