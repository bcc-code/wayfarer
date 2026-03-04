<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query AdminKickOffPage {
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

const { data } = useAdminKickOffPageQuery()
const { executeMutation, fetching } = useBulkEnrollUsersInChallengeMutation()

const teamLeads = computed(() =>
  data.value?.myCurrentProject.myChurchTeams.flatMap((team) =>
    team.members.filter((member) => member.isTeamLead),
  ),
)
const teamLeadIds = computed(() => teamLeads.value?.map((lead) => lead.id))

const readyToLaunch = ref(false)
const hasLaunched = ref(false)

// hacky, but it works for these kinds of custom things
const challengeId = computed(() => {
  const frontendConfig = data.value?.frontendConfig
  if (!frontendConfig) return
  const frontendConfigJson = JSON.parse(frontendConfig)
  const id = frontendConfigJson['team_name_changed_challenge_id']
  if (!id) return
  return id as string
})

const { isActive, start, stop, remaining } = useCountdown(10, {
  immediate: false,
  onComplete: async () => {
    await executeMutation({
      challengeId: challengeId.value ?? '',
      target: {
        userIds: teamLeadIds.value,
      },
    })
    hasLaunched.value = true
  },
})
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
              label: $t('admin.churchHome.kickOff'),
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
        :to="{ name: 'admin-my-church' }"
      >
        <Icon name="lucide:arrow-left" />
        {{ $t('admin.common.back') }}
      </UButton>

      <TransitionGroup
        tag="div"
        enter-active-class="transition duration-300 ease-out"
        enter-from-class="scale-90 opacity-0"
        enter-to-class="scale-100 opacity-100"
        leave-active-class="transition duration-300 ease-out absolute"
        leave-from-class="scale-100 opacity-100"
        leave-to-class="scale-90 opacity-0"
        move-class="transition duration-300 ease-out"
        class="mt-12 max-w-2xl relative flex flex-col items-start gap-4"
      >
        <h2 key="title" class="text-3xl font-semibold">
          {{ $t('admin.churchHome.kickOff') }}
        </h2>
        <p key="description">
          {{ $t('admin.churchHome.kickOffOnboardingDescription') }}
          <br >
          {{ $t('admin.churchHome.kickOffOnboardingDescriptionWarning') }}
        </p>

        <h3 key="explanation-title" class="text-lg font-bold">
          {{ $t('admin.churchHome.kickOffOnboardingExplanationTitle') }}
        </h3>
        <ul class="list-disc pl-6">
          <li>{{ $t('admin.churchHome.kickOffOnboardingExplanation1') }}</li>
          <li>{{ $t('admin.churchHome.kickOffOnboardingExplanation2') }}</li>
        </ul>

        <UCheckbox
          key="switch"
          v-model="readyToLaunch"
          :label="$t('admin.churchHome.kickOffConfirmation')"
          :ui="{
            label: 'text-base font-normal',
            base: 'size-5 ring-current rounded-none',
          }"
          class="my-4"
        />

        <div
          key="button"
          :class="[
            'flex gap-8 items-center',
            { 'opacity-50 pointer-events-none': !readyToLaunch || hasLaunched },
          ]"
        >
          <BigRedButton
            :label="$t('admin.churchHome.kickOffButton')"
            :loading="fetching"
            @click="() => start()"
          />
          <div v-if="isActive">
            <p class="tabular-nums mb-2 font-bold uppercase">
              {{
                $t(
                  'admin.churchHome.kickOffOnboardingCountdown',
                  { seconds: remaining },
                  remaining,
                )
              }}
            </p>
            <UButton color="neutral" variant="subtle" @click="() => stop()">
              {{ $t('admin.common.cancel') }}
            </UButton>
          </div>
        </div>

        <UAlert
          v-if="hasLaunched"
          key="success"
          :title="$t('admin.churchHome.kickOffSuccessTitle')"
          :description="$t('admin.churchHome.kickOffSuccessDescription')"
          color="success"
          variant="subtle"
          icon="lucide:check"
          :ui="{ title: 'text-default' }"
        />
      </TransitionGroup>
    </UContainer>
  </div>
</template>
