<script setup lang="ts">
gql(`
  query StandingsPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter, $first: Int) {
    myCurrentProject {
      id
      leaderboard(entityType: $entityType, filter: $filter, first: $first) {
        edges {
          node {
            id
            name
            description
            score
            image
            rank
            isMe
          }
        }
        me {
          id
          name
          description
          score
          rank
          isMe
          image
        }
      }
    }
  }
`)

const tab = ref<'global' | 'unit'>('global')
</script>

<template>
  <PageLayout :title="$t('pages.standings')">
    <DesignTabs
      v-model="tab"
      :tabs="[
        { label: $t('standings.global'), value: 'global' },
        { label: $t('standings.unit'), value: 'unit' },
      ]"
      class="mb-default"
    />
    <StandingsGlobal v-if="tab == 'global'" />
    <StandingsUnit v-if="tab == 'unit'" />
  </PageLayout>
</template>
