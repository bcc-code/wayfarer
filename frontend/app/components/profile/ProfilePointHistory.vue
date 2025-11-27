<script setup lang="ts">
gql(`
	query PointHistory($first: Int) {
    myCurrentProject {
      journal(first: $first) {
        edges {
          node {
            id
            sourceType
            challenge {
              name
            }
            source {
              __typename
            }
            reason
            points
            createdAt
          }
        }
      }
    }
  }
`)

const open = ref(false)

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = usePointHistoryQuery({
  variables: { first: 100 },
  pause: computed(() => !isAuthReady.value || !open.value),
})
</script>

<template>
  <UModal
    v-model:open="open"
    :ui="{ content: 'bg-background-default' }"
    :transition="false"
    fullscreen
    modal
  >
    <slot />
    <template #content="{ close }">
      <PageLayout :title="$t('pages.pointHistory')">
        <template #action>
          <DesignIconButton icon="lucide:x" @click="close" />
        </template>

        <LoadingState v-if="fetching" />
        <ErrorState v-else-if="error" :error />
        <div
          v-else-if="data?.myCurrentProject.journal.edges.length"
          class="space-y-list-section-gap"
        >
          <div
            v-for="journal in data.myCurrentProject.journal.edges"
            :key="journal.node.id"
          >
            <span>{{ journal.node.reason }}</span>
            <span>{{ journal.node.points }}</span>
          </div>
        </div>
        <EmptyState v-else />
      </PageLayout>
    </template>
  </UModal>
</template>
