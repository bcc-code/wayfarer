<script setup lang="ts">
gql(`
query ProfilePage {
  me {
    id
    name
    image
    church {
      id
      name
    }
    projects {
      id
      achievements {
        id
        name
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
    </div>
  </PageLayout>
</template>
