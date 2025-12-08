<script setup lang="ts">
gql(`
  query ProfilePage {
    me {
      id
      name
      image
      consentStatus {
        pendingConsents {
          id
          key
          version
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
const { data, error, fetching } = useProfilePageQuery({
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
      <!-- <Transition
        enter-from-class="opacity-0 scale-95"
        leave-to-class="opacity-0 scale-95"
        enter-active-class="duration-200 transition ease-out"
        leave-active-class="duration-200 transition ease-out"
      >
        <DesignBanner
          v-if="showBanner"
          :title="$t('consent.bannerTitle')"
          @close="showBanner = false"
        >
          <template #action>
            <NuxtLink :to="{ name: 'settings-consents' }">
              <DesignButton size="small" variant="secondary">
                {{ $t('consent.bannerButton') }}
              </DesignButton>
            </NuxtLink>
          </template>
        </DesignBanner>
      </Transition> -->

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
