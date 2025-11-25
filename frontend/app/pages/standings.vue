<script setup lang="ts">
const params = useUrlSearchParams('history')
const tab = computed({
  get() {
    if (typeof params.tab === 'string') return params.tab
    return 'global'
  },
  set(tab: 'global' | 'unit') {
    params.tab = tab
  },
})
</script>

<template>
  <PageLayout :title="$t('pages.standings')">
    <DesignTabs
      v-model="tab"
      :tabs="[
        { label: $t('standings.global'), value: 'global' },
        { label: $t('standings.unit'), value: 'unit' },
      ]"
      class="mb-default -mt-list-outside"
    />
    <StandingsGlobal v-if="tab == 'global'" />
    <StandingsUnit v-if="tab == 'unit'" />
  </PageLayout>
</template>
