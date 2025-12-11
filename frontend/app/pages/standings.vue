<script setup lang="ts">
const { track } = useAnalytics()

const params = useUrlSearchParams('history')
const tab = computed({
  get() {
    if (typeof params.tab === 'string') return params.tab
    return 'global'
  },
  set(newTab: 'global' | 'unit' | 'local') {
    const oldTab = tab.value
    track(AnalyticsEvent.LeaderboardTabChanged, { from: oldTab, to: newTab })
    params.tab = newTab
  },
})

gql(`
  query StandingsPage {
    myCurrentProject {
      myTeam {
        id
      }
    }
  }
`)

const { data } = await useStandingsPageQuery()
const hasUnit = computed(() => Boolean(data.value?.myCurrentProject.myTeam?.id))
</script>

<template>
  <PageLayout :title="$t('pages.standings')">
    <div class="p-list-outside">
      <DesignTabs
        v-model="tab"
        :tabs="[
          { label: $t('standings.global'), value: 'global' },
          { label: $t('standings.local'), value: 'local' },
          { label: $t('standings.unit'), value: 'unit', enabled: hasUnit },
        ]"
        class="mb-default -mt-list-outside"
      />
      <StandingsGlobal v-if="tab == 'global'" />
      <StandingsLocal v-if="tab == 'local'" />
      <StandingsUnit v-if="tab == 'unit'" />
    </div>
  </PageLayout>
</template>
