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
</script>

<template>
  <PageLayout :title="$t('pages.standings')">
    <DesignTabs
      v-model="tab"
      :tabs="[
        { label: $t('standings.global'), value: 'global' },
        { label: $t('standings.local'), value: 'local' },
        { label: $t('standings.unit'), value: 'unit' },
      ]"
      class="mb-default -mt-list-outside"
    />
    <StandingsGlobal v-if="tab == 'global'" />
    <StandingsLocal v-if="tab == 'local'" />
    <StandingsUnit v-if="tab == 'unit'" />
  </PageLayout>
</template>
