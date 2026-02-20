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
const { t } = useI18n()

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
        <h2 class="text-2xl font-semibold">
          {{ $t('admin.gamenight.beforeYouStart') }}
        </h2>
        <div>
          <UCheckbox
            v-model="state.filmsReady"
            :label="$t('admin.gamenight.ensureFilmsReady')"
          />
          <NuxtLink
            :to="filmsUrl"
            external
            target="_blank"
            class="pl-7 underline text-primary flex items-center gap-1"
          >
            {{ $t('admin.gamenight.findFilmsHere') }}
            <Icon name="lucide:external-link" />
          </NuxtLink>
        </div>
        <div>
          <UCheckbox
            v-model="state.bigScreenReady"
            :label="$t('admin.gamenight.haveBigScreenReady')"
            :description="$t('admin.gamenight.usedInStep5')"
          />
          <NuxtLink
            :to="bigScreenUrl"
            external
            target="_blank"
            :data-loading="!bigScreenUrl"
            class="pl-7 underline text-primary data-[loading=true]:pointer-events-none data-[loading=true]:opacity-50 flex items-center gap-1"
          >
            {{ $t('admin.gamenight.findBigScreenHere') }}
            <Icon name="lucide:external-link" />
          </NuxtLink>
        </div>
      </section>

      <div class="divide-y divide-muted mb-24">
        <AdminStep
          :step="1"
          :title="$t('admin.gamenight.step1Title')"
          :description="$t('admin.gamenight.step1Description')"
          :active="state.step === 1"
        >
          <UAlert
            :title="$t('admin.gamenight.tip')"
            :description="$t('admin.gamenight.fullscreenTip')"
            color="info"
            icon="lucide:lightbulb"
            variant="subtle"
            :ui="{ description: 'text-default' }"
            class="mb-6"
          />
          <UButton
            size="xl"
            trailing-icon="lucide:external-link"
            :to="bigScreenUrl"
            target="_blank"
            external
          >
            {{ $t('admin.gamenight.openBigScreenNewTab') }}
          </UButton>
          <UButton
            class="flex mt-2"
            variant="outline"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            {{ $t('admin.gamenight.goToNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="2"
          :title="$t('admin.gamenight.step2Title')"
          :description="$t('admin.gamenight.step2Description')"
          :active="state.step === 2"
        >
          <UButton
            size="xl"
            :to="`${filmsUrl}?film=1`"
            trailing-icon="lucide:external-link"
          >
            {{ $t('admin.gamenight.goToFilm', { number: 1 }) }}
          </UButton>
          <UButton
            class="flex mt-2"
            variant="outline"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            {{ $t('admin.gamenight.filmFinishedNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="3"
          :title="$t('admin.gamenight.step3Title')"
          :description="$t('admin.gamenight.step3Description')"
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
              {{ $t('admin.gamenight.startBetting') }}
            </UButton>
            <div
              v-if="state.quizSessionState === QuizSessionState.Open"
              class="flex gap-4 items-center"
            >
              <UButton
                size="xl"
                :loading="bettingActionLoading"
                @click="lockBetting"
              >
                {{ $t('admin.gamenight.lockBetting') }}
              </UButton>
              <p>
                {{ $t('admin.gamenight.bettingLockedNotice') }} <br />
                {{ $t('admin.gamenight.checkAllBetBeforeLock') }}
              </p>
            </div>
            <div
              v-if="state.quizSessionState === QuizSessionState.Locked"
              class="flex gap-4 items-center"
            >
              <UButton
                size="xl"
                :loading="bettingActionLoading"
                @click="reopenBetting"
              >
                {{ $t('admin.gamenight.reopenBetting') }}
              </UButton>
              <p>{{ $t('admin.gamenight.reopenBettingHint') }}</p>
            </div>
          </div>
          <div
            v-if="state.quizSessionState === QuizSessionState.Locked"
            class="flex gap-4 items-center"
          >
            <UButton
              trailing-icon="lucide:check"
              size="xl"
              @click="state.step++"
            >
              {{ $t('admin.gamenight.allBetNextStep') }}
            </UButton>
            <p>{{ $t('admin.gamenight.cannotReopenAfterContinue') }}</p>
          </div>
        </AdminStep>
        <AdminStep
          :step="4"
          :title="$t('admin.gamenight.step4Title')"
          :description="$t('admin.gamenight.step4Description')"
          :active="state.step === 4"
        >
          <UButton
            size="xl"
            :to="`${filmsUrl}?film=2`"
            trailing-icon="lucide:external-link"
          >
            {{ $t('admin.gamenight.goToFilm', { number: 2 }) }}
          </UButton>
          <UButton
            class="flex mt-2"
            variant="outline"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            {{ $t('admin.gamenight.filmFinishedNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="5"
          :title="$t('admin.gamenight.step5TitleBigScreen')"
          :active="state.step === 5"
        >
          <UButton size="xl" trailing-icon="lucide:check" @click="state.step++">
            {{ $t('admin.gamenight.goToNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="6"
          :title="$t('admin.gamenight.step5Title')"
          :description="$t('admin.gamenight.step5Description')"
          :active="state.step === 6"
        >
          <UAlert
            :title="$t('admin.gamenight.tip')"
            color="info"
            variant="subtle"
            icon="lucide:lightbulb"
            :ui="{ description: 'text-default' }"
            class="mb-6"
          >
            <template #description>
              <ul class="list-disc list-inside">
                <li>
                  {{ $t('admin.gamenight.unitLeaderMustGoTo') }}
                  <NuxtLink
                    to="https://pc26.bcc.media"
                    target="_blank"
                    external
                    class="underline text-info inline-flex items-center gap-0.5"
                  >
                    pc26.bcc.media
                    <Icon name="lucide:external-link" />
                  </NuxtLink>
                  {{ $t('admin.gamenight.onAPC') }}
                </li>
                <li>
                  {{ $t('admin.gamenight.ifUnitLeaderMissing') }}
                  <NuxtLink
                    :to="{ name: 'admin-my-church-units' }"
                    target="_blank"
                    class="underline text-info inline-flex items-center gap-0.5"
                  >
                    {{ $t('admin.gamenight.unitAdminPage') }}
                    <Icon name="lucide:external-link" />
                  </NuxtLink>
                </li>
              </ul>
            </template>
          </UAlert>
          <UButton size="xl" @click="enrollUnitLeadersInChallenge">
            {{ $t('admin.gamenight.giveUnitLeadersAccess') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="7"
          :title="$t('admin.gamenight.step6Title')"
          :description="$t('admin.gamenight.step6Description')"
          :active="state.step === 7"
        >
          <UButton
            :loading="cryptexUrlStatus !== 'success'"
            size="xl"
            :to="cryptexUrl?.url"
            trailing-icon="lucide:arrow-right"
          >
            {{ $t('admin.gamenight.goToUnitTaskAdmin') }}
          </UButton>
        </AdminStep>
        <div
          class="py-12 pl-22.75 text-2xl font-semibold data-[active=false]:pointer-events-none data-[active=false]:opacity-50"
          :data-active="state.step === 7"
        >
          <p>{{ $t('admin.gamenight.completeStepsOnOtherPage') }}</p>
        </div>
        <AdminStep
          :step="13"
          :title="$t('admin.gamenight.step12Title')"
          :active="state.step === 8"
        >
          <UButton
            size="xl"
            :to="`${filmsUrl}?film=3`"
            trailing-icon="lucide:external-link"
          >
            {{ $t('admin.gamenight.goToFilm', { number: 3 }) }}
          </UButton>
          <UButton
            class="flex mt-2"
            variant="outline"
            trailing-icon="lucide:check"
            size="xl"
            @click="state.step++"
          >
            {{ $t('admin.gamenight.filmFinishedNextStep') }}
          </UButton>
        </AdminStep>
        <AdminStep
          :step="14"
          :title="$t('admin.gamenight.step13Title')"
          :description="$t('admin.gamenight.step13Description')"
          :active="state.step === 9"
        >
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
          :data-active="state.step === 10"
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
