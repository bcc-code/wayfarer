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
        <ProfileProjectCard
          :project-name="data.currentProject.name"
          :points="2000"
          :standing="214"
          :achievements="data.currentProject.achievements"
          highlighted
        />
      </template>
    </div>
  </PageLayout>
</template>
