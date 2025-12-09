<script setup lang="ts">
gql(`
  query ProfilePage {
    me {
      id
      name
      image
      consentStatus {
        pendingConsents {
          __typename
          id
          key
          version
          title
          body {
            html
          }
          managementType
        }
      }
    }
    myCurrentProject {
      id
      name
      achievements {
        id
        name
        description
        image
        hidden
        achievedAt
        points
      }
      leaderboard(entityType: PERSONS) {
        me {
          score
          rank
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const {
  data,
  error,
  fetching,
  executeQuery: refetch,
} = useProfilePageQuery({
  pause: computed(() => !isAuthReady.value),
})

// Consent banner
const showBanner = useLocalStorage('showBanner', false, {
  listenToStorageChanges: true,
})
watch(
  () => data.value?.me.consentStatus.pendingConsents.length,
  (pending) => {
    if (pending && !showBanner.value) {
      showBanner.value = true
    }
  },
)
</script>

<template>
  <PageLayout :title="data?.me.name">
    <template #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="lucide:settings" />
      </NuxtLink>
    </template>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-list-section-gap">
      <template v-if="showBanner">
        <ConsentCard
          v-for="consent in data.me.consentStatus.pendingConsents"
          :key="consent.id"
          :consent
          class="rounded-card!"
          @update="refetch"
        />
      </template>

      <ProfileProjectCard
        v-if="data.myCurrentProject"
        :project-name="data.myCurrentProject.name"
        :score="data.myCurrentProject.leaderboard.me?.score"
        :rank="data.myCurrentProject.leaderboard.me?.rank"
        :achievements="data.myCurrentProject.achievements"
      />
    </div>
  </PageLayout>
</template>
