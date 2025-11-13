<script setup lang="ts">
gql(`
  query ProfilePage {
    me {
      id
      name
      image
      projects {
        id
        achievements {
          id
          name
          description
          image
          hidden
          achievedAt
          points
        }
      }
    }
  }
`)

const { data, error, fetching } = useProfilePageQuery()

const achievements = computed(() => {
  return data.value?.me.projects.flatMap((project) => project.achievements)
})
const completedAchievements = computed(() => {
  return achievements.value?.filter((achievement) => achievement.achievedAt)
})
const notCompletedAchievements = computed(() => {
  return achievements.value?.filter((achievement) => !achievement.achievedAt)
})
</script>

<template>
  <PageLayout :title="$t('pages.profile')">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-list-section-gap">
      <div v-if="data.me.image" class="flex items-center justify-center p-4">
        <div
          class="shadow-large bg-background-raised p-list-section-inset aspect-square size-42 rounded-full"
        >
          <NuxtImg
            :src="data.me.image"
            height="160"
            width="160"
            class="bg-background-default text-accent-contrast rounded-full"
          />
        </div>
      </div>
      <LocaleSelector />

      <template v-if="completedAchievements?.length">
        <h3 class="text-label mt-6 mb-2">
          {{ $t('achievements.completed') }}
        </h3>
        <div class="gap-list-section-gap grid grid-cols-5">
          <AchievementBadge
            v-for="achievement in completedAchievements"
            :key="achievement.id"
            :achievement
          />
        </div>
      </template>
      <template v-if="notCompletedAchievements?.length">
        <h3 class="text-label mt-6 mb-2">
          {{ $t('achievements.notCompleted') }}
        </h3>
        <div class="gap-list-section-gap grid grid-cols-5">
          <AchievementBadge
            v-for="achievement in notCompletedAchievements"
            :key="achievement.id"
            :achievement
          />
        </div>
      </template>
    </div>
  </PageLayout>
</template>
