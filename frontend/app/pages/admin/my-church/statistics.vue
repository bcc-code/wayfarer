<script setup lang="ts">
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

function formatAgeGroup(ageGroup: string): string {
  return `${ageGroup} ${t('admin.statistics.yearsOld')}`
}

function formatScore(score: number): string {
  return Math.round(score).toLocaleString()
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
        <p class="text-dimmed mb-8">
          {{ data.churchAdminStatistics.churchName }} -
          {{ data.churchAdminStatistics.projectName }}
        </p>

        <!-- Age group statistics -->
        <h3 class="text-xl font-semibold mb-4">
          {{ $t('admin.statistics.averageScoresByAgeGroup') }}
        </h3>

        <div
          v-if="data.churchAdminStatistics.ageGroups.length === 0"
          class="text-dimmed text-center py-8"
        >
          {{ $t('admin.statistics.noData') }}
        </div>

        <div
          v-else
          class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-4 w-full"
        >
          <div
            v-for="group in data.churchAdminStatistics.ageGroups"
            :key="group.ageGroup"
            class="p-6 rounded-xl border border-default bg-elevated/50 flex flex-col gap-2"
          >
            <div class="text-sm text-muted">
              {{ formatAgeGroup(group.ageGroup) }}
            </div>
            <div class="text-4xl font-bold tabular-nums">
              {{ formatScore(group.averageScore) }}
            </div>
            <div class="text-xs text-dimmed">
              {{ group.userCount }}
              {{ $t('admin.statistics.users').toLowerCase() }}
            </div>
          </div>
        </div>
      </div>
    </UContainer>
  </div>
</template>
