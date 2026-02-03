<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query AdminKickOffPage {
    myCurrentProject {
      teams {
        members {
          id
          isTeamLead
        }
      }
    }
  }
`)

const { data } = useAdminKickOffPageQuery()
const { executeMutation } = useBulkEnrollUsersInChallengeMutation()

const teamLeads = computed(() =>
  data.value?.myCurrentProject.teams.flatMap((team) =>
    team.members.filter((member) => member.isTeamLead),
  ),
)
const teamLeadIds = computed(() => teamLeads.value?.map((lead) => lead.id))

watch(teamLeadIds, (leads) => {
  console.log(leads)
})

const readyToLaunch = ref(false)

const { isActive, start, stop, remaining } = useCountdown(5, {
  immediate: false,
  onComplete: async () => {
    await executeMutation({
      challengeId: '',
      target: {
        userIds: [],
      },
    })
    readyToLaunch.value = false
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
        <p key="description">{{ $t('admin.churchHome.kickOffDescription') }}</p>

        <USwitch
          key="switch"
          v-model="readyToLaunch"
          :label="$t('admin.churchHome.kickOffConfirmation')"
          class="my-4"
        />
        <div v-if="readyToLaunch" key="button" class="flex gap-8 items-center">
          <BigRedButton
            :label="$t('admin.churchHome.kickOffButton')"
            class="my-6"
            @click="() => start()"
          />
          <div v-if="isActive">
            <p class="tabular-nums mb-2">
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

        <p key="description2" class="text-muted">
          {{ $t('admin.churchHome.kickOffOnboardingDescription') }}
        </p>
      </TransitionGroup>
    </UContainer>
  </div>
</template>
