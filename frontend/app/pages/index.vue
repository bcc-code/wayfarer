<script setup lang="ts">
gql(`
  query ProfilePage {
    me {
      id
      name
      image
    }
    currentProject {
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
          id
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
</script>

<template>
  <PageLayout :title="$t('pages.profile')">
    <template #action>
      <NuxtLink :to="{ name: 'settings' }">
        <button
          class="rounded-button-medium bg-border-default grid size-11 place-items-center"
        >
          <Icon name="lucide:settings" />
        </button>
      </NuxtLink>
    </template>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-list-section-gap">
      <div
        v-if="data.me"
        class="gap-medium flex flex-col items-center justify-center p-4"
      >
        <div
          class="shadow-large bg-background-raised p-list-section-inset flex aspect-square size-35 items-center justify-center rounded-full"
        >
          <NuxtImg
            v-if="data.me.image"
            :src="data.me.image"
            height="160"
            width="160"
            class="bg-background-default text-accent-contrast size-full rounded-full"
          />
          <Icon
            v-else
            name="IconProfile"
            class="text-accent-contrast size-16"
          />
        </div>
        <h2 class="text-heading">{{ data.me.name }}</h2>
      </div>

      <template v-if="data.currentProject">
        <!-- <h2>{{ data.currentProject.name }}</h2> -->
        <DesignPanel class="p-medium space-y-default">
          <template v-if="data.currentProject.leaderboard.me">
            <div class="divide-border-default grid grid-cols-2 divide-x">
              <div class="flex flex-col items-center">
                <p class="text-title">
                  {{ data.currentProject.leaderboard.me.score }}
                </p>
                <p class="text-label text-text-hint">{{ $t('points') }}</p>
              </div>
              <div class="flex flex-col items-center">
                <p class="text-title">
                  {{ data.currentProject.leaderboard.me.rank }}
                </p>
                <p class="text-label text-text-hint">{{ $t('place') }}</p>
              </div>
            </div>
            <DesignButton class="w-full" variant="secondary">
              Points History
            </DesignButton>
          </template>
          <template v-else>
            <EmptyState
              icon="lucide:coins"
              title="Nothing here yet"
              description="You'll see your points here once you start collecting them"
              class="p-small!"
            />
          </template>
        </DesignPanel>

        <AchievementGroup
          title="Test"
          :achievements="data.currentProject.achievements"
        />
      </template>
    </div>
  </PageLayout>
</template>
