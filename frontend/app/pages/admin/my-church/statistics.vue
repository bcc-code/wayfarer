<script setup lang="ts">
import { VisXYContainer, VisGroupedBar, VisAxis } from '@unovis/vue'

definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query ChurchAdminStatisticsPage {
    churchAdminStatistics {
      churchId
      churchName
      projectId
      projectName
      totalUsersInTeams
      ageGroups {
        ageGroup
        userCount
        averageScore
      }
      userScores {
        userId
        totalScore
      }
    }
  }
`)

const { t } = useI18n()

const { data, fetching, error } = useChurchAdminStatisticsPageQuery({})

// Track initial load
const hasLoadedOnce = ref(false)
watch(data, (newData) => {
  if (!newData) return
  hasLoadedOnce.value = true
})

// All age groups in order
const allAgeGroups = ['12 - 18', '19 - 25', '26 - 36', '37 - 59', '60+']

// Ensure all age groups are shown, even if they have no data
const ageGroupsWithDefaults = computed(() => {
  const apiGroups = data.value?.churchAdminStatistics?.ageGroups ?? []
  const groupMap = new Map(apiGroups.map((g) => [g.ageGroup, g]))

  return allAgeGroups.map((ageGroup) => {
    const existing = groupMap.get(ageGroup)
    return {
      ageGroup,
      userCount: existing?.userCount ?? 0,
      averageScore: existing?.averageScore ?? 0,
    }
  })
})

function formatAgeGroup(ageGroup: string): string {
  return `${ageGroup} ${t('admin.statistics.yearsOld')}`
}

function formatScore(score: number): string {
  return formatNumber(Math.round(score))
}

// User scores for bar chart
const userScores = computed(() => {
  return data.value?.churchAdminStatistics?.userScores ?? []
})

type UserScoreData = (typeof userScores.value)[number]

// Bar chart configuration
const barX = (_d: UserScoreData, i: number) => i
const barY = (d: UserScoreData) => d.totalScore
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
              label: $t('admin.churchHome.statistics'),
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
      <div v-else-if="data?.churchAdminStatistics" class="mt-12 w-full">
        <h2 class="text-3xl font-semibold mb-2">
          {{ $t('admin.churchHome.statistics') }}
        </h2>
        <p class="text-muted">
          {{ data.churchAdminStatistics.churchName }} -
          {{ data.churchAdminStatistics.projectName }}
        </p>

        <!-- Age group statistics -->
        <h3 class="mb-4 mt-16">
          {{ $t('admin.statistics.averageScoresByAgeGroup') }}
        </h3>
        <section
          id="averages"
          class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-6 w-full"
        >
          <AdminStatisticCard
            v-for="group in ageGroupsWithDefaults"
            :key="group.ageGroup"
            :title="formatAgeGroup(group.ageGroup)"
            :value="formatScore(group.averageScore)"
            :subtitle="`${group.userCount} ${$t('admin.statistics.users').toLocaleLowerCase()}`"
          />
        </section>

        <!-- Points distribution chart -->
        <h3 class="mt-16">
          {{ $t('admin.statistics.pointsDistribution') }}
        </h3>
        <p class="text-muted text-sm mb-4">
          {{ $t('admin.statistics.pointsDistributionDescription') }}
        </p>
        <section id="points" class="w-full">
          <VisXYContainer
            v-if="userScores.length > 0"
            :data="userScores"
            :height="300"
          >
            <VisGroupedBar
              :x="barX"
              :y="barY"
              color="var(--ui-primary)"
              :rounded-corners="false"
            />
            <VisAxis
              type="y"
              :label="$t('admin.statistics.points')"
              :tick-format="(v: number) => formatScore(v)"
            />
            <VisAxis
              type="x"
              :label="$t('admin.statistics.peopleInAUnit')"
              :tick-format="(i: number) => ''"
            />
          </VisXYContainer>
          <p v-else class="text-muted">
            {{ $t('admin.statistics.noData') }}
          </p>
        </section>
      </div>
    </UContainer>
  </div>
</template>

<style scoped>
.dark #points {
  --vis-axis-grid-color: #333;
}
</style>
