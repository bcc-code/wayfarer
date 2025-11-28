<script setup lang="ts">
gql(`
  query ProfilePage {
    me {
      id
      name
      image
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
</script>

<template>
  <PageLayout :title="$t('pages.profile')">
    <template #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="lucide:settings" />
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
            class="bg-background-default text-accent-contrast size-full rounded-full object-cover object-center"
          />
          <Icon
            v-else
            name="IconProfile"
            class="text-accent-contrast size-16"
          />
        </div>
        <h2 class="text-heading">{{ data.me.name }}</h2>
      </div>

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
